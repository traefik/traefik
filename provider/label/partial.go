package label

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// GetWhiteList Create white list from labels
func GetWhiteList(labels map[string]string) *types.WhiteList {
	if Has(labels, TraefikFrontendWhitelistSourceRange) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", TraefikFrontendWhitelistSourceRange, TraefikFrontendWhiteListSourceRange)
	}

	ranges := GetSliceStringValue(labels, TraefikFrontendWhiteListSourceRange)
	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: GetBoolValue(labels, TraefikFrontendWhiteListUseXForwardedFor, false),
		}
	}

	// TODO: Deprecated
	values := GetSliceStringValue(labels, TraefikFrontendWhitelistSourceRange)
	if len(values) > 0 {
		return &types.WhiteList{
			SourceRange:      values,
			UseXForwardedFor: false,
		}
	}

	return nil
}

// GetRedirect Create redirect from labels
func GetRedirect(labels map[string]string) *types.Redirect {
	permanent := GetBoolValue(labels, TraefikFrontendRedirectPermanent, false)

	if Has(labels, TraefikFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: GetStringValue(labels, TraefikFrontendRedirectEntryPoint, ""),
			Permanent:  permanent,
		}
	}

	if Has(labels, TraefikFrontendRedirectRegex) &&
		Has(labels, TraefikFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       GetStringValue(labels, TraefikFrontendRedirectRegex, ""),
			Replacement: GetStringValue(labels, TraefikFrontendRedirectReplacement, ""),
			Permanent:   permanent,
		}
	}

	return nil
}

// GetTLSClientCert create TLS client header configuration from labels
func GetTLSClientCert(labels map[string]string) *types.TLSClientHeaders {
	if !HasPrefix(labels, TraefikFrontendPassTLSClientCert) {
		return nil
	}

	tlsClientHeaders := &types.TLSClientHeaders{
		PEM: GetBoolValue(labels, TraefikFrontendPassTLSClientCertPem, false),
	}

	if HasPrefix(labels, TraefikFrontendPassTLSClientCertInfos) {
		infos := &types.TLSClientCertificateInfos{
			NotAfter:  GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosNotAfter, false),
			NotBefore: GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosNotBefore, false),
			Sans:      GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSans, false),
		}

		if HasPrefix(labels, TraefikFrontendPassTLSClientCertInfosSubject) {
			subject := &types.TLSCLientCertificateDNInfos{
				CommonName:      GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectCommonName, false),
				Country:         GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectCountry, false),
				DomainComponent: GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent, false),
				Locality:        GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectLocality, false),
				Organization:    GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectOrganization, false),
				Province:        GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectProvince, false),
				SerialNumber:    GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber, false),
			}
			infos.Subject = subject
		}

		if HasPrefix(labels, TraefikFrontendPassTLSClientCertInfosIssuer) {
			issuer := &types.TLSCLientCertificateDNInfos{
				CommonName:      GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerCommonName, false),
				Country:         GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerCountry, false),
				DomainComponent: GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent, false),
				Locality:        GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerLocality, false),
				Organization:    GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerOrganization, false),
				Province:        GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerProvince, false),
				SerialNumber:    GetBoolValue(labels, TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber, false),
			}
			infos.Issuer = issuer
		}
		tlsClientHeaders.Infos = infos
	}
	return tlsClientHeaders
}

// GetAuth Create auth from labels
func GetAuth(labels map[string]string) *types.Auth {
	if !HasPrefix(labels, TraefikFrontendAuth) {
		return nil
	}

	auth := &types.Auth{
		HeaderField: GetStringValue(labels, TraefikFrontendAuthHeaderField, ""),
	}

	if HasPrefix(labels, TraefikFrontendAuthBasic) {
		auth.Basic = getAuthBasic(labels)
	} else if HasPrefix(labels, TraefikFrontendAuthDigest) {
		auth.Digest = getAuthDigest(labels)
	} else if HasPrefix(labels, TraefikFrontendAuthForward) {
		auth.Forward = getAuthForward(labels)
	} else {
		return nil
	}

	return auth
}

// getAuthBasic Create Basic Auth from labels
func getAuthBasic(labels map[string]string) *types.Basic {
	basicAuth := &types.Basic{
		UsersFile:    GetStringValue(labels, TraefikFrontendAuthBasicUsersFile, ""),
		RemoveHeader: GetBoolValue(labels, TraefikFrontendAuthBasicRemoveHeader, false),
	}

	// backward compatibility
	if Has(labels, TraefikFrontendAuthBasic) {
		basicAuth.Users = GetSliceStringValue(labels, TraefikFrontendAuthBasic)
		log.Warnf("Deprecated configuration found: %s. Please use %s.", TraefikFrontendAuthBasic, TraefikFrontendAuthBasicUsers)
	} else {
		basicAuth.Users = GetSliceStringValue(labels, TraefikFrontendAuthBasicUsers)
	}

	return basicAuth
}

// getAuthDigest Create Digest Auth from labels
func getAuthDigest(labels map[string]string) *types.Digest {
	return &types.Digest{
		Users:        GetSliceStringValue(labels, TraefikFrontendAuthDigestUsers),
		UsersFile:    GetStringValue(labels, TraefikFrontendAuthDigestUsersFile, ""),
		RemoveHeader: GetBoolValue(labels, TraefikFrontendAuthDigestRemoveHeader, false),
	}
}

// getAuthForward Create Forward Auth from labels
func getAuthForward(labels map[string]string) *types.Forward {
	forwardAuth := &types.Forward{
		Address:             GetStringValue(labels, TraefikFrontendAuthForwardAddress, ""),
		AuthResponseHeaders: GetSliceStringValue(labels, TraefikFrontendAuthForwardAuthResponseHeaders),
		TrustForwardHeader:  GetBoolValue(labels, TraefikFrontendAuthForwardTrustForwardHeader, false),
	}

	// TLS configuration
	if HasPrefix(labels, TraefikFrontendAuthForwardTLS) {
		forwardAuth.TLS = &types.ClientTLS{
			CA:                 GetStringValue(labels, TraefikFrontendAuthForwardTLSCa, ""),
			CAOptional:         GetBoolValue(labels, TraefikFrontendAuthForwardTLSCaOptional, false),
			Cert:               GetStringValue(labels, TraefikFrontendAuthForwardTLSCert, ""),
			InsecureSkipVerify: GetBoolValue(labels, TraefikFrontendAuthForwardTLSInsecureSkipVerify, false),
			Key:                GetStringValue(labels, TraefikFrontendAuthForwardTLSKey, ""),
		}
	}

	return forwardAuth
}

// GetErrorPages Create error pages from labels
func GetErrorPages(labels map[string]string) map[string]*types.ErrorPage {
	prefix := Prefix + BaseFrontendErrorPage
	return ParseErrorPages(labels, prefix, RegexpFrontendErrorPage)
}

// ParseErrorPages parse error pages to create ErrorPage struct
func ParseErrorPages(labels map[string]string, labelPrefix string, labelRegex *regexp.Regexp) map[string]*types.ErrorPage {
	var errorPages map[string]*types.ErrorPage

	for lblName, value := range labels {
		if strings.HasPrefix(lblName, labelPrefix) {
			submatch := labelRegex.FindStringSubmatch(lblName)
			if len(submatch) != 3 {
				log.Errorf("Invalid page error label: %s, sub-match: %v", lblName, submatch)
				continue
			}

			if errorPages == nil {
				errorPages = make(map[string]*types.ErrorPage)
			}

			pageName := submatch[1]

			ep, ok := errorPages[pageName]
			if !ok {
				ep = &types.ErrorPage{}
				errorPages[pageName] = ep
			}

			switch submatch[2] {
			case SuffixErrorPageStatus:
				ep.Status = SplitAndTrimString(value, ",")
			case SuffixErrorPageQuery:
				ep.Query = value
			case SuffixErrorPageBackend:
				ep.Backend = value
			default:
				log.Errorf("Invalid page error label: %s", lblName)
				continue
			}
		}
	}

	return errorPages
}

// GetRateLimit Create rate limits from labels
func GetRateLimit(labels map[string]string) *types.RateLimit {
	extractorFunc := GetStringValue(labels, TraefikFrontendRateLimitExtractorFunc, "")
	if len(extractorFunc) == 0 {
		return nil
	}

	prefix := Prefix + BaseFrontendRateLimit
	limits := ParseRateSets(labels, prefix, RegexpFrontendRateLimit)

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

// ParseRateSets parse rate limits to create Rate struct
func ParseRateSets(labels map[string]string, labelPrefix string, labelRegex *regexp.Regexp) map[string]*types.Rate {
	var rateSets map[string]*types.Rate

	for lblName, rawValue := range labels {
		if strings.HasPrefix(lblName, labelPrefix) && len(rawValue) > 0 {
			submatch := labelRegex.FindStringSubmatch(lblName)
			if len(submatch) != 3 {
				log.Errorf("Invalid rate limit label: %s, sub-match: %v", lblName, submatch)
				continue
			}

			if rateSets == nil {
				rateSets = make(map[string]*types.Rate)
			}

			limitName := submatch[1]

			ep, ok := rateSets[limitName]
			if !ok {
				ep = &types.Rate{}
				rateSets[limitName] = ep
			}

			switch submatch[2] {
			case "period":
				var d flaeg.Duration
				err := d.Set(rawValue)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Period = d
			case "average":
				value, err := strconv.ParseInt(rawValue, 10, 64)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Average = value
			case "burst":
				value, err := strconv.ParseInt(rawValue, 10, 64)
				if err != nil {
					log.Errorf("Unable to parse %q: %q. %v", lblName, rawValue, err)
					continue
				}
				ep.Burst = value
			default:
				log.Errorf("Invalid rate limit label: %s", lblName)
				continue
			}
		}
	}
	return rateSets
}

// GetHeaders Create headers from labels
func GetHeaders(labels map[string]string) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    GetMapValue(labels, TraefikFrontendRequestHeaders),
		CustomResponseHeaders:   GetMapValue(labels, TraefikFrontendResponseHeaders),
		SSLProxyHeaders:         GetMapValue(labels, TraefikFrontendSSLProxyHeaders),
		AllowedHosts:            GetSliceStringValue(labels, TraefikFrontendAllowedHosts),
		HostsProxyHeaders:       GetSliceStringValue(labels, TraefikFrontendHostsProxyHeaders),
		STSSeconds:              GetInt64Value(labels, TraefikFrontendSTSSeconds, 0),
		SSLRedirect:             GetBoolValue(labels, TraefikFrontendSSLRedirect, false),
		SSLTemporaryRedirect:    GetBoolValue(labels, TraefikFrontendSSLTemporaryRedirect, false),
		SSLForceHost:            GetBoolValue(labels, TraefikFrontendSSLForceHost, false),
		STSIncludeSubdomains:    GetBoolValue(labels, TraefikFrontendSTSIncludeSubdomains, false),
		STSPreload:              GetBoolValue(labels, TraefikFrontendSTSPreload, false),
		ForceSTSHeader:          GetBoolValue(labels, TraefikFrontendForceSTSHeader, false),
		FrameDeny:               GetBoolValue(labels, TraefikFrontendFrameDeny, false),
		ContentTypeNosniff:      GetBoolValue(labels, TraefikFrontendContentTypeNosniff, false),
		BrowserXSSFilter:        GetBoolValue(labels, TraefikFrontendBrowserXSSFilter, false),
		IsDevelopment:           GetBoolValue(labels, TraefikFrontendIsDevelopment, false),
		SSLHost:                 GetStringValue(labels, TraefikFrontendSSLHost, ""),
		CustomFrameOptionsValue: GetStringValue(labels, TraefikFrontendCustomFrameOptionsValue, ""),
		ContentSecurityPolicy:   GetStringValue(labels, TraefikFrontendContentSecurityPolicy, ""),
		PublicKey:               GetStringValue(labels, TraefikFrontendPublicKey, ""),
		ReferrerPolicy:          GetStringValue(labels, TraefikFrontendReferrerPolicy, ""),
		CustomBrowserXSSValue:   GetStringValue(labels, TraefikFrontendCustomBrowserXSSValue, ""),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

// GetMaxConn Create max connection from labels
func GetMaxConn(labels map[string]string) *types.MaxConn {
	amount := GetInt64Value(labels, TraefikBackendMaxConnAmount, math.MinInt64)
	extractorFunc := GetStringValue(labels, TraefikBackendMaxConnExtractorFunc, DefaultBackendMaxconnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

// GetHealthCheck Create health check from labels
func GetHealthCheck(labels map[string]string) *types.HealthCheck {
	path := GetStringValue(labels, TraefikBackendHealthCheckPath, "")
	if len(path) == 0 {
		return nil
	}

	scheme := GetStringValue(labels, TraefikBackendHealthCheckScheme, "")
	port := GetIntValue(labels, TraefikBackendHealthCheckPort, DefaultBackendHealthCheckPort)
	interval := GetStringValue(labels, TraefikBackendHealthCheckInterval, "")
	hostname := GetStringValue(labels, TraefikBackendHealthCheckHostname, "")
	headers := GetMapValue(labels, TraefikBackendHealthCheckHeaders)

	return &types.HealthCheck{
		Scheme:   scheme,
		Path:     path,
		Port:     port,
		Interval: interval,
		Hostname: hostname,
		Headers:  headers,
	}
}

// GetResponseForwarding Create ResponseForwarding from labels
func GetResponseForwarding(labels map[string]string) *types.ResponseForwarding {
	if !HasPrefix(labels, TraefikBackendResponseForwardingFlushInterval) {
		return nil
	}

	value := GetStringValue(labels, TraefikBackendResponseForwardingFlushInterval, "0")

	return &types.ResponseForwarding{
		FlushInterval: value,
	}
}

// GetBuffering Create buffering from labels
func GetBuffering(labels map[string]string) *types.Buffering {
	if !HasPrefix(labels, TraefikBackendBuffering) {
		return nil
	}

	return &types.Buffering{
		MaxRequestBodyBytes:  GetInt64Value(labels, TraefikBackendBufferingMaxRequestBodyBytes, 0),
		MaxResponseBodyBytes: GetInt64Value(labels, TraefikBackendBufferingMaxResponseBodyBytes, 0),
		MemRequestBodyBytes:  GetInt64Value(labels, TraefikBackendBufferingMemRequestBodyBytes, 0),
		MemResponseBodyBytes: GetInt64Value(labels, TraefikBackendBufferingMemResponseBodyBytes, 0),
		RetryExpression:      GetStringValue(labels, TraefikBackendBufferingRetryExpression, ""),
	}
}

// GetCircuitBreaker Create circuit breaker from labels
func GetCircuitBreaker(labels map[string]string) *types.CircuitBreaker {
	circuitBreaker := GetStringValue(labels, TraefikBackendCircuitBreakerExpression, "")
	if len(circuitBreaker) == 0 {
		return nil
	}
	return &types.CircuitBreaker{Expression: circuitBreaker}
}

// GetLoadBalancer Create load balancer from labels
func GetLoadBalancer(labels map[string]string) *types.LoadBalancer {
	if !HasPrefix(labels, TraefikBackendLoadBalancer) {
		return nil
	}

	method := GetStringValue(labels, TraefikBackendLoadBalancerMethod, DefaultBackendLoadBalancerMethod)

	lb := &types.LoadBalancer{
		Method: method,
		Sticky: getSticky(labels),
	}

	if GetBoolValue(labels, TraefikBackendLoadBalancerStickiness, false) {
		cookieName := GetStringValue(labels, TraefikBackendLoadBalancerStickinessCookieName, DefaultBackendLoadbalancerStickinessCookieName)
		lb.Stickiness = &types.Stickiness{CookieName: cookieName}
	}

	return lb
}

// TODO: Deprecated
// replaced by Stickiness
// Deprecated
func getSticky(labels map[string]string) bool {
	if Has(labels, TraefikBackendLoadBalancerSticky) {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", TraefikBackendLoadBalancerSticky, TraefikBackendLoadBalancerStickiness)
	}

	return GetBoolValue(labels, TraefikBackendLoadBalancerSticky, false)
}
