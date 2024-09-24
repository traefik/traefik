package k8s

import (
	"testing"

	v1 "k8s.io/api/discovery/v1"
)

func TestEndpointServing(t *testing.T) {
	valTrue := true
	valFalse := false
	tests := []struct {
		name     string
		endpoint v1.Endpoint
		want     bool
	}{
		{name: "test1.1", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: nil}}, want: true},
		{name: "test1.2", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: &valTrue}}, want: true},
		{name: "test1.3", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: &valFalse}}, want: true},
		{name: "test2.1", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: nil}}, want: false},
		{name: "test2.2", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: &valTrue}}, want: true},
		{name: "test2.3", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: &valFalse}}, want: false},
		{name: "test3.1", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: nil}}, want: false},
		{name: "test3.2", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: &valTrue}}, want: false},
		{name: "test3.3", endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: &valFalse}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndpointServing(tt.endpoint); got != tt.want {
				t.Errorf("EndpointServing() = %v, want %v", got, tt.want)
			}
		})
	}
}
