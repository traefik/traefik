package k8s

import (
	"fmt"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

const EndpointSliceServiceNameIndex = "endpointSliceServiceName"

var EndpointSliceServiceNameIndexers = cache.Indexers{
	EndpointSliceServiceNameIndex: endpointSliceServiceNameIndexFunc,
}

func EndpointSlicesByServiceName(indexer cache.Indexer, namespace, serviceName string) ([]*discoveryv1.EndpointSlice, error) {
	objects, err := indexer.ByIndex(EndpointSliceServiceNameIndex, endpointSliceServiceNameIndexKey(namespace, serviceName))
	if err != nil {
		return nil, fmt.Errorf("listing endpoint slices by service name index: %w", err)
	}

	endpointSlices := make([]*discoveryv1.EndpointSlice, 0, len(objects))
	for _, object := range objects {
		endpointSlice, ok := object.(*discoveryv1.EndpointSlice)
		if !ok {
			return nil, fmt.Errorf("unexpected endpoint slice object type %T", object)
		}

		endpointSlices = append(endpointSlices, endpointSlice)
	}

	return endpointSlices, nil
}

func endpointSliceServiceNameIndexFunc(obj any) ([]string, error) {
	endpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		return nil, fmt.Errorf("unexpected endpoint slice object type %T", obj)
	}

	serviceName := endpointSlice.Labels[discoveryv1.LabelServiceName]
	if serviceName == "" {
		return nil, nil
	}

	return []string{endpointSliceServiceNameIndexKey(endpointSlice.Namespace, serviceName)}, nil
}

func endpointSliceServiceNameIndexKey(namespace, serviceName string) string {
	return namespace + "/" + serviceName
}
