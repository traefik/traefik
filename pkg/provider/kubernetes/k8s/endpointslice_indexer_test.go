package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestEndpointSlicesByServiceName(t *testing.T) {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, EndpointSliceServiceNameIndexers)

	matching := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "matching",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "app",
			},
		},
	}
	require.NoError(t, indexer.Add(matching))

	// A service routinely has more than one EndpointSlice, so the lookup must return all of them.
	matchingSecond := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "matching-second",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "app",
			},
		},
	}
	require.NoError(t, indexer.Add(matchingSecond))

	require.NoError(t, indexer.Add(&discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-service",
			Namespace: "default",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "other",
			},
		},
	}))

	require.NoError(t, indexer.Add(&discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-namespace",
			Namespace: "other",
			Labels: map[string]string{
				discoveryv1.LabelServiceName: "app",
			},
		},
	}))

	endpointSlices, err := EndpointSlicesByServiceName(indexer, "default", "app")
	require.NoError(t, err)
	// ByIndex returns objects in map-iteration order, so the result is not ordered.
	require.ElementsMatch(t, []*discoveryv1.EndpointSlice{matching, matchingSecond}, endpointSlices)

	unknown, err := EndpointSlicesByServiceName(indexer, "default", "unknown")
	require.NoError(t, err)
	require.Empty(t, unknown)
}
