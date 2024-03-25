package connratelimit

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mailgun/ttlmap"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"golang.org/x/time/rate"
)

const (
	typeName   = "ConnRateLimitTCP"
	maxSources = 65536
)

// Mainly derived from HTTP RateLimiter
type connRateLimitTCP struct {
	name string
	next tcp.Handler

	rate  rate.Limit // conns/s
	burst int64

	// maxDelay is the maximum duration we're willing to wait for a bucket reservation to become effective, in nanoseconds.
	// For now it is somewhat arbitrarily set to 1/(2*rate).
	maxDelay time.Duration

	mu      sync.Mutex
	ttl     int
	buckets *ttlmap.TtlMap
}

// New creates a max connections middleware.
// The connections are identified and grouped by remote IP.
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPConnRateLimit, name string) (tcp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	buckets, err := ttlmap.NewConcurrent(maxSources)
	if err != nil {
		return nil, err
	}

	burst := config.Burst
	if burst < 1 {
		burst = 1
	}

	period := time.Duration(config.Period)
	if period < 0 {
		return nil, fmt.Errorf("negative value not valid for period: %v", period)
	}
	if period == 0 {
		period = time.Second
	}

	// Initialized at rate.Inf to enforce no rate limiting when config.Average == 0
	rtl := float64(rate.Inf)
	// No need to set any particular value for maxDelay as the reservation's delay
	// will be <= 0 in the Inf case (i.e. the average == 0 case).
	var maxDelay time.Duration

	if config.Average > 0 {
		rtl = float64(config.Average*int64(time.Second)) / float64(period)
		// maxDelay does not scale well for rates below 1,
		// so we just cap it to the corresponding value, i.e. 0.5s, in order to keep the effective rate predictable.
		// One alternative would be to switch to a no-reservation mode (Allow() method) whenever we are in such a low rate regime.
		if rtl < 1 {
			maxDelay = 500 * time.Millisecond
		} else {
			maxDelay = time.Second / (time.Duration(rtl) * 2)
		}
	}

	// Make the ttl inversely proportional to how often a rate limiter is supposed to see any activity (when maxed out),
	// for low rate limiters.
	// Otherwise just make it a second for all the high rate limiters.
	// Add an extra second in both cases for continuity between the two cases.
	ttl := 1
	if rtl >= 1 {
		ttl++
	} else if rtl > 0 {
		ttl += int(1 / rtl)
	}

	return &connRateLimitTCP{
		name: name,
		next: next,

		rate:     rate.Limit(rtl),
		maxDelay: maxDelay,
		burst:    burst,
		ttl:      ttl,

		mu:      sync.Mutex{},
		buckets: buckets,
	}, nil
}

// ServeTCP serves the given TCP connection.
func (rl *connRateLimitTCP) ServeTCP(conn tcp.WriteCloser) {
	logger := middlewares.GetLogger(context.Background(), rl.name, typeName)

	ip, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		logger.Error().Err(err).Msg("Cannot parse IP from remote addr")
		conn.Close()
		return
	}

	var bucket *rate.Limiter
	if rlSource, exists := rl.buckets.Get(ip); exists {
		bucket = rlSource.(*rate.Limiter)
	} else {
		bucket = rate.NewLimiter(rl.rate, int(rl.burst))
	}

	// We Set even in the case where the source already exists,
	// because we want to update the expiryTime everytime we get the source,
	// as the expiryTime is supposed to reflect the activity (or lack thereof) on that source.
	if err := rl.buckets.Set(ip, bucket, rl.ttl); err != nil {
		logger.Error().Err(err).Msg("Could not insert/update bucket")
		conn.Close()
		return
	}

	res := bucket.Reserve()
	if !res.OK() {
		logger.Debug().Msgf("Dropper bursty traffic from %s:%s", ip, port)
		conn.Close()
		return
	}

	delay := res.Delay()
	if delay > rl.maxDelay {
		res.Cancel()
		logger.Debug().Msgf("Connection from %s:%s rejected", ip, port)
		conn.Close()
		return
	}
	logger.Debug().Msgf("Connection from %s:%s accepted", ip, port)

	time.Sleep(delay)
	rl.next.ServeTCP(conn)
}
