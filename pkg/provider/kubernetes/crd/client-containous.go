package crd

import (
	"fmt"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned/scheme"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/informers/externalversions"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *clientWrapper) appendContainousIngressRoutes(result []*v1alpha1.IngressRoute) []*v1alpha1.IngressRoute {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.TraefikContainous().V1alpha1().IngressRoutes().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list ingress routes in namespace %s: %v", ns, err)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 ingress route (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert ingress route in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.IngressRoute))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousIngressRouteTCPs(result []*v1alpha1.IngressRouteTCP) []*v1alpha1.IngressRouteTCP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.TraefikContainous().V1alpha1().IngressRouteTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list tcp ingress routes in namespace %s: %v", ns, err)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 tcp ingress route (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert tcp ingress route in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.IngressRouteTCP))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousIngressRouteUDPs(result []*v1alpha1.IngressRouteUDP) []*v1alpha1.IngressRouteUDP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.TraefikContainous().V1alpha1().IngressRouteUDPs().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list udp ingress routes in namespace %s: %v", ns, err)
		}

		for _, ing := range ings {
			key := objectKey(ing.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 udp ingress route (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(ing, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert udp ingress route in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.IngressRouteUDP))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousMiddlewares(result []*v1alpha1.Middleware) []*v1alpha1.Middleware {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.TraefikContainous().V1alpha1().Middlewares().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list middlewares in namespace %s: %v", ns, err)
		}

		for _, middleware := range middlewares {
			key := objectKey(middleware.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 middleware (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(middleware, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert middleware in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.Middleware))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousMiddlewareTCPs(result []*v1alpha1.MiddlewareTCP) []*v1alpha1.MiddlewareTCP {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.TraefikContainous().V1alpha1().MiddlewareTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list tcp middlewares in namespace %s: %v", ns, err)
		}

		for _, middleware := range middlewares {
			key := objectKey(middleware.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 middleware (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(middleware, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert tcp middleware in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.MiddlewareTCP))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousTraefikServices(result []*v1alpha1.TraefikService) []*v1alpha1.TraefikService {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		traefikServices, err := factory.TraefikContainous().V1alpha1().TraefikServices().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list Traefik services in namespace %s: %v", ns, err)
		}

		for _, traefikService := range traefikServices {
			key := objectKey(traefikService.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 Traefik service (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(traefikService, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert Traefik service in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.TraefikService))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousServersTransport(result []*v1alpha1.ServersTransport) []*v1alpha1.ServersTransport {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.TraefikContainous().V1alpha1().ServersTransports().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list servers transports in namespace %s: %v", ns, err)
		}

		for _, serversTransport := range serversTransports {
			key := objectKey(serversTransport.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 servers transport (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(serversTransport, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert servers transport in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.ServersTransport))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousTLSOptions(result []*v1alpha1.TLSOption) []*v1alpha1.TLSOption {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		options, err := factory.TraefikContainous().V1alpha1().TLSOptions().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list tls options in namespace %s: %v", ns, err)
		}

		for _, option := range options {
			key := objectKey(option.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 tls option (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(option, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert tls option in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.TLSOption))
		}
	}

	return result
}

func (c *clientWrapper) appendContainousTLSStores(result []*v1alpha1.TLSStore) []*v1alpha1.TLSStore {
	listed := map[string]struct{}{}
	for _, obj := range result {
		listed[objectKey(obj.ObjectMeta)] = struct{}{}
	}

	for ns, factory := range c.factoriesCrd {
		stores, err := factory.TraefikContainous().V1alpha1().TLSStores().Lister().List(labels.Everything())
		if err != nil {
			log.Errorf("Failed to list tls stores in namespace %s: %v", ns, err)
		}

		for _, store := range stores {
			key := objectKey(store.ObjectMeta)
			if _, ok := listed[key]; ok {
				log.Debugf("Ignoring traefik.containo.us/v1alpha1 tls store (%s) already listed within traefik.io/v1alpha1 API GroupVersion", key)
				continue
			}

			toVersion, err := scheme.Scheme.ConvertToVersion(store, GroupVersioner)
			if err != nil {
				log.Errorf("Failed to convert tls store in namespace %s: %v", ns, err)
				continue
			}

			result = append(result, toVersion.(*v1alpha1.TLSStore))
		}
	}

	return result
}

func (c *clientWrapper) getContainousTraefikService(namespace, name string) (*v1alpha1.TraefikService, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesCrd[c.lookupNamespace(namespace)].TraefikContainous().V1alpha1().TraefikServices().Lister().TraefikServices(namespace).Get(name)
	exist, err := translateNotFoundError(err)

	if !exist {
		return nil, false, err
	}

	toVersion, err := scheme.Scheme.ConvertToVersion(service, GroupVersioner)
	if err != nil {
		log.Errorf("Failed to convert Traefik service in namespace %s: %v", namespace, err)
	}

	return toVersion.(*v1alpha1.TraefikService), exist, err
}

func addContainousInformers(factoryCrd externalversions.SharedInformerFactory, eventHandler *k8s.ResourceEventHandler) {
	factoryCrd.TraefikContainous().V1alpha1().IngressRoutes().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().Middlewares().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().MiddlewareTCPs().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().IngressRouteTCPs().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().IngressRouteUDPs().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().TLSOptions().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().ServersTransports().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().TLSStores().Informer().AddEventHandler(eventHandler)
	factoryCrd.TraefikContainous().V1alpha1().TraefikServices().Informer().AddEventHandler(eventHandler)
}

func objectKey(meta metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
