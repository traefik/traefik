package aggregator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestProviderAggregator_Provide(t *testing.T) {
	aggregator := ProviderAggregator{
		internalProvider: &providerMock{"internal"},
		fileProvider:     &providerMock{"file"},
		providers: []provider.Provider{
			&providerMock{"salad"},
			&providerMock{"tomato"},
			&providerMock{"onion"},
		},
	}

	cfgCh := make(chan dynamic.Message)
	errCh := make(chan error)
	pool := safe.NewPool(t.Context())

	t.Cleanup(pool.Stop)

	go func() {
		errCh <- aggregator.Provide(cfgCh, pool)
	}()

	// Make sure the file provider is always called first.
	requireReceivedMessageFromProviders(t, cfgCh, []string{"file"})

	// Check if all providers have been called, the order doesn't matter.
	requireReceivedMessageFromProviders(t, cfgCh, []string{"salad", "tomato", "onion", "internal"})

	require.NoError(t, <-errCh)
}

// requireReceivedMessageFromProviders makes sure the given providers have emitted a message on the given message channel.
// Providers order is not enforced.
func requireReceivedMessageFromProviders(t *testing.T, cfgCh <-chan dynamic.Message, names []string) {
	t.Helper()

	var msg dynamic.Message
	var receivedMessagesFrom []string

	for range names {
		select {
		case <-time.After(100 * time.Millisecond):
			require.Fail(t, "Timeout while waiting for configuration.")
		case msg = <-cfgCh:
			receivedMessagesFrom = append(receivedMessagesFrom, msg.ProviderName)
		}
	}

	require.ElementsMatch(t, names, receivedMessagesFrom)
}

type providerMock struct {
	Name string
}

func (p *providerMock) Init() error {
	return nil
}

func (p *providerMock) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	configurationChan <- dynamic.Message{
		ProviderName:  p.Name,
		Configuration: &dynamic.Configuration{},
	}

	return nil
}

// mockNamespaceProvider is a mock implementation of NamespaceProvider for testing
type mockNamespaceProvider struct {
	namespace string
}

func (m *mockNamespaceProvider) GetNamespace() string {
	return m.namespace
}

func (m *mockNamespaceProvider) Provide(_ chan<- dynamic.Message, _ *safe.Pool) error {
	return nil
}

func (m *mockNamespaceProvider) Init() error {
	return nil
}

func TestLaunchProviderWithNamespace(t *testing.T) {
	// Test that providers implementing NamespaceProvider are correctly identified
	providerWithNamespace := &mockNamespaceProvider{namespace: "test-namespace"}

	// Verify the interface implementation
	var _ provider.NamespaceProvider = providerWithNamespace
	var _ provider.Provider = providerWithNamespace

	// Test GetNamespace method
	assert.Equal(t, "test-namespace", providerWithNamespace.GetNamespace())

	// Test with empty namespace
	providerEmptyNamespace := &mockNamespaceProvider{namespace: ""}
	assert.Empty(t, providerEmptyNamespace.GetNamespace())
}
