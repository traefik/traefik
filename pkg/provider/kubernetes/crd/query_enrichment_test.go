package crd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikcrdfake "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/fake"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestQueryEnrichmentCRDRuntimeAndReferences(t *testing.T) {
	crdClient := traefikcrdfake.NewClientset(
		&traefikv1alpha1.Middleware{
			Name: "metadata", Namespace: "default",
			Spec: traefikv1alpha1.MiddlewareSpec{QueryEnrichment: &traefikv1alpha1.QueryEnrichment{
				RuntimeValues: map[string]traefikv1alpha1.RuntimeValueSource{"reader": {SecretKeyRef: &traefikv1alpha1.RuntimeObjectKeySelector{Name: "lookup", Key: "arbitrary"}}},
				Queries:       []traefikv1alpha1.EnrichmentQuery{{Name: "query", AllowedOrigins: []string{"https://metadata.example"}, HTTP: dynamic.ExternalActionHTTP{URL: "https://metadata.example", Auth: &dynamic.ExternalActionAuth{Type: "bearer", Token: "runtime://reader"}}}},
			}},
		},
		&traefikv1alpha1.Middleware{
			Name: "report", Namespace: "default",
			Spec: traefikv1alpha1.MiddlewareSpec{ExternalAction: &traefikv1alpha1.ExternalAction{Enrichment: []string{"metadata"}}},
		},
	)
	kubeClient := kubefake.NewClientset(&corev1.Secret{Name: "lookup", Namespace: "default", ResourceVersion: "41", Data: map[string][]byte{"arbitrary": []byte("private-lookup-token")}})
	client := newClientImpl(kubeClient, crdClient)
	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })
	_, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)
	conf := (&Provider{}).loadConfigurationFromCRD(t.Context(), client)
	query := conf.HTTP.Middlewares["default-metadata"].QueryEnrichment
	require.NotNil(t, query)
	assert.Equal(t, "private-lookup-token", query.RuntimeValues["reader"].Value)
	assert.Equal(t, []string{"default-metadata"}, conf.HTTP.Middlewares["default-report"].ExternalAction.Enrichment)
	safe := (&Provider{nameBuilder: nameBuilder{safe: true}}).loadConfigurationFromCRD(t.Context(), client)
	require.NotNil(t, safe.HTTP.Middlewares["default_metadata"].QueryEnrichment)
	assert.Equal(t, []string{"default_metadata"}, safe.HTTP.Middlewares["default_report"].ExternalAction.Enrichment)
	encoded, err := json.Marshal(query)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "private-lookup-token")
	for _, reference := range []string{"", "other/metadata", "metadata@file"} {
		_, err := createExternalActionMiddlewareWithRuntime(client, "default", &traefikv1alpha1.ExternalAction{Enrichment: []string{reference}})
		require.Error(t, err)
	}
}

func TestQueryEnrichmentDurationConversion(t *testing.T) {
	var source traefikv1alpha1.QueryEnrichment
	require.NoError(t, json.Unmarshal([]byte(`{"queries":[{"name":"metadata","timeout":"5s","cacheTTL":"30s"}]}`), &source))
	converted, err := createQueryEnrichmentMiddlewareWithRuntime(nil, "default", &source)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, time.Duration(converted.Queries[0].Timeout))
	assert.Equal(t, 30*time.Second, time.Duration(converted.Queries[0].CacheTTL))
	invalid := intstr.FromString("not-a-duration")
	source.Queries[0].Timeout = &invalid
	_, err = createQueryEnrichmentMiddlewareWithRuntime(nil, "default", &source)
	require.Error(t, err)
}
