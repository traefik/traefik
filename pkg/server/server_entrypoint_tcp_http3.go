package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	tcpmuxer "github.com/traefik/traefik/v2/pkg/muxer/tcp"
	tcprouter "github.com/traefik/traefik/v2/pkg/server/router/tcp"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

type http3server struct {
	*http3.Server

	http3conn net.PacketConn

	lock   sync.RWMutex
	getter func(data tcpmuxer.ConnData) (*tls.Config, string, error)
}

func newHTTP3Server(ctx context.Context, configuration *static.EntryPoint, httpsServer *httpServer) (*http3server, error) {
	if configuration.HTTP3 == nil {
		return nil, nil
	}

	if configuration.HTTP3.AdvertisedPort < 0 {
		return nil, errors.New("advertised port must be greater than or equal to zero")
	}

	conn, err := net.ListenPacket("udp", configuration.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("starting listener: %w", err)
	}

	h3 := &http3server{
		http3conn: conn,
		getter: func(data tcpmuxer.ConnData) (*tls.Config, string, error) {
			return nil, "", errors.New("no tls config")
		},
	}

	h3.Server = &http3.Server{
		Addr:      configuration.GetAddress(),
		Port:      configuration.HTTP3.AdvertisedPort,
		Handler:   httpsServer.Server.(*http.Server).Handler,
		TLSConfig: &tls.Config{GetConfigForClient: h3.getConfigForClient},
		QUICConfig: &quic.Config{
			Allow0RTT: false,
		},
		ConnContext: func(ctx context.Context, c *quic.Conn) context.Context {
			name, err := h3.getConfigName(c)
			if err != nil {
				log.WithoutContext().Errorf("Error getting name for client: %v", err)
				return ctx
			}
			return tcp.AddTLSOptionsNameInContext(ctx, name)
		},
	}

	previousHandler := httpsServer.Server.(*http.Server).Handler

	httpsServer.Server.(*http.Server).Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := h3.Server.SetQUICHeaders(rw.Header()); err != nil {
			log.FromContext(ctx).Errorf("Failed to set HTTP3 headers: %v", err)
		}

		previousHandler.ServeHTTP(rw, req)
	})

	return h3, nil
}

func (e *http3server) Start() error {
	return e.Serve(e.http3conn)
}

func (e *http3server) Switch(rt *tcprouter.Router) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.getter = rt.GetTLSConfigMatcherFunc()
}

func (e *http3server) Shutdown(_ context.Context) error {
	// TODO: use e.Server.CloseGracefully() when available.
	return e.Server.Close()
}

func (e *http3server) getConfigForClient(info *tls.ClientHelloInfo) (*tls.Config, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	connData, err := tcpmuxer.NewConnData(info.ServerName, info.Conn.RemoteAddr(), info.SupportedProtos)
	if err != nil {
		return nil, fmt.Errorf("creating ConnData from client hello: %w", err)
	}

	conf, _, err := e.getter(connData)
	return conf, err
}

func (e *http3server) getConfigName(c *quic.Conn) (string, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	connData, err := tcpmuxer.NewConnData(c.ConnectionState().TLS.ServerName, c.RemoteAddr(), []string{c.ConnectionState().TLS.NegotiatedProtocol})
	if err != nil {
		return "", fmt.Errorf("creating ConnData from quic Conn: %w", err)
	}

	_, name, err := e.getter(connData)
	return name, err
}
