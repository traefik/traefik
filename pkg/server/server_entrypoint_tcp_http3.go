package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/static"
	tcprouter "github.com/traefik/traefik/v2/pkg/server/router/tcp"
)

type http3server struct {
	*http3.Server

	http3conn net.PacketConn

	lock   sync.RWMutex
	getter func(info *tls.ClientHelloInfo) (*tls.Config, error)
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
		getter: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			return nil, errors.New("no tls config")
		},
	}

	h3.Server = &http3.Server{
		Addr:      configuration.GetAddress(),
		Port:      configuration.HTTP3.AdvertisedPort,
		Handler:   httpsServer.Server.(*http.Server).Handler,
		TLSConfig: &tls.Config{GetConfigForClient: h3.getGetConfigForClient},
	}

	previousHandler := httpsServer.Server.(*http.Server).Handler

	httpsServer.Server.(*http.Server).Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := h3.Server.SetQuicHeaders(rw.Header()); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Failed to set HTTP3 headers")
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

	e.getter = rt.GetTLSGetClientInfo()
}

func (e *http3server) getGetConfigForClient(info *tls.ClientHelloInfo) (*tls.Config, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.getter(info)
}

func (e *http3server) Shutdown(_ context.Context) error {
	// TODO: use e.Server.CloseGracefully() when available.
	return e.Server.Close()
}
