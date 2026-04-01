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

func Test_parseGatewayAnnotations(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		want        *StatusAddress
		wantErr     bool
	}{
		{
			desc:        "nil annotations",
			annotations: nil,
			want:        nil,
		},
		{
			desc:        "empty annotations",
			annotations: map[string]string{},
			want:        nil,
		},
		{
			desc: "no statusaddress annotations",
			annotations: map[string]string{
				"traefik.io/service.nativelb": "true",
			},
			want: nil,
		},
		{
			desc: "IP annotation",
			annotations: map[string]string{
				"traefik.io/statusaddress.ip": "10.0.0.1",
			},
			want: &StatusAddress{
				IP: "10.0.0.1",
			},
		},
		{
			desc: "hostname annotation",
			annotations: map[string]string{
				"traefik.io/statusaddress.hostname": "internal.example.com",
			},
			want: &StatusAddress{
				Hostname: "internal.example.com",
			},
		},
		{
			desc: "service annotations",
			annotations: map[string]string{
				"traefik.io/statusaddress.service.name":      "traefik-public",
				"traefik.io/statusaddress.service.namespace": "traefik-system",
			},
			want: &StatusAddress{
				Service: ServiceRef{
					Name:      "traefik-public",
					Namespace: "traefik-system",
				},
			},
		},
		{
			desc: "mixed with unrelated annotations",
			annotations: map[string]string{
				"traefik.io/statusaddress.ip": "10.0.0.1",
				"traefik.io/service.nativelb": "true",
				"kubernetes.io/ingress.class": "traefik",
			},
			want: &StatusAddress{
				IP: "10.0.0.1",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, err := parseGatewayAnnotations(test.annotations)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
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
