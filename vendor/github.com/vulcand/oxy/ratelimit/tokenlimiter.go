// Tokenbucket based request rate limiter
package ratelimit

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mailgun/timetools"
	"github.com/mailgun/ttlmap"
	"github.com/vulcand/oxy/utils"
)

const DefaultCapacity = 65536

// RateSet maintains a set of rates. It can contain only one rate per period at a time.
type RateSet struct {
	m map[time.Duration]*rate
}

// NewRateSet crates an empty `RateSet` instance.
func NewRateSet() *RateSet {
	rs := new(RateSet)
	rs.m = make(map[time.Duration]*rate)
	return rs
}

// Add adds a rate to the set. If there is a rate with the same period in the
// set then the new rate overrides the old one.
func (rs *RateSet) Add(period time.Duration, average int64, burst int64) error {
	if period <= 0 {
		return fmt.Errorf("Invalid period: %v", period)
	}
	if average <= 0 {
		return fmt.Errorf("Invalid average: %v", average)
	}
	if burst <= 0 {
		return fmt.Errorf("Invalid burst: %v", burst)
	}
	rs.m[period] = &rate{period, average, burst}
	return nil
}

func (rs *RateSet) String() string {
	return fmt.Sprint(rs.m)
}

type RateExtractor interface {
	Extract(r *http.Request) (*RateSet, error)
}

type RateExtractorFunc func(r *http.Request) (*RateSet, error)

func (e RateExtractorFunc) Extract(r *http.Request) (*RateSet, error) {
	return e(r)
}

// TokenLimiter implements rate limiting middleware.
type TokenLimiter struct {
	defaultRates *RateSet
	extract      utils.SourceExtractor
	extractRates RateExtractor
	clock        timetools.TimeProvider
	mutex        sync.Mutex
	bucketSets   *ttlmap.TtlMap
	errHandler   utils.ErrorHandler
	log          utils.Logger
	capacity     int
	next         http.Handler
}

// New constructs a `TokenLimiter` middleware instance.
func New(next http.Handler, extract utils.SourceExtractor, defaultRates *RateSet, opts ...TokenLimiterOption) (*TokenLimiter, error) {
	if defaultRates == nil || len(defaultRates.m) == 0 {
		return nil, fmt.Errorf("Provide default rates")
	}
	if extract == nil {
		return nil, fmt.Errorf("Provide extract function")
	}
	tl := &TokenLimiter{
		next:         next,
		defaultRates: defaultRates,
		extract:      extract,
	}

	for _, o := range opts {
		if err := o(tl); err != nil {
			return nil, err
		}
	}
	setDefaults(tl)
	bucketSets, err := ttlmap.NewMapWithProvider(tl.capacity, tl.clock)
	if err != nil {
		return nil, err
	}
	tl.bucketSets = bucketSets
	return tl, nil
}

func (tl *TokenLimiter) Wrap(next http.Handler) {
	tl.next = next
}

func (tl *TokenLimiter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	source, amount, err := tl.extract.Extract(req)
	if err != nil {
		tl.errHandler.ServeHTTP(w, req, err)
		return
	}

	if err := tl.consumeRates(req, source, amount); err != nil {
		tl.log.Infof("limiting request %v %v, limit: %v", req.Method, req.URL, err)
		tl.errHandler.ServeHTTP(w, req, err)
		return
	}

	tl.next.ServeHTTP(w, req)
}

func (tl *TokenLimiter) consumeRates(req *http.Request, source string, amount int64) error {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	effectiveRates := tl.resolveRates(req)
	bucketSetI, exists := tl.bucketSets.Get(source)
	var bucketSet *TokenBucketSet

	if exists {
		bucketSet = bucketSetI.(*TokenBucketSet)
		bucketSet.Update(effectiveRates)
	} else {
		bucketSet = NewTokenBucketSet(effectiveRates, tl.clock)
		// We set ttl as 10 times rate period. E.g. if rate is 100 requests/second per client ip
		// the counters for this ip will expire after 10 seconds of inactivity
		tl.bucketSets.Set(source, bucketSet, int(bucketSet.maxPeriod/time.Second)*10+1)
	}
	delay, err := bucketSet.Consume(amount)
	if err != nil {
		return err
	}
	if delay > 0 {
		return &MaxRateError{delay: delay}
	}
	return nil
}

// effectiveRates retrieves rates to be applied to the request.
func (tl *TokenLimiter) resolveRates(req *http.Request) *RateSet {
	// If configuration mapper is not specified for this instance, then return
	// the default bucket specs.
	if tl.extractRates == nil {
		return tl.defaultRates
	}

	rates, err := tl.extractRates.Extract(req)
	if err != nil {
		tl.log.Errorf("Failed to retrieve rates: %v", err)
		return tl.defaultRates
	}

	// If the returned rate set is empty then used the default one.
	if len(rates.m) == 0 {
		return tl.defaultRates
	}

	return rates
}

type MaxRateError struct {
	delay time.Duration
}

func (m *MaxRateError) Error() string {
	return fmt.Sprintf("max rate reached: retry-in %v", m.delay)
}

type RateErrHandler struct {
}

func (e *RateErrHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	if rerr, ok := err.(*MaxRateError); ok {
		w.Header().Set("X-Retry-In", rerr.delay.String())
		w.WriteHeader(429)
		w.Write([]byte(err.Error()))
		return
	}
	utils.DefaultHandler.ServeHTTP(w, req, err)
}

type TokenLimiterOption func(l *TokenLimiter) error

// Logger sets the logger that will be used by this middleware.
func Logger(l utils.Logger) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.log = l
		return nil
	}
}

// ErrorHandler sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.errHandler = h
		return nil
	}
}

func ExtractRates(e RateExtractor) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.extractRates = e
		return nil
	}
}

func Clock(clock timetools.TimeProvider) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.clock = clock
		return nil
	}
}

func Capacity(cap int) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		if cap <= 0 {
			return fmt.Errorf("bad capacity: %v", cap)
		}
		cl.capacity = cap
		return nil
	}
}

var defaultErrHandler = &RateErrHandler{}

func setDefaults(tl *TokenLimiter) {
	if tl.log == nil {
		tl.log = utils.NullLogger
	}
	if tl.capacity <= 0 {
		tl.capacity = DefaultCapacity
	}
	if tl.clock == nil {
		tl.clock = &timetools.RealTime{}
	}
	if tl.errHandler == nil {
		tl.errHandler = defaultErrHandler
	}
}
