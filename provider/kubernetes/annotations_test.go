package kubernetes

import (
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/stretchr/testify/assert"
)

func TestGetAnnotationName(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		name        string
		expected    string
	}{
		{
			desc: "with standard annotation",
			name: annotationKubernetesPreserveHost,
			annotations: map[string]string{
				annotationKubernetesPreserveHost: "true",
			},
			expected: annotationKubernetesPreserveHost,
		},
		{
			desc: "with prefixed annotation",
			name: annotationKubernetesPreserveHost,
			annotations: map[string]string{
				label.Prefix + annotationKubernetesPreserveHost: "true",
			},
			expected: label.Prefix + annotationKubernetesPreserveHost,
		},
		{
			desc: "with label",
			name: annotationKubernetesPreserveHost,
			annotations: map[string]string{
				label.TraefikFrontendPassHostHeader: "true",
			},
			expected: label.TraefikFrontendPassHostHeader,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getAnnotationName(test.annotations, test.name)
			assert.Equal(t, test.expected, actual)
		})
	}
}
