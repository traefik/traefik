package kv

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/abronan/valkeyrie/store"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration() (*types.Configuration, error) {
	templateObjects := struct {
		Prefix string
	}{
		// Allow `/traefik/alias` to supersede `p.Prefix`
		Prefix: strings.TrimSuffix(p.get(p.Prefix, p.Prefix+pathAlias), pathSeparator),
	}

	kvFuncMap := template.FuncMap{
		"List":        p.list,
		"ListServers": p.listServers,
		"Get":         p.get,
		"GetBool":     p.getBool,
		"GetInt":      p.getInt,
		"GetInt64":    p.getInt64,
		"GetList":     p.getList,
		"SplitGet":    p.splitGet,
		"Last":        p.last,
		"Has":         p.has,

		"getTLSSection": p.getTLSSection,

		// Frontend functions
		"getBackendName":       p.getFuncString(pathFrontendBackend, ""),
		"getPriority":          p.getFuncInt(pathFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":    p.getPassHostHeader(),
		"getPassTLSCert":       p.getFuncBool(pathFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPassTLSClientCert": p.getTLSClientCert,
		"getEntryPoints":       p.getFuncList(pathFrontendEntryPoints),
		"getBasicAuth":         p.getFuncList(pathFrontendBasicAuth), // Deprecated
		"getAuth":              p.getAuth,
		"getRoutes":            p.getRoutes,
		"getRedirect":          p.getRedirect,
		"getErrorPages":        p.getErrorPages,
		"getRateLimit":         p.getRateLimit,
		"getHeaders":           p.getHeaders,
		"getWhiteList":         p.getWhiteList,

		// Backend functions
		"getServers":              p.getServers,
		"getCircuitBreaker":       p.getCircuitBreaker,
		"getResponseForwarding":   p.getResponseForwarding,
		"getLoadBalancer":         p.getLoadBalancer,
		"getMaxConn":              p.getMaxConn,
		"getHealthCheck":          p.getHealthCheck,
		"getBuffering":            p.getBuffering,
		"getSticky":               p.getSticky,               // Deprecated [breaking]
		"hasStickinessLabel":      p.hasStickinessLabel,      // Deprecated [breaking]
		"getStickinessCookieName": p.getStickinessCookieName, // Deprecated [breaking]
	}

	configuration, err := p.safeGetConfiguration("templates/kv.tmpl", kvFuncMap, templateObjects)
	if err != nil {
		return nil, err
	}

	for key, frontend := range configuration.Frontends {
		if _, ok := configuration.Backends[frontend.Backend]; !ok {
			delete(configuration.Frontends, key)
		}
	}

	return configuration, nil
}

func (p *Provider) safeGetConfiguration(defaultTemplate string, funcMap template.FuncMap, templateObjects interface{}) (configuration *types.Configuration, err error) {
	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("error while getting the configuration: %v", e)
		}
	}()

	configuration, err = p.GetConfiguration("templates/kv.tmpl", funcMap, templateObjects)
	return
}

// Deprecated
func (p *Provider) getPassHostHeader() func(rootPath string) bool {
	return func(rootPath string) bool {
		rawValue := p.get("", rootPath, pathFrontendPassHostHeader)

		if len(rawValue) > 0 {
			value, err := strconv.ParseBool(rawValue)
			if err != nil {
				log.Errorf("Invalid value for %s %s: %s", rootPath, pathFrontendPassHostHeader, rawValue)
				return label.DefaultPassHostHeader
			}
			return value
		}

		return p.getBool(label.DefaultPassHostHeader, rootPath, pathFrontendPassHostHeaderDeprecated)
	}
}

// Deprecated
func (p *Provider) getSticky(rootPath string) bool {
	stickyValue := p.get("", rootPath, pathBackendLoadBalancerSticky)
	if len(stickyValue) > 0 {
		log.Warnf("Deprecated configuration found: %s. Please use %s.", pathBackendLoadBalancerSticky, pathBackendLoadBalancerStickiness)
	} else {
		return false
	}

	sticky, err := strconv.ParseBool(stickyValue)
	if err != nil {
		log.Warnf("Invalid %s value: %s.", pathBackendLoadBalancerSticky, stickyValue)
	}

	return sticky
}

// Deprecated
func (p *Provider) hasStickinessLabel(rootPath string) bool {
	return p.getBool(false, rootPath, pathBackendLoadBalancerStickiness)
}

// Deprecated
func (p *Provider) getStickinessCookieName(rootPath string) string {
	return p.get("", rootPath, pathBackendLoadBalancerStickinessCookieName)
}

func (p *Provider) getWhiteList(rootPath string) *types.WhiteList {
	ranges := p.getList(rootPath, pathFrontendWhiteListSourceRange)

	if len(ranges) > 0 {
		return &types.WhiteList{
			SourceRange:      ranges,
			UseXForwardedFor: p.getBool(false, rootPath, pathFrontendWhiteListUseXForwardedFor),
		}
	}

	return nil
}

func (p *Provider) getRedirect(rootPath string) *types.Redirect {
	permanent := p.getBool(false, rootPath, pathFrontendRedirectPermanent)

	if p.has(rootPath, pathFrontendRedirectEntryPoint) {
		return &types.Redirect{
			EntryPoint: p.get("", rootPath, pathFrontendRedirectEntryPoint),
			Permanent:  permanent,
		}
	}

	if p.has(rootPath, pathFrontendRedirectRegex) && p.has(rootPath, pathFrontendRedirectReplacement) {
		return &types.Redirect{
			Regex:       p.get("", rootPath, pathFrontendRedirectRegex),
			Replacement: p.get("", rootPath, pathFrontendRedirectReplacement),
			Permanent:   permanent,
		}
	}

	return nil
}

func (p *Provider) getErrorPages(rootPath string) map[string]*types.ErrorPage {
	var errorPages map[string]*types.ErrorPage

	pathErrors := p.list(rootPath, pathFrontendErrorPages)

	for _, pathPage := range pathErrors {
		if errorPages == nil {
			errorPages = make(map[string]*types.ErrorPage)
		}

		pageName := p.last(pathPage)

		errorPages[pageName] = &types.ErrorPage{
			Backend: p.get("", pathPage, pathFrontendErrorPagesBackend),
			Query:   p.get("", pathPage, pathFrontendErrorPagesQuery),
			Status:  p.getList(pathPage, pathFrontendErrorPagesStatus),
		}
	}

	return errorPages
}

func (p *Provider) getRateLimit(rootPath string) *types.RateLimit {
	extractorFunc := p.get("", rootPath, pathFrontendRateLimitExtractorFunc)
	if len(extractorFunc) == 0 {
		return nil
	}

	var limits map[string]*types.Rate

	pathRateSet := p.list(rootPath, pathFrontendRateLimitRateSet)
	for _, pathLimits := range pathRateSet {
		if limits == nil {
			limits = make(map[string]*types.Rate)
		}

		rawPeriod := p.get("", pathLimits+pathFrontendRateLimitPeriod)

		var period flaeg.Duration
		err := period.Set(rawPeriod)
		if err != nil {
			log.Errorf("Invalid %q value: %q", pathLimits+pathFrontendRateLimitPeriod, rawPeriod)
			continue
		}

		limitName := p.last(pathLimits)

		limits[limitName] = &types.Rate{
			Average: p.getInt64(0, pathLimits+pathFrontendRateLimitAverage),
			Burst:   p.getInt64(0, pathLimits+pathFrontendRateLimitBurst),
			Period:  period,
		}
	}

	return &types.RateLimit{
		ExtractorFunc: extractorFunc,
		RateSet:       limits,
	}
}

func (p *Provider) getHeaders(rootPath string) *types.Headers {
	headers := &types.Headers{
		CustomRequestHeaders:    p.getMap(rootPath, pathFrontendCustomRequestHeaders),
		CustomResponseHeaders:   p.getMap(rootPath, pathFrontendCustomResponseHeaders),
		SSLProxyHeaders:         p.getMap(rootPath, pathFrontendSSLProxyHeaders),
		AllowedHosts:            p.getList("", rootPath, pathFrontendAllowedHosts),
		HostsProxyHeaders:       p.getList(rootPath, pathFrontendHostsProxyHeaders),
		SSLForceHost:            p.getBool(false, rootPath, pathFrontendSSLForceHost),
		SSLRedirect:             p.getBool(false, rootPath, pathFrontendSSLRedirect),
		SSLTemporaryRedirect:    p.getBool(false, rootPath, pathFrontendSSLTemporaryRedirect),
		SSLHost:                 p.get("", rootPath, pathFrontendSSLHost),
		STSSeconds:              p.getInt64(0, rootPath, pathFrontendSTSSeconds),
		STSIncludeSubdomains:    p.getBool(false, rootPath, pathFrontendSTSIncludeSubdomains),
		STSPreload:              p.getBool(false, rootPath, pathFrontendSTSPreload),
		ForceSTSHeader:          p.getBool(false, rootPath, pathFrontendForceSTSHeader),
		FrameDeny:               p.getBool(false, rootPath, pathFrontendFrameDeny),
		CustomFrameOptionsValue: p.get("", rootPath, pathFrontendCustomFrameOptionsValue),
		ContentTypeNosniff:      p.getBool(false, rootPath, pathFrontendContentTypeNosniff),
		BrowserXSSFilter:        p.getBool(false, rootPath, pathFrontendBrowserXSSFilter),
		CustomBrowserXSSValue:   p.get("", rootPath, pathFrontendCustomBrowserXSSValue),
		ContentSecurityPolicy:   p.get("", rootPath, pathFrontendContentSecurityPolicy),
		PublicKey:               p.get("", rootPath, pathFrontendPublicKey),
		ReferrerPolicy:          p.get("", rootPath, pathFrontendReferrerPolicy),
		IsDevelopment:           p.getBool(false, rootPath, pathFrontendIsDevelopment),
	}

	if !headers.HasSecureHeadersDefined() && !headers.HasCustomHeadersDefined() {
		return nil
	}

	return headers
}

func (p *Provider) getLoadBalancer(rootPath string) *types.LoadBalancer {
	lb := &types.LoadBalancer{
		Method: p.get(label.DefaultBackendLoadBalancerMethod, rootPath, pathBackendLoadBalancerMethod),
		Sticky: p.getSticky(rootPath),
	}

	if p.getBool(false, rootPath, pathBackendLoadBalancerStickiness) {
		lb.Stickiness = &types.Stickiness{
			CookieName: p.get("", rootPath, pathBackendLoadBalancerStickinessCookieName),
		}
	}

	return lb
}

func (p *Provider) getResponseForwarding(rootPath string) *types.ResponseForwarding {
	if !p.has(rootPath, pathBackendResponseForwardingFlushInterval) {
		return nil
	}
	value := p.get("", rootPath, pathBackendResponseForwardingFlushInterval)
	if len(value) == 0 {
		return nil
	}

	return &types.ResponseForwarding{
		FlushInterval: value,
	}
}

func (p *Provider) getCircuitBreaker(rootPath string) *types.CircuitBreaker {
	if !p.has(rootPath, pathBackendCircuitBreakerExpression) {
		return nil
	}

	circuitBreaker := p.get(label.DefaultCircuitBreakerExpression, rootPath, pathBackendCircuitBreakerExpression)
	if len(circuitBreaker) == 0 {
		return nil
	}

	return &types.CircuitBreaker{Expression: circuitBreaker}
}

func (p *Provider) getMaxConn(rootPath string) *types.MaxConn {
	amount := p.getInt64(math.MinInt64, rootPath, pathBackendMaxConnAmount)
	extractorFunc := p.get(label.DefaultBackendMaxconnExtractorFunc, rootPath, pathBackendMaxConnExtractorFunc)

	if amount == math.MinInt64 || len(extractorFunc) == 0 {
		return nil
	}

	return &types.MaxConn{
		Amount:        amount,
		ExtractorFunc: extractorFunc,
	}
}

func (p *Provider) getHealthCheck(rootPath string) *types.HealthCheck {
	path := p.get("", rootPath, pathBackendHealthCheckPath)

	if len(path) == 0 {
		return nil
	}

	scheme := p.get("", rootPath, pathBackendHealthCheckScheme)
	port := p.getInt(label.DefaultBackendHealthCheckPort, rootPath, pathBackendHealthCheckPort)
	interval := p.get("30s", rootPath, pathBackendHealthCheckInterval)
	hostname := p.get("", rootPath, pathBackendHealthCheckHostname)
	headers := p.getMap(rootPath, pathBackendHealthCheckHeaders)

	return &types.HealthCheck{
		Scheme:   scheme,
		Path:     path,
		Port:     port,
		Interval: interval,
		Hostname: hostname,
		Headers:  headers,
	}
}

func (p *Provider) getBuffering(rootPath string) *types.Buffering {
	pathsBuffering := p.list(rootPath, pathBackendBuffering)

	var buffering *types.Buffering
	if len(pathsBuffering) > 0 {
		if buffering == nil {
			buffering = &types.Buffering{}
		}

		buffering.MaxRequestBodyBytes = p.getInt64(0, rootPath, pathBackendBufferingMaxRequestBodyBytes)
		buffering.MaxResponseBodyBytes = p.getInt64(0, rootPath, pathBackendBufferingMaxResponseBodyBytes)
		buffering.MemRequestBodyBytes = p.getInt64(0, rootPath, pathBackendBufferingMemRequestBodyBytes)
		buffering.MemResponseBodyBytes = p.getInt64(0, rootPath, pathBackendBufferingMemResponseBodyBytes)
		buffering.RetryExpression = p.get("", rootPath, pathBackendBufferingRetryExpression)
	}

	return buffering
}

func (p *Provider) getTLSSection(prefix string) []*tls.Configuration {
	var tlsSection []*tls.Configuration

	for _, tlsConfPath := range p.list(prefix, pathTLS) {
		certFile := p.get("", tlsConfPath, pathTLSCertFile)
		keyFile := p.get("", tlsConfPath, pathTLSKeyFile)

		if len(certFile) == 0 && len(keyFile) == 0 {
			log.Warnf("Invalid TLS configuration (no cert and no key): %s", tlsConfPath)
			continue
		}

		entryPoints := p.getList(tlsConfPath, pathTLSEntryPoints)
		if len(entryPoints) == 0 {
			log.Warnf("Invalid TLS configuration (no entry points): %s", tlsConfPath)
			continue
		}

		tlsConf := &tls.Configuration{
			EntryPoints: entryPoints,
			Certificate: &tls.Certificate{
				CertFile: tls.FileOrContent(certFile),
				KeyFile:  tls.FileOrContent(keyFile),
			},
		}

		tlsSection = append(tlsSection, tlsConf)
	}

	return tlsSection
}

// getTLSClientCert create TLS client header configuration from labels
func (p *Provider) getTLSClientCert(rootPath string) *types.TLSClientHeaders {
	if !p.hasPrefix(rootPath, pathFrontendPassTLSClientCert) {
		return nil
	}

	tlsClientHeaders := &types.TLSClientHeaders{
		PEM: p.getBool(false, rootPath, pathFrontendPassTLSClientCertPem),
	}

	if p.hasPrefix(rootPath, pathFrontendPassTLSClientCertInfos) {
		infos := &types.TLSClientCertificateInfos{
			NotAfter:  p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosNotAfter),
			NotBefore: p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosNotBefore),
			Sans:      p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSans),
		}

		if p.hasPrefix(rootPath, pathFrontendPassTLSClientCertInfosSubject) {
			subject := &types.TLSCLientCertificateDNInfos{
				CommonName:      p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectCommonName),
				Country:         p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectCountry),
				DomainComponent: p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectDomainComponent),
				Locality:        p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectLocality),
				Organization:    p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectOrganization),
				Province:        p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectProvince),
				SerialNumber:    p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosSubjectSerialNumber),
			}
			infos.Subject = subject
		}

		if p.hasPrefix(rootPath, pathFrontendPassTLSClientCertInfosIssuer) {
			issuer := &types.TLSCLientCertificateDNInfos{
				CommonName:      p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerCommonName),
				Country:         p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerCountry),
				DomainComponent: p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerDomainComponent),
				Locality:        p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerLocality),
				Organization:    p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerOrganization),
				Province:        p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerProvince),
				SerialNumber:    p.getBool(false, rootPath, pathFrontendPassTLSClientCertInfosIssuerSerialNumber),
			}
			infos.Issuer = issuer
		}

		tlsClientHeaders.Infos = infos
	}
	return tlsClientHeaders
}

// hasDeprecatedBasicAuth check if the frontend basic auth use the deprecated configuration
func (p *Provider) hasDeprecatedBasicAuth(rootPath string) bool {
	return len(p.getList(rootPath, pathFrontendBasicAuth)) > 0
}

// GetAuth Create auth from path
func (p *Provider) getAuth(rootPath string) *types.Auth {
	hasDeprecatedBasicAuth := p.hasDeprecatedBasicAuth(rootPath)
	if p.hasPrefix(rootPath, pathFrontendAuth) || hasDeprecatedBasicAuth {
		auth := &types.Auth{
			HeaderField: p.get("", rootPath, pathFrontendAuthHeaderField),
		}

		if p.hasPrefix(rootPath, pathFrontendAuthBasic) || hasDeprecatedBasicAuth {
			auth.Basic = p.getAuthBasic(rootPath)
		} else if p.hasPrefix(rootPath, pathFrontendAuthDigest) {
			auth.Digest = p.getAuthDigest(rootPath)
		} else if p.hasPrefix(rootPath, pathFrontendAuthForward) {
			auth.Forward = p.getAuthForward(rootPath)
		}

		return auth
	}
	return nil
}

// getAuthBasic Create Basic Auth from path
func (p *Provider) getAuthBasic(rootPath string) *types.Basic {
	basicAuth := &types.Basic{
		UsersFile:    p.get("", rootPath, pathFrontendAuthBasicUsersFile),
		RemoveHeader: p.getBool(false, rootPath, pathFrontendAuthBasicRemoveHeader),
	}

	// backward compatibility
	if p.hasDeprecatedBasicAuth(rootPath) {
		basicAuth.Users = p.getList(rootPath, pathFrontendBasicAuth)
		log.Warnf("Deprecated configuration found: %s. Please use %s.", pathFrontendBasicAuth, pathFrontendAuthBasic)
	} else {
		basicAuth.Users = p.getList(rootPath, pathFrontendAuthBasicUsers)
	}

	return basicAuth
}

// getAuthDigest Create Digest Auth from path
func (p *Provider) getAuthDigest(rootPath string) *types.Digest {
	return &types.Digest{
		Users:        p.getList(rootPath, pathFrontendAuthDigestUsers),
		UsersFile:    p.get("", rootPath, pathFrontendAuthDigestUsersFile),
		RemoveHeader: p.getBool(false, rootPath, pathFrontendAuthDigestRemoveHeader),
	}
}

// getAuthForward Create Forward Auth from path
func (p *Provider) getAuthForward(rootPath string) *types.Forward {
	forwardAuth := &types.Forward{
		Address:             p.get("", rootPath, pathFrontendAuthForwardAddress),
		TrustForwardHeader:  p.getBool(false, rootPath, pathFrontendAuthForwardTrustForwardHeader),
		AuthResponseHeaders: p.getList(rootPath, pathFrontendAuthForwardAuthResponseHeaders),
	}

	// TLS configuration
	if len(p.getList(rootPath, pathFrontendAuthForwardTLS)) > 0 {
		forwardAuth.TLS = &types.ClientTLS{
			CA:                 p.get("", rootPath, pathFrontendAuthForwardTLSCa),
			CAOptional:         p.getBool(false, rootPath, pathFrontendAuthForwardTLSCaOptional),
			Cert:               p.get("", rootPath, pathFrontendAuthForwardTLSCert),
			InsecureSkipVerify: p.getBool(false, rootPath, pathFrontendAuthForwardTLSInsecureSkipVerify),
			Key:                p.get("", rootPath, pathFrontendAuthForwardTLSKey),
		}
	}

	return forwardAuth
}

func (p *Provider) getRoutes(rootPath string) map[string]types.Route {
	var routes map[string]types.Route

	rts := p.list(rootPath, pathFrontendRoutes)
	for _, rt := range rts {

		rule := p.get("", rt, pathFrontendRule)
		if len(rule) == 0 {
			continue
		}

		if routes == nil {
			routes = make(map[string]types.Route)
		}

		routeName := p.last(rt)
		routes[routeName] = types.Route{
			Rule: rule,
		}
	}

	return routes
}

func (p *Provider) getServers(rootPath string) map[string]types.Server {
	var servers map[string]types.Server

	serverKeys := p.listServers(rootPath)
	for _, serverKey := range serverKeys {
		serverURL := p.get("", serverKey, pathBackendServerURL)
		if len(serverURL) == 0 {
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		serverName := p.last(serverKey)
		servers[serverName] = types.Server{
			URL:    serverURL,
			Weight: p.getInt(label.DefaultWeight, serverKey, pathBackendServerWeight),
		}
	}

	return servers
}

func (p *Provider) listServers(backend string) []string {
	serverNames := p.list(backend, pathBackendServers)
	return fun.Filter(p.serverFilter, serverNames).([]string)
}

func (p *Provider) serverFilter(serverName string) bool {
	key := fmt.Sprint(serverName, pathBackendServerURL)
	if _, err := p.kvClient.Get(key, nil); err != nil {
		log.Errorf("Failed to retrieve value for key %s: %s", key, err)
		checkError(err)

		return false
	}
	return p.checkConstraints(serverName, pathTags)
}

func (p *Provider) checkConstraints(keys ...string) bool {
	joinedKeys := strings.Join(keys, "")
	keyPair, err := p.kvClient.Get(joinedKeys, nil)
	if err != nil {
		checkError(err)
	}

	value := ""
	if err == nil && keyPair != nil && keyPair.Value != nil {
		value = string(keyPair.Value)
	}

	constraintTags := label.SplitAndTrimString(value, ",")
	ok, failingConstraint := p.MatchConstraints(constraintTags)
	if !ok {
		if failingConstraint != nil {
			log.Debugf("Constraint %v not matching with following tags: %v", failingConstraint.String(), value)
		}
		return false
	}
	return true
}

func (p *Provider) getFuncString(key string, defaultValue string) func(rootPath string) string {
	return func(rootPath string) string {
		return p.get(defaultValue, rootPath, key)
	}
}

func (p *Provider) getFuncBool(key string, defaultValue bool) func(rootPath string) bool {
	return func(rootPath string) bool {
		return p.getBool(defaultValue, rootPath, key)
	}
}

func (p *Provider) getFuncInt(key string, defaultValue int) func(rootPath string) int {
	return func(rootPath string) int {
		return p.getInt(defaultValue, rootPath, key)
	}
}

func (p *Provider) getFuncList(key string) func(rootPath string) []string {
	return func(rootPath string) []string {
		return p.getList(rootPath, key)
	}
}

func (p *Provider) get(defaultValue string, keyParts ...string) string {
	key := strings.Join(keyParts, "")

	if p.storeType == store.ETCD {
		key = strings.TrimPrefix(key, pathSeparator)
	}

	keyPair, err := p.kvClient.Get(key, nil)
	if err != nil || keyPair == nil {
		log.Debugf("Cannot get key %s %s", key, err)
		checkError(err)

		log.Debugf("Setting %s to default: %s", key, defaultValue)
		return defaultValue
	}

	return string(keyPair.Value)
}

func (p *Provider) getBool(defaultValue bool, keyParts ...string) bool {
	rawValue := p.get(strconv.FormatBool(defaultValue), keyParts...)

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %v: %s", keyParts, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) has(keyParts ...string) bool {
	value := p.get("", keyParts...)
	return len(value) > 0
}

func (p *Provider) hasPrefix(keyParts ...string) bool {
	baseKey := strings.Join(keyParts, "")
	if !strings.HasSuffix(baseKey, "/") {
		baseKey += "/"
	}

	listKeys, err := p.kvClient.List(baseKey, nil)
	if err != nil {
		log.Debugf("Cannot list keys under %q: %v", baseKey, err)
		checkError(err)

		return false
	}

	return len(listKeys) > 0
}

func (p *Provider) getInt(defaultValue int, keyParts ...string) int {
	rawValue := p.get("", keyParts...)

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil {
		log.Errorf("Invalid value for %v: %s", keyParts, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) getInt64(defaultValue int64, keyParts ...string) int64 {
	rawValue := p.get("", keyParts...)

	if len(rawValue) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil {
		log.Errorf("Invalid value for %v: %s", keyParts, rawValue)
		return defaultValue
	}
	return value
}

func (p *Provider) list(keyParts ...string) []string {
	rootKey := strings.Join(keyParts, "")

	keysPairs, err := p.kvClient.List(rootKey, nil)
	if err != nil {
		log.Debugf("Cannot list keys under %q: %v", rootKey, err)
		checkError(err)

		return nil
	}

	directoryKeys := make(map[string]string)
	for _, key := range keysPairs {
		directory := strings.Split(strings.TrimPrefix(key.Key, rootKey), pathSeparator)[0]
		directoryKeys[directory] = rootKey + directory
	}

	keys := fun.Values(directoryKeys).([]string)
	sort.Strings(keys)
	return keys
}

func (p *Provider) getList(keyParts ...string) []string {
	values := p.splitGet(keyParts...)
	if len(values) > 0 {
		return values
	}

	return p.getSlice(keyParts...)
}

// get sub keys. ex: foo/0, foo/1, foo/2
func (p *Provider) getSlice(keyParts ...string) []string {
	baseKey := strings.Join(keyParts, "")
	if !strings.HasSuffix(baseKey, "/") {
		baseKey += "/"
	}

	listKeys := p.list(baseKey)

	var values []string
	for _, entryKey := range listKeys {
		val := p.get("", entryKey)
		if len(val) > 0 {
			values = append(values, val)
		}
	}
	return values
}

func (p *Provider) splitGet(keyParts ...string) []string {
	value := p.get("", keyParts...)

	if len(value) == 0 {
		return nil
	}
	return label.SplitAndTrimString(value, ",")
}

func (p *Provider) last(key string) string {
	index := strings.LastIndex(key, pathSeparator)
	return key[index+1:]
}

func (p *Provider) getMap(keyParts ...string) map[string]string {
	var mapData map[string]string

	list := p.list(keyParts...)
	for _, name := range list {
		if mapData == nil {
			mapData = make(map[string]string)
		}

		mapData[http.CanonicalHeaderKey(p.last(name))] = p.get("", name)
	}

	return mapData
}

func checkError(err error) {
	if err != nil && err != store.ErrKeyNotFound {
		panic(err)
	}
}
