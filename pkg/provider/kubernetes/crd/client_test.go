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

	kubeClient := kubefake.NewSimpleClientset(helmSecret, secret)
	crdClient := traefikcrdfake.NewSimpleClientset()

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
