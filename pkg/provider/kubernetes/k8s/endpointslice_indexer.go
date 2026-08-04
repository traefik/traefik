package k8s

import (
	"fmt"
	"slices"
	"strings"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/tools/cache"
)

// endpointSliceByServiceNameIndexName is the name of the Kubernetes indexer indexing endpoint slices by service name.
const endpointSliceByServiceNameIndexName = "endpointSliceByServiceName"

// EndpointSliceByServiceNameIndexers is the indexers for indexing endpoint slices by service name.
// The index key is in the format "<namespace>/<serviceName>".
var EndpointSliceByServiceNameIndexers = cache.Indexers{
	endpointSliceByServiceNameIndexName: func(obj any) ([]string, error) {
		endpointSlice, ok := obj.(*discoveryv1.EndpointSlice)
		if !ok {
			return nil, fmt.Errorf("unexpected endpoint slice object type %T", obj)
		}

		serviceName := endpointSlice.Labels[discoveryv1.LabelServiceName]

		// In case the service name is empty nothing needs to be indexed.
		if serviceName == "" {
			return nil, nil
		}

		objectName := cache.NewObjectName(endpointSlice.Namespace, serviceName)
		return []string{objectName.String()}, nil
	},
}

// EndpointSlicesByServiceName returns the EndpointSlices for the given service name in the given namespace from the indexer.
// The returned endpoint slices are sorted by creation time and name to ensure a consistent order.
func EndpointSlicesByServiceName(indexer cache.Indexer, namespace, serviceName string) ([]*discoveryv1.EndpointSlice, error) {
	objectName := cache.NewObjectName(namespace, serviceName)
	objects, err := indexer.ByIndex(endpointSliceByServiceNameIndexName, objectName.String())
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

	// As Kubernetes indexer does not guarantee a consistent order of the returned objects,
	// we sort them by creation time then name to keep backend IP assignment consistent across
	// configuration rebuilds.
	slices.SortStableFunc(endpointSlices, func(a, b *discoveryv1.EndpointSlice) int {
		cmpTime := a.CreationTimestamp.Time.Compare(b.CreationTimestamp.Time)
		if cmpTime == 0 {
			return strings.Compare(a.Name, b.Name)
		}
		return cmpTime
	})

	return endpointSlices, nil
}
