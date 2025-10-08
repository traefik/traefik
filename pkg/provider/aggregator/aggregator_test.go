package aggregator

import (
	"bytes"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func TestLaunchNamespacedProvider(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	originalLogger := log.Logger
	log.Logger = zerolog.New(&buf).Level(zerolog.InfoLevel)

	providerWithNamespace := &mockNamespacedProvider{namespace: "test-namespace"}

	aggregator := ProviderAggregator{
		internalProvider: providerWithNamespace,
	}

	cfgCh := make(chan dynamic.Message)
	pool := safe.NewPool(t.Context())

	t.Cleanup(func() {
		pool.Stop()
		log.Logger = originalLogger
	})

	err := aggregator.Provide(cfgCh, pool)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Starting provider *aggregator.mockNamespacedProvider (namespace: test-namespace)")
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

// mockNamespacedProvider is a mock implementation of NamespacedProvider for testing.
type mockNamespacedProvider struct {
	namespace string
}

func (m *mockNamespacedProvider) Namespace() string {
	return m.namespace
}

func (m *mockNamespacedProvider) Provide(_ chan<- dynamic.Message, _ *safe.Pool) error {
	return nil
}

func (m *mockNamespacedProvider) Init() error {
	return nil
}
