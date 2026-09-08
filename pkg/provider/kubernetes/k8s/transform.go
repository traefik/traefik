package k8s

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
