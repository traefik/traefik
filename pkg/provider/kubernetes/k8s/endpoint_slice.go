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

func EndpointSliceServiceNameIndexKey(namespace, serviceName string) string {
	return fmt.Sprintf("%s/%s", namespace, serviceName)
}

func EndpointSlicesByServiceName(indexer cache.Indexer, namespace, serviceName string) ([]*discoveryv1.EndpointSlice, error) {
	objects, err := indexer.ByIndex(EndpointSliceServiceNameIndex, EndpointSliceServiceNameIndexKey(namespace, serviceName))
	if err != nil {
		return nil, err
	}

	endpointSlices := make([]*discoveryv1.EndpointSlice, 0, len(objects))
	for _, object := range objects {
		endpointSlice, ok := object.(*discoveryv1.EndpointSlice)
		if !ok {
			continue
		}
		endpointSlices = append(endpointSlices, endpointSlice)
	}

	return endpointSlices, nil
}

func endpointSliceServiceNameIndexFunc(obj any) ([]string, error) {
	endpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		return nil, nil
	}

	serviceName := endpointSlice.Labels[discoveryv1.LabelServiceName]
	if serviceName == "" {
		return nil, nil
	}

	return []string{EndpointSliceServiceNameIndexKey(endpointSlice.Namespace, serviceName)}, nil
}
