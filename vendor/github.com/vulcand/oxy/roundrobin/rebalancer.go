package roundrobin

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/mailgun/timetools"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/memmetrics"
	"github.com/vulcand/oxy/utils"
)

// RebalancerOption - functional option setter for rebalancer
type RebalancerOption func(*Rebalancer) error

// Meter measures server performance and returns it's relative value via rating
type Meter interface {
	Rating() float64
	Record(int, time.Duration)
	IsReady() bool
}

// NewMeterFn type of functions to create new Meter
type NewMeterFn func() (Meter, error)

// Rebalancer increases weights on servers that perform better than others. It also rolls back to original weights
// if the servers have changed. It is designed as a wrapper on top of the roundrobin.
type Rebalancer struct {
	// mutex
	mtx *sync.Mutex
	// As usual, control time in tests
	clock timetools.TimeProvider
	// Time that freezes state machine to accumulate stats after updating the weights
	backoffDuration time.Duration
	// Timer is set to give probing some time to take place
	timer time.Time
	// server records that remember original weights
	servers []*rbServer
	// next is  internal load balancer next in chain
	next balancerHandler
	// errHandler is HTTP handler called in case of errors
	errHandler utils.ErrorHandler

	ratings []float64

	// creates new meters
	newMeter NewMeterFn

	// sticky session object
	stickySession *StickySession

	requestRewriteListener RequestRewriteListener

	log *log.Logger
}

// RebalancerClock sets a clock
func RebalancerClock(clock timetools.TimeProvider) RebalancerOption {
	return func(r *Rebalancer) error {
		r.clock = clock
		return nil
	}
}

// RebalancerBackoff sets a beck off duration
func RebalancerBackoff(d time.Duration) RebalancerOption {
	return func(r *Rebalancer) error {
		r.backoffDuration = d
		return nil
	}
}

// RebalancerMeter sets a Meter builder function
func RebalancerMeter(newMeter NewMeterFn) RebalancerOption {
	return func(r *Rebalancer) error {
		r.newMeter = newMeter
		return nil
	}
}

// RebalancerErrorHandler is a functional argument that sets error handler of the server
func RebalancerErrorHandler(h utils.ErrorHandler) RebalancerOption {
	return func(r *Rebalancer) error {
		r.errHandler = h
		return nil
	}
}

// RebalancerStickySession sets a sticky session
func RebalancerStickySession(stickySession *StickySession) RebalancerOption {
	return func(r *Rebalancer) error {
		r.stickySession = stickySession
		return nil
	}
}

// RebalancerRequestRewriteListener is a functional argument that sets error handler of the server
func RebalancerRequestRewriteListener(rrl RequestRewriteListener) RebalancerOption {
	return func(r *Rebalancer) error {
		r.requestRewriteListener = rrl
		return nil
	}
}

// NewRebalancer creates a new Rebalancer
func NewRebalancer(handler balancerHandler, opts ...RebalancerOption) (*Rebalancer, error) {
	rb := &Rebalancer{
		mtx:           &sync.Mutex{},
		next:          handler,
		stickySession: nil,

		log: log.StandardLogger(),
	}
	for _, o := range opts {
		if err := o(rb); err != nil {
			return nil, err
		}
	}
	if rb.clock == nil {
		rb.clock = &timetools.RealTime{}
	}
	if rb.backoffDuration == 0 {
		rb.backoffDuration = 10 * time.Second
	}
	if rb.newMeter == nil {
		rb.newMeter = func() (Meter, error) {
			rc, err := memmetrics.NewRatioCounter(10, time.Second, memmetrics.RatioClock(rb.clock))
			if err != nil {
				return nil, err
			}
			return &codeMeter{
				r:     rc,
				codeS: http.StatusInternalServerError,
				codeE: http.StatusGatewayTimeout + 1,
			}, nil
		}
	}
	if rb.errHandler == nil {
		rb.errHandler = utils.DefaultHandler
	}
	return rb, nil
}

// RebalancerLogger defines the logger the rebalancer will use.
//
// It defaults to logrus.StandardLogger(), the global logger used by logrus.
func RebalancerLogger(l *log.Logger) RebalancerOption {
	return func(rb *Rebalancer) error {
		rb.log = l
		return nil
	}
}

// Servers gets all servers
func (rb *Rebalancer) Servers() []*url.URL {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	return rb.next.Servers()
}

func (rb *Rebalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if rb.log.Level >= log.DebugLevel {
		logEntry := rb.log.WithField("Request", utils.DumpHttpRequest(req))
		logEntry.Debug("vulcand/oxy/roundrobin/rebalancer: begin ServeHttp on request")
		defer logEntry.Debug("vulcand/oxy/roundrobin/rebalancer: completed ServeHttp on request")
	}

	pw := utils.NewProxyWriter(w)
	start := rb.clock.UtcNow()

	// make shallow copy of request before changing anything to avoid side effects
	newReq := *req
	stuck := false

	if rb.stickySession != nil {
		cookieUrl, present, err := rb.stickySession.GetBackend(&newReq, rb.Servers())

		if err != nil {
			log.Warnf("vulcand/oxy/roundrobin/rebalancer: error using server from cookie: %v", err)
		}

		if present {
			newReq.URL = cookieUrl
			stuck = true
		}
	}

	if !stuck {
		fwdURL, err := rb.next.NextServer()
		if err != nil {
			rb.errHandler.ServeHTTP(w, req, err)
			return
		}

		if log.GetLevel() >= log.DebugLevel {
			// log which backend URL we're sending this request to
			log.WithFields(log.Fields{"Request": utils.DumpHttpRequest(req), "ForwardURL": fwdURL}).Debugf("vulcand/oxy/roundrobin/rebalancer: Forwarding this request to URL")
		}

		if rb.stickySession != nil {
			rb.stickySession.StickBackend(fwdURL, &w)
		}

		newReq.URL = fwdURL
	}

	// Emit event to a listener if one exists
	if rb.requestRewriteListener != nil {
		rb.requestRewriteListener(req, &newReq)
	}

	rb.next.Next().ServeHTTP(pw, &newReq)

	rb.recordMetrics(newReq.URL, pw.StatusCode(), rb.clock.UtcNow().Sub(start))
	rb.adjustWeights()
}

func (rb *Rebalancer) recordMetrics(u *url.URL, code int, latency time.Duration) {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()
	if srv, i := rb.findServer(u); i != -1 {
		srv.meter.Record(code, latency)
	}
}

func (rb *Rebalancer) reset() {
	for _, s := range rb.servers {
		s.curWeight = s.origWeight
		rb.next.UpsertServer(s.url, Weight(s.origWeight))
	}
	rb.timer = rb.clock.UtcNow().Add(-1 * time.Second)
	rb.ratings = make([]float64, len(rb.servers))
}

// Wrap sets the next handler to be called by rebalancer handler.
func (rb *Rebalancer) Wrap(next balancerHandler) error {
	if rb.next != nil {
		return fmt.Errorf("already bound to %T", rb.next)
	}
	rb.next = next
	return nil
}

// UpsertServer upsert a server
func (rb *Rebalancer) UpsertServer(u *url.URL, options ...ServerOption) error {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	if err := rb.next.UpsertServer(u, options...); err != nil {
		return err
	}
	weight, _ := rb.next.ServerWeight(u)
	if err := rb.upsertServer(u, weight); err != nil {
		rb.next.RemoveServer(u)
		return err
	}
	rb.reset()
	return nil
}

// RemoveServer remove a server
func (rb *Rebalancer) RemoveServer(u *url.URL) error {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	return rb.removeServer(u)
}

func (rb *Rebalancer) removeServer(u *url.URL) error {
	_, i := rb.findServer(u)
	if i == -1 {
		return fmt.Errorf("%v not found", u)
	}
	if err := rb.next.RemoveServer(u); err != nil {
		return err
	}
	rb.servers = append(rb.servers[:i], rb.servers[i+1:]...)
	rb.reset()
	return nil
}

func (rb *Rebalancer) upsertServer(u *url.URL, weight int) error {
	if s, i := rb.findServer(u); i != -1 {
		s.origWeight = weight
	}
	meter, err := rb.newMeter()
	if err != nil {
		return err
	}
	rbSrv := &rbServer{
		url:        utils.CopyURL(u),
		origWeight: weight,
		curWeight:  weight,
		meter:      meter,
	}
	rb.servers = append(rb.servers, rbSrv)
	return nil
}

func (rb *Rebalancer) findServer(u *url.URL) (*rbServer, int) {
	if len(rb.servers) == 0 {
		return nil, -1
	}
	for i, s := range rb.servers {
		if sameURL(u, s.url) {
			return s, i
		}
	}
	return nil, -1
}

// adjustWeights Called on every load balancer ServeHTTP call, returns the suggested weights
// on every call, can adjust weights if needed.
func (rb *Rebalancer) adjustWeights() {
	rb.mtx.Lock()
	defer rb.mtx.Unlock()

	// In this case adjusting weights would have no effect, so do nothing
	if len(rb.servers) < 2 {
		return
	}
	// Metrics are not ready
	if !rb.metricsReady() {
		return
	}
	if !rb.timerExpired() {
		return
	}
	if rb.markServers() {
		if rb.setMarkedWeights() {
			rb.setTimer()
		}
	} else { // No servers that are different by their quality, so converge weights
		if rb.convergeWeights() {
			rb.setTimer()
		}
	}
}

func (rb *Rebalancer) applyWeights() {
	for _, srv := range rb.servers {
		rb.log.Debugf("upsert server %v, weight %v", srv.url, srv.curWeight)
		rb.next.UpsertServer(srv.url, Weight(srv.curWeight))
	}
}

func (rb *Rebalancer) setMarkedWeights() bool {
	changed := false
	// Increase weights on servers marked as good
	for _, srv := range rb.servers {
		if srv.good {
			weight := increase(srv.curWeight)
			if weight <= FSMMaxWeight {
				rb.log.Debugf("increasing weight of %v from %v to %v", srv.url, srv.curWeight, weight)
				srv.curWeight = weight
				changed = true
			}
		}
	}
	if changed {
		rb.normalizeWeights()
		rb.applyWeights()
		return true
	}
	return false
}

func (rb *Rebalancer) setTimer() {
	rb.timer = rb.clock.UtcNow().Add(rb.backoffDuration)
}

func (rb *Rebalancer) timerExpired() bool {
	return rb.timer.Before(rb.clock.UtcNow())
}

func (rb *Rebalancer) metricsReady() bool {
	for _, s := range rb.servers {
		if !s.meter.IsReady() {
			return false
		}
	}
	return true
}

// markServers splits servers into two groups of servers with bad and good failure rate.
// It does compare relative performances of the servers though, so if all servers have approximately the same error rate
// this function returns the result as if all servers are equally good.
func (rb *Rebalancer) markServers() bool {
	for i, srv := range rb.servers {
		rb.ratings[i] = srv.meter.Rating()
	}
	g, b := memmetrics.SplitFloat64(splitThreshold, 0, rb.ratings)
	for i, srv := range rb.servers {
		if g[rb.ratings[i]] {
			srv.good = true
		} else {
			srv.good = false
		}
	}
	if len(g) != 0 && len(b) != 0 {
		rb.log.Debugf("bad: %v good: %v, ratings: %v", b, g, rb.ratings)
	}
	return len(g) != 0 && len(b) != 0
}

func (rb *Rebalancer) convergeWeights() bool {
	// If we have previously changed servers try to restore weights to the original state
	changed := false
	for _, s := range rb.servers {
		if s.origWeight == s.curWeight {
			continue
		}
		changed = true
		newWeight := decrease(s.origWeight, s.curWeight)
		log.Debugf("decreasing weight of %v from %v to %v", s.url, s.curWeight, newWeight)
		s.curWeight = newWeight
	}
	if !changed {
		return false
	}
	rb.normalizeWeights()
	rb.applyWeights()
	return true
}

func (rb *Rebalancer) weightsGcd() int {
	divisor := -1
	for _, w := range rb.servers {
		if divisor == -1 {
			divisor = w.curWeight
		} else {
			divisor = gcd(divisor, w.curWeight)
		}
	}
	return divisor
}

func (rb *Rebalancer) normalizeWeights() {
	gcd := rb.weightsGcd()
	if gcd <= 1 {
		return
	}
	for _, s := range rb.servers {
		s.curWeight = s.curWeight / gcd
	}
}

func increase(weight int) int {
	return weight * FSMGrowFactor
}

func decrease(target, current int) int {
	adjusted := current / FSMGrowFactor
	if adjusted < target {
		return target
	}
	return adjusted
}

// rebalancer server record that keeps track of the original weight supplied by user
type rbServer struct {
	url        *url.URL
	origWeight int // original weight supplied by user
	curWeight  int // current weight
	good       bool
	meter      Meter
}

const (
	// FSMMaxWeight is the maximum weight that handler will set for the server
	FSMMaxWeight = 4096
	// FSMGrowFactor Multiplier for the server weight
	FSMGrowFactor = 4
)

type codeMeter struct {
	r     *memmetrics.RatioCounter
	codeS int
	codeE int
}

// Rating gets ratio
func (n *codeMeter) Rating() float64 {
	return n.r.Ratio()
}

// Record records a meter
func (n *codeMeter) Record(code int, d time.Duration) {
	if code >= n.codeS && code < n.codeE {
		n.r.IncA(1)
	} else {
		n.r.IncB(1)
	}
}

// IsReady returns true if the counter is ready
func (n *codeMeter) IsReady() bool {
	return n.r.IsReady()
}

// splitThreshold tells how far the value should go from the median + median absolute deviation before it is considered an outlier
const splitThreshold = 1.5
