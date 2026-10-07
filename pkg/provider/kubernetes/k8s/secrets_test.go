package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestSecretInformers_Get(t *testing.T) {
	newSecret := func(namespace, name string) *corev1.Secret {
		return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	}

	testCases := []struct {
		desc       string
		namespaces []string
		allowed    map[string][]string
		namespace  string
		name       string
		expectErr  bool
	}{
		{
			desc:       "no restriction",
			namespaces: []string{metav1.NamespaceAll},
			namespace:  "ns1",
			name:       "foo",
		},
		{
			desc:       "allowed secret",
			namespaces: []string{metav1.NamespaceAll},
			allowed:    map[string][]string{"ns1": {"foo"}},
			namespace:  "ns1",
			name:       "foo",
		},
		{
			desc:       "secret not allowed",
			namespaces: []string{metav1.NamespaceAll},
			allowed:    map[string][]string{"ns1": {"foo"}},
			namespace:  "ns1",
			name:       "bar",
			expectErr:  true,
		},
		{
			desc:       "namespace not allowed",
			namespaces: []string{metav1.NamespaceAll},
			allowed:    map[string][]string{"ns1": {"foo"}},
			namespace:  "ns2",
			name:       "foo",
			expectErr:  true,
		},
		{
			desc:       "allowed secret in watched namespace",
			namespaces: []string{"ns1"},
			allowed:    map[string][]string{"ns1": {"foo", "bar"}, "ns2": {"foo"}},
			namespace:  "ns1",
			name:       "bar",
		},
		{
			desc:       "allowed secret in unwatched namespace",
			namespaces: []string{"ns1"},
			allowed:    map[string][]string{"ns1": {"foo"}, "ns2": {"foo"}},
			namespace:  "ns2",
			name:       "foo",
			expectErr:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := kubefake.NewSimpleClientset(newSecret("ns1", "foo"), newSecret("ns1", "bar"), newSecret("ns2", "foo"))

			informers, err := NewSecretInformers(client, time.Minute, test.namespaces, test.allowed, &ResourceEventHandler{Ev: make(chan any, 100)})
			require.NoError(t, err)

			stopCh := make(chan struct{})
			t.Cleanup(func() { close(stopCh) })

			informers.Start(stopCh)
			require.NoError(t, informers.WaitForCacheSync(stopCh))

			lookupNamespace := test.namespaces[0]

			secret, err := informers.Get(lookupNamespace, test.namespace, test.name)
			if test.expectErr {
				require.Error(t, err)
				assert.True(t, kerror.IsNotFound(err))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.name, secret.Name)
		})
	}
}
