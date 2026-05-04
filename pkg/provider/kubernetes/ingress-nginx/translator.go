package ingressnginx

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

// translate converts a model produced by Phase 1 into a Traefik dynamic.Configuration.
func (p *Provider) translate(ctx context.Context, mc *model) *dynamic.Configuration {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{},
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()
	conf.HTTP.Services[unavailableServiceName] = &dynamic.Service{LoadBalancer: lb}

	for cert, key := range mc.Certs {
		conf.TLS.Certificates = append(conf.TLS.Certificates, &tls.CertAndStores{
			Certificate: tls.Certificate{
				CertFile: types.FileOrContent(cert),
				KeyFile:  types.FileOrContent(key),
			},
		})
	}

	if mc.DefaultBackend != nil {
		obs := &dynamic.RouterObservabilityConfig{
			Metadata: &dynamic.ObservabilityMetadata{
				Ingress: &dynamic.KubernetesIngressMetadata{
					Namespace:   mc.DefaultBackend.Namespace,
					ServiceName: mc.DefaultBackend.ServiceName,
				},
			},
		}

		var serversTransportName string
		if loc := mc.DefaultBackendLocation; loc != nil {
			obs.Metadata.Ingress.IngressName = loc.IngressName
			obs.Metadata.Ingress.ServicePort = loc.ServicePort
			if loc.ServersTransport != nil && loc.ServersTransportName != "" {
				serversTransportName = loc.ServersTransportName
				conf.HTTP.ServersTransports[loc.ServersTransportName] = loc.ServersTransport
			}
		}

		conf.HTTP.Services[defaultBackendName] = buildService(mc.DefaultBackend, serversTransportName)

		rt := &dynamic.Router{
			EntryPoints:   p.NonTLSEntryPoints,
			Rule:          `PathPrefix("/")`,
			RuleSyntax:    "default",
			Priority:      math.MinInt32,
			Service:       defaultBackendName,
			Observability: obs,
		}
		rtTLS := &dynamic.Router{
			EntryPoints:   p.TLSEntryPoints,
			Rule:          `PathPrefix("/")`,
			RuleSyntax:    "default",
			Priority:      math.MinInt32,
			Service:       defaultBackendName,
			TLS:           &dynamic.RouterTLSConfig{},
			Observability: obs,
		}

		if loc := mc.DefaultBackendLocation; loc != nil && loc.Retry != nil {
			retryName := defaultBackendName + "-retry"
			retryTLSName := defaultBackendTLSName + "-retry"
			conf.HTTP.Middlewares[retryName] = &dynamic.Middleware{Retry: loc.Retry}
			conf.HTTP.Middlewares[retryTLSName] = &dynamic.Middleware{Retry: loc.Retry}
			rt.Middlewares = []string{retryName}
			rtTLS.Middlewares = []string{retryTLSName}
		}

		conf.HTTP.Routers[defaultBackendName] = rt
		conf.HTTP.Routers[defaultBackendTLSName] = rtTLS
	}

	for _, pt := range mc.PassthroughBackends {
		backend, ok := mc.Backends[pt.BackendName]
		if !ok {
			log.Ctx(ctx).Error().Msgf("Passthrough backend %q not found in metamodel", pt.BackendName)
			continue
		}

		lb := &dynamic.TCPServersLoadBalancer{}
		for _, ep := range backend.Endpoints {
			lb.Servers = append(lb.Servers, dynamic.TCPServer{Address: ep.Address})
		}

		conf.TCP.Services[pt.BackendName] = &dynamic.TCPService{LoadBalancer: lb}
		conf.TCP.Routers[pt.RouterKey] = &dynamic.TCPRouter{
			EntryPoints: p.TLSEntryPoints,
			Rule:        fmt.Sprintf("HostSNI(%q)", pt.Hostname),
			RuleSyntax:  "default",
			Service:     pt.BackendName,
			TLS:         &dynamic.RouterTCPTLSConfig{Passthrough: true},
		}
	}

	for _, srv := range mc.Servers {
		for _, loc := range srv.Locations {
			if loc.ServersTransport != nil && loc.ServersTransportName != "" {
				if _, exists := conf.HTTP.ServersTransports[loc.ServersTransportName]; !exists {
					conf.HTTP.ServersTransports[loc.ServersTransportName] = loc.ServersTransport
				}
			}

			if loc.TLSOption != nil && loc.TLSOptionName != "" {
				if conf.TLS.Options == nil {
					conf.TLS.Options = make(map[string]tls.Options)
				}
				if _, exists := conf.TLS.Options[loc.TLSOptionName]; !exists {
					conf.TLS.Options[loc.TLSOptionName] = *loc.TLSOption
				}
			}

			backend, ok := mc.Backends[loc.BackendName]
			if !ok {
				log.Ctx(ctx).Error().Msgf("Backend %q not found for location %s%s", loc.BackendName, srv.Hostname, loc.Path)
				continue
			}

			primarySvcName := loc.BackendName
			conf.HTTP.Services[primarySvcName] = buildServiceWithLocConfig(backend, loc.ServersTransportName, loc.Config)

			obs := &dynamic.RouterObservabilityConfig{
				Metadata: &dynamic.ObservabilityMetadata{
					Ingress: &dynamic.KubernetesIngressMetadata{
						Namespace:   loc.Namespace,
						IngressName: loc.IngressName,
						ServiceName: loc.ServiceName,
						ServicePort: loc.ServicePort,
					},
				},
			}

			routerSvcName := primarySvcName
			canarySvcName := primarySvcName + "-canary"
			if loc.Canary != nil {
				canaryBackend, ok := mc.Backends[loc.Canary.BackendName]
				if ok {
					canaryWRRName := primarySvcName + "-wrr"

					canarySvc := buildServiceWithLocConfig(canaryBackend, loc.ServersTransportName, loc.Config)
					conf.HTTP.Services[canarySvcName] = canarySvc

					conf.HTTP.Services[canaryWRRName] = &dynamic.Service{
						Weighted: &dynamic.WeightedRoundRobin{
							Sticky: buildSticky(loc.Config, "wrr"),
							Services: []dynamic.WRRService{
								{Name: primarySvcName, Weight: ptr.To(loc.Canary.WeightTotal - loc.Canary.Weight)},
								{Name: canarySvcName, Weight: ptr.To(loc.Canary.Weight)},
							},
						},
					}
					routerSvcName = canaryWRRName
				}
			}

			rule := buildRule(srv.Hostname, loc)

			var routerKey string
			if loc.IsIngressDefaultBackend {
				routerKey = provider.Normalize(fmt.Sprintf("%s-%s-default-backend", loc.Namespace, loc.IngressName))
			} else {
				routerKey = provider.Normalize(fmt.Sprintf("%s-%s-rule-%d-path-%d", loc.Namespace, loc.IngressName, loc.RuleIndex, loc.LocationIndex))
			}

			rt := &dynamic.Router{
				EntryPoints:   p.NonTLSEntryPoints,
				Rule:          rule,
				RuleSyntax:    "default",
				Service:       routerSvcName,
				Observability: obs,
			}

			rtTLS := &dynamic.Router{
				EntryPoints: p.TLSEntryPoints,
				Rule:        rule,
				RuleSyntax:  "default",
				Service:     routerSvcName,
				TLS: &dynamic.RouterTLSConfig{
					Options: loc.TLSOptionName,
				},
				Observability: obs,
			}

			// TODO: in case we want to add the unavailable service only when it is used this should be done here.
			if loc.Error {
				rt.Service = unavailableServiceName
				rtTLS.Service = unavailableServiceName
			}

			conf.HTTP.Routers[routerKey] = rt
			conf.HTTP.Routers[routerKey+"-tls"] = rtTLS

			if !loc.Error {
				p.applyMiddlewares(mc, loc, routerKey, rt, conf)
				p.applyMiddlewares(mc, loc, routerKey+"-tls", rtTLS, conf)
				applyFromToWwwRedirect(loc, routerKey, rt, obs, conf)
				applyFromToWwwRedirect(loc, routerKey+"-tls", rtTLS, obs, conf)
			}

			if loc.Canary != nil && loc.Canary.RequiresCanaryRouter() {
				canaryKey := routerKey + "-canary"
				canaryRouter := &dynamic.Router{
					EntryPoints:   rt.EntryPoints,
					Rule:          appendCanaryRule(rule, loc.Canary),
					RuleSyntax:    rt.RuleSyntax,
					Service:       canarySvcName,
					Observability: obs,
				}
				conf.HTTP.Routers[canaryKey] = canaryRouter
				p.applyMiddlewares(mc, loc, canaryKey, canaryRouter, conf)

				canaryKeyTLS := canaryKey + "-tls"
				canaryRouterTLS := &dynamic.Router{
					EntryPoints:   rtTLS.EntryPoints,
					Rule:          appendCanaryRule(rule, loc.Canary),
					RuleSyntax:    rtTLS.RuleSyntax,
					Service:       canarySvcName,
					TLS:           rtTLS.TLS,
					Observability: obs,
				}
				conf.HTTP.Routers[canaryKeyTLS] = canaryRouterTLS
				p.applyMiddlewares(mc, loc, canaryKeyTLS, canaryRouterTLS, conf)
			}

			if loc.Canary != nil && loc.Canary.RequiresNonCanaryRouter() {
				nonCanaryKey := routerKey + "-non-canary"
				nonCanaryRouter := &dynamic.Router{
					EntryPoints:   rt.EntryPoints,
					Rule:          appendNonCanaryRule(rule, loc.Canary),
					RuleSyntax:    rt.RuleSyntax,
					Service:       primarySvcName,
					Observability: obs,
				}
				conf.HTTP.Routers[nonCanaryKey] = nonCanaryRouter
				p.applyMiddlewares(mc, loc, nonCanaryKey, nonCanaryRouter, conf)

				nonCanaryKeyTLS := nonCanaryKey + "-tls"
				nonCanaryRouterTLS := &dynamic.Router{
					EntryPoints:   rtTLS.EntryPoints,
					Rule:          appendNonCanaryRule(rule, loc.Canary),
					RuleSyntax:    rtTLS.RuleSyntax,
					Service:       primarySvcName,
					TLS:           rtTLS.TLS,
					Observability: obs,
				}
				conf.HTTP.Routers[nonCanaryKeyTLS] = nonCanaryRouterTLS
				p.applyMiddlewares(mc, loc, nonCanaryKeyTLS, nonCanaryRouterTLS, conf)
			}
		}
	}

	return conf
}

func buildService(backend *backend, serversTransportName string) *dynamic.Service {
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	if serversTransportName != "" {
		lb.ServersTransport = serversTransportName
	}

	svc := &dynamic.Service{LoadBalancer: lb}
	for _, ep := range backend.Endpoints {
		svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.Server{
			URL: fmt.Sprintf("http://%s", ep.Address),
		})
	}

	return svc
}

func buildServiceWithLocConfig(backend *backend, serversTransportName string, locCfg IngressConfig) *dynamic.Service {
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	if serversTransportName != "" && len(backend.Endpoints) > 0 {
		lb.ServersTransport = serversTransportName
	}

	lb.Sticky = buildSticky(locCfg, "")

	if upstreamHashBy := ptr.Deref(locCfg.UpstreamHashBy, ""); upstreamHashBy != "" {
		lb.Strategy = dynamic.BalancerStrategyHRW
		lb.NginxUpstreamHashBy = upstreamHashBy
	}

	scheme := parseBackendProtocol(ptr.Deref(locCfg.BackendProtocol, "HTTP"))
	svc := &dynamic.Service{LoadBalancer: lb}
	for _, ep := range backend.Endpoints {
		svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", scheme, ep.Address),
		})
	}

	return svc
}

func appendCanaryRule(rule string, c *canaryConfig) string {
	var rules []string
	if c.Header != "" {
		switch {
		case c.HeaderValue == "" && c.HeaderPattern != "":
			rules = append(rules, fmt.Sprintf("HeaderRegexp(%q, %q)", c.Header, c.HeaderPattern))
		case c.HeaderValue != "":
			rules = append(rules, fmt.Sprintf("Header(%q, %q)", c.Header, c.HeaderValue))
		default:
			rules = append(rules, fmt.Sprintf(`Header(%q, "always")`, c.Header))
		}
	}

	if c.Cookie != "" {
		cookieRule := fmt.Sprintf(`HeaderRegexp("Cookie", %q)`,
			fmt.Sprintf(`(^|;\s*)%s=always(;|$)`, regexp.QuoteMeta(c.Cookie)))
		if c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "" {
			cookieRule = fmt.Sprintf("(%s && !%s)", cookieRule, fmt.Sprintf(`Header(%q, "never")`, c.Header))
		}
		rules = append(rules, cookieRule)
	}

	return fmt.Sprintf("(%s) && (%s)", rule, strings.Join(rules, " || "))
}

func appendNonCanaryRule(rule string, c *canaryConfig) string {
	var rules []string
	if c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "" {
		rules = append(rules, fmt.Sprintf(`Header(%q, "never")`, c.Header))
	}
	if c.Cookie != "" {
		rules = append(rules, fmt.Sprintf(`HeaderRegexp("Cookie", %q)`,
			fmt.Sprintf(`(^|;\s*)%s=never(;|$)`, regexp.QuoteMeta(c.Cookie))))
	}
	return fmt.Sprintf("(%s) && (%s)", rule, strings.Join(rules, " || "))
}

// buildSticky returns a Sticky model if the affinity model is set to "cookie" and nil otherwise.
// It also appends the given nameSuffix to the cookie name if not empty.
func buildSticky(cfg IngressConfig, nameSuffix string) *dynamic.Sticky {
	if ptr.Deref(cfg.Affinity, "") != "cookie" {
		return nil
	}

	name := ptr.Deref(cfg.SessionCookieName, "INGRESSCOOKIE")
	if nameSuffix != "" {
		name += "-" + nameSuffix
	}

	return &dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name:     name,
			Secure:   ptr.Deref(cfg.SessionCookieSecure, false),
			HTTPOnly: true,
			SameSite: strings.ToLower(ptr.Deref(cfg.SessionCookieSameSite, "")),
			MaxAge:   ptr.Deref(cfg.SessionCookieMaxAge, 0),
			Expires:  ptr.Deref(cfg.SessionCookieExpires, 0),
			Path:     ptr.To(ptr.Deref(cfg.SessionCookiePath, "/")),
			Domain:   ptr.Deref(cfg.SessionCookieDomain, ""),
		},
	}
}

func (p *Provider) applyMiddlewares(mc *model, loc *location, routerKey string, rt *dynamic.Router, conf *dynamic.Configuration) {
	if loc.SSLRedirectOnly && rt.TLS == nil {
		name := routerKey + "-redirect-scheme"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{
			RedirectScheme: &dynamic.RedirectScheme{Scheme: "https", ForcePermanentRedirect: true},
		}
		rt.Middlewares = []string{name}
		rt.Service = "noop@internal"
		return
	}

	if loc.AccessLog != nil {
		if rt.Observability == nil {
			rt.Observability = &dynamic.RouterObservabilityConfig{}
		}
		rt.Observability.AccessLogs = ptr.To(*loc.AccessLog)
	}

	if loc.CustomHTTPErrors != nil {
		e := loc.CustomHTTPErrors
		errorSvcName := e.ErrorServiceName
		if e.ErrorBackendName != "" {
			errorSvcName = "default-backend-" + routerKey
			if errBackend, ok := mc.Backends[e.ErrorBackendName]; ok {
				conf.HTTP.Services[errorSvcName] = buildService(errBackend, "")
			}
		}
		headers := http.Header{
			"X-Namespaces":   {e.Namespace},
			"X-Ingress-Name": {e.IngressName},
			"X-Service-Name": {e.ServiceName},
			"X-Service-Port": {e.ServicePort},
		}
		name := routerKey + "-custom-http-errors"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{
			Errors: &dynamic.ErrorPage{
				Status:       e.Status,
				Service:      errorSvcName,
				NginxHeaders: &headers,
			},
		}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.AppRoot != nil {
		name := routerKey + "-app-root"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{
			RedirectRegex: &dynamic.RedirectRegex{
				Regex:       `^(https?://[^/]+)/(\?.*)?$`,
				Replacement: "$1" + *loc.AppRoot,
			},
		}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.Redirect != nil {
		name := routerKey + "-redirect"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{RedirectRegex: loc.Redirect}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.BasicAuth != nil {
		name := routerKey + "-basic-auth"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{BasicAuth: loc.BasicAuth}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.DigestAuth != nil {
		name := routerKey + "-digest-auth"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{DigestAuth: loc.DigestAuth}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.Buffering != nil {
		name := routerKey + "-buffering"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{Buffering: loc.Buffering}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.IPAllowList != nil {
		name := routerKey + "-allowed-source-range"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{IPAllowList: loc.IPAllowList}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.CORS != nil {
		name := routerKey + "-cors"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{Headers: loc.CORS}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.RewriteTarget != nil {
		name := routerKey + "-rewrite-target"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{RewriteTarget: loc.RewriteTarget}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.UpstreamVhost != nil {
		name := routerKey + "-vhost"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{UpstreamVHost: loc.UpstreamVhost}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.RateLimitRPM != nil {
		name := routerKey + "-limit-rpm"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{RateLimit: loc.RateLimitRPM}
		rt.Middlewares = append(rt.Middlewares, name)
	}
	if loc.RateLimitRPS != nil {
		name := routerKey + "-limit-rps"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{RateLimit: loc.RateLimitRPS}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.LimitConnections != nil {
		name := routerKey + "-limit-connections"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{InFlightReq: loc.LimitConnections}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.AuthTLSPassCert != nil && rt.TLS != nil {
		name := routerKey + "-pass-certificate-to-upstream"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{AuthTLSPassCertificateToUpstream: loc.AuthTLSPassCert}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if len(loc.ResolvedCustomHeaders) > 0 {
		name := routerKey + "-custom-headers"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{
			Headers: &dynamic.Headers{CustomResponseHeaders: loc.ResolvedCustomHeaders},
		}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.SnippetAuth != nil {
		name := routerKey + "-snippet"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{Snippet: loc.SnippetAuth}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if loc.Retry != nil {
		name := routerKey + "-retry"
		conf.HTTP.Middlewares[name] = &dynamic.Middleware{Retry: loc.Retry}
		rt.Middlewares = append(rt.Middlewares, name)
	}

	if p.applyMiddlewareFunc != nil {
		if err := p.applyMiddlewareFunc(routerKey, rt, conf, loc.Config); err != nil {
			log.Error().Err(err).Str("router", routerKey).Msg("Error in ApplyMiddlewareFunc")
		}
	} else if ptr.Deref(loc.Config.EnableModSecurity, false) ||
		ptr.Deref(loc.Config.EnableOWASPCoreRules, false) ||
		ptr.Deref(loc.Config.ModSecuritySnippet, "") != "" ||
		ptr.Deref(loc.Config.ModSecurityTransactionID, "") != "" {
		log.Error().Str("router", routerKey).Msg("mod-security annotations require ApplyMiddlewareFunc to be set")
	}
}

func applyFromToWwwRedirect(loc *location, routerKey string, rt *dynamic.Router, obs *dynamic.RouterObservabilityConfig, conf *dynamic.Configuration) {
	if loc.FromToWwwRedirect == nil {
		return
	}

	f := loc.FromToWwwRedirect
	mwName := routerKey + "-from-to-www-redirect"
	conf.HTTP.Middlewares[mwName] = &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
			Replacement: fmt.Sprintf("$1://%s$2/$3", f.TargetHostname),
			StatusCode:  ptr.To(http.StatusPermanentRedirect),
		},
	}

	conf.HTTP.Routers[routerKey+"-from-to-www-redirect"] = &dynamic.Router{
		EntryPoints:   rt.EntryPoints,
		Rule:          f.ExtraRouterRule,
		Priority:      rt.Priority,
		RuleSyntax:    "default",
		Middlewares:   []string{mwName},
		Service:       rt.Service,
		TLS:           rt.TLS,
		Observability: obs,
	}
}

func buildRule(host string, loc *location) string {
	var rules []string

	if host != "" {
		hosts := append([]string{host}, loc.Aliases...)
		hostRules := make([]string, 0, len(hosts))
		for _, h := range hosts {
			hostRules = append(hostRules, fmt.Sprintf("Host(%q)", h))
		}
		if len(hostRules) > 1 {
			rules = append(rules, "("+strings.Join(hostRules, " || ")+")")
		} else {
			rules = append(rules, hostRules[0])
		}
	}

	if len(loc.Path) > 0 {
		pathType := ptr.Deref(loc.PathType, netv1.PathTypePrefix)
		if pathType == netv1.PathTypeImplementationSpecific {
			pathType = netv1.PathTypePrefix
		}

		switch pathType {
		case netv1.PathTypeExact:
			rules = append(rules, fmt.Sprintf("Path(%q)", loc.Path))
		case netv1.PathTypePrefix:
			if loc.UseRegex {
				rules = append(rules, fmt.Sprintf("PathRegexp(%q)", "(?i)^"+loc.Path))
			} else {
				rules = append(rules, buildPrefixRule(loc.Path))
			}
		}
	}

	return strings.Join(rules, " && ")
}

// buildPrefixRule is a helper function to build a path prefix rule that matches path prefix split by `/`.
// For example, the paths `/abc`, `/abc/`, and `/abc/def` would all match the prefix `/abc`,
// but the path `/abcd` would not. See TestStrictPrefixMatchingRule() for more examples.
//
// "PathPrefix" in Kubernetes Gateway API is semantically equivalent to the "Prefix" path type in the
// Kubernetes Ingress API.
func buildPrefixRule(path string) string {
	if path == "/" {
		return `PathPrefix("/")`
	}
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("(Path(%q) || PathPrefix(%q))", path, path+"/")
}
