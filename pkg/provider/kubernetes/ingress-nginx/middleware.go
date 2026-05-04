package ingressnginx

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/utils/ptr"
)

// buildMiddlewares populates all middleware-related fields on loc from its
// IngressConfig and provider defaults. It is called once per location in Phase 1,
// after all k8s resources (secrets, configmaps, endpoints) are resolved.
func (p *Provider) buildMiddlewares(ctx context.Context, loc *location, hostname string, allHosts map[string]bool, endpointCount int) {
	// SSL redirect only suppresses other middlewares on the non-TLS router; the
	// TLS router still gets the full middleware stack.
	p.buildSSLRedirect(loc)

	p.buildAccessLog(loc)
	p.buildCustomHTTPErrors(loc)
	p.buildAppRoot(loc)
	p.buildFromToWwwRedirect(loc, hostname, allHosts)
	p.buildRedirect(loc)
	p.buildBuffering(ctx, loc)
	p.buildIPAllowList(loc)
	p.buildCORS(loc)
	p.buildRewriteTarget(loc)
	p.buildUpstreamVhost(loc)
	p.buildRateLimit(loc)
	p.buildLimitConnections(loc)
	p.buildAuthTLSPassCert(loc)
	p.buildCustomHeaders(loc)
	p.buildSnippetAuth(loc)
	p.buildRetry(loc, endpointCount)
}

func (p *Provider) buildSSLRedirect(loc *location) {
	// force-ssl-redirect redirects to HTTPS regardless of whether the Ingress has a TLS block.
	forceSSLRedirect := ptr.Deref(loc.Config.ForceSSLRedirect, false)
	// When an Ingress has a TLS block, by default SSL redirect should be applied.
	// When an Ingress does not have a TLS block, SSL redirect should not be applied.
	sslRedirect := loc.HasTLS && ptr.Deref(loc.Config.SSLRedirect, true)

	loc.SSLRedirectOnly = forceSSLRedirect || sslRedirect
}

func (p *Provider) buildAccessLog(loc *location) {
	loc.AccessLog = loc.Config.EnableAccessLog
}

func (p *Provider) buildCustomHTTPErrors(loc *location) {
	status := ptr.Deref(loc.Config.CustomHTTPErrors, p.CustomHTTPErrors)
	if len(status) == 0 {
		return
	}

	mw := &middlewareCustomHTTPErrors{
		Status:      status,
		Namespace:   loc.Namespace,
		IngressName: loc.IngressName,
		ServiceName: loc.ServiceName,
		ServicePort: loc.ServicePort,
	}

	switch {
	case loc.ResolvedHTTPErrorBackendName != "":
		// Per-ingress default-backend annotation: translator will create per-router services.
		mw.ErrorBackendName = loc.ResolvedHTTPErrorBackendName
	case p.defaultBackendServiceName != "":
		// Provider-level default backend: translator uses the shared "default-backend" service.
		mw.ErrorServiceName = defaultBackendName
	default:
		// No error-page service available — skip (mirrors nginx behavior).
		return
	}

	loc.CustomHTTPErrors = mw
}

func (p *Provider) buildAppRoot(loc *location) {
	if loc.Config.AppRoot == nil || !strings.HasPrefix(*loc.Config.AppRoot, "/") {
		return
	}
	loc.AppRoot = loc.Config.AppRoot
}

func (p *Provider) buildFromToWwwRedirect(loc *location, hostname string, allHosts map[string]bool) {
	if !ptr.Deref(loc.Config.FromToWwwRedirect, false) {
		return
	}

	wwwType := strings.HasPrefix(hostname, "www.")
	wildcardType := strings.HasPrefix(hostname, "*.")
	bypass := (wwwType && allHosts[strings.TrimPrefix(hostname, "www.")]) ||
		(!wwwType && allHosts["www."+hostname]) || wildcardType
	if bypass {
		return
	}

	extraRule := fmt.Sprintf("Host(%q)", "www."+hostname)
	target := hostname
	if wwwType {
		// hostname is www.host — capture non-www traffic and redirect TO www.host.
		// Only the extra router rule changes; target stays as hostname (www.host).
		extraRule = fmt.Sprintf("Host(%q)", strings.TrimPrefix(hostname, "www."))
	}

	loc.FromToWwwRedirect = &middlewareFromToWwwRedirect{
		ExtraRouterRule: extraRule,
		TargetHostname:  target,
	}
}

func (p *Provider) buildRedirect(loc *location) {
	if loc.Config.PermanentRedirect == nil && loc.Config.TemporalRedirect == nil {
		return
	}

	url, code := "", 0

	if loc.Config.PermanentRedirect != nil {
		url = *loc.Config.PermanentRedirect
		code = ptr.Deref(loc.Config.PermanentRedirectCode, http.StatusMovedPermanently)
		// NGINX only accepts valid redirect codes and defaults to 301.
		if code < 300 || code > 308 {
			code = http.StatusMovedPermanently
		}
	}

	// TemporalRedirect takes precedence over the PermanentRedirect.
	if loc.Config.TemporalRedirect != nil {
		url = *loc.Config.TemporalRedirect
		code = ptr.Deref(loc.Config.TemporalRedirectCode, http.StatusFound)
		// NGINX only accepts valid redirect codes and defaults to 302.
		if code < 300 || code > 308 {
			code = http.StatusFound
		}
	}

	loc.Redirect = &dynamic.RedirectRegex{
		Regex:       ".*",
		Replacement: url,
		StatusCode:  &code,
	}
}

func (p *Provider) buildBuffering(ctx context.Context, loc *location) {
	disableReq := !p.ProxyRequestBuffering
	if loc.Config.ProxyRequestBuffering != nil {
		disableReq = *loc.Config.ProxyRequestBuffering != "on"
	}

	disableResp := !p.ProxyBuffering
	if loc.Config.ProxyBuffering != nil {
		disableResp = *loc.Config.ProxyBuffering != "on"
	}

	if disableReq && disableResp {
		return
	}

	buf := &dynamic.Buffering{
		DisableRequestBuffer:  disableReq,
		DisableResponseBuffer: disableResp,
		MemRequestBodyBytes:   p.ClientBodyBufferSize,
		MaxRequestBodyBytes:   p.ProxyBodySize,
		MemResponseBodyBytes:  p.ProxyBufferSize * int64(p.ProxyBuffersNumber),
	}

	if !disableReq {
		if s := ptr.Deref(loc.Config.ClientBodyBufferSize, ""); s != "" {
			if v, err := nginxSizeToBytes(s); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("client-body-buffer-size invalid, using provider default")
			} else {
				buf.MemRequestBodyBytes = v
			}
		}
		if s := ptr.Deref(loc.Config.ProxyBodySize, ""); s != "" {
			if v, err := nginxSizeToBytes(s); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("proxy-body-size invalid, using provider default")
			} else {
				buf.MaxRequestBodyBytes = v
			}
		}
	}

	if !disableResp {
		if loc.Config.ProxyBufferSize != nil || loc.Config.ProxyBuffersNumber != nil {
			bufSize := p.ProxyBufferSize
			if s := ptr.Deref(loc.Config.ProxyBufferSize, ""); s != "" {
				if v, err := nginxSizeToBytes(s); err != nil {
					log.Ctx(ctx).Warn().Err(err).Msg("proxy-buffer-size invalid, using provider default")
				} else {
					bufSize = v
				}
			}
			buf.MemResponseBodyBytes = bufSize * int64(ptr.Deref(loc.Config.ProxyBuffersNumber, p.ProxyBuffersNumber))
		}

		maxTmp := int64(1024 * 1024 * 1024) // 1 GB nginx default
		if loc.Config.ProxyMaxTempFileSize != nil {
			if v, err := nginxSizeToBytes(*loc.Config.ProxyMaxTempFileSize); err != nil {
				log.Ctx(ctx).Warn().Err(err).Msg("proxy-max-temp-file-size invalid, using 1GB default")
			} else {
				maxTmp = v
			}
		}
		buf.MaxResponseBodyBytes = buf.MemResponseBodyBytes + maxTmp
	}

	loc.Buffering = buf
}

func (p *Provider) buildIPAllowList(loc *location) {
	allowed := ptr.Deref(loc.Config.AllowlistSourceRange, ptr.Deref(loc.Config.WhitelistSourceRange, ""))
	if allowed == "" {
		return
	}

	var ranges []string
	for r := range strings.SplitSeq(allowed, ",") {
		ranges = append(ranges, strings.TrimSpace(r))
	}

	loc.IPAllowList = &dynamic.IPAllowList{SourceRange: ranges}
}

func (p *Provider) buildCORS(loc *location) {
	if !ptr.Deref(loc.Config.EnableCORS, false) {
		return
	}

	loc.CORS = &dynamic.Headers{
		AccessControlAllowCredentials: ptr.Deref(loc.Config.EnableCORSAllowCredentials, true),
		AccessControlExposeHeaders:    ptr.Deref(loc.Config.CORSExposeHeaders, []string{}),
		AccessControlAllowHeaders:     ptr.Deref(loc.Config.CORSAllowHeaders, []string{"DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range,Authorization"}),
		AccessControlAllowMethods:     ptr.Deref(loc.Config.CORSAllowMethods, []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"}),
		AccessControlAllowOriginList:  ptr.Deref(loc.Config.CORSAllowOrigin, []string{"*"}),
		AccessControlMaxAge:           int64(ptr.Deref(loc.Config.CORSMaxAge, 1728000)),
	}
}

func (p *Provider) buildRewriteTarget(loc *location) {
	rewrite := ptr.Deref(loc.Config.RewriteTarget, "")
	if rewrite == "" || rewrite == loc.Path {
		return
	}

	regex := loc.Path

	xfp := ""
	if loc.Config.XForwardedPrefix != nil && regexPathWithCapture.MatchString(*loc.Config.XForwardedPrefix) {
		xfp = *loc.Config.XForwardedPrefix
	}

	loc.RewriteTarget = &dynamic.RewriteTarget{
		Regex:            regex,
		Replacement:      rewrite,
		XForwardedPrefix: xfp,
	}
}

func (p *Provider) buildUpstreamVhost(loc *location) {
	if loc.Config.UpstreamVHost == nil {
		return
	}

	// ingress-nginx exposes per-location variables (set at the NGINX location scope)
	// to upstream-vhost: $namespace, $ingress_name, $service_name, $service_port,
	// $location_path. They are static at config-build time.
	loc.UpstreamVhost = &dynamic.UpstreamVHost{
		VHost: *loc.Config.UpstreamVHost,
		Vars: map[string]string{
			"$namespace":     loc.Namespace,
			"$ingress_name":  loc.IngressName,
			"$location_path": loc.Path,
			"$service_name":  loc.ServiceName,
			"$service_port":  loc.ServicePort,
		},
	}
}

func (p *Provider) buildRateLimit(loc *location) {
	rpm := ptr.Deref(loc.Config.LimitRPM, 0)
	rps := ptr.Deref(loc.Config.LimitRPS, 0)
	if rpm == 0 && rps == 0 {
		return
	}

	burst := getLimitBurstMultiplier(loc.Config)
	if rpm > 0 {
		loc.RateLimitRPM = &dynamic.RateLimit{
			Average: int64(rpm),
			Period:  ptypes.Duration(time.Minute),
			Burst:   int64(rpm) * burst,
		}
	}
	if rps > 0 {
		loc.RateLimitRPS = &dynamic.RateLimit{
			Average: int64(rps),
			Period:  ptypes.Duration(time.Second),
			Burst:   int64(rps) * burst,
		}
	}
}

func (p *Provider) buildLimitConnections(loc *location) {
	limit := ptr.Deref(loc.Config.LimitConnections, 0)
	if limit <= 0 {
		return
	}

	loc.LimitConnections = &dynamic.InFlightReq{
		Amount: int64(limit),
		SourceCriterion: &dynamic.SourceCriterion{
			IPStrategy: &dynamic.IPStrategy{},
		},
	}
}

func (p *Provider) buildAuthTLSPassCert(loc *location) {
	if !ptr.Deref(loc.Config.AuthTLSPassCertificateToUpstream, false) || loc.Config.AuthTLSSecret == nil {
		return
	}

	loc.AuthTLSPassCert = &dynamic.AuthTLSPassCertificateToUpstream{
		ClientAuthType: clientAuthTypeFromString(loc.Config.AuthTLSVerifyClient),
	}
}

func (p *Provider) buildCustomHeaders(loc *location) {
	// ResolvedCustomHeaders is already populated. Validate header values here so the
	// translator receives only valid entries.
	if len(loc.ResolvedCustomHeaders) == 0 {
		return
	}

	for _, v := range loc.ResolvedCustomHeaders {
		if !headerValueRegexp.MatchString(v) {
			loc.ResolvedCustomHeaders = nil
			loc.Error = true
			return
		}
	}
}

func (p *Provider) buildSnippetAuth(loc *location) {
	snippet := ptr.Deref(loc.Config.ConfigurationSnippet, "")
	serverSnippet := loc.ServerSnippet

	authURL := ptr.Deref(loc.Config.AuthURL, "")
	if authURL == "" && p.GlobalAuthURL != "" && ptr.Deref(loc.Config.EnableGlobalAuth, true) {
		authURL = p.GlobalAuthURL
	}

	if serverSnippet == "" && snippet == "" && authURL == "" {
		return
	}

	sa := &dynamic.Snippet{
		ServerSnippet:        serverSnippet,
		ConfigurationSnippet: snippet,
	}

	if authURL != "" {
		auth := &dynamic.Auth{
			Address:       authURL,
			AuthSigninURL: ptr.Deref(loc.Config.AuthSignin, ""),
			Method:        ptr.Deref(loc.Config.AuthMethod, ""),
			Snippet:       ptr.Deref(loc.Config.AuthSnippet, ""),
		}
		if raw := ptr.Deref(loc.Config.AuthResponseHeaders, ""); raw != "" {
			for h := range strings.SplitSeq(raw, ",") {
				if trimmed := strings.TrimSpace(h); trimmed != "" {
					auth.AuthResponseHeaders = append(auth.AuthResponseHeaders, trimmed)
				}
			}
		}
		sa.Auth = auth
	}

	loc.SnippetAuth = sa
}

func (p *Provider) buildRetry(loc *location, endpointCount int) {
	attempts := ptr.Deref(loc.Config.ProxyNextUpstreamTries, p.ProxyNextUpstreamTries)
	// Safeguard to deactivate retry when the value is less than 0.
	if attempts < 0 {
		return
	}

	upstream := ptr.Deref(loc.Config.ProxyNextUpstream, p.ProxyNextUpstream)
	if upstream == "" {
		return
	}

	conditions := strings.Fields(upstream)
	// "off" disables the retry entirely.
	if slices.Contains(conditions, "off") {
		return
	}

	// proxy-next-upstream-tries = 0 on NGINX means unlimited tries, which maps to try every available server.
	// To avoid infinite retries, put the number of servers as the attempts limit.
	if attempts == 0 {
		attempts = endpointCount
	}
	if attempts <= 0 {
		return
	}

	retry := &dynamic.Retry{Attempts: attempts}

	hasError := slices.Contains(conditions, "error")
	hasTimeout := slices.Contains(conditions, "timeout")
	if !hasError && !hasTimeout {
		retry.DisableRetryOnNetworkError = true
	}

	for _, sc := range conditions {
		if code, ok := strings.CutPrefix(sc, "http_"); ok {
			retry.Status = append(retry.Status, code)
		}
	}

	if slices.Contains(conditions, "non_idempotent") {
		retry.RetryNonIdempotentMethod = true
	}

	if timeoutSec := ptr.Deref(loc.Config.ProxyNextUpstreamTimeout, p.ProxyNextUpstreamTimeout); timeoutSec > 0 {
		retry.Timeout = ptypes.Duration(time.Duration(timeoutSec) * time.Second)
	}

	loc.Retry = retry
}

// ---- utilities ---------------------------------------------------------------

func getLimitBurstMultiplier(cfg IngressConfig) int64 {
	m := ptr.Deref(cfg.LimitBurstMultiplier, defaultLimitBurstMultiplier)
	if m < 1 {
		m = defaultLimitBurstMultiplier
	}
	return int64(m)
}

// nginxSizeToBytes convert nginx size to memory bytes as defined in https://nginx.org/en/docs/syntax.html.
func nginxSizeToBytes(nginxSize string) (int64, error) {
	units := map[string]int64{
		"g": 1024 * 1024 * 1024,
		"m": 1024 * 1024,
		"k": 1024,
		"b": 1,
		"":  1,
	}

	if !nginxSizeRegexp.MatchString(nginxSize) {
		return 0, fmt.Errorf("unable to parse number %s", nginxSize)
	}
	size := nginxSizeRegexp.FindStringSubmatch(nginxSize)
	bytes, err := strconv.ParseInt(size[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return bytes * units[strings.ToLower(size[2])], nil
}
