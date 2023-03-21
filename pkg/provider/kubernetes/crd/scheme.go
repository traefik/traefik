package crd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/scheme"
	containousv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var GroupVersioner k8sruntime.GroupVersioner

func init() {
	GroupVersioner = k8sruntime.NewMultiGroupVersioner(
		v1alpha1.SchemeGroupVersion,
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.IngressRoute{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.IngressRouteTCP{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.IngressRouteUDP{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.Middleware{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.MiddlewareTCP{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.TLSOption{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.TLSStore{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.ServersTransport{}.Kind},
		schema.GroupKind{Group: containousv1alpha1.GroupName, Kind: containousv1alpha1.TraefikService{}.Kind},
	)

	convert := map[interface{}]interface{}{}
	convert[&containousv1alpha1.IngressRoute{}] = &v1alpha1.IngressRoute{}
	convert[&containousv1alpha1.IngressRouteTCP{}] = &v1alpha1.IngressRouteTCP{}
	convert[&containousv1alpha1.IngressRouteUDP{}] = &v1alpha1.IngressRouteUDP{}
	convert[&containousv1alpha1.Middleware{}] = &v1alpha1.Middleware{}
	convert[&containousv1alpha1.MiddlewareTCP{}] = &v1alpha1.MiddlewareTCP{}
	convert[&containousv1alpha1.TLSOption{}] = &v1alpha1.TLSOption{}
	convert[&containousv1alpha1.TLSStore{}] = &v1alpha1.TLSStore{}
	convert[&containousv1alpha1.ServersTransport{}] = &v1alpha1.ServersTransport{}
	convert[&containousv1alpha1.TraefikService{}] = &v1alpha1.TraefikService{}

	for interfaceA, interfaceB := range convert {
		err := scheme.Scheme.AddConversionFunc(interfaceA, interfaceB, func(a, b interface{}, scope conversion.Scope) error {
			unstruct, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(a)
			if err != nil {
				return fmt.Errorf("failed to unstruct interface: %w", err)
			}

			u := &unstructured.Unstructured{Object: unstruct}
			u.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind(u.GetKind()))

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
