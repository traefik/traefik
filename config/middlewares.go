package config

import (
	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/ip"
)

// Middleware holds the Middleware configuration.
type Middleware struct {
	AddPrefix         *AddPrefix         `json:"addPrefix,omitempty"`
	StripPrefix       *StripPrefix       `json:"stripPrefix,omitempty"`
	StripPrefixRegex  *StripPrefixRegex  `json:"stripPrefixRegex,omitempty"`
	ReplacePath       *ReplacePath       `json:"replacePath,omitempty"`
	ReplacePathRegex  *ReplacePathRegex  `json:"replacePathRegex,omitempty"`
	Chain             *Chain             `json:"chain,omitempty"`
	IPWhiteList       *IPWhiteList       `json:"ipWhiteList,omitempty"`
	Headers           *Headers           `json:"headers,omitempty"`
	Errors            *ErrorPage         `json:"errors,omitempty"`
	RateLimit         *RateLimit         `json:"rateLimit,omitempty"`
	Redirect          *Redirect          `json:"redirect,omitempty"`
	BasicAuth         *BasicAuth         `json:"basicAuth,omitempty"`
	DigestAuth        *DigestAuth        `json:"digestAuth,omitempty"`
	ForwardAuth       *ForwardAuth       `json:"forwardAuth,omitempty"`
	MaxConn           *MaxConn           `json:"maxConn,omitempty"`
	Buffering         *Buffering         `json:"buffering,omitempty"`
	CircuitBreaker    *CircuitBreaker    `json:"circuitBreaker,omitempty"`
	Compress          *Compress          `json:"compress,omitempty" label:"allowEmpty"`
	PassTLSClientCert *PassTLSClientCert `json:"passTLSClientCert,omitempty"`
	Retry             *Retry             `json:"retry,omitempty"`
}

// AddPrefix holds the AddPrefix configuration.
type AddPrefix struct {
	Prefix string `json:"prefix,omitempty"`
}

// Auth holds the authentication configuration (BASIC, DIGEST, users).
type Auth struct {
	Basic   *BasicAuth   `json:"basic,omitempty" export:"true"`
	Digest  *DigestAuth  `json:"digest,omitempty" export:"true"`
	Forward *ForwardAuth `json:"forward,omitempty" export:"true"`
}

// BasicAuth holds the HTTP basic authentication configuration.
type BasicAuth struct {
	Users        `json:"users,omitempty" mapstructure:","`
	UsersFile    string `json:"usersFile,omitempty"`
	Realm        string `json:"realm,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	HeaderField  string `json:"headerField,omitempty" export:"true"`
}

// Buffering holds the request/response buffering configuration.
type Buffering struct {
	MaxRequestBodyBytes  int64  `json:"maxRequestBodyBytes,omitempty"`
	MemRequestBodyBytes  int64  `json:"memRequestBodyBytes,omitempty"`
	MaxResponseBodyBytes int64  `json:"maxResponseBodyBytes,omitempty"`
	MemResponseBodyBytes int64  `json:"memResponseBodyBytes,omitempty"`
	RetryExpression      string `json:"retryExpression,omitempty"`
}

// Chain holds a chain of middlewares
type Chain struct {
	Middlewares []string `json:"middlewares"`
}

// CircuitBreaker holds the circuit breaker configuration.
type CircuitBreaker struct {
	Expression string `json:"expression,omitempty"`
}

// Compress holds the compress configuration.
type Compress struct{}

// DigestAuth holds the Digest HTTP authentication configuration.
type DigestAuth struct {
	Users        `json:"users,omitempty" mapstructure:","`
	UsersFile    string `json:"usersFile,omitempty"`
	RemoveHeader bool   `json:"removeHeader,omitempty"`
	Realm        string `json:"realm,omitempty" mapstructure:","`
	HeaderField  string `json:"headerField,omitempty" export:"true"`
}

// ErrorPage holds the custom error page configuration.
type ErrorPage struct {
	Status  []string `json:"status,omitempty"`
	Service string   `json:"service,omitempty"`
	Query   string   `json:"query,omitempty"`
}

// ForwardAuth holds the http forward authentication configuration.
type ForwardAuth struct {
	Address             string     `description:"Authentication server address" json:"address,omitempty"`
	TLS                 *ClientTLS `description:"Enable TLS support" json:"tls,omitempty" export:"true"`
	TrustForwardHeader  bool       `description:"Trust X-Forwarded-* headers" json:"trustForwardHeader,omitempty" export:"true"`
	AuthResponseHeaders []string   `description:"Headers to be forwarded from auth response" json:"authResponseHeaders,omitempty"`
}

// Headers holds the custom header configuration.
type Headers struct {
	CustomRequestHeaders  map[string]string `json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty"`

	AllowedHosts            []string          `json:"allowedHosts,omitempty"`
	HostsProxyHeaders       []string          `json:"hostsProxyHeaders,omitempty"`
	SSLRedirect             bool              `json:"sslRedirect,omitempty"`
	SSLTemporaryRedirect    bool              `json:"sslTemporaryRedirect,omitempty"`
	SSLHost                 string            `json:"sslHost,omitempty"`
	SSLProxyHeaders         map[string]string `json:"sslProxyHeaders,omitempty"`
	SSLForceHost            bool              `json:"sslForceHost,omitempty"`
	STSSeconds              int64             `json:"stsSeconds,omitempty"`
	STSIncludeSubdomains    bool              `json:"stsIncludeSubdomains,omitempty"`
	STSPreload              bool              `json:"stsPreload,omitempty"`
	ForceSTSHeader          bool              `json:"forceSTSHeader,omitempty"`
	FrameDeny               bool              `json:"frameDeny,omitempty"`
	CustomFrameOptionsValue string            `json:"customFrameOptionsValue,omitempty"`
	ContentTypeNosniff      bool              `json:"contentTypeNosniff,omitempty"`
	BrowserXSSFilter        bool              `json:"browserXssFilter,omitempty"`
	CustomBrowserXSSValue   string            `json:"customBrowserXSSValue,omitempty"`
	ContentSecurityPolicy   string            `json:"contentSecurityPolicy,omitempty"`
	PublicKey               string            `json:"publicKey,omitempty"`
	ReferrerPolicy          string            `json:"referrerPolicy,omitempty"`
	IsDevelopment           bool              `json:"isDevelopment,omitempty"`
}

// HasCustomHeadersDefined checks to see if any of the custom header elements have been set
func (h *Headers) HasCustomHeadersDefined() bool {
	return h != nil && (len(h.CustomResponseHeaders) != 0 ||
		len(h.CustomRequestHeaders) != 0)
}

// HasSecureHeadersDefined checks to see if any of the secure header elements have been set
func (h *Headers) HasSecureHeadersDefined() bool {
	return h != nil && (len(h.AllowedHosts) != 0 ||
		len(h.HostsProxyHeaders) != 0 ||
		h.SSLRedirect ||
		h.SSLTemporaryRedirect ||
		h.SSLForceHost ||
		h.SSLHost != "" ||
		len(h.SSLProxyHeaders) != 0 ||
		h.STSSeconds != 0 ||
		h.STSIncludeSubdomains ||
		h.STSPreload ||
		h.ForceSTSHeader ||
		h.FrameDeny ||
		h.CustomFrameOptionsValue != "" ||
		h.ContentTypeNosniff ||
		h.BrowserXSSFilter ||
		h.CustomBrowserXSSValue != "" ||
		h.ContentSecurityPolicy != "" ||
		h.PublicKey != "" ||
		h.ReferrerPolicy != "" ||
		h.IsDevelopment)
}

// IPStrategy holds the ip strategy configuration.
type IPStrategy struct {
	Depth       int      `json:"depth,omitempty" export:"true"`
	ExcludedIPs []string `json:"excludedIPs,omitempty"`
}

// Get an IP selection strategy
// if nil return the RemoteAddr strategy
// else return a strategy base on the configuration using the X-Forwarded-For Header.
// Depth override the ExcludedIPs
func (s *IPStrategy) Get() (ip.Strategy, error) {
	if s == nil {
		return &ip.RemoteAddrStrategy{}, nil
	}

	if s.Depth > 0 {
		return &ip.DepthStrategy{
			Depth: s.Depth,
		}, nil
	}

	if len(s.ExcludedIPs) > 0 {
		checker, err := ip.NewChecker(s.ExcludedIPs)
		if err != nil {
			return nil, err
		}
		return &ip.CheckerStrategy{
			Checker: checker,
		}, nil
	}

	return &ip.RemoteAddrStrategy{}, nil
}

// IPWhiteList holds the ip white list configuration.
type IPWhiteList struct {
	SourceRange []string    `json:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty" label:"allowEmpty"`
}

// MaxConn holds maximum connection configuration.
type MaxConn struct {
	Amount        int64  `json:"amount,omitempty"`
	ExtractorFunc string `json:"extractorFunc,omitempty"`
}

// SetDefaults Default values for a MaxConn.
func (m *MaxConn) SetDefaults() {
	m.ExtractorFunc = "request.host"
}

// PassTLSClientCert holds the TLS client cert headers configuration.
type PassTLSClientCert struct {
	PEM  bool                      `description:"Enable header with escaped client pem" json:"pem"`
	Info *TLSClientCertificateInfo `description:"Enable header with configured client cert info" json:"info,omitempty"`
}

// Rate holds the rate limiting configuration for a specific time period.
type Rate struct {
	Period  parse.Duration `json:"period,omitempty"`
	Average int64          `json:"average,omitempty"`
	Burst   int64          `json:"burst,omitempty"`
}

// RateLimit holds the rate limiting configuration for a given frontend.
type RateLimit struct {
	RateSet map[string]*Rate `json:"rateset,omitempty"`
	// FIXME replace by ipStrategy see oxy and replace
	ExtractorFunc string `json:"extractorFunc,omitempty"`
}

// SetDefaults Default values for a MaxConn.
func (r *RateLimit) SetDefaults() {
	r.ExtractorFunc = "request.host"
}

// Redirect holds the redirection configuration of an entry point to another, or to an URL.
type Redirect struct {
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
	Permanent   bool   `json:"permanent,omitempty"`
}

// SchemeRedirect holds the scheme redirection configuration.
type SchemeRedirect struct {
	Scheme    string `json:"scheme,omitempty"`
	Port      string `json:"port,omitempty"`
	Permanent bool   `json:"permanent,omitempty"`
}

// ReplacePath holds the ReplacePath configuration.
type ReplacePath struct {
	Path string `json:"path,omitempty"`
}

// ReplacePathRegex holds the ReplacePathRegex configuration.
type ReplacePathRegex struct {
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Retry holds the retry configuration.
type Retry struct {
	Attempts int `description:"Number of attempts" export:"true"`
}

// StripPrefix holds the StripPrefix configuration.
type StripPrefix struct {
	Prefixes []string `json:"prefixes,omitempty"`
}

// StripPrefixRegex holds the StripPrefixRegex configuration.
type StripPrefixRegex struct {
	Regex []string `json:"regex,omitempty"`
}

// TLSClientCertificateInfo holds the client TLS certificate info configuration.
type TLSClientCertificateInfo struct {
	NotAfter  bool                        `description:"Add NotAfter info in header" json:"notAfter"`
	NotBefore bool                        `description:"Add NotBefore info in header" json:"notBefore"`
	Sans      bool                        `description:"Add Sans info in header" json:"sans"`
	Subject   *TLSCLientCertificateDNInfo `description:"Add Subject info in header" json:"subject,omitempty"`
	Issuer    *TLSCLientCertificateDNInfo `description:"Add Issuer info in header" json:"issuer,omitempty"`
}

// TLSCLientCertificateDNInfo holds the client TLS certificate distinguished name info configuration
// cf https://tools.ietf.org/html/rfc3739
type TLSCLientCertificateDNInfo struct {
	Country         bool `description:"Add Country info in header" json:"country"`
	Province        bool `description:"Add Province info in header" json:"province"`
	Locality        bool `description:"Add Locality info in header" json:"locality"`
	Organization    bool `description:"Add Organization info in header" json:"organization"`
	CommonName      bool `description:"Add CommonName info in header" json:"commonName"`
	SerialNumber    bool `description:"Add SerialNumber info in header" json:"serialNumber"`
	DomainComponent bool `description:"Add Domain Component info in header" json:"domainComponent"`
}

// Users holds a list of users
type Users []string
