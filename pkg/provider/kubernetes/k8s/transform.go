package k8s

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

// StripManagedFields drops metadata.managedFields as objects enter the informer cache.
// Traefik never reads them, and they inflate the cache footprint and the cost of copying
// and comparing cached objects, which matters under heavy resource churn.
func StripManagedFields(obj any) (any, error) {
	object, ok := obj.(metav1.Object)
	if !ok {
		return obj, nil
	}

	object.SetManagedFields(nil)
	return obj, nil
}

// StripManagedFieldsAndSetGVK strips the managed fields, and sets back the GVK of the objects registered in scheme,
// which the typed clients drop when decoding, as the metadata sources record it.
func StripManagedFieldsAndSetGVK(scheme *runtime.Scheme) cache.TransformFunc {
	return func(obj any) (any, error) {
		obj, err := StripManagedFields(obj)
		if err != nil {
			return nil, err
		}

		object, ok := obj.(runtime.Object)
		if !ok {
			return obj, nil
		}

		gvks, _, err := scheme.ObjectKinds(object)
		if err != nil {
			return nil, fmt.Errorf("getting object kinds: %w", err)
		}

		object.GetObjectKind().SetGroupVersionKind(gvks[0])

		return object, nil
	}
}
