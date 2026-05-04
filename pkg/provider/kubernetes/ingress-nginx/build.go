package ingressnginx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

type ingressEntry struct {
	*netv1.Ingress

	config IngressConfig
}

type pathEntry struct {
	netv1.HTTPIngressPath

	config IngressConfig
}

type namedServersTransport struct {
	*dynamic.ServersTransport

	name string
}

// certBlocks holds the raw TLS material extracted from a Kubernetes Secret.
type certBlocks struct {
	ca   []byte
	cert []byte
	key  []byte
}

// build reads all Ingress resources visible to the client and produces a model.
//
//nolint:funlen // multi-phase ingress processing kept inline for readability
func (p *Provider) build(ctx context.Context, ingressClasses []*netv1.IngressClass) *model {
	mc := &model{
		Backends: make(map[string]*backend),
		Servers:  make(map[string]*server),
		Certs:    make(map[string]string),
	}

	// Builder-local cache of TLS options resolved per ingress. Each Location
	// that needs an option carries a pointer to the cached entry; the translator
	// registers each unique option once in conf.TLS.Options.
	tlsOptionCache := make(map[string]*tls.Options)

	allHosts := make(map[string]bool)
	hostsWithUseRegex := make(map[string]bool)
	claimedAliases := make(map[string]string)

	// Provider-level default backend.
	if p.defaultBackendServiceNamespace != "" && p.defaultBackendServiceName != "" {
		backend, err := p.resolveBackend(
			p.defaultBackendServiceNamespace,
			netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: p.defaultBackendServiceName}},
			IngressConfig{},
		)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Cannot resolve default backend service")
		} else {
			mc.Backends[backend.Name] = backend
			mc.DefaultBackend = backend
		}
	}

	// First pass: collect all regular and canary ingresses.
	var (
		regularIngresses []ingressEntry
		canaryIngresses  []ingressEntry
	)

	// ingressPaths tracks existing paths for canary matching.
	// key: namespace/host+path+pathType → (backend service, IngressConfig)
	ingressPaths := make(map[string]pathEntry)

	serverSnippets := make(map[string]string) // host → first server-snippet seen

	// Sort ingresses by creation timestamp (ascending). Ties are broken by
	// descending namespace/name lexicographic order, matching ingress-nginx
	// behavior. This ordering is what makes server-alias conflict resolution
	// deterministic: the first ingress in this order to claim an alias owns it.
	ingresses := p.k8sClient.ListIngresses()
	sort.SliceStable(ingresses, func(a, b int) bool {
		ta, tb := ingresses[a].CreationTimestamp, ingresses[b].CreationTimestamp
		if ta.Equal(&tb) {
			ia := ingresses[a].Namespace + "/" + ingresses[a].Name
			ib := ingresses[b].Namespace + "/" + ingresses[b].Name
			return ia > ib
		}
		return ta.Before(&tb)
	})

	for _, ingress := range ingresses {
		if !p.shouldProcess(ingress, ingressClasses) {
			continue
		}

		logger := log.Ctx(ctx).With().
			Str("ingress", ingress.Name).
			Str("namespace", ingress.Namespace).
			Logger()

		cfg := parseIngressConfig(ingress)
		entry := ingressEntry{Ingress: ingress, config: cfg}

		if err := p.validateIngress(entry.Ingress, entry.config); err != nil {
			logger.Error().Err(err).Msg("Invalid Ingress, skipping")
			continue
		}

		if ptr.Deref(cfg.Canary, false) {
			canaryIngresses = append(canaryIngresses, entry)
			continue
		}

		for _, alias := range ptr.Deref(cfg.ServerAlias, nil) {
			a := strings.ToLower(alias)
			if _, exists := claimedAliases[a]; !exists {
				claimedAliases[a] = ingress.Namespace + "/" + ingress.Name
			}
		}

		for _, rule := range ingress.Spec.Rules {
			allHosts[rule.Host] = true

			if ptr.Deref(cfg.UseRegex, false) || ptr.Deref(cfg.RewriteTarget, "") != "" {
				hostsWithUseRegex[rule.Host] = true
			}

			if srvSnippet := ptr.Deref(cfg.ServerSnippet, ""); srvSnippet != "" {
				if serverSnippets[rule.Host] == "" {
					serverSnippets[rule.Host] = srvSnippet
				}
			}

			if rule.HTTP != nil {
				for _, pa := range rule.HTTP.Paths {
					if pa.Backend.Service == nil {
						continue
					}
					key := ingressPathKey(ingress.Namespace, rule.Host, pa)
					ingressPaths[key] = pathEntry{HTTPIngressPath: pa, config: cfg}
				}
			}
		}

		regularIngresses = append(regularIngresses, entry)
	}

	// Second pass: resolve canary backends.
	// canaryMap: key = namespace/svcName/port → CanaryConfig
	canaryConfigs := make(map[string]canaryConfig) // primary backend key → canary info
	alreadyMatchedPaths := make(map[string]struct{})

	for _, canaryIngress := range canaryIngresses {
		logger := log.With().
			Str("ingress", canaryIngress.Name).
			Str("namespace", canaryIngress.Namespace).
			Logger()

		for _, rule := range canaryIngress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, pa := range rule.HTTP.Paths {
				pathKey := ingressPathKey(canaryIngress.Namespace, rule.Host, pa)
				prodIngressPath, ok := ingressPaths[pathKey]
				if !ok {
					logger.Error().Msgf("Canary ingress does not match any Ingress rule for host=%s path=%s", rule.Host, pa.Path)
					continue
				}

				if _, ok := alreadyMatchedPaths[pathKey]; ok {
					logger.Error().Msgf("A canary ingress is already matching Ingress rule host=%s path=%s", rule.Host, pa.Path)
					continue
				}

				if pa.Backend.Service == nil {
					continue
				}

				// Skip if the canary path uses the same service as the production path — this
				// is a "non-matching" canary ingress where the matching path doesn't actually
				// route to a different backend.
				prodSvc := prodIngressPath.Backend.Service
				if pa.Backend.Service.Name == prodSvc.Name &&
					portString(pa.Backend.Service.Port) == portString(prodSvc.Port) {
					continue
				}

				// The canary backend's addresses are checked using the original ingress config.
				endpoints, err := p.getBackendEndpoints(canaryIngress.Namespace, pa.Backend, prodIngressPath.config)
				if err != nil || len(endpoints) == 0 {
					continue
				}

				alreadyMatchedPaths[pathKey] = struct{}{}

				weightTotal := max(ptr.Deref(canaryIngress.config.CanaryWeightTotal, 0), 100)
				weight := min(max(ptr.Deref(canaryIngress.config.CanaryWeight, 0), 0), weightTotal)

				// Resolve canary backend.
				canaryBackendName := provider.Normalize(canaryIngress.Namespace + "-" + pa.Backend.Service.Name + "-" + portString(pa.Backend.Service.Port))
				if _, exists := mc.Backends[canaryBackendName]; !exists {
					mc.Backends[canaryBackendName] = &backend{
						Name:      canaryBackendName,
						Namespace: canaryIngress.Namespace,
						Endpoints: endpoints,
					}
				}

				// Primary backend key for the original ingress path.
				primaryKey := canaryPrimaryKey(canaryIngress.Namespace, *prodIngressPath.Backend.Service)
				canaryConfigs[primaryKey] = canaryConfig{
					BackendName:   canaryBackendName,
					Weight:        weight,
					WeightTotal:   weightTotal,
					Header:        ptr.Deref(canaryIngress.config.CanaryHeader, ""),
					HeaderValue:   ptr.Deref(canaryIngress.config.CanaryHeaderValue, ""),
					HeaderPattern: ptr.Deref(canaryIngress.config.CanaryHeaderPattern, ""),
					Cookie:        ptr.Deref(canaryIngress.config.CanaryCookie, ""),
				}
			}
		}
	}

	// Third pass: build Servers and Locations from regular ingresses.
	loadedSecrets := make(map[string]bool) // cross-ingress secret-load dedup

	for _, ing := range regularIngresses {
		logger := log.Ctx(ctx).With().
			Str("ingress", ing.Name).
			Str("namespace", ing.Namespace).
			Logger()
		ctxIng := logger.WithContext(ctx)

		// ssl-passthrough: handled per-rule. serversTransport is not needed for passthrough.
		if ptr.Deref(ing.config.SSLPassthrough, false) {
			for _, rule := range ing.Spec.Rules {
				if rule.Host == "" {
					logger.Error().Msg("Cannot process ssl-passthrough: rule has no host")
					continue
				}

				var ingBackend *netv1.IngressBackend
				if rule.HTTP != nil {
					for _, pa := range rule.HTTP.Paths {
						if pa.Path == "/" {
							ingBackend = &pa.Backend
							break
						}
					}
				} else if ing.Spec.DefaultBackend != nil {
					ingBackend = ing.Spec.DefaultBackend
				}

				if ingBackend == nil || ingBackend.Service == nil {
					logger.Error().Msgf("No backend found for ssl-passthrough on host %q", rule.Host)
					continue
				}

				endpoints, err := p.getBackendEndpoints(ing.Namespace, *ingBackend, ing.config)
				if err != nil {
					logger.Error().Err(err).Msgf("Cannot resolve passthrough backend for host %q", rule.Host)
					continue
				}

				ptBackendName := provider.Normalize(ing.Namespace + "-" + ingBackend.Service.Name + "-" + portString(ingBackend.Service.Port))
				if _, exists := mc.Backends[ptBackendName]; !exists {
					mc.Backends[ptBackendName] = &backend{
						Name:      ptBackendName,
						Namespace: ing.Namespace,
						Endpoints: endpoints,
					}
				}

				routerKey := strings.TrimPrefix(provider.Normalize(ing.Namespace+"-"+ing.Name+"-"+rule.Host), "-")
				mc.PassthroughBackends = append(mc.PassthroughBackends, &sslPassthroughBackend{
					BackendName: ptBackendName,
					Hostname:    rule.Host,
					RouterKey:   routerKey,
				})
			}
			continue
		}

		// Normal ingress: build serversTransport, TLS options, TLS certs, and Locations.
		nst, err := p.buildServersTransport(ctxIng, ing.Namespace, ing.Name, ing.config)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot build serversTransport, skipping ingress")
			continue
		}
		// Resolve TLS option (auth-tls-secret).
		var tlsOptionName string
		var tlsOption *tls.Options
		if ing.config.AuthTLSSecret != nil {
			optName := provider.Normalize(ing.Namespace + "-" + ing.Name + "-" + *ing.config.AuthTLSSecret)
			if cached, exists := tlsOptionCache[optName]; exists {
				tlsOptionName = optName
				tlsOption = cached
			} else {
				tlsOpt, err := p.buildClientAuthTLSOption(ing.Namespace, ing.config)
				if err != nil {
					logger.Error().Err(err).Msg("Cannot build client auth TLS option")
				} else {
					tlsOptionName = optName
					tlsOption = tlsOpt
					tlsOptionCache[optName] = tlsOption
				}
			}
		}

		// Load TLS certificates into the shared mc.Certs map (keyed by cert PEM,
		// which naturally deduplicates certs reused across ingresses). When loading
		// fails, hasTLS still signals that the ingress has a TLS section so the
		// translator can fall back to the default cert.
		hasTLS := len(ing.Spec.TLS) > 0
		if hasTLS {
			if err := p.loadCertificates(ctxIng, ing.Ingress, mc.Certs, loadedSecrets); err != nil {
				logger.Warn().Err(err).Msg("Error loading TLS certificates, defaulting to default certificate")
			}
		}

		for ri, rule := range ing.Spec.Rules {
			srv := getOrCreateServer(mc.Servers, rule.Host)

			if rule.HTTP == nil {
				continue
			}
			for pi, pa := range rule.HTTP.Paths {
				if pa.Backend.Service == nil {
					continue
				}

				// Resolve primary backend. Missing/unavailable services are tolerated:
				// the location is still emitted with an empty backend so the router
				// (and its middlewares) exist, mirroring ingress-nginx's 503-on-no-servers
				// behavior.
				endpoints, err := p.getBackendEndpoints(ing.Namespace, pa.Backend, ing.config)
				if err != nil {
					logger.Warn().
						Str("service", pa.Backend.Service.Name).
						Err(err).
						Msg("Cannot resolve backend addresses, emitting empty backend")
					endpoints = nil
				}

				backendName := provider.Normalize(ing.Namespace + "-" + ing.Name + "-" + pa.Backend.Service.Name + "-" + portString(pa.Backend.Service.Port))
				if _, exists := mc.Backends[backendName]; !exists {
					mc.Backends[backendName] = &backend{
						Name:      backendName,
						Namespace: ing.Namespace,
						Endpoints: endpoints,
					}
				}

				loc := &location{
					Path:                 pa.Path,
					PathType:             pa.PathType,
					BackendName:          backendName,
					ServersTransportName: nst.name,
					ServersTransport:     nst.ServersTransport,
					LocationIndex:        pi,
					RuleIndex:            ri,
					Config:               ing.config,
					TLSOptionName:        tlsOptionName,
					TLSOption:            tlsOption,
					HasTLS:               hasTLS,
					ServerSnippet:        serverSnippets[rule.Host],
					Namespace:            ing.Namespace,
					IngressName:          ing.Name,
					ServiceName:          pa.Backend.Service.Name,
					ServicePort:          portString(pa.Backend.Service.Port),
				}

				// Attach canary config if one exists for this primary backend.
				primaryKey := canaryPrimaryKey(ing.Namespace, *pa.Backend.Service)
				if cc, ok := canaryConfigs[primaryKey]; ok {
					loc.Canary = &cc
				}

				// Pre-resolve basic auth.
				if ing.config.AuthType != nil {
					basic, digest, err := p.resolveBasicAuth(ing.Namespace, ing.config)
					if err != nil {
						logger.Error().
							Err(err).
							Str("ingress", fmt.Sprintf("%s/%s rule-%d path-%d", ing.Namespace, ing.Name, ri, pi)).
							Msg("Cannot resolve auth secret, skipping auth middleware")
					} else {
						loc.BasicAuth = basic
						loc.DigestAuth = digest
					}
				}

				// Pre-resolve custom headers ConfigMap.
				if ing.config.CustomHeaders != nil {
					headers, err := p.resolveCustomHeaders(ing.Namespace, *ing.config.CustomHeaders)
					if err != nil {
						logger.Error().Err(err).Msg("Cannot resolve custom-headers ConfigMap")
						loc.Error = true
					} else {
						loc.ResolvedCustomHeaders = headers
					}
				}

				// Pre-resolve custom-http-errors default backend.
				customErrors := ptr.Deref(ing.config.CustomHTTPErrors, nil)
				if len(customErrors) > 0 {
					if errBackendName := p.resolveHTTPErrorBackend(ing.Namespace, ing.config); errBackendName != "" {
						loc.ResolvedHTTPErrorBackendName = errBackendName
						if _, exists := mc.Backends[errBackendName]; !exists {
							// Build the error backend using defaultBackend annotation or provider default.
							defaultSvcName := ptr.Deref(ing.config.DefaultBackend, "")
							if defaultSvcName != "" {
								endpoints, err := p.getBackendEndpoints(ing.Namespace,
									netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: defaultSvcName}},
									ing.config)
								if err == nil {
									mc.Backends[errBackendName] = &backend{
										Name:      errBackendName,
										Namespace: ing.Namespace,
										Endpoints: endpoints,
									}
								}
							}
						}
					}
				}

				loc.Aliases = resolveAliases(ctx, loc, allHosts, claimedAliases)
				loc.UseRegex = hostsWithUseRegex[rule.Host]

				// Build all middleware configurations so the translator only registers them.
				endpointCount := 0
				if backend, ok := mc.Backends[backendName]; ok {
					endpointCount = len(backend.Endpoints)
				}
				p.buildMiddlewares(ctx, loc, rule.Host, allHosts, endpointCount)

				srv.Locations = append(srv.Locations, loc)
			}
		}

		// Ingress-level spec.defaultBackend: emit one host-only catch-all location
		// per distinct host in the ingress rules. Mirrors ingress-nginx which uses
		// the ingress defaultBackend as the fallback location for each server block.
		// When the ingress has no rules, register the backend as the global catch-all
		// (default-backend / default-backend-tls routers) so it still serves traffic.
		if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
			db := ing.Spec.DefaultBackend
			endpoints, err := p.getBackendEndpoints(ing.Namespace, *db, ing.config)
			if err != nil {
				logger.Warn().
					Str("service", db.Service.Name).
					Err(err).
					Msg("Cannot resolve ingress default backend, emitting empty backend")
				endpoints = nil
			}

			if len(ing.Spec.Rules) == 0 {
				if mc.DefaultBackend == nil {
					bk := &backend{
						Name:        defaultBackendName,
						Namespace:   ing.Namespace,
						ServiceName: db.Service.Name,
						Endpoints:   endpoints,
					}
					mc.Backends[defaultBackendName] = bk
					mc.DefaultBackend = bk

					loc := &location{
						BackendName:          defaultBackendName,
						ServersTransportName: nst.name,
						ServersTransport:     nst.ServersTransport,
						Config:               ing.config,
						Namespace:            ing.Namespace,
						IngressName:          ing.Name,
						ServiceName:          db.Service.Name,
						ServicePort:          portString(db.Service.Port),
					}
					p.buildMiddlewares(ctx, loc, "", allHosts, len(endpoints))
					mc.DefaultBackendLocation = loc
				}
				continue
			}

			ingDefaultBackendName := provider.Normalize(ing.Namespace + "-" + ing.Name + "-default-backend")
			if _, exists := mc.Backends[ingDefaultBackendName]; !exists {
				mc.Backends[ingDefaultBackendName] = &backend{
					Name:      ingDefaultBackendName,
					Namespace: ing.Namespace,
					Endpoints: endpoints,
				}
			}

			seenHosts := map[string]struct{}{}
			for _, rule := range ing.Spec.Rules {
				if _, ok := seenHosts[rule.Host]; ok {
					continue
				}
				seenHosts[rule.Host] = struct{}{}

				srv := getOrCreateServer(mc.Servers, rule.Host)

				loc := &location{
					Path:                    "",
					BackendName:             ingDefaultBackendName,
					ServersTransportName:    nst.name,
					ServersTransport:        nst.ServersTransport,
					Config:                  ing.config,
					TLSOptionName:           tlsOptionName,
					TLSOption:               tlsOption,
					HasTLS:                  hasTLS,
					ServerSnippet:           serverSnippets[rule.Host],
					Namespace:               ing.Namespace,
					IngressName:             ing.Name,
					ServiceName:             db.Service.Name,
					ServicePort:             portString(db.Service.Port),
					IsIngressDefaultBackend: true,
				}

				loc.Aliases = resolveAliases(ctx, loc, allHosts, claimedAliases)
				loc.UseRegex = hostsWithUseRegex[rule.Host]

				endpointCount := 0
				if backend, ok := mc.Backends[ingDefaultBackendName]; ok {
					endpointCount = len(backend.Endpoints)
				}
				p.buildMiddlewares(ctx, loc, rule.Host, allHosts, endpointCount)

				srv.Locations = append(srv.Locations, loc)
			}
		}
	}

	return mc
}

func (p *Provider) buildServersTransport(ctx context.Context, namespace, name string, cfg IngressConfig) (namedServersTransport, error) {
	nst := namedServersTransport{
		name: provider.Normalize(namespace + "-" + name),
		ServersTransport: &dynamic.ServersTransport{
			ForwardingTimeouts: &dynamic.ForwardingTimeouts{
				DialTimeout:     ptypes.Duration(time.Duration(ptr.Deref(cfg.ProxyConnectTimeout, p.ProxyConnectTimeout)) * time.Second),
				ReadTimeout:     ptypes.Duration(time.Duration(ptr.Deref(cfg.ProxyReadTimeout, p.ProxyReadTimeout)) * time.Second),
				WriteTimeout:    ptypes.Duration(time.Duration(ptr.Deref(cfg.ProxySendTimeout, p.ProxySendTimeout)) * time.Second),
				IdleConnTimeout: ptypes.Duration(time.Duration(p.UpstreamKeepaliveTimeout) * time.Second),
			},
		},
	}

	if proxyHTTPVersion := ptr.Deref(cfg.ProxyHTTPVersion, ""); proxyHTTPVersion != "" {
		switch proxyHTTPVersion {
		case "1.1":
			nst.DisableHTTP2 = true
		case "1.0":
			log.Ctx(ctx).Warn().Msg("Value '1.0' is not supported with proxy-http-version, ignoring")
		default:
			log.Ctx(ctx).Warn().Msgf("Invalid proxy-http-version value: %q, ignoring", proxyHTTPVersion)
		}
	}

	if scheme := parseBackendProtocol(ptr.Deref(cfg.BackendProtocol, "HTTP")); scheme != "https" {
		return nst, nil
	}

	nst.ServerName = ptr.Deref(cfg.ProxySSLName, ptr.Deref(cfg.ProxySSLServerName, ""))
	nst.InsecureSkipVerify = strings.ToLower(ptr.Deref(cfg.ProxySSLVerify, "off")) != "on"

	if sslSecret := ptr.Deref(cfg.ProxySSLSecret, ""); sslSecret != "" {
		parts := strings.Split(sslSecret, "/")
		if len(parts) != 2 {
			return namedServersTransport{}, fmt.Errorf("malformed proxy SSL secret: %s", sslSecret)
		}

		secretNamespace, secretName := parts[0], parts[1]
		if !p.AllowCrossNamespaceResources && secretNamespace != namespace {
			return namedServersTransport{}, fmt.Errorf("cross-namespace proxy SSL secret not allowed: %s/%s", secretNamespace, secretName)
		}

		blocks, err := p.certificateBlocks(secretNamespace, secretName)
		if err != nil {
			return namedServersTransport{}, fmt.Errorf("getting certificate blocks: %w", err)
		}

		if len(blocks.ca) > 0 {
			nst.RootCAs = []types.FileOrContent{types.FileOrContent(blocks.ca)}
		}
		if len(blocks.cert) > 0 && len(blocks.key) > 0 {
			nst.Certificates = []tls.Certificate{{
				CertFile: types.FileOrContent(blocks.cert),
				KeyFile:  types.FileOrContent(blocks.key),
			}}
		}
	}

	return nst, nil
}

func (p *Provider) getBackendEndpoints(namespace string, backend netv1.IngressBackend, cfg IngressConfig) ([]endpoint, error) {
	service, err := p.k8sClient.GetService(namespace, backend.Service.Name)
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}

	if p.DisableSvcExternalName && service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, errors.New("externalName services not allowed")
	}

	servicePort, match := getServicePort(service, backend)
	if !match {
		return nil, errors.New("service port not found")
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return []endpoint{{Address: net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(servicePort.TargetPort.IntValue()))}}, nil
	}

	if ptr.Deref(cfg.ServiceUpstream, false) {
		return []endpoint{{Address: net.JoinHostPort(service.Spec.ClusterIP, strconv.Itoa(int(servicePort.Port)))}}, nil
	}

	endpoints, err := p.getEndpointsFromEndpointSlices(namespace, backend.Service.Name, servicePort.Name)
	if err != nil {
		return nil, fmt.Errorf("getting backend endpoints: %w", err)
	}

	// Fall back to default-backend annotation if no endpoints.
	defaultBackend := ptr.Deref(cfg.DefaultBackend, "")
	if defaultBackend == "" || defaultBackend == backend.Service.Name || len(endpoints) > 0 {
		return endpoints, nil
	}

	fallbackSvc, err := p.k8sClient.GetService(namespace, defaultBackend)
	if err != nil {
		return nil, fmt.Errorf("getting fallback service: %w", err)
	}

	if p.DisableSvcExternalName && fallbackSvc.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, errors.New("externalName services not allowed")
	}

	servicePort, match = getServicePort(fallbackSvc, netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: defaultBackend}})
	if !match {
		return nil, errors.New("fallback service port not found")
	}

	return p.getEndpointsFromEndpointSlices(namespace, defaultBackend, servicePort.Name)
}

func (p *Provider) getEndpointsFromEndpointSlices(namespace, name, portName string) ([]endpoint, error) {
	endpointSlices, err := p.k8sClient.GetEndpointSlicesForService(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("getting endpointslices: %w", err)
	}

	var endpoints []endpoint
	seen := map[string]struct{}{}

	for _, es := range endpointSlices {
		var port int32
		for _, p := range es.Ports {
			if p.Name != nil && *p.Name == portName {
				if p.Port != nil {
					port = *p.Port
				}
				break
			}
		}
		if port == 0 {
			continue
		}

		for _, ep := range es.Endpoints {
			if !ptr.Deref(ep.Conditions.Serving, true) {
				continue
			}

			for _, addr := range ep.Addresses {
				if _, ok := seen[addr]; ok {
					continue
				}
				seen[addr] = struct{}{}
				endpoints = append(endpoints, endpoint{
					Address: net.JoinHostPort(addr, strconv.Itoa(int(port))),
					Fenced:  ptr.Deref(ep.Conditions.Terminating, false),
				})
			}
		}
	}

	return endpoints, nil
}

func (p *Provider) resolveBackend(namespace string, ingBackend netv1.IngressBackend, cfg IngressConfig) (*backend, error) {
	endpoints, err := p.getBackendEndpoints(namespace, ingBackend, cfg)
	if err != nil {
		return nil, err
	}

	name := provider.Normalize(namespace + "-" + ingBackend.Service.Name + "-" + portString(ingBackend.Service.Port))
	return &backend{
		Name:        name,
		Namespace:   namespace,
		ServiceName: ingBackend.Service.Name,
		Endpoints:   endpoints,
	}, nil
}

func (p *Provider) certificateBlocks(namespace, name string) (*certBlocks, error) {
	secret, err := p.k8sClient.GetSecret(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("fetching secret %s/%s: %w", namespace, name, err)
	}

	certBytes, hasCert := secret.Data[corev1.TLSCertKey]
	keyBytes, hasKey := secret.Data[corev1.TLSPrivateKeyKey]
	caBytes, hasCA := secret.Data[corev1.ServiceAccountRootCAKey]

	if !hasCert && !hasKey && !hasCA {
		return nil, errors.New("secret does not contain a keypair or CA certificate")
	}

	var blocks certBlocks
	if hasCA {
		if len(caBytes) == 0 {
			return nil, errors.New("secret contains an empty CA certificate")
		}
		blocks.ca = caBytes
	}

	if hasKey && hasCert {
		if len(certBytes) == 0 {
			return nil, errors.New("secret contains an empty certificate")
		}
		if len(keyBytes) == 0 {
			return nil, errors.New("secret contains an empty key")
		}
		blocks.cert = certBytes
		blocks.key = keyBytes
	}

	return &blocks, nil
}

// loadCertificates loads TLS certificates for an ingress into mcCerts (keyed by cert PEM).
// The loaded set is shared across ingresses to avoid re-reading the same secret multiple times.
func (p *Provider) loadCertificates(ctx context.Context, ing *netv1.Ingress, mcCerts map[string]string, loaded map[string]bool) error {
	for _, t := range ing.Spec.TLS {
		if t.SecretName == "" {
			log.Ctx(ctx).Debug().Msg("Skipping TLS section: no secret name")
			continue
		}

		secretKey := ing.Namespace + "-" + t.SecretName
		if loaded[secretKey] {
			continue
		}
		loaded[secretKey] = true

		blocks, err := p.certificateBlocks(ing.Namespace, t.SecretName)
		if err != nil {
			return fmt.Errorf("getting certificate blocks: %w", err)
		}

		if blocks.cert == nil || blocks.key == nil {
			return fmt.Errorf("no keypair found in secret %s/%s", ing.Namespace, t.SecretName)
		}

		mcCerts[string(blocks.cert)] = string(blocks.key)
	}

	return nil
}

func (p *Provider) buildClientAuthTLSOption(ingressNamespace string, cfg IngressConfig) (*tls.Options, error) {
	if cfg.AuthTLSSecret == nil {
		return nil, errors.New("auth-tls-secret is nil")
	}

	parts := strings.SplitN(*cfg.AuthTLSSecret, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, errors.New("auth-tls-secret must be in namespace/name format")
	}

	secretNamespace, secretName := parts[0], parts[1]
	if !p.AllowCrossNamespaceResources && secretNamespace != ingressNamespace {
		return nil, fmt.Errorf("cross-namespace auth-tls-secret not allowed: %s/%s", secretNamespace, secretName)
	}

	blocks, err := p.certificateBlocks(secretNamespace, secretName)
	if err != nil {
		return nil, fmt.Errorf("reading client certificate: %w", err)
	}

	if blocks.ca == nil {
		return nil, errors.New("secret does not contain a CA certificate")
	}

	opt := &tls.Options{}
	opt.SetDefaults()
	opt.ClientAuth = tls.ClientAuth{
		CAFiles:        []types.FileOrContent{types.FileOrContent(blocks.ca)},
		ClientAuthType: clientAuthTypeFromString(cfg.AuthTLSVerifyClient),
	}
	return opt, nil
}

func (p *Provider) resolveBasicAuth(namespace string, cfg IngressConfig) (*dynamic.BasicAuth, *dynamic.DigestAuth, error) {
	if cfg.AuthType == nil {
		return nil, nil, nil
	}

	authType := *cfg.AuthType
	if authType != "basic" && authType != "digest" {
		return nil, nil, fmt.Errorf("invalid auth-type %q", authType)
	}

	authSecret := ptr.Deref(cfg.AuthSecret, "")
	if authSecret == "" {
		return nil, nil, errors.New("auth-secret must not be empty")
	}

	parts := strings.Split(authSecret, "/")
	if len(parts) > 2 {
		return nil, nil, fmt.Errorf("invalid auth secret format %q", authSecret)
	}

	secretName := parts[0]
	secretNamespace := namespace
	if len(parts) == 2 {
		secretNamespace = parts[0]
		secretName = parts[1]
	}

	if !p.AllowCrossNamespaceResources && secretNamespace != namespace {
		return nil, nil, fmt.Errorf("cross-namespace auth secret not allowed: %s/%s", secretNamespace, secretName)
	}

	secret, err := p.k8sClient.GetSecret(secretNamespace, secretName)
	if err != nil {
		return nil, nil, fmt.Errorf("getting secret %s: %w", authSecret, err)
	}

	authSecretType := ptr.Deref(cfg.AuthSecretType, "auth-file")
	if authSecretType != "auth-file" && authSecretType != "auth-map" {
		return nil, nil, fmt.Errorf("invalid auth-secret-type %q", authSecretType)
	}

	users, err := basicAuthUsers(secret, authSecretType)
	if err != nil {
		return nil, nil, fmt.Errorf("getting users from secret: %w", err)
	}

	realm := ptr.Deref(cfg.AuthRealm, "")
	if authType == "digest" {
		return nil, &dynamic.DigestAuth{Users: dynamic.Users(users), Realm: realm}, nil
	}
	return &dynamic.BasicAuth{Users: dynamic.Users(users), Realm: realm}, nil, nil
}

func (p *Provider) resolveCustomHeaders(namespace, customHeadersRef string) (map[string]string, error) {
	parts := strings.Split(customHeadersRef, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid custom-headers configmap reference %q", customHeadersRef)
	}

	configMapNamespace := parts[0]
	configMapName := parts[1]

	if !p.AllowCrossNamespaceResources && configMapNamespace != namespace {
		return nil, fmt.Errorf("cross-namespace custom-headers is not allowed: configMap %s/%s is not from ingress namespace %q", configMapNamespace, configMapName, namespace)
	}

	configMap, err := p.k8sClient.GetConfigMap(configMapNamespace, configMapName)
	if err != nil {
		return nil, fmt.Errorf("getting configmap %s: %w", customHeadersRef, err)
	}

	headers := make(map[string]string)
	for key, value := range configMap.Data {
		if !slices.Contains(p.allowedHeaders, key) {
			return nil, fmt.Errorf("header %q is not in GlobalAllowedResponseHeaders", key)
		}
		headers[key] = value
	}

	return headers, nil
}

// resolveHTTPErrorBackend returns the metamodel backend key for a per-ingress
// default-backend annotation. It returns "" for the provider-level default backend
// case (handled separately in buildCustomHTTPErrors) and when no backend is available.
func (p *Provider) resolveHTTPErrorBackend(namespace string, cfg IngressConfig) string {
	// Per-ingress annotation only — produces a real backend key stored in mc.Backends.
	if defaultBackend := ptr.Deref(cfg.DefaultBackend, ""); defaultBackend != "" {
		return fmt.Sprintf("default-backend-%s-%s", namespace, defaultBackend)
	}

	// Provider-level default backend is handled by buildCustomHTTPErrors directly
	// (via p.defaultBackendServiceName / defaultBackendName).
	return ""
}

func (p *Provider) shouldProcess(ing *netv1.Ingress, ingressClasses []*netv1.IngressClass) bool {
	if len(ingressClasses) > 0 && ing.Spec.IngressClassName != nil {
		return slices.ContainsFunc(ingressClasses, func(ic *netv1.IngressClass) bool {
			return *ing.Spec.IngressClassName == ic.ObjectMeta.Name
		})
	}

	if class, ok := ing.Annotations[annotationIngressClass]; ok {
		return class == p.IngressClass
	}

	return p.WatchIngressWithoutClass
}

func (p *Provider) validateIngress(ing *netv1.Ingress, cfg IngressConfig) error {
	if !p.AllowSnippetAnnotations &&
		(ptr.Deref(cfg.ServerSnippet, "") != "" ||
			ptr.Deref(cfg.ConfigurationSnippet, "") != "" ||
			ptr.Deref(cfg.AuthSnippet, "") != "") {
		return errors.New("snippet annotations are not allowed when allowSnippetAnnotations is disabled")
	}

	if p.StrictValidatePathType {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, pa := range rule.HTTP.Paths {
				if len(pa.Path) > 0 {
					pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
					if pathType != netv1.PathTypeImplementationSpecific && !strictPathTypeRegexp.MatchString(pa.Path) {
						return fmt.Errorf("regex characters not allowed for pathType %s when strictValidatePathType is enabled", pathType)
					}
				}
			}
		}
	}

	return nil
}

func getOrCreateServer(m map[string]*server, hostname string) *server {
	if srv, ok := m[hostname]; ok {
		return srv
	}
	srv := &server{Hostname: hostname}
	m[hostname] = srv
	return srv
}

func getServicePort(service *corev1.Service, backend netv1.IngressBackend) (corev1.ServicePort, bool) {
	for _, p := range service.Spec.Ports {
		if (backend.Service.Port.Number == 0 && backend.Service.Port.Name == "") ||
			(backend.Service.Port.Number == p.Port || (backend.Service.Port.Name == p.Name && len(p.Name) > 0)) {
			return p, true
		}
	}

	// For ExternalName services, the backend port may not be declared in the
	// service spec. Synthesize a ServicePort whose TargetPort echoes the backend
	// port value (named ports parse to 0, matching ingress-nginx behavior).
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return corev1.ServicePort{TargetPort: intstr.Parse(portString(backend.Service.Port))}, true
	}

	return corev1.ServicePort{}, false
}

func portString(port netv1.ServiceBackendPort) string {
	if port.Name == "" {
		return strconv.Itoa(int(port.Number))
	}
	return port.Name
}

func ingressPathKey(namespace, host string, pa netv1.HTTPIngressPath) string {
	pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
	return namespace + "/" + host + pa.Path + "/" + string(pathType)
}

func canaryPrimaryKey(namespace string, backend netv1.IngressServiceBackend) string {
	return namespace + "/" + backend.Name + "/" + portString(backend.Port)
}

// clientAuthTypeFromString maps an ingress-nginx auth-tls-verify-client value to the corresponding ClientAuthType.
// Default is "on" (RequireAndVerifyClientCert) when verifyClient is nil.
func clientAuthTypeFromString(verifyClient *string) string {
	if verifyClient == nil {
		return tls.RequireAndVerifyClientCert
	}
	switch *verifyClient {
	// off means that client certificate is not requested and no verification will be passed.
	case "off":
		return tls.NoClientCert
	// optional means that the client certificate is requested, but not required.
	// If the certificate is present, it needs to be verified.
	case "optional":
		return tls.VerifyClientCertIfGiven
	// optional_no_ca means that the client certificate is requested, but does not require it to be signed by a trusted CA certificate.
	case "optional_no_ca":
		return tls.RequestClientCert
	default:
		return tls.RequireAndVerifyClientCert
	}
}

func basicAuthUsers(secret *corev1.Secret, authSecretType string) ([]string, error) {
	var users []string
	if authSecretType == "auth-map" {
		if len(secret.Data) == 0 {
			return nil, fmt.Errorf("secret %s/%s contains no user credentials", secret.Namespace, secret.Name)
		}
		for user, pass := range secret.Data {
			users = append(users, user+":"+string(pass))
		}
		return users, nil
	}

	authFileContent, ok := secret.Data["auth"]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s missing key 'auth'", secret.Namespace, secret.Name)
	}

	for rawLine := range strings.SplitSeq(string(authFileContent), "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" && !strings.HasPrefix(line, "#") {
			users = append(users, line)
		}
	}

	return users, nil
}

// resolveAliases returns the server-alias hostnames that are valid additions
// to the location's host rule. Aliases that collide with another host or were
// claimed by a different ingress are dropped (with a debug log).
func resolveAliases(ctx context.Context, loc *location, allHosts map[string]bool, claimedAliases map[string]string) []string {
	if loc.Config.ServerAlias == nil {
		return nil
	}

	ingressKey := loc.Namespace + "/" + loc.IngressName
	var aliases []string
	for _, alias := range *loc.Config.ServerAlias {
		a := strings.ToLower(alias)
		if _, ok := allHosts[a]; ok {
			log.Ctx(ctx).Debug().Str("alias", alias).Msg("Skipping server-alias already defined as a host")
			continue
		}
		if owner, ok := claimedAliases[a]; ok && owner != ingressKey {
			log.Ctx(ctx).Debug().
				Str("alias", alias).
				Str("ingress", ingressKey).
				Msgf("Skipping server-alias because it is already claimed by %s", owner)
			continue
		}
		aliases = append(aliases, alias)
	}
	return aliases
}
