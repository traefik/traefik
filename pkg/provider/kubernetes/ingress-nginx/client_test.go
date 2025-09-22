package ingressnginx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kversion "k8s.io/apimachinery/pkg/version"
	discoveryfake "k8s.io/client-go/discovery/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestIsLoadBalancerIngressEquals(t *testing.T) {
	testCases := []struct {
		desc          string
		aSlice        []netv1.IngressLoadBalancerIngress
		bSlice        []netv1.IngressLoadBalancerIngress
		expectedEqual bool
	}{
		{
			desc:          "both slices are empty",
			expectedEqual: true,
		},
		{
			desc: "not the same length",
			bSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
			},
			expectedEqual: false,
		},
		{
			desc: "same ordered content",
			aSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
			},
			bSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
			},
			expectedEqual: true,
		},
		{
			desc: "same unordered content",
			aSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
				{IP: "192.168.1.2", Hostname: "traefik2"},
			},
			bSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.2", Hostname: "traefik2"},
				{IP: "192.168.1.1", Hostname: "traefik"},
			},
			expectedEqual: true,
		},
		{
			desc: "different ordered content",
			aSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
				{IP: "192.168.1.2", Hostname: "traefik2"},
			},
			bSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
				{IP: "192.168.1.2", Hostname: "traefik"},
			},
			expectedEqual: false,
		},
		{
			desc: "different unordered content",
			aSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.1", Hostname: "traefik"},
				{IP: "192.168.1.2", Hostname: "traefik2"},
			},
			bSlice: []netv1.IngressLoadBalancerIngress{
				{IP: "192.168.1.2", Hostname: "traefik3"},
				{IP: "192.168.1.1", Hostname: "traefik"},
			},
			expectedEqual: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotEqual := isLoadBalancerIngressEquals(test.aSlice, test.bSlice)
			assert.Equal(t, test.expectedEqual, gotEqual)
		})
	}
}

func TestClientIgnoresHelmOwnedSecrets(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "secret",
		},
	}
	helmSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "helm-secret",
			Labels: map[string]string{
				"owner": "helm",
			},
		},
	}

	kubeClient := kubefake.NewClientset(helmSecret, secret)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{
		GitVersion: "v1.19",
	}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "")
	require.NoError(t, err)

	select {
	case event := <-eventCh:
		secret, ok := event.(*corev1.Secret)
		require.True(t, ok)

		assert.NotEqual(t, "helm-secret", secret.Name)
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "expected to receive event for secret")
	}

	select {
	case <-eventCh:
		assert.Fail(t, "received more than one event")
	case <-time.After(50 * time.Millisecond):
	}

	_, err = client.GetSecret("default", "secret")
	require.NoError(t, err)

	_, err = client.GetSecret("default", "helm-secret")
	assert.True(t, kerror.IsNotFound(err))
}

func TestClientIgnoresEmptyEndpointSliceUpdates(t *testing.T) {
	emptyEndpointSlice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "empty-endpointslice",
			Namespace:       "test",
			ResourceVersion: "1244",
			Annotations: map[string]string{
				"test-annotation": "_",
			},
		},
	}

	samplePortName := "testing"
	samplePortNumber := int32(1337)
	samplePortProtocol := corev1.ProtocolTCP
	sampleAddressReady := true
	filledEndpointSlice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "filled-endpointslice",
			Namespace:       "test",
			ResourceVersion: "1234",
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{{
			Addresses: []string{"10.13.37.1"},
			Conditions: discoveryv1.EndpointConditions{
				Ready: &sampleAddressReady,
			},
		}},
		Ports: []discoveryv1.EndpointPort{{
			Name:     &samplePortName,
			Port:     &samplePortNumber,
			Protocol: &samplePortProtocol,
		}},
	}

	kubeClient := kubefake.NewClientset(emptyEndpointSlice, filledEndpointSlice)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{
		GitVersion: "v1.19",
	}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "")
	require.NoError(t, err)

	select {
	case event := <-eventCh:
		ep, ok := event.(*discoveryv1.EndpointSlice)
		require.True(t, ok)

		assert.True(t, ep.Name == "empty-endpointslice" || ep.Name == "filled-endpointslice")
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "expected to receive event for endpointslices")
	}

	emptyEndpointSlice, err = kubeClient.DiscoveryV1().EndpointSlices("test").Get(t.Context(), "empty-endpointslice", metav1.GetOptions{})
	assert.NoError(t, err)

	// Update endpoint annotation and resource version (apparently not done by fake client itself)
	// to show an update that should not trigger an update event on our eventCh.
	// This reflects the behavior of kubernetes controllers which use endpoint annotations for leader election.
	emptyEndpointSlice.Annotations["test-annotation"] = "___"
	emptyEndpointSlice.ResourceVersion = "1245"
	_, err = kubeClient.DiscoveryV1().EndpointSlices("test").Update(t.Context(), emptyEndpointSlice, metav1.UpdateOptions{})
	require.NoError(t, err)

	select {
	case event := <-eventCh:
		ep, ok := event.(*discoveryv1.EndpointSlice)
		require.True(t, ok)

		assert.Fail(t, "didn't expect to receive event for empty endpointslice update", ep.Name)
	case <-time.After(50 * time.Millisecond):
	}

	filledEndpointSlice, err = kubeClient.DiscoveryV1().EndpointSlices("test").Get(t.Context(), "filled-endpointslice", metav1.GetOptions{})
	assert.NoError(t, err)

	filledEndpointSlice.Endpoints[0].Addresses[0] = "10.13.37.2"
	filledEndpointSlice.ResourceVersion = "1235"
	_, err = kubeClient.DiscoveryV1().EndpointSlices("test").Update(t.Context(), filledEndpointSlice, metav1.UpdateOptions{})
	require.NoError(t, err)

	select {
	case event := <-eventCh:
		ep, ok := event.(*discoveryv1.EndpointSlice)
		require.True(t, ok)

		assert.Equal(t, "filled-endpointslice", ep.Name)
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "expected to receive event for filled endpointslice")
	}

	select {
	case <-eventCh:
		assert.Fail(t, "received more than one event")
	case <-time.After(50 * time.Millisecond):
	}

	newPortNumber := int32(42)
	filledEndpointSlice.Ports[0].Port = &newPortNumber
	filledEndpointSlice.ResourceVersion = "1236"
	_, err = kubeClient.DiscoveryV1().EndpointSlices("test").Update(t.Context(), filledEndpointSlice, metav1.UpdateOptions{})
	require.NoError(t, err)

	select {
	case event := <-eventCh:
		ep, ok := event.(*discoveryv1.EndpointSlice)
		require.True(t, ok)

		assert.Equal(t, "filled-endpointslice", ep.Name)
	case <-time.After(50 * time.Millisecond):
		assert.Fail(t, "expected to receive event for filled endpointslice")
	}

	select {
	case <-eventCh:
		assert.Fail(t, "received more than one event")
	case <-time.After(50 * time.Millisecond):
	}
}
