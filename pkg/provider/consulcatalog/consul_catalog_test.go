package consulcatalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestProviderInitDatacenters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		datacenter  string
		datacenters []string
		expected    []string
		expectedErr string
	}{
		{
			desc:     "default Consul datacenter",
			expected: []string{""},
		},
		{
			desc:       "legacy singular datacenter",
			datacenter: "dc1",
			expected:   []string{"dc1"},
		},
		{
			desc:        "multiple datacenters",
			datacenters: []string{"dc1", "dc2"},
			expected:    []string{"dc1", "dc2"},
		},
		{
			desc:        "singular and plural datacenters",
			datacenter:  "dc1",
			datacenters: []string{"dc2"},
			expectedErr: "datacenter and datacenters are mutually exclusive",
		},
		{
			desc:        "empty datacenter",
			datacenters: []string{"dc1", ""},
			expectedErr: "datacenters cannot contain an empty value",
		},
		{
			desc:        "duplicate datacenter",
			datacenters: []string{"dc1", "dc1"},
			expectedErr: `datacenters contains duplicate value "dc1"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var config Configuration
			config.SetDefaults()
			config.Endpoint.DataCenter = test.datacenter
			config.Endpoint.DataCenters = test.datacenters

			provider := Provider{Configuration: config}
			err := provider.Init()
			if test.expectedErr != "" {
				require.EqualError(t, err, test.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, provider.datacenters())
		})
	}
}

func TestAggregateCatalogUpdatesRetainsLastSuccessfulDatacenterSnapshot(t *testing.T) {
	t.Parallel()

	var config Configuration
	config.SetDefaults()

	provider := Provider{Configuration: config}
	require.NoError(t, provider.Init())

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	updates := make(chan catalogUpdate)
	configurationChan := make(chan dynamic.Message)
	go provider.aggregateCatalogUpdates(ctx, []string{"dc1", "dc2"}, updates, configurationChan)

	updates <- catalogUpdate{
		datacenter: "dc1",
		snapshot:   catalogSnapshot{data: []itemData{catalogTestItem("dc1", "127.0.0.1")}},
	}
	require.Len(t, catalogServers(t, receiveCatalogMessage(t, configurationChan)), 1)

	updates <- catalogUpdate{
		datacenter: "dc2",
		snapshot:   catalogSnapshot{data: []itemData{catalogTestItem("dc2", "127.0.0.2")}},
	}
	require.Len(t, catalogServers(t, receiveCatalogMessage(t, configurationChan)), 2)

	// A successful refresh from dc1 must not discard the last successful dc2 snapshot.
	updates <- catalogUpdate{
		datacenter: "dc1",
		snapshot:   catalogSnapshot{data: []itemData{catalogTestItem("dc1", "127.0.0.3")}},
	}
	assert.Equal(t, []dynamic.Server{
		{URL: "http://127.0.0.3:80"},
		{URL: "http://127.0.0.2:80"},
	}, catalogServers(t, receiveCatalogMessage(t, configurationChan)))
}

func TestGetConsulServicesDataFromMultipleDatacenters(t *testing.T) {
	t.Parallel()

	addresses := map[string]string{
		"dc1": "127.0.0.1",
		"dc2": "127.0.0.2",
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		datacenter := req.URL.Query().Get("dc")
		address, ok := addresses[datacenter]
		if !ok {
			http.Error(rw, "unexpected datacenter", http.StatusBadRequest)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		switch req.URL.Path {
		case "/v1/catalog/services":
			assert.NoError(t, json.NewEncoder(rw).Encode(map[string][]string{"Test": {}}))

		case "/v1/health/service/Test":
			assert.NoError(t, json.NewEncoder(rw).Encode([]*api.ServiceEntry{
				{
					Node: &api.Node{
						ID:         "node-" + datacenter,
						Node:       "node-" + datacenter,
						Address:    address,
						Datacenter: datacenter,
					},
					Service: &api.AgentService{
						ID:      "test-" + datacenter,
						Service: "Test",
						Address: address,
						Port:    80,
					},
					Checks: api.HealthChecks{
						&api.HealthCheck{Status: api.HealthPassing},
					},
				},
			}))

		default:
			http.NotFound(rw, req)
		}
	}))
	t.Cleanup(server.Close)

	var config Configuration
	config.SetDefaults()
	config.Endpoint.Address = strings.TrimPrefix(server.URL, "http://")
	config.Endpoint.Scheme = "http"

	provider := Provider{Configuration: config}
	require.NoError(t, provider.Init())

	var items []itemData
	for _, datacenter := range []string{"dc1", "dc2"} {
		client, err := createClient("", provider.Endpoint, datacenter)
		require.NoError(t, err)

		data, err := provider.getConsulServicesData(t.Context(), client)
		require.NoError(t, err)
		require.Len(t, data, 1)
		assert.Equal(t, datacenter, data[0].Datacenter)
		assert.Equal(t, addresses[datacenter], data[0].Address)
		items = append(items, data...)
	}

	dynamicConfiguration := provider.buildConfiguration(t.Context(), items, nil)
	assert.Equal(t, []dynamic.Server{
		{URL: "http://127.0.0.1:80"},
		{URL: "http://127.0.0.2:80"},
	}, dynamicConfiguration.HTTP.Services["Test"].LoadBalancer.Servers)
}

func catalogTestItem(datacenter, address string) itemData {
	return itemData{
		ID:         "id",
		Node:       "node",
		Datacenter: datacenter,
		Name:       "Test",
		Namespace:  "default",
		Address:    address,
		Port:       "80",
		Status:     api.HealthPassing,
		ExtraConf:  configuration{Enable: true},
	}
}

func receiveCatalogMessage(t *testing.T, messages <-chan dynamic.Message) dynamic.Message {
	t.Helper()

	select {
	case message := <-messages:
		return message
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for catalog configuration")
		return dynamic.Message{}
	}
}

func catalogServers(t *testing.T, message dynamic.Message) []dynamic.Server {
	t.Helper()

	require.NotNil(t, message.Configuration)
	require.Contains(t, message.Configuration.HTTP.Services, "Test")
	return message.Configuration.HTTP.Services["Test"].LoadBalancer.Servers
}
