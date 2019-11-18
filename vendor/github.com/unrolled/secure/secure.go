package secure

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type secureCtxKey string

const (
	stsHeader            = "Strict-Transport-Security"
	stsSubdomainString   = "; includeSubDomains"
	stsPreloadString     = "; preload"
	frameOptionsHeader   = "X-Frame-Options"
	frameOptionsValue    = "DENY"
	contentTypeHeader    = "X-Content-Type-Options"
	contentTypeValue     = "nosniff"
	xssProtectionHeader  = "X-XSS-Protection"
	xssProtectionValue   = "1; mode=block"
	cspHeader            = "Content-Security-Policy"
	cspReportOnlyHeader  = "Content-Security-Policy-Report-Only"
	hpkpHeader           = "Public-Key-Pins"
	referrerPolicyHeader = "Referrer-Policy"
	featurePolicyHeader  = "Feature-Policy"
	expectCTHeader       = "Expect-CT"

	ctxSecureHeaderKey = secureCtxKey("SecureResponseHeader")
	cspNonceSize       = 16
)

// SSLHostFunc a type whose pointer is the type of field `SSLHostFunc` of `Options` struct
type SSLHostFunc func(host string) (newHost string)

func defaultBadHostHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad Host", http.StatusInternalServerError)
}

// Options is a struct for specifying configuration options for the secure.Secure middleware.
type Options struct {
	// If BrowserXssFilter is true, adds the X-XSS-Protection header with the value `1; mode=block`. Default is false.
	BrowserXssFilter bool // nolint: golint
	// If ContentTypeNosniff is true, adds the X-Content-Type-Options header with the value `nosniff`. Default is false.
	ContentTypeNosniff bool
	// If ForceSTSHeader is set to true, the STS header will be added even when the connection is HTTP. Default is false.
	ForceSTSHeader bool
	// If FrameDeny is set to true, adds the X-Frame-Options header with the value of `DENY`. Default is false.
	FrameDeny bool
	// When developing, the AllowedHosts, SSL, and STS options can cause some unwanted effects. Usually testing happens on http, not https, and on localhost, not your production domain... so set this to true for dev environment.
	// If you would like your development environment to mimic production with complete Host blocking, SSL redirects, and STS headers, leave this as false. Default if false.
	IsDevelopment bool
	// nonceEnabled is used internally for dynamic nouces.
	nonceEnabled bool
	// If SSLRedirect is set to true, then only allow https requests. Default is false.
	SSLRedirect bool
	// If SSLForceHost is true and SSLHost is set, requests will be forced to use SSLHost even the ones that are already using SSL. Default is false.
	SSLForceHost bool
	// If SSLTemporaryRedirect is true, the a 302 will be used while redirecting. Default is false (301).
	SSLTemporaryRedirect bool
	// If STSIncludeSubdomains is set to true, the `includeSubdomains` will be appended to the Strict-Transport-Security header. Default is false.
	STSIncludeSubdomains bool
	// If STSPreload is set to true, the `preload` flag will be appended to the Strict-Transport-Security header. Default is false.
	STSPreload bool
	// ContentSecurityPolicy allows the Content-Security-Policy header value to be set with a custom value. Default is "".
	ContentSecurityPolicy string
	// ContentSecurityPolicyReportOnly allows the Content-Security-Policy-Report-Only header value to be set with a custom value. Default is "".
	ContentSecurityPolicyReportOnly string
	// CustomBrowserXssValue allows the X-XSS-Protection header value to be set with a custom value. This overrides the BrowserXssFilter option. Default is "".
	CustomBrowserXssValue string // nolint: golint
	// Passing a template string will replace `$NONCE` with a dynamic nonce value of 16 bytes for each request which can be later retrieved using the Nonce function.
	// Eg: script-src $NONCE -> script-src 'nonce-a2ZobGFoZg=='
	// CustomFrameOptionsValue allows the X-Frame-Options header value to be set with a custom value. This overrides the FrameDeny option. Default is "".
	CustomFrameOptionsValue string
	// PublicKey implements HPKP to prevent MITM attacks with forged certificates. Default is "".
	PublicKey string
	// ReferrerPolicy allows sites to control when browsers will pass the Referer header to other sites. Default is "".
	ReferrerPolicy string
	// FeaturePolicy allows to selectively enable and disable use of various browser features and APIs. Default is "".
	FeaturePolicy string
	// SSLHost is the host name that is used to redirect http requests to https. Default is "", which indicates to use the same host.
	SSLHost string
	// AllowedHosts is a list of fully qualified domain names that are allowed. Default is empty list, which allows any and all host names.
	AllowedHosts []string
	// AllowedHostsAreRegex determines, if the provided slice contains valid regular expressions. If this flag is set to true, every request's
	// host will be checked against these expressions. Default is false for backwards compatibility.
	AllowedHostsAreRegex bool
	// HostsProxyHeaders is a set of header keys that may hold a proxied hostname value for the request.
	HostsProxyHeaders []string
	// SSLHostFunc is a function pointer, the return value of the function is the host name that has same functionality as `SSHost`. Default is nil.
	// If SSLHostFunc is nil, the `SSLHost` option will be used.
	SSLHostFunc *SSLHostFunc
	// SSLProxyHeaders is set of header keys with associated values that would indicate a valid https request. Useful when using Nginx: `map[string]string{"X-Forwarded-Proto": "https"}`. Default is blank map.
	SSLProxyHeaders map[string]string
	// STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
	STSSeconds int64
	// ExpectCTHeader allows the Expect-CT header value to be set with a custom value. Default is "".
	ExpectCTHeader string
}

// Secure is a middleware that helps setup a few basic security features. A single secure.Options struct can be
// provided to configure which features should be enabled, and the ability to override a few of the default values.
type Secure struct {
	// Customize Secure with an Options struct.
	opt Options

	// badHostHandler is the handler used when an incorrect host is passed in.
	badHostHandler http.Handler

	// cRegexAllowedHosts saves the compiled regular expressions of the AllowedHosts
	// option for subsequent use in processRequest
	cRegexAllowedHosts []*regexp.Regexp
}

// New constructs a new Secure instance with the supplied options.
func New(options ...Options) *Secure {
	var o Options
	if len(options) == 0 {
		o = Options{}
	} else {
		o = options[0]
	}

	o.ContentSecurityPolicy = strings.Replace(o.ContentSecurityPolicy, "$NONCE", "'nonce-%[1]s'", -1)
	o.ContentSecurityPolicyReportOnly = strings.Replace(o.ContentSecurityPolicyReportOnly, "$NONCE", "'nonce-%[1]s'", -1)

	o.nonceEnabled = strings.Contains(o.ContentSecurityPolicy, "%[1]s") || strings.Contains(o.ContentSecurityPolicyReportOnly, "%[1]s")

	s := &Secure{
		opt:            o,
		badHostHandler: http.HandlerFunc(defaultBadHostHandler),
	}

	if s.opt.AllowedHostsAreRegex {
		// Test for invalid regular expressions in AllowedHosts
		for _, allowedHost := range o.AllowedHosts {
			regex, err := regexp.Compile(fmt.Sprintf("^%s$", allowedHost))
			if err != nil {
				panic(fmt.Sprintf("Error parsing AllowedHost: %s", err))
			}
			s.cRegexAllowedHosts = append(s.cRegexAllowedHosts, regex)
		}
	}

	return s
}

// SetBadHostHandler sets the handler to call when secure rejects the host name.
func (s *Secure) SetBadHostHandler(handler http.Handler) {
	s.badHostHandler = handler
}

// Handler implements the http.HandlerFunc for integration with the standard net/http lib.
func (s *Secure) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let secure process the request. If it returns an error,
		// that indicates the request should not continue.
		responseHeader, r, err := s.processRequest(w, r)
		addResponseHeaders(responseHeader, w)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// HandlerForRequestOnly implements the http.HandlerFunc for integration with the standard net/http lib.
// Note that this is for requests only and will not write any headers.
func (s *Secure) HandlerForRequestOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let secure process the request. If it returns an error,
		// that indicates the request should not continue.
		responseHeader, r, err := s.processRequest(w, r)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		// Save response headers in the request context.
		ctx := context.WithValue(r.Context(), ctxSecureHeaderKey, responseHeader)

		// No headers will be written to the ResponseWriter.
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandlerFuncWithNext is a special implementation for Negroni, but could be used elsewhere.
func (s *Secure) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Let secure process the request. If it returns an error,
	// that indicates the request should not continue.
	responseHeader, r, err := s.processRequest(w, r)
	addResponseHeaders(responseHeader, w)

	// If there was an error, do not call next.
	if err == nil && next != nil {
		next(w, r)
	}
}

// HandlerFuncWithNextForRequestOnly is a special implementation for Negroni, but could be used elsewhere.
// Note that this is for requests only and will not write any headers.
func (s *Secure) HandlerFuncWithNextForRequestOnly(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Let secure process the request. If it returns an error,
	// that indicates the request should not continue.
	responseHeader, r, err := s.processRequest(w, r)

	// If there was an error, do not call next.
	if err == nil && next != nil {
		// Save response headers in the request context
		ctx := context.WithValue(r.Context(), ctxSecureHeaderKey, responseHeader)

		// No headers will be written to the ResponseWriter.
		next(w, r.WithContext(ctx))
	}
}

// addResponseHeaders Adds the headers from 'responseHeader' to the response.
func addResponseHeaders(responseHeader http.Header, w http.ResponseWriter) {
	for key, values := range responseHeader {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}
}

// Process runs the actual checks and writes the headers in the ResponseWriter.
func (s *Secure) Process(w http.ResponseWriter, r *http.Request) error {
	responseHeader, _, err := s.processRequest(w, r)
	addResponseHeaders(responseHeader, w)

	return err
}

// ProcessAndReturnNonce runs the actual checks and writes the headers in the ResponseWriter.
// In addition, the generated nonce for the request is returned as well as the error value.
func (s *Secure) ProcessAndReturnNonce(w http.ResponseWriter, r *http.Request) (string, error) {
	responseHeader, newR, err := s.processRequest(w, r)
	addResponseHeaders(responseHeader, w)

	return CSPNonce(newR.Context()), err
}

// ProcessNoModifyRequest runs the actual checks but does not write the headers in the ResponseWriter.
func (s *Secure) ProcessNoModifyRequest(w http.ResponseWriter, r *http.Request) (http.Header, *http.Request, error) {
	return s.processRequest(w, r)
}

// processRequest runs the actual checks on the request and returns an error if the middleware chain should stop.
func (s *Secure) processRequest(w http.ResponseWriter, r *http.Request) (http.Header, *http.Request, error) {
	// Setup nonce if required.
	if s.opt.nonceEnabled {
		r = withCSPNonce(r, cspRandNonce())
	}

	// Resolve the host for the request, using proxy headers if present.
	host := r.Host
	for _, header := range s.opt.HostsProxyHeaders {
		if h := r.Header.Get(header); h != "" {
			host = h
			break
		}
	}

	// Allowed hosts check.
	if len(s.opt.AllowedHosts) > 0 && !s.opt.IsDevelopment {
		isGoodHost := false
		if s.opt.AllowedHostsAreRegex {
			for _, allowedHost := range s.cRegexAllowedHosts {
				if match := allowedHost.MatchString(host); match {
					isGoodHost = true
					break
				}
			}
		} else {
			for _, allowedHost := range s.opt.AllowedHosts {
				if strings.EqualFold(allowedHost, host) {
					isGoodHost = true
					break
				}
			}
		}

		if !isGoodHost {
			s.badHostHandler.ServeHTTP(w, r)
			return nil, nil, fmt.Errorf("bad host name: %s", host)
		}
	}

	// Determine if we are on HTTPS.
	ssl := s.isSSL(r)

	// SSL check.
	if s.opt.SSLRedirect && !ssl && !s.opt.IsDevelopment {
		url := r.URL
		url.Scheme = "https"
		url.Host = host

		if s.opt.SSLHostFunc != nil {
			if h := (*s.opt.SSLHostFunc)(host); len(h) > 0 {
				url.Host = h
			}
		} else if len(s.opt.SSLHost) > 0 {
			url.Host = s.opt.SSLHost
		}

		status := http.StatusMovedPermanently
		if s.opt.SSLTemporaryRedirect {
			status = http.StatusTemporaryRedirect
		}

		http.Redirect(w, r, url.String(), status)
		return nil, nil, fmt.Errorf("redirecting to HTTPS")
	}

	if s.opt.SSLForceHost {
		var SSLHost = host
		if s.opt.SSLHostFunc != nil {
			if h := (*s.opt.SSLHostFunc)(host); len(h) > 0 {
				SSLHost = h
			}
		} else if len(s.opt.SSLHost) > 0 {
			SSLHost = s.opt.SSLHost
		}
		if SSLHost != host {
			url := r.URL
			url.Scheme = "https"
			url.Host = SSLHost

			status := http.StatusMovedPermanently
			if s.opt.SSLTemporaryRedirect {
				status = http.StatusTemporaryRedirect
			}

			http.Redirect(w, r, url.String(), status)
			return nil, nil, fmt.Errorf("redirecting to HTTPS")
		}
	}

	// Create our header container.
	responseHeader := make(http.Header)

	// Strict Transport Security header. Only add header when we know it's an SSL connection.
	// See https://tools.ietf.org/html/rfc6797#section-7.2 for details.
	if s.opt.STSSeconds != 0 && (ssl || s.opt.ForceSTSHeader) && !s.opt.IsDevelopment {
		stsSub := ""
		if s.opt.STSIncludeSubdomains {
			stsSub = stsSubdomainString
		}

		if s.opt.STSPreload {
			stsSub += stsPreloadString
		}

		responseHeader.Set(stsHeader, fmt.Sprintf("max-age=%d%s", s.opt.STSSeconds, stsSub))
	}

	// Frame Options header.
	if len(s.opt.CustomFrameOptionsValue) > 0 {
		responseHeader.Set(frameOptionsHeader, s.opt.CustomFrameOptionsValue)
	} else if s.opt.FrameDeny {
		responseHeader.Set(frameOptionsHeader, frameOptionsValue)
	}

	// Content Type Options header.
	if s.opt.ContentTypeNosniff {
		responseHeader.Set(contentTypeHeader, contentTypeValue)
	}

	// XSS Protection header.
	if len(s.opt.CustomBrowserXssValue) > 0 {
		responseHeader.Set(xssProtectionHeader, s.opt.CustomBrowserXssValue)
	} else if s.opt.BrowserXssFilter {
		responseHeader.Set(xssProtectionHeader, xssProtectionValue)
	}

	// HPKP header.
	if len(s.opt.PublicKey) > 0 && ssl && !s.opt.IsDevelopment {
		responseHeader.Set(hpkpHeader, s.opt.PublicKey)
	}

	// Content Security Policy header.
	if len(s.opt.ContentSecurityPolicy) > 0 {
		if s.opt.nonceEnabled {
			responseHeader.Set(cspHeader, fmt.Sprintf(s.opt.ContentSecurityPolicy, CSPNonce(r.Context())))
		} else {
			responseHeader.Set(cspHeader, s.opt.ContentSecurityPolicy)
		}
	}

	// Content Security Policy Report Only header.
	if len(s.opt.ContentSecurityPolicyReportOnly) > 0 {
		if s.opt.nonceEnabled {
			responseHeader.Set(cspReportOnlyHeader, fmt.Sprintf(s.opt.ContentSecurityPolicyReportOnly, CSPNonce(r.Context())))
		} else {
			responseHeader.Set(cspReportOnlyHeader, s.opt.ContentSecurityPolicyReportOnly)
		}
	}

	// Referrer Policy header.
	if len(s.opt.ReferrerPolicy) > 0 {
		responseHeader.Set(referrerPolicyHeader, s.opt.ReferrerPolicy)
	}

	// Feature Policy header.
	if len(s.opt.FeaturePolicy) > 0 {
		responseHeader.Set(featurePolicyHeader, s.opt.FeaturePolicy)
	}

	// Expect-CT header.
	if len(s.opt.ExpectCTHeader) > 0 {
		responseHeader.Set(expectCTHeader, s.opt.ExpectCTHeader)
	}

	return responseHeader, r, nil
}

// isSSL determine if we are on HTTPS.
func (s *Secure) isSSL(r *http.Request) bool {
	ssl := strings.EqualFold(r.URL.Scheme, "https") || r.TLS != nil
	if !ssl {
		for k, v := range s.opt.SSLProxyHeaders {
			if r.Header.Get(k) == v {
				ssl = true
				break
			}
		}
	}
	return ssl
}

// ModifyResponseHeaders modifies the Response.
// Used by http.ReverseProxy.
func (s *Secure) ModifyResponseHeaders(res *http.Response) error {
	if res != nil && res.Request != nil {
		// Fix Location response header http to https:
		// When SSL is enabled,
		// And SSLHost is defined,
		// And the response location header includes the SSLHost as the domain with a trailing slash,
		// Or an exact match to the SSLHost.
		location := res.Header.Get("Location")
		if s.isSSL(res.Request) &&
			len(s.opt.SSLHost) > 0 &&
			(strings.HasPrefix(location, fmt.Sprintf("http://%s/", s.opt.SSLHost)) || location == fmt.Sprintf("http://%s", s.opt.SSLHost)) {
			location = strings.Replace(location, "http:", "https:", 1)
			res.Header.Set("Location", location)
		}

		responseHeader := res.Request.Context().Value(ctxSecureHeaderKey)
		if responseHeader != nil {
			for header, values := range responseHeader.(http.Header) {
				if len(values) > 0 {
					res.Header.Set(header, strings.Join(values, ","))
				}
			}
		}
	}
	return nil
}
