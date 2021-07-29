package consul

import (
	"fmt"
	"github.com/go-playground/form"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type consulTarget struct {
	Addr              string        `form:"-"`
	User              string        `form:"-"`
	Password          string        `form:"-"`
	Service           string        `form:"-"`
	Wait              time.Duration `form:"wait"`
	Timeout           time.Duration `form:"timeout"`
	MaxBackoff        time.Duration `form:"max-backoff"`
	Tag               string        `form:"tag"`
	Near              string        `form:"near"`
	Healthy           bool          `form:"healthy"`
	TLSInsecure       bool          `form:"insecure"`
	Token             string        `form:"token"`
	Dc                string        `form:"dc"`
	AllowStale        bool          `form:"allow-stale"`
	RequireConsistent bool          `form:"require-consistent"`
}

func (c *consulTarget) String() string {
	return fmt.Sprintf("addr='%s' service='%s' healthy='%t' tag='%s'", c.Addr, c.Service, c.Healthy, c.Tag)
}

//  parseURL with parameters
// see README.md for the actual format
// URL schema will stay stable in the future for backward compatibility
func parseURL(u string) (consulTarget, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return consulTarget{}, errors.Wrap(err, "Malformed URL")
	}

	if rawURL.Scheme != "consul" ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return consulTarget{},
			errors.Errorf("Malformed URL('%s'). Must be in the next format: 'consul://[user:passwd]@host/service?param=value'", u)
	}

	var target consulTarget
	target.User = rawURL.User.Username()
	target.Password, _ = rawURL.User.Password()
	target.Addr = rawURL.Host
	target.Service = strings.TrimLeft(rawURL.Path, "/")
	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.ParseDuration(vals[0])
	}, time.Duration(0))

	err = decoder.Decode(&target, rawURL.Query())
	if err != nil {
		return consulTarget{}, errors.Wrap(err, "Malformed URL parameters")
	}
	if len(target.Near) == 0 {
		target.Near = "_agent"
	}
	if target.MaxBackoff == 0 {
		target.MaxBackoff = 30 * time.Second
	}
	if target.Wait == 0 {
		target.Wait = 5 * time.Second
	}
	return target, nil
}

// consulConfig returns config based on the parsed target.
// It uses custom http-client.
func (c *consulTarget) consulConfig() *api.Config {
	var creds *api.HttpBasicAuth
	if len(c.User) > 0 && len(c.Password) > 0 {
		creds = new(api.HttpBasicAuth)
		creds.Password = c.Password
		creds.Username = c.User
	}
	// custom http.Client
	client := &http.Client{
		Timeout: c.Timeout,
	}
	return &api.Config{
		Address:    c.Addr,
		HttpAuth:   creds,
		WaitTime:   c.Wait,
		HttpClient: client,
		TLSConfig: api.TLSConfig{
			InsecureSkipVerify: c.TLSInsecure,
		},
		Token: c.Token,
	}
}
