package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func Test_parseServiceConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		expected    ServiceConfig
	}{
		{
			desc: "service annotations",
			annotations: map[string]string{
				"ingress.kubernetes.io/foo":   "bar",
				"traefik.io/foo":              "bar",
				"traefik.io/service.nativelb": "true",
			},
			expected: ServiceConfig{
				Service: Service{
					NativeLB: ptr.To(true),
				},
			},
		},
		{
			desc:        "empty map",
			annotations: map[string]string{},
			expected:    ServiceConfig{},
		},
		{
			desc:        "nil map",
			annotations: nil,
			expected:    ServiceConfig{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg, err := parseServiceAnnotations(test.annotations)
			require.NoError(t, err)

			assert.Equal(t, test.expected, cfg)
		})
	}
}

func Test_convertAnnotations(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		expected    map[string]string
	}{
		{
			desc: "service annotations",
			annotations: map[string]string{
				"traefik.io/service.nativelb": "true",
			},
			expected: map[string]string{
				"traefik.service.nativelb": "true",
			},
		},
		{
			desc:        "empty map",
			annotations: map[string]string{},
			expected:    nil,
		},
		{
			desc:        "nil map",
			annotations: nil,
			expected:    nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			labels := convertAnnotations(test.annotations)

			assert.Equal(t, test.expected, labels)
		})
	}
}
