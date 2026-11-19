package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/provider/acme"
	"github.com/traefik/traefik/v3/pkg/provider/aggregator"
	"github.com/traefik/traefik/v3/pkg/provider/docker"
	"github.com/traefik/traefik/v3/pkg/provider/file"
	httpprovider "github.com/traefik/traefik/v3/pkg/provider/http"
	"github.com/traefik/traefik/v3/pkg/safe"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

func TestInitialRequiredProviders(t *testing.T) {
	config := &static.Configuration{Providers: &static.Providers{File: &file.Provider{}, HTTP: &httpprovider.Provider{}, Docker: &docker.Provider{}, Swarm: &docker.SwarmProvider{}}}
	require.ElementsMatch(t, []string{"internal", "file", "http", "docker", "swarm", "le.acme", "dns.acme"}, initialRequiredProviders(config, []*acme.Provider{{ResolverName: "le"}, {ResolverName: "dns"}}))
	require.Equal(t, []string{"internal"}, initialRequiredProviders(&static.Configuration{Providers: &static.Providers{}}, nil))
}

func TestInvalidACMEResolverFailsInitialization(t *testing.T) {
	pool := safe.NewPool(t.Context())
	defer pool.Stop()
	config := &static.Configuration{CertificatesResolvers: map[string]static.CertificateResolver{"invalid": {ACME: &acme.Configuration{}}}}
	resolvers, err := initACMEProvider(config, aggregator.NewProviderAggregator(static.Providers{}), traefiktls.NewManager(nil), nil, nil, pool)
	require.ErrorContains(t, err, "initializing ACME resolver invalid")
	require.Empty(t, resolvers)
}
