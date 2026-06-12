package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestEndpointSlicesByServiceName(t *testing.T) {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, EndpointSliceServiceNameIndexers)

	endpointSlices := []*discoveryv1.EndpointSlice{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-1",
				Namespace: "default",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "foo",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar-1",
				Namespace: "default",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "bar",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-2",
				Namespace: "other",
				Labels: map[string]string{
					discoveryv1.LabelServiceName: "foo",
				},
			},
		},
	}

	for _, endpointSlice := range endpointSlices {
		require.NoError(t, indexer.Add(endpointSlice))
	}

	got, err := EndpointSlicesByServiceName(indexer, "default", "foo")
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "foo-1", got[0].Name)
}
