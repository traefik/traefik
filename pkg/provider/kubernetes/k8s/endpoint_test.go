package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/discovery/v1"
)

func TestEndpointServing(t *testing.T) {
	tests := []struct {
		name     string
		endpoint v1.Endpoint
		want     bool
	}{
		{
			name: "no status",
			endpoint: v1.Endpoint{
				Conditions: v1.EndpointConditions{
					Ready:   nil,
					Serving: nil,
				},
			},
			want: false,
		},
		{
			name: "ready",
			endpoint: v1.Endpoint{
				Conditions: v1.EndpointConditions{
					Ready:   pointer(true),
					Serving: nil,
				},
			},
			want: true,
		},
		{
			name: "not ready",
			endpoint: v1.Endpoint{
				Conditions: v1.EndpointConditions{
					Ready:   pointer(false),
					Serving: nil,
				},
			},
			want: false,
		},
		{
			name: "not ready and serving",
			endpoint: v1.Endpoint{
				Conditions: v1.EndpointConditions{
					Ready:   pointer(false),
					Serving: pointer(true),
				},
			},
			want: true,
		},
		{
			name: "not ready and not serving",
			endpoint: v1.Endpoint{
				Conditions: v1.EndpointConditions{
					Ready:   pointer(false),
					Serving: pointer(false),
				},
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EndpointServing(test.endpoint)
			assert.Equal(t, test.want, got)
		})
	}
}

func pointer[T any](v T) *T { return &v }
