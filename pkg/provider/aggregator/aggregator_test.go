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

func TestProviderAggregator_Provide_throttled(t *testing.T) {
	aggregator := ProviderAggregator{
		internalProvider:          &unthrottledProviderMock{"internal"},
		fileProvider:              &providerMock{"file"},
		providersThrottleDuration: 10 * time.Millisecond,
	}

	// Buffered like the ConfigurationWatcher channel, so that the internal provider never waits for a reader.
	cfgCh := make(chan dynamic.Message, 100)
	pool := safe.NewPool(t.Context())

	t.Cleanup(pool.Stop)

	require.NoError(t, aggregator.Provide(cfgCh, pool))

	// The file provider initial configuration must be sent before the internal provider one, even when throttled,
	// as the ConfigurationWatcher only waits for the internal provider before applying the configuration.
	requireReceivedMessageFromProviders(t, cfgCh, []string{"file"})
	requireReceivedMessageFromProviders(t, cfgCh, []string{"internal"})
}

func TestMaybeThrottledProvide_startupMessages(t *testing.T) {
	cfgCh := make(chan dynamic.Message)
	errCh := make(chan error, 1)
	pool := safe.NewPool(t.Context())

	t.Cleanup(func() {
		// Unblocks the routines still sending on cfgCh, so that the pool can stop.
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-done:
					return
				case <-cfgCh:
				}
			}
		}()

		pool.Stop()
		close(done)
	})

	go func() {
		errCh <- maybeThrottledProvide(&providerMock{"file"}, 10*time.Millisecond)(cfgCh, pool)
	}()

	// Provide must not return until the configuration sent while starting has been delivered.
	select {
	case <-errCh:
		require.Fail(t, "Provide returned before the startup configuration was delivered.")
	case <-time.After(100 * time.Millisecond):
	}

	msg := <-cfgCh
	assert.Equal(t, "file", msg.ProviderName)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.Fail(t, "Timeout while waiting for Provide to return.")
	}
}

func TestMaybeThrottledProvide_throttlesAfterStartup(t *testing.T) {
	throttleDuration := 100 * time.Millisecond

	cfgCh := make(chan dynamic.Message, 10)
	pool := safe.NewPool(t.Context())

	t.Cleanup(pool.Stop)

	prd := &watchingProviderMock{}
	require.NoError(t, maybeThrottledProvide(prd, throttleDuration)(cfgCh, pool))

	msg := <-cfgCh
	assert.Equal(t, "startup", msg.ProviderName)
	startupReceived := time.Now()

	prd.send("update")

	select {
	case msg = <-cfgCh:
		assert.Equal(t, "update", msg.ProviderName)
		assert.GreaterOrEqual(t, time.Since(startupReceived), throttleDuration)
	case <-time.After(time.Second):
		require.Fail(t, "Timeout while waiting for configuration.")
	}
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

// watchingProviderMock sends a configuration while starting, and keeps the channel to send updates later.
type watchingProviderMock struct {
	configurationChan chan<- dynamic.Message
}

func (p *watchingProviderMock) Init() error {
	return nil
}

func (p *watchingProviderMock) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	p.configurationChan = configurationChan
	p.send("startup")

	return nil
}

func (p *watchingProviderMock) send(name string) {
	p.configurationChan <- dynamic.Message{
		ProviderName:  name,
		Configuration: &dynamic.Configuration{},
	}
}

// unthrottledProviderMock mimics the internal provider, which disables throttling.
type unthrottledProviderMock struct {
	Name string
}

func (p *unthrottledProviderMock) Init() error {
	return nil
}

func (p *unthrottledProviderMock) Provide(configurationChan chan<- dynamic.Message, _ *safe.Pool) error {
	configurationChan <- dynamic.Message{
		ProviderName:  p.Name,
		Configuration: &dynamic.Configuration{},
	}

	return nil
}

func (p *unthrottledProviderMock) ThrottleDuration() time.Duration {
	return 0
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
