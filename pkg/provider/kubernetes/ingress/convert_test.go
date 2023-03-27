package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

func Test_convertSlice_corev1_to_networkingv1(t *testing.T) {
	g := []corev1.LoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []corev1.PortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	actual, err := convertSlice[networkingv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []networkingv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []networkingv1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_convertSlice_networkingv1beta1_to_networkingv1(t *testing.T) {
	g := []networkingv1beta1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []networkingv1beta1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	actual, err := convertSlice[networkingv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []networkingv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []networkingv1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_convertSlice_networkingv1_to_networkingv1beta1(t *testing.T) {
	g := []networkingv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []networkingv1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	actual, err := convertSlice[networkingv1beta1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []networkingv1beta1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []networkingv1beta1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func Test_convert(t *testing.T) {
	g := &corev1.LoadBalancerIngress{
		IP:       "132456",
		Hostname: "foo",
		Ports: []corev1.PortStatus{
			{
				Port:     123,
				Protocol: "https",
				Error:    ptr("test"),
			},
		},
	}

	actual, err := convert[networkingv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := &networkingv1.IngressLoadBalancerIngress{
		IP:       "132456",
		Hostname: "foo",
		Ports: []networkingv1.IngressPortStatus{
			{
				Port:     123,
				Protocol: "https",
				Error:    ptr("test"),
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func ptr[T any](v T) *T {
	return &v
}
