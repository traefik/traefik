// Package network is the extension point for the networks Traefik serves on
// besides the host's own.
//
// Out of the box an entryPoint listens on a host socket, and nothing here is
// involved. An Extension adds places to listen that the host does not have: a
// VPN node embedded in the process, for one, has no socket to bind.
//
// The package holds no implementation and depends on none. An Extension is
// built by whatever links one in, from its own part of the configuration, and
// handed to the entryPoints and the server. With none, which is a nil
// Extension, Traefik serves on the host network only.
package network

import (
	"context"
	"fmt"
	"net"
)

// Extension supplies what the host network does not: places for entryPoints
// to accept on.
//
// Which entryPoints it serves is the Extension's own business, decided from
// its own configuration: the rest of Traefik asks, and falls back to the host
// network when the answer is no.
type Extension interface {
	// Sources returns where the named entryPoint accepts besides, or instead
	// of, its own address on the host. The zero Sources means the entryPoint
	// is on the host only.
	//
	// A configuration that cannot be honored is refused when the Extension
	// is built, not here.
	Sources(entryPoint string) Sources

	// Start is called once, before the entryPoints start accepting. It must
	// not wait for the network: an Extension that has something to bring up
	// does so in the background.
	//
	// ctx is done as soon as the shutdown begins, which is before Withdraw
	// and before the entryPoints drain. It ends what Start is still bringing
	// up, and must not end what the Extension serves: that lasts until
	// Close.
	Start(ctx context.Context)

	// Withdraw is the first step of a graceful shutdown, called before any
	// entryPoint stops accepting. It asks the Extension to stop attracting
	// new connections while it still serves the ones that arrive, so that
	// peers have moved elsewhere by the time the listeners close. It is
	// bounded by ctx.
	Withdraw(ctx context.Context)

	// Close is the last step of a shutdown, called once every entryPoint has
	// stopped. It runs within the time the server allows itself to stop,
	// past which the process is killed.
	Close()
}

// Nop is the Extension that supplies nothing: every entryPoint stays on the
// host network. It is there to be embedded, by an Extension that answers for
// only part of what the interface covers.
type Nop struct{}

// Sources returns no source.
func (Nop) Sources(string) Sources { return Sources{} }

// Start does nothing.
func (Nop) Start(context.Context) {}

// Withdraw does nothing.
func (Nop) Withdraw(context.Context) {}

// Close does nothing.
func (Nop) Close() {}

// Sources are the places an Extension gives one entryPoint to accept on.
type Sources struct {
	// Primary takes the place of the entryPoint's own address: when it is
	// set, nothing is bound on the host for the entryPoint.
	Primary Source

	// Additional sources accept alongside the entryPoint's own address,
	// wherever that is, and feed the same entryPoint.
	Additional []Source
}

// Source is one place an entryPoint accepts on that is not a host socket.
//
// A Source is asked for a stream listener by a TCP entryPoint.
type Source interface {
	// String names the source, for logs and errors.
	fmt.Stringer

	// Listen returns a stream listener on the source. It must not wait for
	// the network: a listener that cannot bind yet is returned all the same,
	// and binds within Accept. A network that is slow or unreachable at
	// startup then costs only the entryPoints that are on it.
	//
	// The connections it accepts must be able to close their write half, with
	// a CloseWrite method as *net.TCPConn has: the entryPoint closes the ones
	// that cannot.
	Listen(ctx context.Context) net.Listener
}
