// Package model defines interfaces and structures controlling the proxy configuration.
package engine

import (
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vulcand/oxy/memmetrics"
	"github.com/vulcand/oxy/stream"
	"github.com/vulcand/route"
	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/router"
)

// StatsProvider provides realtime stats abount endpoints, backends and locations
type StatsProvider interface {
	FrontendStats(FrontendKey) (*RoundTripStats, error)
	ServerStats(ServerKey) (*RoundTripStats, error)
	BackendStats(BackendKey) (*RoundTripStats, error)

	// TopFrontends returns locations sorted by criteria (faulty, slow, most used)
	// if hostname or backendId is present, will filter out locations for that host or backendId
	TopFrontends(*BackendKey) ([]Frontend, error)

	// TopServers returns endpoints sorted by criteria (faulty, slow, mos used)
	// if backendId is not empty, will filter out endpoints for that backendId
	TopServers(*BackendKey) ([]Server, error)
}

type KeyPair struct {
	Key  []byte
	Cert []byte
}

func NewKeyPair(cert, key []byte) (*KeyPair, error) {
	if len(cert) == 0 || len(key) == 0 {
		return nil, fmt.Errorf("Provide non-empty certificate and a private key")
	}
	if _, err := tls.X509KeyPair(cert, key); err != nil {
		return nil, err
	}
	return &KeyPair{Cert: cert, Key: key}, nil
}

func (c *KeyPair) Equals(o *KeyPair) bool {
	return (len(c.Cert) == len(o.Cert)) &&
		(len(c.Key) == len(o.Key)) &&
		subtle.ConstantTimeCompare(c.Cert, o.Cert) == 1 &&
		subtle.ConstantTimeCompare(c.Key, o.Key) == 1
}

type Address struct {
	Network string
	Address string
}

type ListenerKey struct {
	Id string
}

func (l ListenerKey) String() string {
	return l.Id
}

// Listener specifies the listening point - the network and interface for each host. Host can have multiple interfaces.
type Listener struct {
	Id string
	// HTTP or HTTPS
	Protocol string
	// Adddress specifies network (tcp or unix) and address (ip:port or path to unix socket)
	Address Address
	// Scope is optional expression that limits the operational scope of this listener
	Scope string
	// Settings provides listener-type specific settings, e.g. TLS settings for HTTPS listener
	Settings *HTTPSListenerSettings `json:",omitempty"`
}

func (l *Listener) TLSConfig() (*tls.Config, error) {
	if l.Protocol != HTTPS {
		return nil, fmt.Errorf("wrong listener proto: %v", l.Protocol)
	}
	if l.Settings == nil {
		return NewTLSConfig(&TLSSettings{})
	}
	return NewTLSConfig(&l.Settings.TLS)
}

func (l *Listener) String() string {
	return fmt.Sprintf("Listener(%s, %s://%s, scope=%s)", l.Protocol, l.Address.Network, l.Address.Address, l.Scope)
}

func (a *Address) Equals(o Address) bool {
	return a.Network == o.Network && a.Address == o.Address
}

func (l *Listener) SettingsEquals(o *Listener) bool {
	if l.Settings == nil && o.Settings == nil {
		return true
	}
	ls, os := l.Settings, o.Settings
	if (ls == nil && os != nil) || (ls != nil && os == nil) {
		return false
	}
	return (&os.TLS).Equals(&ls.TLS)
}

type HTTPSListenerSettings struct {
	TLS TLSSettings
}

// Sets up OCSP stapling, see http://en.wikipedia.org/wiki/OCSP_stapling
type OCSPSettings struct {
	Enabled bool
	Period  string
	// Optional responders. Responder is the CA-operated HTTP server that responds with revocation status
	// If set, this field will override
	Responders         []string
	SkipSignatureCheck bool
}

func (o *OCSPSettings) RefreshPeriod() (time.Duration, error) {
	if o.Period == "" {
		return time.Hour, nil
	}
	return time.ParseDuration(o.Period)
}

func (o *OCSPSettings) Equals(other *OCSPSettings) bool {
	if o.Enabled != other.Enabled {
		return false
	}
	p, err := o.RefreshPeriod()
	if err != nil {
		return false
	}
	p2, err := other.RefreshPeriod()
	if err != nil {
		return false
	}
	if p != p2 {
		return false
	}
	if len(o.Responders) != len(other.Responders) {
		return false
	}
	if len(o.Responders) == 0 || len(other.Responders) == 0 {
		return true
	}
	for i := range o.Responders {
		if o.Responders[i] != other.Responders[i] {
			return false
		}
	}
	return true
}

type HostSettings struct {
	Default bool
	KeyPair *KeyPair
	OCSP    OCSPSettings
}

type HostKey struct {
	Name string
}

func (h HostKey) String() string {
	return h.Name
}

// Incoming requests are matched by their hostname first. Hostname is defined by incoming 'Host' header.
// E.g. curl http://example.com/alice will be matched by the host example.com first.
type Host struct {
	Name     string
	Settings HostSettings
}

func NewHost(name string, settings HostSettings) (*Host, error) {
	if name == "" {
		return nil, fmt.Errorf("Hostname can not be empty")
	}
	return &Host{
		Name:     name,
		Settings: settings,
	}, nil
}

func (h *Host) String() string {
	return fmt.Sprintf("Host(%s, keyPair=%t, ocsp=%t)", h.Name, h.Settings.KeyPair != nil, h.Settings.OCSP.Enabled)
}

func (h *Host) GetId() string {
	return h.Name
}

// Frontend is connected to a backend and vulcand will use the servers from this backend.
type Frontend struct {
	Id        string
	Route     string
	Type      string
	BackendId string

	Stats    *RoundTripStats `json:",omitempty"`
	Settings interface{}     `json:",omitempty"`
}

// Limits contains various limits one can supply for a location.
type HTTPFrontendLimits struct {
	MaxMemBodyBytes int64 // Maximum size to keep in memory before buffering to disk
	MaxBodyBytes    int64 // Maximum size of a request body in bytes
}

// Additional options to control this location, such as timeouts
type HTTPFrontendSettings struct {
	// Limits contains various limits one can supply for a location.
	Limits HTTPFrontendLimits
	// Predicate that defines when requests are allowed to failover
	FailoverPredicate string
	// Used in forwarding headers
	Hostname string
	// In this case appends new forward info to the existing header
	TrustForwardHeader bool
	// Should host header be forwarded as-is?
	PassHostHeader bool
}

func NewAddress(network, address string) (*Address, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("supply a non empty address")
	}

	network = strings.ToLower(network)
	if network != TCP && network != UNIX {
		return nil, fmt.Errorf("unsupported network '%s', supported networks are tcp and unix", network)
	}

	return &Address{Network: network, Address: address}, nil
}

func NewListener(id, protocol, network, address, scope string, settings *HTTPSListenerSettings) (*Listener, error) {
	protocol = strings.ToLower(protocol)
	if protocol != HTTP && protocol != HTTPS {
		return nil, fmt.Errorf("unsupported protocol '%s', supported protocols are http and https", protocol)
	}

	if scope != "" {
		if !route.IsValid(scope) {
			return nil, fmt.Errorf("Scope should be a valid route expression")
		}
	}

	a, err := NewAddress(network, address)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Scope:    scope,
		Id:       id,
		Address:  *a,
		Protocol: protocol,
		Settings: settings,
	}, nil
}

func NewHTTPFrontend(router router.Router, id, backendId string, routeExpr string, settings HTTPFrontendSettings) (*Frontend, error) {
	if len(id) == 0 || len(backendId) == 0 {
		return nil, fmt.Errorf("supply valid  route, id, and backendId")
	}

	// Make sure location path is a valid route expression
	if !router.IsValid(routeExpr) {
		return nil, fmt.Errorf("route should be a valid route expression: %s", routeExpr)
	}

	if settings.FailoverPredicate != "" && !stream.IsValidExpression(settings.FailoverPredicate) {
		return nil, fmt.Errorf("invalid failover predicate: %s", settings.FailoverPredicate)
	}

	return &Frontend{
		Id:        id,
		BackendId: backendId,
		Route:     routeExpr,
		Type:      HTTP,
		Settings:  settings,
	}, nil
}

func (f *Frontend) HTTPSettings() HTTPFrontendSettings {
	return (f.Settings).(HTTPFrontendSettings)
}

func (l HTTPFrontendSettings) Equals(o HTTPFrontendSettings) bool {
	return (l.Limits.MaxMemBodyBytes == o.Limits.MaxMemBodyBytes &&
		l.Limits.MaxBodyBytes == o.Limits.MaxBodyBytes &&
		l.FailoverPredicate == o.FailoverPredicate &&
		l.Hostname == o.Hostname &&
		l.TrustForwardHeader == o.TrustForwardHeader)
}

func (f *Frontend) String() string {
	return fmt.Sprintf("Frontend(%v, %v, %v)", f.Type, f.Id, f.BackendId)
}

func (l *Frontend) GetId() string {
	return l.Id
}

func (l *Frontend) GetKey() FrontendKey {
	return FrontendKey{Id: l.Id}
}

type HTTPBackendTimeouts struct {
	// Socket read timeout (before we receive the first reply header)
	Read string
	// Socket connect timeout
	Dial string
	// TLS handshake timeout
	TLSHandshake string
}

type HTTPBackendKeepAlive struct {
	// Keepalive period
	Period string
	// How many idle connections will be kept per host
	MaxIdleConnsPerHost int
}

type HTTPBackendSettings struct {
	// Timeouts provides timeout settings for backend servers
	Timeouts HTTPBackendTimeouts
	// KeepAlive controls keep-alive settings for backend servers
	KeepAlive HTTPBackendKeepAlive
	// TLS provides optional TLS settings for HTTP backend
	TLS *TLSSettings `json:",omitempty"`
}

func (s *HTTPBackendSettings) Equals(o HTTPBackendSettings) bool {
	return (s.Timeouts.Read == o.Timeouts.Read &&
		s.Timeouts.Dial == o.Timeouts.Dial &&
		s.Timeouts.TLSHandshake == o.Timeouts.TLSHandshake &&
		s.KeepAlive.Period == o.KeepAlive.Period &&
		s.KeepAlive.MaxIdleConnsPerHost == o.KeepAlive.MaxIdleConnsPerHost &&
		((s.TLS == nil && o.TLS == nil) ||
			((s.TLS != nil && o.TLS != nil) && s.TLS.Equals(o.TLS))))
}

type MiddlewareKey struct {
	FrontendKey FrontendKey
	Id          string
}

func (m MiddlewareKey) String() string {
	return fmt.Sprintf("%v.%v", m.FrontendKey, m.Id)
}

// Middleware contains information about this middleware backend-specific data used for serialization/deserialization
type Middleware struct {
	Id         string
	Priority   int
	Type       string
	Middleware plugin.Middleware
}

// Backend is a collection of endpoints. Each location is assigned an backend. Changing assigned backend
// of the location gracefully redirects the traffic to the new endpoints of the backend.
type Backend struct {
	Id       string
	Type     string
	Stats    *RoundTripStats `json:",omitempty"`
	Settings interface{}
}

// NewBackend creates a new instance of the backend object
func NewHTTPBackend(id string, s HTTPBackendSettings) (*Backend, error) {
	if _, err := transportSettings(s); err != nil {
		return nil, err
	}
	return &Backend{
		Id:       id,
		Type:     HTTP,
		Settings: s,
	}, nil
}

func (b *Backend) HTTPSettings() HTTPBackendSettings {
	return b.Settings.(HTTPBackendSettings)
}

func (b *Backend) String() string {
	return fmt.Sprintf("Backend(id=%s)", b.Id)
}

func (b *Backend) GetId() string {
	return b.Id
}

func (b *Backend) GetUniqueId() BackendKey {
	return BackendKey{Id: b.Id}
}

func (b *Backend) TransportSettings() (*TransportSettings, error) {
	return transportSettings(b.Settings.(HTTPBackendSettings))
}

func transportSettings(s HTTPBackendSettings) (*TransportSettings, error) {
	t := &TransportSettings{}
	var err error
	// Connection timeouts
	if len(s.Timeouts.Read) != 0 {
		if t.Timeouts.Read, err = time.ParseDuration(s.Timeouts.Read); err != nil {
			return nil, fmt.Errorf("invalid read timeout: %s", err)
		}
	}
	if len(s.Timeouts.Dial) != 0 {
		if t.Timeouts.Dial, err = time.ParseDuration(s.Timeouts.Dial); err != nil {
			return nil, fmt.Errorf("invalid dial timeout: %s", err)
		}
	}
	if len(s.Timeouts.TLSHandshake) != 0 {
		if t.Timeouts.TLSHandshake, err = time.ParseDuration(s.Timeouts.TLSHandshake); err != nil {
			return nil, fmt.Errorf("invalid tls handshake timeout: %s", err)
		}
	}

	// Keep Alive parameters
	if len(s.KeepAlive.Period) != 0 {
		if t.KeepAlive.Period, err = time.ParseDuration(s.KeepAlive.Period); err != nil {
			return nil, fmt.Errorf("invalid tls handshake timeout: %s", err)
		}
	}
	t.KeepAlive.MaxIdleConnsPerHost = s.KeepAlive.MaxIdleConnsPerHost

	if s.TLS != nil {
		config, err := NewTLSConfig(s.TLS)
		if err != nil {
			return nil, err
		}
		t.TLS = config
	}
	return t, nil
}

// Server is a final destination of the request
type Server struct {
	Id    string
	URL   string
	Stats *RoundTripStats `json:",omitempty"`
}

func NewServer(id, u string) (*Server, error) {
	if _, err := url.ParseRequestURI(u); err != nil {
		return nil, fmt.Errorf("endpoint url '%s' is not valid", u)
	}
	return &Server{
		Id:  id,
		URL: u,
	}, nil
}

func (e *Server) String() string {
	return fmt.Sprintf("HTTPServer(%s, %s, %s)", e.Id, e.URL, e.Stats)
}

func (e *Server) GetId() string {
	return e.Id
}

type LatencyBrackets []Bracket

func (l LatencyBrackets) GetQuantile(q float64) (*Bracket, error) {
	if len(l) == 0 {
		return nil, fmt.Errorf("quantile %f not found in %v", q, l)
	}
	for _, b := range l {
		if b.Quantile == q {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("quantile %f not found", q)
}

// RoundTripStats contain real time statistics about performance of Server or Frontend
// such as latency, processed and failed requests.
type RoundTripStats struct {
	Verdict         Verdict
	Counters        Counters
	LatencyBrackets LatencyBrackets
}

func NewRoundTripStats(m *memmetrics.RTMetrics) (*RoundTripStats, error) {
	codes := m.StatusCodesCounts()

	sc := make([]StatusCode, 0, len(codes))
	for k, v := range codes {
		if v != 0 {
			sc = append(sc, StatusCode{Code: k, Count: v})
		}
	}

	h, err := m.LatencyHistogram()
	if err != nil {
		return nil, err
	}

	return &RoundTripStats{
		Counters: Counters{
			NetErrors:   m.NetworkErrorCount(),
			Total:       m.TotalCount(),
			Period:      m.CounterWindowSize(),
			StatusCodes: sc,
		},
		LatencyBrackets: NewBrackets(h),
	}, nil
}

// NetErroRate calculates the amont of ntwork errors such as time outs and dropped connection
// that occured in the given time window
func (e *RoundTripStats) NetErrorRatio() float64 {
	if e.Counters.Total == 0 {
		return 0
	}
	return (float64(e.Counters.NetErrors) / float64(e.Counters.Total))
}

// AppErrorRate calculates the ratio of 500 responses that designate internal server errors
// to success responses - 2xx, it specifically not counts 4xx or any other than 500 error to avoid noisy results.
func (e *RoundTripStats) AppErrorRatio() float64 {
	return e.ResponseCodeRatio(http.StatusInternalServerError, http.StatusInternalServerError+1, 200, 300)
}

// ResponseCodeRatio calculates ratio of count(startA to endA) / count(startB to endB)
func (e *RoundTripStats) ResponseCodeRatio(startA, endA, startB, endB int) float64 {
	a := int64(0)
	b := int64(0)
	for _, status := range e.Counters.StatusCodes {
		if status.Code < endA && status.Code >= startA {
			a += status.Count
		}
		if status.Code < endB && status.Code >= startB {
			b += status.Count
		}
	}
	if b != 0 {
		return float64(a) / float64(b)
	}
	return 0
}

func (e *RoundTripStats) RequestsPerSecond() float64 {
	if e.Counters.Period == 0 {
		return 0
	}
	return float64(e.Counters.Total) / float64(e.Counters.Period/time.Second)
}

func (e *RoundTripStats) String() string {
	return fmt.Sprintf("%.2f requests/sec, %.2f failures/sec", e.RequestsPerSecond(), e.NetErrorRatio())
}

type Verdict struct {
	IsBad     bool
	Anomalies []Anomaly
}

func (v Verdict) String() string {
	return fmt.Sprintf("verdict[bad=%t, anomalies=%v]", v.IsBad, v.Anomalies)
}

type Anomaly struct {
	Code    int
	Message string
}

func (a Anomaly) String() string {
	return fmt.Sprintf("(%d) %s", a.Code, a.Message)
}

type NotFoundError struct {
	Message string
}

func (n *NotFoundError) Error() string {
	if n.Message != "" {
		return n.Message
	} else {
		return "Object not found"
	}
}

type InvalidFormatError struct {
	Message string
}

func (n *InvalidFormatError) Error() string {
	if n.Message != "" {
		return n.Message
	} else {
		return "Invalid format"
	}
}

type AlreadyExistsError struct {
	Message string
}

func (n *AlreadyExistsError) Error() string {
	return n.Message
}

type Counters struct {
	Period      time.Duration
	NetErrors   int64
	Total       int64
	StatusCodes []StatusCode
}

type StatusCode struct {
	Code  int
	Count int64
}

type Bracket struct {
	Quantile float64
	Value    time.Duration
}

func NewBrackets(h *memmetrics.HDRHistogram) []Bracket {
	quantiles := []float64{50, 75, 95, 99, 99.9}
	brackets := make([]Bracket, len(quantiles))

	for i, v := range quantiles {
		brackets[i] = Bracket{
			Quantile: v,
			Value:    time.Duration(h.ValueAtQuantile(v)) * time.Microsecond,
		}
	}
	return brackets
}

type FrontendKey struct {
	Id string
}

func (f FrontendKey) String() string {
	return f.Id
}

type ServerKey struct {
	BackendKey BackendKey
	Id         string
}

func (e ServerKey) String() string {
	return fmt.Sprintf("%v.%v", e.BackendKey, e.Id)
}

func ParseServerKey(v string) (*ServerKey, error) {
	out := strings.SplitN(v, ".", 2)
	if len(out) != 2 {
		return nil, fmt.Errorf("invalid id: '%s'", v)
	}
	return &ServerKey{BackendKey: BackendKey{Id: out[0]}, Id: out[1]}, nil
}

func MustParseServerKey(v string) ServerKey {
	k, err := ParseServerKey(v)
	if err != nil {
		panic(err)
	}
	return *k
}

type BackendKey struct {
	Id string
}

func (u BackendKey) String() string {
	return u.Id
}

const (
	HTTP  = "http"
	HTTPS = "https"
	TCP   = "tcp"
	UNIX  = "unix"
	NoTTL = 0
)

type TransportTimeouts struct {
	// Socket read timeout (before we receive the first reply header)
	Read time.Duration
	// Socket connect timeout
	Dial time.Duration
	// TLS handshake timeout
	TLSHandshake time.Duration
}

type TransportKeepAlive struct {
	// Keepalive period
	Period time.Duration
	// How many idle connections will be kept per host
	MaxIdleConnsPerHost int
}

type TransportSettings struct {
	Timeouts  TransportTimeouts
	KeepAlive TransportKeepAlive
	TLS       *tls.Config
}
