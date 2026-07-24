package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestLoadCertificatesContinuesAfterMissingSecret(t *testing.T) {
	t.Parallel()

	const namespace = "default"

	validFirst := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "valid-first", Namespace: namespace},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("first-certificate"),
			corev1.TLSPrivateKeyKey: []byte("first-key"),
		},
	}
	validLast := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "valid-last", Namespace: namespace},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("last-certificate"),
			corev1.TLSPrivateKeyKey: []byte("last-key"),
		},
	}

	kubeClient := kubefake.NewClientset(validFirst, validLast)
	client := newClient(kubeClient)
	_, err := client.WatchAll(t.Context(), "", "")
	require.NoError(t, err)

	provider := Provider{k8sClient: client}
	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "testing", Namespace: namespace},
		Spec: netv1.IngressSpec{
			TLS: []netv1.IngressTLS{
				{SecretName: validFirst.Name},
				{SecretName: "missing"},
				{SecretName: validLast.Name},
			},
		},
	}

	certs := make(map[string]certPair)
	loaded := make(map[string]bool)
	err = provider.loadCertificates(t.Context(), ingress, certs, loaded)

	require.ErrorContains(t, err, "default/missing")
	assert.Equal(t, map[string]certPair{
		"default/valid-first": {Cert: "first-certificate", Key: "first-key"},
		"default/valid-last":  {Cert: "last-certificate", Key: "last-key"},
	}, certs)
	assert.Equal(t, map[string]bool{
		"default/valid-first": true,
		"default/valid-last":  true,
	}, loaded)
}
