package crd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	traefikcrdfake "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

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
	crdClient := traefikcrdfake.NewClientset()

	client := newClientImpl(kubeClient, crdClient)

	stopCh := make(chan struct{})

	eventCh, err := client.WatchAll(nil, stopCh)
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

	_, found, err := client.GetSecret("default", "secret")
	require.NoError(t, err)
	assert.True(t, found)

	_, found, err = client.GetSecret("default", "helm-secret")
	require.NoError(t, err)
	assert.False(t, found)
}

func Test_aggregateWatchedNamespaces(t *testing.T) {
	nsProd1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod-1",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}
	nsProd2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "prod-2",
			Labels: map[string]string{
				"env": "prod",
			},
		},
	}
	nsDev := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dev",
			Labels: map[string]string{
				"env": "dev",
			},
		},
	}

	testCases := []struct {
		desc              string
		namespaces        []string
		namespaceSelector string
		expected          []string
		expectError       bool
	}{
		{
			desc:     "empty selector and nil namespaces",
			expected: nil,
		},
		{
			desc:       "empty selector with static namespaces",
			namespaces: []string{"foo", "bar"},
			expected:   []string{"foo", "bar"},
		},
		{
			desc:              "matching selector with nil namespaces",
			namespaceSelector: "env=prod",
			expected:          []string{"prod-1", "prod-2"},
		},
		{
			desc:              "matching selector with static namespaces and deduplication",
			namespaces:        []string{"default", "prod-1"},
			namespaceSelector: "env=prod",
			expected:          []string{"default", "prod-1", "prod-2"},
		},
		{
			desc:              "selector matching no namespaces",
			namespaceSelector: "env=staging",
			expected:          nil,
		},
		{
			desc:              "invalid label selector",
			namespaceSelector: "invalid selector %%%",
			expectError:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			kubeClient := kubefake.NewClientset(nsProd1, nsProd2, nsDev)
			crdClient := traefikcrdfake.NewClientset()
			client := newClientImpl(kubeClient, crdClient)

			actual, err := client.aggregateWatchedNamespaces(t.Context(), test.namespaces, test.namespaceSelector)
			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
