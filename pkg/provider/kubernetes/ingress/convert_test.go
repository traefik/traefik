package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
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

	actual, err := convertSlice[netv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []netv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []netv1.IngressPortStatus{
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
	g := []netv1beta1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []netv1beta1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	actual, err := convertSlice[netv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []netv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []netv1.IngressPortStatus{
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
	g := []netv1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []netv1.IngressPortStatus{
				{
					Port:     123,
					Protocol: "https",
					Error:    ptr("test"),
				},
			},
		},
	}

	actual, err := convertSlice[netv1beta1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := []netv1beta1.IngressLoadBalancerIngress{
		{
			IP:       "132456",
			Hostname: "foo",
			Ports: []netv1beta1.IngressPortStatus{
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

	actual, err := convert[netv1.IngressLoadBalancerIngress](g)
	require.NoError(t, err)

	expected := &netv1.IngressLoadBalancerIngress{
		IP:       "132456",
		Hostname: "foo",
		Ports: []netv1.IngressPortStatus{
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
