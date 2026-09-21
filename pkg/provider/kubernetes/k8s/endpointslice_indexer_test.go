package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func Test_EndpointSlicesByServiceName(t *testing.T) {
	tests := []struct {
		desc           string
		namespace      string
		svcName        string
		endpointSlices []*discoveryv1.EndpointSlice
		expected       []*discoveryv1.EndpointSlice
	}{
		{
			desc:      "returns all endpoint slices for the same service",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-b",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-b",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
		},
		{
			desc:      "returns empty for unknown service",
			namespace: "default",
			svcName:   "unknown",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{},
		},
		{
			desc:      "filters endpoint slices from other namespace",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "other",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{},
		},
		{
			desc:      "filters endpoint slices from other service",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "other"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{},
		},
		{
			desc:      "sorts endpoint slices by name when creation timestamps are equal",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-b",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-a",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "slice-b",
						Namespace: "default",
						Labels:    map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
		},
		{
			desc:      "sorts endpoint slices by creation timestamp",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-a",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-b",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-b",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-a",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
		},
		{
			desc:      "creation timestamp takes precedence over name in sort",
			namespace: "default",
			svcName:   "app",
			endpointSlices: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-a",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-z",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
			expected: []*discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-z",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "slice-a",
						Namespace:         "default",
						CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
						Labels:            map[string]string{discoveryv1.LabelServiceName: "app"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, EndpointSliceByServiceNameIndexers)
			for _, es := range test.endpointSlices {
				require.NoError(t, indexer.Add(es))
			}

			got, err := EndpointSlicesByServiceName(indexer, test.namespace, test.svcName)
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}
