package crd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	traefikscheme "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/scheme"
	traefikinformers "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/informers/externalversions"
	traefikv1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *clientWrapper) appendV1alpha1IngressRoutes(result []*traefikv1.IngressRoute) []*traefikv1.IngressRoute {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1alpha1().IngressRoutes().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list ingress routes in namespace %s", ns)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 ingress route (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert ingress route in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.IngressRoute))
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1IngressRouteTCPs(result []*traefikv1.IngressRouteTCP) []*traefikv1.IngressRouteTCP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1alpha1().IngressRouteTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tcp ingress routes in namespace %s", ns)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 tcp ingress route (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert tcp ingress route in namespace %s", ns)
				continue
			}

			v1Route := toVersion.(*traefikv1.IngressRouteTCP)

			// Handle deprecated options.
			for routeIdx, route := range ing.Spec.Routes {
				for serviceIdx, service := range route.Services {
					v1Route.Spec.Routes[routeIdx].Services[serviceIdx].TerminationDelay = service.TerminationDelay
				}
			}

			result = append(result, v1Route)
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1IngressRouteUDPs(result []*traefikv1.IngressRouteUDP) []*traefikv1.IngressRouteUDP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1alpha1().IngressRouteUDPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list udp ingress routes in namespace %s", ns)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 udp ingress route (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert udp ingress route in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.IngressRouteUDP))
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1Middlewares(result []*traefikv1.Middleware) []*traefikv1.Middleware {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Traefik().V1alpha1().Middlewares().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list middlewares in namespace %s", ns)
		}

		for _, middleware := range middlewares {
			key := objectKey(middleware.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 middleware (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(middleware, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert middleware in namespace %s", ns)
				continue
			}

			v1Middleware := toVersion.(*traefikv1.Middleware)

			if middleware.Spec.ContentType != nil {
				v1Middleware.Spec.ContentType.AutoDetect = middleware.Spec.ContentType.AutoDetect
			}

			// Handle deprecated options.
			if middleware.Spec.Headers != nil {
				v1Middleware.Spec.Headers.SSLForceHost = middleware.Spec.Headers.SSLForceHost
				v1Middleware.Spec.Headers.SSLRedirect = middleware.Spec.Headers.SSLRedirect
				v1Middleware.Spec.Headers.SSLTemporaryRedirect = middleware.Spec.Headers.SSLTemporaryRedirect
				v1Middleware.Spec.Headers.SSLHost = middleware.Spec.Headers.SSLHost
				v1Middleware.Spec.Headers.FeaturePolicy = middleware.Spec.Headers.FeaturePolicy
			}

			if middleware.Spec.StripPrefix != nil {
				v1Middleware.Spec.StripPrefix.ForceSlash = middleware.Spec.StripPrefix.ForceSlash
			}

			if middleware.Spec.IPWhiteList != nil {
				v1Middleware.Spec.IPWhiteList = middleware.Spec.IPWhiteList
			}

			if middleware.Spec.ForwardAuth != nil && middleware.Spec.ForwardAuth.TLS != nil && middleware.Spec.ForwardAuth.TLS.CAOptional != nil {
				if v1Middleware.Spec.ForwardAuth.TLS != nil {
					v1Middleware.Spec.ForwardAuth.TLS.CAOptional = middleware.Spec.ForwardAuth.TLS.CAOptional
				}
			}

			result = append(result, v1Middleware)
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1MiddlewareTCPs(result []*traefikv1.MiddlewareTCP) []*traefikv1.MiddlewareTCP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Traefik().V1alpha1().MiddlewareTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tcp middlewares in namespace %s", ns)
		}

		for _, middleware := range middlewares {
			key := objectKey(middleware.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 middleware (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(middleware, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert tcp middleware in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.MiddlewareTCP))
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1TraefikServices(result []*traefikv1.TraefikService) []*traefikv1.TraefikService {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		traefikServices, err := factory.Traefik().V1alpha1().TraefikServices().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list Traefik services in namespace %s", ns)
		}

		for _, traefikService := range traefikServices {
			key := objectKey(traefikService.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 Traefik service (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(traefikService, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert Traefik service in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.TraefikService))
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1ServersTransport(result []*traefikv1.ServersTransport) []*traefikv1.ServersTransport {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.Traefik().V1alpha1().ServersTransports().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list servers transports in namespace %s", ns)
		}

		for _, serversTransport := range serversTransports {
			key := objectKey(serversTransport.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 servers transport (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(serversTransport, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert servers transport in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.ServersTransport))
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1TLSOptions(result []*traefikv1.TLSOption) []*traefikv1.TLSOption {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		options, err := factory.Traefik().V1alpha1().TLSOptions().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls options in namespace %s", ns)
		}

		for _, option := range options {
			key := objectKey(option.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 tls option (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(option, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert tls option in namespace %s", ns)
				continue
			}

			v1TLSOption := toVersion.(*traefikv1.TLSOption)

			// Handle deprecated options.
			if option.Spec.PreferServerCipherSuites != nil {
				v1TLSOption.Spec.PreferServerCipherSuites = option.Spec.PreferServerCipherSuites
			}

			result = append(result, v1TLSOption)
		}
	}

	return result
}

func (c *clientWrapper) appendV1alpha1TLSStores(result []*traefikv1.TLSStore) []*traefikv1.TLSStore {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		stores, err := factory.Traefik().V1alpha1().TLSStores().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls stores in namespace %s", ns)
		}

		for _, store := range stores {
			key := objectKey(store.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debug().Msgf("Ignoring traefik.io/v1alpha1 tls store (%s) already listed within traefik.io/v1 API GroupVersion", key)
				continue
			}

			toVersion, err := traefikscheme.Scheme.ConvertToVersion(store, GroupVersioner)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to convert tls store in namespace %s", ns)
				continue
			}

			result = append(result, toVersion.(*traefikv1.TLSStore))
		}
	}

	return result
}

func (c *clientWrapper) getV1alpha1TraefikService(namespace, name string) (*traefikv1.TraefikService, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesCrd[c.lookupNamespace(namespace)].Traefik().V1alpha1().TraefikServices().Lister().TraefikServices(namespace).Get(name)
	exist, err := translateNotFoundError(err)

	if !exist {
		return nil, false, err
	}

	toVersion, err := traefikscheme.Scheme.ConvertToVersion(service, GroupVersioner)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to convert Traefik service in namespace %s", namespace)
	}

	return toVersion.(*traefikv1.TraefikService), exist, err
}

func addV1alpha1Informers(factoryCrd traefikinformers.SharedInformerFactory, eventHandler *k8s.ResourceEventHandler) error {
	_, err := factoryCrd.Traefik().V1alpha1().IngressRoutes().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().Middlewares().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().MiddlewareTCPs().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().IngressRouteTCPs().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().IngressRouteUDPs().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().TLSOptions().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().ServersTransports().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().TLSStores().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	_, err = factoryCrd.Traefik().V1alpha1().TraefikServices().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return err
	}

	return nil
}

func objectKey(meta metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
