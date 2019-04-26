package api

import (
	"go.uber.org/ratelimit"

	"net/http"
	"sync"
)

// RateLimitRoundTripper 秒間アクセス数を制限するためのhttp.RoundTripper実装
type RateLimitRoundTripper struct {
	// Transport 親となるhttp.RoundTripper、nilの場合http.DefaultTransportが利用される
	Transport http.RoundTripper
	// RateLimitPerSec 秒あたりのリクエスト数
	RateLimitPerSec int

	once      sync.Once
	rateLimit ratelimit.Limiter
}

// RoundTrip http.RoundTripperの実装
func (r *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.once.Do(func() {
		r.rateLimit = ratelimit.New(r.RateLimitPerSec)
	})
	if r.Transport == nil {
		r.Transport = http.DefaultTransport
	}

	r.rateLimit.Take()
	return r.Transport.RoundTrip(req)
}
