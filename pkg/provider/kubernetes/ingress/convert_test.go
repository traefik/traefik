package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
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
					Error:    pointer("test"),
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
					Error:    pointer("test"),
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
				Error:    pointer("test"),
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
				Error:    pointer("test"),
			},
		},
	}

	assert.Equal(t, expected, actual)
}
