package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/network"
)

// fakeSource is a place to accept on that an entryPoint did not open itself.
// It stands in for a network that is not the host's, and uses loopback sockets to do it.
type fakeSource struct {
	name string

	// listener is what Listen returns.
	listener net.Listener
}

func (s *fakeSource) String() string { return s.name }

func (s *fakeSource) Listen(context.Context) net.Listener { return s.listener }

// streamSource returns a source whose listener is a loopback one.
func streamSource(t *testing.T, name string) *fakeSource {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	return &fakeSource{name: name, listener: listener}
}

// pendingSource returns a source on a network that has not come up: its listener never accepts.
func pendingSource(name string) *fakeSource {
	return &fakeSource{name: name, listener: newStubListener()}
}

func sourcesEntryPoint(address string) *static.EntryPoint {
	transport := &static.EntryPointsTransport{}
	transport.SetDefaults()

	return &static.EntryPoint{
		Address:          address,
		Transport:        transport,
		ForwardedHeaders: &static.ForwardedHeaders{},
		HTTP2:            &static.HTTP2Config{},
	}
}

// requireAccepts dials address, and requires listener to accept the connection.
func requireAccepts(t *testing.T, listener net.Listener, address string) {
	t.Helper()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := net.DialTimeout("tcp", address, 5*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	select {
	case conn := <-accepted:
		t.Cleanup(func() { _ = conn.Close() })
	case <-time.After(5 * time.Second):
		t.Fatalf("no connection was accepted on %s", address)
	}
}

// A primary source takes the place of the entryPoint's own address: nothing is bound on the host.
func TestBuildListenerOnAPrimarySource(t *testing.T) {
	source := streamSource(t, "primary")

	// The address is one nothing could bind, so a listener at all proves the host was left alone.
	listener, err := buildListener(t.Context(), "web", sourcesEntryPoint("192.0.2.1:80"), network.Sources{Primary: source})
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	assert.Equal(t, source.listener.Addr(), listener.Addr())

	requireAccepts(t, listener, source.listener.Addr().String())
}

// The entryPoint's proxyProtocol judges the peer of its own address, wherever that address is.
func TestBuildListenerAppliesProxyProtocolToAPrimarySource(t *testing.T) {
	source := streamSource(t, "primary")

	config := sourcesEntryPoint("192.0.2.1:80")
	config.ProxyProtocol = &static.ProxyProtocol{Insecure: true}

	listener, err := buildListener(t.Context(), "web", config, network.Sources{Primary: source})
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := net.DialTimeout("tcp", source.listener.Addr().String(), 5*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Write([]byte("PROXY TCP4 203.0.113.7 198.51.100.1 4242 80\r\n"))
	require.NoError(t, err)

	select {
	case conn := <-accepted:
		t.Cleanup(func() { _ = conn.Close() })
		assert.Equal(t, "203.0.113.7:4242", conn.RemoteAddr().String())
	case <-time.After(5 * time.Second):
		t.Fatal("no connection was accepted")
	}
}

// namedAddr is the address of a network that is not TCP/IP, or does not say so with a *net.TCPAddr.
type namedAddr string

func (a namedAddr) Network() string { return "fake" }

func (a namedAddr) String() string { return string(a) }

// remoteAddrListener accepts the connections of a listener, and gives them the remote address of another network.
type remoteAddrListener struct {
	net.Listener

	remoteAddr net.Addr
}

func (l remoteAddrListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return remoteAddrConn{Conn: conn, remoteAddr: l.remoteAddr}, nil
}

type remoteAddrConn struct {
	net.Conn

	remoteAddr net.Addr
}

func (c remoteAddrConn) RemoteAddr() net.Addr { return c.remoteAddr }

func TestAddrIP(t *testing.T) {
	testCases := []struct {
		desc   string
		addr   net.Addr
		expect string
	}{
		{desc: "a TCP address", addr: &net.TCPAddr{IP: net.ParseIP("203.0.113.7"), Port: 4242}, expect: "203.0.113.7"},
		{desc: "an IP and port of another type", addr: namedAddr("203.0.113.7:4242"), expect: "203.0.113.7"},
		{desc: "an IPv6 and port of another type", addr: namedAddr("[2001:db8::7]:4242"), expect: "2001:db8::7"},
		{desc: "an IP alone", addr: namedAddr("203.0.113.7"), expect: "203.0.113.7"},
		{desc: "a name", addr: namedAddr("tunnel-7")},
		{desc: "no address"},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			ip, ok := addrIP(test.addr)

			assert.Equal(t, test.expect != "", ok)
			if test.expect != "" {
				assert.True(t, ip.Equal(net.ParseIP(test.expect)), ip.String())
			}
		})
	}
}

// The proxyProtocol of the entryPoint judges the peer by its IP, which a primary source may not report as a host socket does.
// A peer it cannot judge is not trusted, and must not end the listener either: an error from the policy does.
func TestBuildListenerProxyProtocolOnAPrimarySourceWithAnotherAddressType(t *testing.T) {
	testCases := []struct {
		desc         string
		remoteAddr   net.Addr
		expectRemote string
	}{
		{
			desc:         "a trusted IP of another type",
			remoteAddr:   namedAddr("10.0.0.7:4242"),
			expectRemote: "203.0.113.7:4242",
		},
		{
			desc:         "an untrusted IP of another type",
			remoteAddr:   namedAddr("192.0.2.9:4242"),
			expectRemote: "192.0.2.9:4242",
		},
		{
			desc:         "an address with no IP",
			remoteAddr:   namedAddr("tunnel-7"),
			expectRemote: "tunnel-7",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			source := streamSource(t, "primary")
			source.listener = remoteAddrListener{Listener: source.listener, remoteAddr: test.remoteAddr}

			config := sourcesEntryPoint("192.0.2.1:80")
			config.ProxyProtocol = &static.ProxyProtocol{TrustedIPs: []string{"10.0.0.0/8"}}

			listener, err := buildListener(t.Context(), "web", config, network.Sources{Primary: source})
			require.NoError(t, err)
			t.Cleanup(func() { _ = listener.Close() })

			type accepted struct {
				conn net.Conn
				err  error
			}

			acceptedCh := make(chan accepted, 1)
			go func() {
				conn, err := listener.Accept()
				acceptedCh <- accepted{conn: conn, err: err}
			}()

			client, err := net.DialTimeout("tcp", source.listener.Addr().String(), 5*time.Second)
			require.NoError(t, err)
			t.Cleanup(func() { _ = client.Close() })

			_, err = client.Write([]byte("PROXY TCP4 203.0.113.7 198.51.100.1 4242 80\r\n"))
			require.NoError(t, err)

			select {
			case got := <-acceptedCh:
				require.NoError(t, got.err)
				t.Cleanup(func() { _ = got.conn.Close() })

				assert.Equal(t, test.expectRemote, got.conn.RemoteAddr().String())
			case <-time.After(5 * time.Second):
				t.Fatal("no connection was accepted")
			}
		})
	}
}

// The reason additional sources exist: the entryPoint's host listener and its sources are one entryPoint,
// and the host side serves while a source's network is still unreachable.
func TestBuildListenerMergesAdditionalSources(t *testing.T) {
	up := streamSource(t, "up")
	pending := pendingSource("pending")

	listener, err := buildListener(t.Context(), "web", sourcesEntryPoint("127.0.0.1:0"), network.Sources{Additional: []network.Source{up, pending}})
	require.NoError(t, err)

	// The merged listener is known by the entryPoint's own address.
	host, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)
	assert.NotEqual(t, up.listener.Addr().String(), host.String())

	requireAccepts(t, listener, host.String())
	requireAccepts(t, listener, up.listener.Addr().String())

	require.NoError(t, listener.Close())

	_, err = listener.Accept()
	require.ErrorIs(t, err, net.ErrClosed)

	// Closing the entryPoint's listener closes the sources' too.
	_, err = up.listener.Accept()
	require.ErrorIs(t, err, net.ErrClosed)
}

// The entryPoint's proxyProtocol wraps its own listener only:
// a peer on an additional source must not be able to claim an address with a header.
func TestBuildListenerKeepsProxyProtocolOffAdditionalSources(t *testing.T) {
	source := streamSource(t, "additional")

	config := sourcesEntryPoint("127.0.0.1:0")
	config.ProxyProtocol = &static.ProxyProtocol{Insecure: true}

	listener, err := buildListener(t.Context(), "web", config, network.Sources{Additional: []network.Source{source}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := net.DialTimeout("tcp", source.listener.Addr().String(), 5*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.Write([]byte("PROXY TCP4 203.0.113.7 198.51.100.1 4242 80\r\n"))
	require.NoError(t, err)

	select {
	case conn := <-accepted:
		t.Cleanup(func() { _ = conn.Close() })
		assert.Equal(t, client.LocalAddr().String(), conn.RemoteAddr().String())
	case <-time.After(5 * time.Second):
		t.Fatal("no connection was accepted")
	}
}

// HTTP/3 is served on the entryPoint's own address on the host, which a primary source takes the place of.
func TestNewTCPEntryPointRefusesHTTP3OnAPrimarySource(t *testing.T) {
	config := sourcesEntryPoint("192.0.2.1:443")
	config.HTTP3 = &static.HTTP3Config{}

	_, err := newTCPEntryPoint(t.Context(), "websecure", config, nil, nil, network.Sources{Primary: streamSource(t, "primary")})
	require.ErrorContains(t, err, "HTTP/3 cannot be served on primary")
}
