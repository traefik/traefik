package server

import (
	"context"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/network"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// recordingNetworks is a networks extension that puts every entryPoint on one source,
// and records what happens to itself and to that source's listener, in order.
type recordingNetworks struct {
	network.Nop

	mu     sync.Mutex
	events []string

	closed    chan struct{}
	closeOnce sync.Once
}

func (n *recordingNetworks) Sources(string) network.Sources { return network.Sources{Primary: n} }

func (n *recordingNetworks) Start(context.Context) { n.record("extension started") }

func (n *recordingNetworks) Withdraw(context.Context) { n.record("extension withdrawn") }

func (n *recordingNetworks) Close() { n.record("extension closed") }

func (n *recordingNetworks) String() string { return "recording source" }

func (n *recordingNetworks) Listen(context.Context) net.Listener { return recordingListener{n} }

func (n *recordingNetworks) record(event string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// A listener is accepted on in a loop, and shut down by more than one path.
	if !slices.Contains(n.events, event) {
		n.events = append(n.events, event)
	}
}

func (n *recordingNetworks) recorded() []string {
	n.mu.Lock()
	defer n.mu.Unlock()

	return slices.Clone(n.events)
}

type recordingListener struct {
	networks *recordingNetworks
}

func (l recordingListener) Accept() (net.Conn, error) {
	l.networks.record("entryPoint accepting")

	<-l.networks.closed

	return nil, net.ErrClosed
}

func (l recordingListener) Close() error {
	l.networks.record("entryPoint stopped")
	l.networks.closeOnce.Do(func() { close(l.networks.closed) })

	return nil
}

func (l recordingListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(192, 0, 2, 1), Port: 80} }

// The server brackets the entryPoints with the networks extension:
// started before they accept, withdrawn before they stop, and closed after.
func TestServerDrivesTheNetworksExtension(t *testing.T) {
	networks := &recordingNetworks{closed: make(chan struct{})}

	entryPoints, err := NewTCPEntryPoints(static.EntryPoints{"web": sourcesEntryPoint("192.0.2.1:80")}, nil, metrics.NewVoidRegistry(), networks)
	require.NoError(t, err)

	routinesPool := safe.NewPool(t.Context())

	provider := &mockProvider{messages: []dynamic.Message{{ProviderName: "mock", Configuration: &dynamic.Configuration{}}}}
	watcher := NewConfigurationWatcher(routinesPool, provider, []string{}, "", false)

	srv := NewServer(routinesPool, entryPoints, UDPEntryPoints{}, watcher, nil, networks)

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	srv.Start(ctx)

	require.Eventually(t, func() bool {
		return slices.Contains(networks.recorded(), "entryPoint accepting")
	}, 5*time.Second, 10*time.Millisecond)

	cancel()
	srv.Wait()
	srv.Close()

	assert.Equal(t, []string{
		"extension started",
		"entryPoint accepting",
		"extension withdrawn",
		"entryPoint stopped",
		"extension closed",
	}, networks.recorded())
}
