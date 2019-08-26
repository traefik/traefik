package ratelimiter

import (
	"context"
	"net/http"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/vulcand/oxy/ratelimit"
	"github.com/vulcand/oxy/utils"
)

const (
	typeName = "RateLimiterType"
)

type rateLimiter struct {
	handler http.Handler
	name    string
}

// New creates rate limiter middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RateLimit, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	extractFunc, err := utils.NewExtractor(config.ExtractorFunc)
	if err != nil {
		return nil, err
	}

	rateSet := ratelimit.NewRateSet()
	for _, rate := range config.RateSet {
		if err = rateSet.Add(time.Duration(rate.Period), rate.Average, rate.Burst); err != nil {
			return nil, err
		}
	}

	rl, err := ratelimit.New(next, extractFunc, rateSet)
	if err != nil {
		return nil, err
	}
	return &rateLimiter{handler: rl, name: name}, nil
}

func (r *rateLimiter) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *rateLimiter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(rw, req)
}
