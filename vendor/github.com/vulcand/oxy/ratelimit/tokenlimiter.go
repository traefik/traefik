// Package ratelimit Tokenbucket based request rate limiter
package ratelimit

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mailgun/timetools"
	"github.com/mailgun/ttlmap"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/utils"
)

// DefaultCapacity default capacity
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
		return fmt.Errorf("invalid period: %v", period)
	}
	if average <= 0 {
		return fmt.Errorf("invalid average: %v", average)
	}
	if burst <= 0 {
		return fmt.Errorf("invalid burst: %v", burst)
	}
	rs.m[period] = &rate{period: period, average: average, burst: burst}
	return nil
}

func (rs *RateSet) String() string {
	return fmt.Sprint(rs.m)
}

// RateExtractor rate extractor
type RateExtractor interface {
	Extract(r *http.Request) (*RateSet, error)
}

// RateExtractorFunc rate extractor function type
type RateExtractorFunc func(r *http.Request) (*RateSet, error)

// Extract extract from request
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
	capacity     int
	next         http.Handler

	log *log.Logger
}

// New constructs a `TokenLimiter` middleware instance.
func New(next http.Handler, extract utils.SourceExtractor, defaultRates *RateSet, opts ...TokenLimiterOption) (*TokenLimiter, error) {
	if defaultRates == nil || len(defaultRates.m) == 0 {
		return nil, fmt.Errorf("provide default rates")
	}
	if extract == nil {
		return nil, fmt.Errorf("provide extract function")
	}
	tl := &TokenLimiter{
		next:         next,
		defaultRates: defaultRates,
		extract:      extract,

		log: log.StandardLogger(),
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

// Logger defines the logger the token limiter will use.
//
// It defaults to logrus.StandardLogger(), the global logger used by logrus.
func Logger(l *log.Logger) TokenLimiterOption {
	return func(tl *TokenLimiter) error {
		tl.log = l
		return nil
	}
}

// Wrap sets the next handler to be called by token limiter handler.
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
		tl.log.Warnf("limiting request %v %v, limit: %v", req.Method, req.URL, err)
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

// MaxRateError max rate error
type MaxRateError struct {
	delay time.Duration
}

func (m *MaxRateError) Error() string {
	return fmt.Sprintf("max rate reached: retry-in %v", m.delay)
}

// RateErrHandler error handler
type RateErrHandler struct{}

func (e *RateErrHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	if rerr, ok := err.(*MaxRateError); ok {
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rerr.delay.Seconds()))
		w.Header().Set("X-Retry-In", rerr.delay.String())
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(err.Error()))
		return
	}
	utils.DefaultHandler.ServeHTTP(w, req, err)
}

// TokenLimiterOption token limiter option type
type TokenLimiterOption func(l *TokenLimiter) error

// ErrorHandler sets error handler of the server
func ErrorHandler(h utils.ErrorHandler) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.errHandler = h
		return nil
	}
}

// ExtractRates sets the rate extractor
func ExtractRates(e RateExtractor) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.extractRates = e
		return nil
	}
}

// Clock sets the clock
func Clock(clock timetools.TimeProvider) TokenLimiterOption {
	return func(cl *TokenLimiter) error {
		cl.clock = clock
		return nil
	}
}

// Capacity sets the capacity
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
