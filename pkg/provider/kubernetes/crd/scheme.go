package crd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	traefikscheme "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/scheme"
	traefikv1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	kschema "k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersioner k8sruntime.GroupVersioner

func init() {
	GroupVersioner = k8sruntime.NewMultiGroupVersioner(
		traefikv1.SchemeGroupVersion,
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.IngressRoute{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.IngressRouteTCP{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.IngressRouteUDP{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.Middleware{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.MiddlewareTCP{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.TLSOption{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.TLSStore{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.ServersTransport{}.Kind},
		kschema.GroupKind{Group: traefikv1alpha1.GroupName, Kind: traefikv1alpha1.TraefikService{}.Kind},
	)

	convert := map[interface{}]interface{}{}
	convert[&traefikv1alpha1.IngressRoute{}] = &traefikv1.IngressRoute{}
	convert[&traefikv1alpha1.IngressRouteTCP{}] = &traefikv1.IngressRouteTCP{}
	convert[&traefikv1alpha1.IngressRouteUDP{}] = &traefikv1.IngressRouteUDP{}
	convert[&traefikv1alpha1.Middleware{}] = &traefikv1.Middleware{}
	convert[&traefikv1alpha1.MiddlewareTCP{}] = &traefikv1.MiddlewareTCP{}
	convert[&traefikv1alpha1.TLSOption{}] = &traefikv1.TLSOption{}
	convert[&traefikv1alpha1.TLSStore{}] = &traefikv1.TLSStore{}
	convert[&traefikv1alpha1.ServersTransport{}] = &traefikv1.ServersTransport{}
	convert[&traefikv1alpha1.TraefikService{}] = &traefikv1.TraefikService{}

	for interfaceA, interfaceB := range convert {
		err := traefikscheme.Scheme.AddConversionFunc(interfaceA, interfaceB, func(a, b interface{}, scope conversion.Scope) error {
			unstruct, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(a)
			if err != nil {
				return fmt.Errorf("failed to unstruct interface: %w", err)
			}

			u := &unstructured.Unstructured{Object: unstruct}
			u.SetGroupVersionKind(traefikv1.SchemeGroupVersion.WithKind(u.GetKind()))

			if err = k8sruntime.DefaultUnstructuredConverter.FromUnstructured(u.Object, b); err != nil {
				return fmt.Errorf("failed to convert interface: %w", err)
			}

			return nil
		})
		if err != nil {
			log.Error().Msg("Failed to add conversion func.")
		}
	}
}
