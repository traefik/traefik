package crd

import (
	"context"
	"fmt"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/containous/traefik/v2/pkg/tls"
)

const (
	roundRobinStrategy = "RoundRobin"
	https              = "https"
	http               = "http"
)

func (p *Provider) loadIngressRouteConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, ingressRoute := range client.GetIngressRoutes() {
		ctxRt := log.With(ctx, log.Str("ingress", ingressRoute.Name), log.Str("namespace", ingressRoute.Namespace))
		logger := log.FromContext(ctxRt)

		// TODO keep the name ingressClass?
		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		err := getTLSHTTP(ctx, ingressRoute, client, tlsConfigs)
		if err != nil {
			logger.Errorf("Error configuring TLS: %v", err)
		}

		ingressName := ingressRoute.Name
		if len(ingressName) == 0 {
			ingressName = ingressRoute.GenerateName
		}

		for _, route := range ingressRoute.Spec.Routes {
			if route.Kind != "Rule" {
				logger.Errorf("Unsupported match kind: %s. Only \"Rule\" is supported for now.", route.Kind)
				continue
			}

			if len(route.Match) == 0 {
				logger.Errorf("Empty match rule")
				continue
			}

			if err := checkStringQuoteValidity(route.Match); err != nil {
				logger.Errorf("Invalid syntax for match rule: %s", route.Match)
				continue
			}

			serviceKey, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error(err)
				continue
			}

			var mds []string
			for _, mi := range route.Middlewares {
				if strings.Contains(mi.Name, "@") {
					if len(mi.Namespace) > 0 {
						logger.
							WithField(log.MiddlewareName, mi.Name).
							Warnf("namespace %q is ignored in cross-provider context", mi.Namespace)
					}
					mds = append(mds, mi.Name)
					continue
				}

				ns := mi.Namespace
				if len(ns) == 0 {
					ns = ingressRoute.Namespace
				}
				mds = append(mds, makeID(ns, mi.Name))
			}

			normalized := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))
			serviceName := normalized

			if len(route.Services) > 1 {
				spec := v1alpha1.ServiceSpec{
					Weighted: &v1alpha1.WeightedRoundRobin{
						Services: route.Services,
					},
				}

				errBuild := buildServicesLB(ctx, client, ingressRoute.Namespace, spec, serviceName, conf.Services)
				if errBuild != nil {
					logger.Error(err)
					continue
				}
			} else if len(route.Services) == 1 {
				fullName, serversLB, err := foo(ctx, client, ingressRoute.Namespace, route.Services[0])
				if err != nil {
					logger.Error(err)
					continue
				}

				if serversLB != nil {
					conf.Services[serviceName] = serversLB
				} else {
					serviceName = fullName
				}
			}

			conf.Routers[normalized] = &dynamic.Router{
				Middlewares: mds,
				Priority:    route.Priority,
				EntryPoints: ingressRoute.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     serviceName,
			}

			if ingressRoute.Spec.TLS != nil {
				tlsConf := &dynamic.RouterTLSConfig{
					CertResolver: ingressRoute.Spec.TLS.CertResolver,
					Domains:      ingressRoute.Spec.TLS.Domains,
				}

				if ingressRoute.Spec.TLS.Options != nil && len(ingressRoute.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRoute.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference, (i.e. not a cross-provider reference)
					ns := ingressRoute.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, "@") {
						if len(ns) == 0 {
							ns = ingressRoute.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.
							WithField("TLSoptions", ingressRoute.Spec.TLS.Options.Name).
							Warnf("namespace %q is ignored in cross-provider context", ns)
					}

					tlsConf.Options = tlsOptionsName
				}
				conf.Routers[normalized].TLS = tlsConf
			}
		}
	}

	return conf
}

// configBuilder holds parameters to help recursively build a dynamic config from a CRD.
type configBuilder struct {
	toplevel bool   // whether we're in the first buildServicesLB call.
	parent   string // to help with infinite recursion detection.

	conf   *dynamic.HTTPConfiguration // the configuration we're building.
	client Client
	// seen keeps track of the parent->child relations we've already seen, to detect
	// infinite recursions. it is keyed by "parent:child".
	seen map[string]struct{}
}

func splitSvcNameProvider(name string) (string, string) {
	parts := strings.Split(name, "@")
	provider := parts[len(parts)-1]
	svc := strings.Join(parts[:len(parts)-1], "@")
	return svc, provider
}

func fullServiceName(ctx context.Context, namespace, serviceName string, port int32) string {
	if port != 0 {
		return provider.Normalize(fmt.Sprintf("%s-%s-%d", namespace, serviceName, port))
	}

	if !strings.Contains(serviceName, "@") {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, serviceName))
	}

	name, providerName := splitSvcNameProvider(serviceName)
	if providerName == "kubernetescrd" {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, name))
	}

	// At this point, if namespace == "default", we do not know whether it had been
	// intentionnally set as such, or if we're simply hitting the value set by default.
	// But as it is most likely very much the latter, and we do not want to systematically log spam
	// users in that case, we skip logging whenever the namespace is "default".
	if namespace != "default" {
		log.FromContext(ctx).
			WithField(log.ServiceName, serviceName).
			Warnf("namespace %q is ignored in cross-provider context", namespace)
	}

	return provider.Normalize(name) + "@" + providerName
}

func namespaceOrFallback(lb v1alpha1.LoadBalancerSpec, fallback string) string {
	if lb.Namespace != "" {
		return lb.Namespace
	}
	return fallback
}

func getTLSHTTP(ctx context.Context, ingressRoute *v1alpha1.IngressRoute, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		tlsConf, err := getTLS(k8sClient, ingressRoute.Spec.TLS.SecretName, ingressRoute.Namespace)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = tlsConf
	}

	return nil
}
