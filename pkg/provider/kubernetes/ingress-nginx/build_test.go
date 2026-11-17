package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	corev1 "k8s.io/api/core/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func Test_basicAuthUsers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc           string
		authSecretType string
		data           map[string][]byte
		expected       []secretUser
	}{
		{
			desc:           "auth-file users are read from the auth key, without comments and blank lines",
			authSecretType: "auth-file",
			data:           map[string][]byte{"auth": []byte("# comment\nuser1:hash1\n\n user2:hash2 \n")},
			expected: []secretUser{
				{key: "auth", value: "user1:hash1"},
				{key: "auth", value: "user2:hash2"},
			},
		},
		{
			desc:           "auth-map users are sorted by key",
			authSecretType: "auth-map",
			data:           map[string][]byte{"user2": []byte("hash2"), "user1": []byte("hash1"), "user3": []byte("hash3")},
			expected: []secretUser{
				{key: "user1", value: "user1:hash1"},
				{key: "user2", value: "user2:hash2"},
				{key: "user3", value: "user3:hash3"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			users, err := basicAuthUsers(&corev1.Secret{Data: test.data}, test.authSecretType)
			require.NoError(t, err)

			assert.Equal(t, test.expected, users)
		})
	}
}

func TestProvider_resolveBasicAuth_authMap(t *testing.T) {
	t.Parallel()

	kubeClient := kubefake.NewClientset(&corev1.Secret{
		Name: "basic-auth", Namespace: "default",
		Data: map[string][]byte{"user2": []byte("hash2"), "user1": []byte("hash1")},
	})
	client := newClient(kubeClient)

	_, err := client.WatchAll(t.Context(), "", "")
	require.NoError(t, err)

	p := Provider{k8sClient: client}

	basic, digest, urnRefs, err := p.resolveBasicAuth("default", IngressConfig{
		AuthType:       new("basic"),
		AuthSecret:     new("basic-auth"),
		AuthSecretType: new("auth-map"),
	})
	require.NoError(t, err)

	assert.Equal(t, &dynamic.BasicAuth{Users: dynamic.Users{"urn:k8s:secret:basic-auth:user1", "urn:k8s:secret:basic-auth:user2"}}, basic)
	assert.Nil(t, digest)
	assert.Equal(t, map[string]string{"/basicAuth/users/0": "user1:hash1", "/basicAuth/users/1": "user2:hash2"}, urnRefs)
}
