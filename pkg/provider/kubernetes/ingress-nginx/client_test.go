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

func TestClientDynamicNamespaceDiscovery(t *testing.T) {
	t.Parallel()

	// Start with a namespace that matches the selector.
	existingNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "existing",
			Labels: map[string]string{"watch": "true"},
		},
	}

	kubeClient := kubefake.NewClientset(existingNs)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	// Drain initial events from the namespace informer seeing the existing namespace.
	drainEvents(eventCh, 100*time.Millisecond)

	// Verify only "existing" is watched initially.
	assert.True(t, client.isWatchedNamespace("existing"))
	assert.False(t, client.isWatchedNamespace("new-ns"))

	// Create a new namespace that matches the selector.
	newNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "new-ns",
			Labels: map[string]string{"watch": "true"},
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(t.Context(), newNs, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for the namespace informer to pick up the new namespace.
	waitForCondition(t, 2*time.Second, func() bool {
		return client.isWatchedNamespace("new-ns")
	})

	// Verify we received an event (the ingress informer for the new namespace fires).
	select {
	case <-eventCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected event after namespace was added")
	}
}

func TestClientDynamicNamespaceRemoval(t *testing.T) {
	t.Parallel()

	// Start with a namespace that matches the selector.
	existingNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "removable",
			Labels: map[string]string{"watch": "true"},
		},
	}

	kubeClient := kubefake.NewClientset(existingNs)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	drainEvents(eventCh, 100*time.Millisecond)

	assert.True(t, client.isWatchedNamespace("removable"))

	// Delete the namespace.
	err = kubeClient.CoreV1().Namespaces().Delete(t.Context(), "removable", metav1.DeleteOptions{})
	require.NoError(t, err)

	// Wait for the namespace to be removed from watching.
	waitForCondition(t, 2*time.Second, func() bool {
		return !client.isWatchedNamespace("removable")
	})
}

func TestClientNamespaceLabelUpdateStartsWatching(t *testing.T) {
	t.Parallel()

	// Start with a namespace that does NOT match the selector.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "will-match-later",
			Labels: map[string]string{"watch": "false"},
		},
	}

	kubeClient := kubefake.NewClientset(ns)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	drainEvents(eventCh, 100*time.Millisecond)

	assert.False(t, client.isWatchedNamespace("will-match-later"))

	// Update the namespace labels to match.
	ns.Labels["watch"] = "true"
	ns.ResourceVersion = "2"
	_, err = kubeClient.CoreV1().Namespaces().Update(t.Context(), ns, metav1.UpdateOptions{})
	require.NoError(t, err)

	waitForCondition(t, 2*time.Second, func() bool {
		return client.isWatchedNamespace("will-match-later")
	})
}

func TestClientNamespaceLabelUpdateStopsWatching(t *testing.T) {
	t.Parallel()

	// Start with a namespace that matches the selector.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "will-unmatch",
			Labels: map[string]string{"watch": "true"},
		},
	}

	kubeClient := kubefake.NewClientset(ns)

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	drainEvents(eventCh, 100*time.Millisecond)

	assert.True(t, client.isWatchedNamespace("will-unmatch"))

	// Update the namespace labels to stop matching.
	ns.Labels["watch"] = "false"
	ns.ResourceVersion = "2"
	_, err = kubeClient.CoreV1().Namespaces().Update(t.Context(), ns, metav1.UpdateOptions{})
	require.NoError(t, err)

	waitForCondition(t, 2*time.Second, func() bool {
		return !client.isWatchedNamespace("will-unmatch")
	})
}

func TestClientNonMatchingNamespaceIgnored(t *testing.T) {
	t.Parallel()

	kubeClient := kubefake.NewClientset()

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	drainEvents(eventCh, 100*time.Millisecond)

	// Create a namespace that does NOT match the selector.
	nonMatchingNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "non-matching",
			Labels: map[string]string{"watch": "false"},
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(t.Context(), nonMatchingNs, metav1.CreateOptions{})
	require.NoError(t, err)

	// Give it time to not be added.
	time.Sleep(200 * time.Millisecond)

	assert.False(t, client.isWatchedNamespace("non-matching"))
}

func TestClientIngressDiscoveredInNewNamespace(t *testing.T) {
	t.Parallel()

	kubeClient := kubefake.NewClientset()

	discovery, _ := kubeClient.Discovery().(*discoveryfake.FakeDiscovery)
	discovery.FakedServerVersion = &kversion.Info{GitVersion: "v1.19"}

	client := newClient(kubeClient)

	eventCh, err := client.WatchAll(t.Context(), "", "watch=true")
	require.NoError(t, err)

	drainEvents(eventCh, 100*time.Millisecond)

	// Create a matching namespace.
	newNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "dynamic-ns",
			Labels: map[string]string{"watch": "true"},
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(t.Context(), newNs, metav1.CreateOptions{})
	require.NoError(t, err)

	waitForCondition(t, 2*time.Second, func() bool {
		return client.isWatchedNamespace("dynamic-ns")
	})

	drainEvents(eventCh, 100*time.Millisecond)

	// Create an ingress in the dynamically discovered namespace.
	ing := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "dynamic-ns",
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{Host: "test.example.com"},
			},
		},
	}
	_, err = kubeClient.NetworkingV1().Ingresses("dynamic-ns").Create(t.Context(), ing, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for the event from the ingress creation.
	select {
	case <-eventCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected event after ingress was created in dynamic namespace")
	}

	// Verify the ingress is listed.
	ingresses := client.ListIngresses()
	var found bool
	for _, i := range ingresses {
		if i.Name == "test-ingress" && i.Namespace == "dynamic-ns" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// drainEvents reads all pending events from the channel until no event arrives within the timeout.
func drainEvents(ch <-chan any, timeout time.Duration) {
	for {
		select {
		case <-ch:
		case <-time.After(timeout):
			return
		}
	}
}

// waitForCondition polls the condition function until it returns true or the timeout expires.
func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition not met within timeout")
}
