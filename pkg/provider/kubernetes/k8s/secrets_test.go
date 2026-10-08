package k8s

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func TestNewSecretInformers_multipleNamespaces(t *testing.T) {
	newSecret := func(namespace, name string) *corev1.Secret {
		return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	}

	testCases := []struct {
		desc         string
		namespaces   []string
		allowed      map[string][]string
		expectedKeys []string
		found        []string
		notFound     []string
	}{
		{
			desc:         "each namespace has its own allowed secrets",
			namespaces:   []string{"default", "infra"},
			allowed:      map[string][]string{"default": {"cert1", "cert2"}, "infra": {"wildcard"}},
			expectedKeys: []string{"default/cert1", "default/cert2", "infra/wildcard"},
			found:        []string{"default/cert1", "default/cert2", "infra/wildcard"},
			notFound:     []string{"default/wildcard", "infra/cert1", "infra/cert2", "other/cert1", "other/wildcard"},
		},
		{
			desc:         "same secret name allowed in several namespaces",
			namespaces:   []string{"default", "infra"},
			allowed:      map[string][]string{"default": {"cert1"}, "infra": {"cert1"}},
			expectedKeys: []string{"default/cert1", "infra/cert1"},
			found:        []string{"default/cert1", "infra/cert1"},
			notFound:     []string{"default/cert2", "infra/cert2", "other/cert1"},
		},
		{
			desc:         "allowed secrets of an unwatched namespace are ignored",
			namespaces:   []string{"default", "infra"},
			allowed:      map[string][]string{"default": {"cert1"}, "infra": {"wildcard"}, "other": {"cert1"}},
			expectedKeys: []string{"default/cert1", "infra/wildcard"},
			found:        []string{"default/cert1", "infra/wildcard"},
			notFound:     []string{"other/cert1"},
		},
		{
			desc:         "watched namespace without allowed secrets",
			namespaces:   []string{"default", "infra"},
			allowed:      map[string][]string{"default": {"cert1"}},
			expectedKeys: []string{"default/cert1"},
			found:        []string{"default/cert1"},
			notFound:     []string{"infra/cert1", "infra/wildcard"},
		},
		{
			desc:         "all namespaces watched",
			namespaces:   []string{metav1.NamespaceAll},
			allowed:      map[string][]string{"default": {"cert1", "cert2"}, "infra": {"wildcard"}},
			expectedKeys: []string{"default/cert1", "default/cert2", "infra/wildcard"},
			found:        []string{"default/cert1", "default/cert2", "infra/wildcard"},
			notFound:     []string{"default/wildcard", "infra/cert1", "other/cert1"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var secrets []runtime.Object
			for _, ns := range []string{"default", "infra", "other"} {
				for _, name := range []string{"cert1", "cert2", "wildcard"} {
					secrets = append(secrets, newSecret(ns, name))
				}
			}

			client := kubefake.NewSimpleClientset(secrets...)

			informers, err := NewSecretInformers(client, time.Minute, test.namespaces, test.allowed, &ResourceEventHandler{Ev: make(chan any, 100)})
			require.NoError(t, err)

			assert.Equal(t, test.expectedKeys, slices.Sorted(maps.Keys(informers.factories)))

			stopCh := make(chan struct{})
			t.Cleanup(func() { close(stopCh) })

			informers.Start(stopCh)
			require.NoError(t, informers.WaitForCacheSync(stopCh))

			for _, key := range test.found {
				namespace, name, _ := strings.Cut(key, "/")

				secret, err := informers.Get(namespace, namespace, name)
				require.NoError(t, err, key)
				assert.Equal(t, namespace, secret.Namespace)
				assert.Equal(t, name, secret.Name)
			}

			for _, key := range test.notFound {
				namespace, name, _ := strings.Cut(key, "/")

				_, err := informers.Get(namespace, namespace, name)
				require.Error(t, err, key)
				assert.True(t, kerror.IsNotFound(err), key)
			}
		})
	}
}
