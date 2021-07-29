package consul

import (
	"context"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"github.com/linhoi/mq/external/gclient/resolver/backoff"
	"github.com/linhoi/mq/external/log"
	"github.com/linhoi/mq/external/trace"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	minResRate = 5 * time.Second
)

// init function needs for  auto-register in resolvers registry
func init() {
	resolver.Register(&consulBuilder{})
}

// consulBuilder implements resolver.Builder and use for constructing all consul resolvers
type consulBuilder struct{}

func (b *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	dsn := strings.Join([]string{"consul:/", target.Authority, target.Endpoint}, "/")
	consulTarget, err := parseURL(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "consul地址配置错误")
	}
	cli, err := consul.NewClient(consulTarget.consulConfig())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("连接consul失败 %s", consulTarget.String()))
	}

	ctx, cancel := context.WithCancel(trace.NewCtxWithTraceId(context.Background()))
	r := &consulResolver{
		serviceResolver:      cli.Health(),
		target:               consulTarget,
		backoff:              backoff.Exponential{MaxDelay: consulTarget.MaxBackoff},
		ctx:                  ctx,
		cancel:               cancel,
		cc:                   cc,
		t:                    time.NewTimer(0),
		rn:                   make(chan struct{}, 1),
		disableServiceConfig: opts.DisableServiceConfig,
		initialized:          false,
	}

	r.wg.Add(1)
	r.ResolveNow(resolver.ResolveNowOptions{})
	go r.watcher()
	return r, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *consulBuilder) Scheme() string {
	return "consul"
}

type serviceResolver interface {
	Service(service, tag string, passingOnly bool, queryOpts *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
}

// consulResolver implements resolver.Resolver from the gRPC package.
// It watches for endpoints changes and pushes them to the underlying gRPC connection.
type consulResolver struct {
	serviceResolver serviceResolver
	target          consulTarget
	lastIndex       uint64
	retryCount      int
	backoff         backoff.Exponential
	ctx             context.Context
	cancel          context.CancelFunc
	cc              resolver.ClientConn
	// rn channel is used by ResolveNow() to force an immediate resolution of the target.
	rn chan struct{}
	t  *time.Timer
	// wg is used to enforce Close() to return after the watcher() goroutine has finished.
	// Otherwise, data race will be possible. [Race Example] in dns_resolver_test we
	// replace the real lookup functions with mocked ones to facilitate testing.
	// If Close() doesn't wait for watcher() goroutine finishes, race detector sometimes
	// will warns lookup (READ the lookup function pointers) inside watcher() goroutine
	// has data race with replaceNetFunc (WRITE the lookup function pointers).
	wg                   sync.WaitGroup
	disableServiceConfig bool
	initialized          bool
}

// ResolveNow will be skipped due unnecessary in this case
func (c *consulResolver) ResolveNow(resolver.ResolveNowOptions) {
	select {
	case c.rn <- struct{}{}:
	default:
	}
}

// Close closes the resolver.
func (c *consulResolver) Close() {
	c.cancel()
}

func (c *consulResolver) watcher() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.t.C:
		case <-c.rn:
			if !c.t.Stop() {
				// Before resetting a timer, it should be stopped to prevent racing with
				// reads on it's channel.
				<-c.t.C
			}
		}

		state := c.lookup()
		// Next lookup should happen within an interval defined by c.freq. It may be
		// more often due to exponential retry on empty address list.
		if len(state.Addresses) == 0 {
			c.retryCount++
		} else {
			c.retryCount = 0
			c.cc.UpdateState(state)
			c.initialized = true
		}
		c.t.Reset(c.backoff.Backoff(c.retryCount))

		// Sleep to prevent excessive re-resolutions. Incoming resolution requests
		// will be queued in c.rn.
		t := time.NewTimer(minResRate)
		select {
		case <-t.C:
		case <-c.ctx.Done():
			t.Stop()
			return
		}
	}
}

func (c *consulResolver) lookup() resolver.State {
	log.L(c.ctx).Info("开始从consul发现服务", zap.String("target", c.target.String()))
	services, meta, err := c.serviceResolver.Service(
		c.target.Service,
		c.target.Tag,
		c.target.Healthy,
		&consul.QueryOptions{
			WaitIndex:         c.lastIndex,
			Near:              c.target.Near,
			WaitTime:          c.target.Wait,
			Datacenter:        c.target.Dc,
			AllowStale:        c.target.AllowStale,
			RequireConsistent: c.target.RequireConsistent,
		},
	)

	if err != nil {
		log.L(c.ctx).Error("发现服务失败", zap.Error(err))
		if c.initialized == false {
			panic("发现服务失败: " + err.Error())
		}
		return resolver.State{}
	}

	if len(services) == 0 {
		log.L(c.ctx).Error("虽然请求consul没有报错，但是返回的服务列表为空")
		if c.initialized == false {
			panic("未发现有效服务: " + c.target.Service)
		}
		return resolver.State{}
	}

	c.lastIndex = meta.LastIndex

	var addresses []resolver.Address
	for i, s := range services {
		addr := fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port)
		log.L(c.ctx).Info(fmt.Sprintf("成功发现服务%s[%d] %s", c.target.Service, i, addr))
		addresses = append(addresses, resolver.Address{
			Addr:       addr,
			ServerName: c.target.Service,
		})
	}
	sort.Sort(byAddressString(addresses))
	return resolver.State{Addresses: addresses}
}

// byAddressString sorts resolver.Address by Address Field  sorting in increasing order.
type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
