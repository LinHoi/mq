package backoff

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r  = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu sync.Mutex
)

// Strategy defines the methodology for backing off after a grpc connection
// failure.
//
type Strategy interface {
	// Backoff returns the amount of time to wait before the next retry given
	// the number of consecutive failures.
	Backoff(retries int) time.Duration
}

const (
	// baseDelay is the amount of time to wait before retrying after the first
	// failure.
	baseDelay = 10 * time.Second
	// factor is applied to the backoff after each retry.
	factor = 1.6
	// jitter provides a range to randomize backoff delays.
	jitter = 0.2
)

// Exponential implements exponential backoff algorithm as defined in
// https://github.com/grpc/grpc/blob/master/doc/connection-backoff.md.
type Exponential struct {
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

// Backoff returns the amount of time to wait before the next retry given the
// number of retries.
func (bc Exponential) Backoff(retries int) time.Duration {
	if retries == 0 {
		return baseDelay
	}
	backoff, max := float64(baseDelay), float64(bc.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= factor
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + jitter*(randFloat64()*2-1)
	if backoff < 0 {
		return 0
	}
	return time.Duration(backoff)
}

// randFloat64 implements rand.randFloat64 on the grpcrand global source.
func randFloat64() float64 {
	mu.Lock()
	res := r.Float64()
	mu.Unlock()
	return res
}
