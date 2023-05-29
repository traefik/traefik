package tcp

import (
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const DefaultName = "default"

var listeners = map[string]Listener{}

func Provide(listener Listener) {
	listeners[listener.Name()] = listener
}

func Accept(hello Hello, conn tcp.WriteCloser) tcp.WriteCloser {
	if nil == hello || "" == hello.ServerName() {
		return conn
	}
	if listener, ok := listeners[hello.ServerName()]; ok && nil != listener {
		return listener.Accept(hello, conn)
	}
	if listener, ok := listeners[DefaultName]; ok && nil != listener {
		return listener.Accept(hello, conn)
	}
	return conn
}

type Hello interface {
	// ServerName is SNI server name
	ServerName() string

	// Protos is ALPN protocols list
	Protos() []string

	// IsTLS is whether we are a TLS handshake
	IsTLS() bool

	// Peeked the bytes peeked from the hello while getting the info
	Peeked() string
}

func (that *clientHello) ServerName() string {
	return that.serverName
}

func (that *clientHello) Protos() []string {
	return that.protos
}

func (that *clientHello) IsTLS() bool {
	return that.isTLS
}

func (that *clientHello) Peeked() string {
	return that.peeked
}

type Listener interface {

	// Name is the provider name
	Name() string

	// Accept a tcp connection
	Accept(hello Hello, conn tcp.WriteCloser) tcp.WriteCloser
}
