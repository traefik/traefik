package k8s

import (
	"testing"

	v1 "k8s.io/api/discovery/v1"
)

func TestEndpointServing(t *testing.T) {
	type args struct {
		endpoint v1.Endpoint
	}
	valTrue := true
	valFalse := false
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test1.1", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: nil}}}, want: true},
		{name: "test1.2", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: &valTrue}}}, want: true},
		{name: "test1.3", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valTrue, Serving: &valFalse}}}, want: true},
		{name: "test2.1", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: nil}}}, want: false},
		{name: "test2.2", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: &valTrue}}}, want: true},
		{name: "test2.3", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: &valFalse, Serving: &valFalse}}}, want: false},
		{name: "test3.1", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: nil}}}, want: false},
		{name: "test3.2", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: &valTrue}}}, want: false},
		{name: "test3.3", args: args{endpoint: v1.Endpoint{Conditions: v1.EndpointConditions{Ready: nil, Serving: &valFalse}}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndpointServing(tt.args.endpoint); got != tt.want {
				t.Errorf("EndpointServing() = %v, want %v", got, tt.want)
			}
		})
	}
}
