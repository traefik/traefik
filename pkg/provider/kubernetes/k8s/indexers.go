package k8s

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

// EndpointSliceServiceNameIndex is the name of the informer indexer that maps
// EndpointSlices to their owning Service via the kubernetes.io/service-name
// label. The index key has the form "namespace/serviceName".
const EndpointSliceServiceNameIndex = "EndpointSliceServiceName"

// EndpointSliceServiceNameIndexers is the cache.Indexers value to register on
// a Discovery/V1/EndpointSlices SharedIndexInformer so that callers can look
// up slices for a given Service in O(1) instead of doing a namespace-wide
// label-selector scan.
var EndpointSliceServiceNameIndexers = cache.Indexers{
	EndpointSliceServiceNameIndex: endpointSliceServiceNameIndexFunc,
}

func endpointSliceServiceNameIndexFunc(obj any) ([]string, error) {
	es, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		return nil, nil
	}
	svc, ok := es.Labels[discoveryv1.LabelServiceName]
	if !ok || svc == "" {
		return nil, nil
	}
	return []string{es.Namespace + "/" + svc}, nil
}

// EndpointSliceServiceNameIndexKey builds the index key used by
// EndpointSliceServiceNameIndex for a given namespace/serviceName pair.
func EndpointSliceServiceNameIndexKey(namespace, serviceName string) string {
	return namespace + "/" + serviceName
}
