package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/tcp"
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

	conn, err := net.ListenPacket("udp", configuration.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("error while starting http3 listener: %w", err)
	}

	h3 := &http3server{
		http3conn: conn,
		getter: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			return nil, errors.New("no tls config")
		},
	}

	h3.Server = &http3.Server{
		Server: &http.Server{
			Addr:         configuration.GetAddress(),
			Handler:      httpsServer.Server.(*http.Server).Handler,
			ErrorLog:     httpServerLogger,
			ReadTimeout:  time.Duration(configuration.Transport.RespondingTimeouts.ReadTimeout),
			WriteTimeout: time.Duration(configuration.Transport.RespondingTimeouts.WriteTimeout),
			IdleTimeout:  time.Duration(configuration.Transport.RespondingTimeouts.IdleTimeout),
			TLSConfig:    &tls.Config{GetConfigForClient: h3.getGetConfigForClient},
		},
	}

	previousHandler := httpsServer.Server.(*http.Server).Handler

	setQuicHeaders := getQuicHeadersSetter(configuration)

	httpsServer.Server.(*http.Server).Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := setQuicHeaders(rw.Header())
		if err != nil {
			log.FromContext(ctx).Errorf("failed to set HTTP3 headers: %v", err)
		}

		previousHandler.ServeHTTP(rw, req)
	})

	return h3, nil
}

// TODO: rewrite if at some point `port` become an exported field of http3.Server.
func getQuicHeadersSetter(configuration *static.EntryPoint) func(header http.Header) error {
	advertisedAddress := configuration.GetAddress()
	if configuration.HTTP3.AdvertisedPort != 0 {
		advertisedAddress = fmt.Sprintf(`:%d`, configuration.HTTP3.AdvertisedPort)
	}

	// if `QuickConfig` of h3.server happens to be configured,
	// it should also be configured identically in the headerServer
	headerServer := &http3.Server{
		Server: &http.Server{
			Addr: advertisedAddress,
		},
	}

	// set quic headers with the "header" http3 server instance
	return headerServer.SetQuicHeaders
}

func (e *http3server) Start() error {
	return e.Serve(e.http3conn)
}

func (e *http3server) Switch(rt *tcp.Router) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.getter = rt.GetTLSGetClientInfo()
}

func (e *http3server) getGetConfigForClient(info *tls.ClientHelloInfo) (*tls.Config, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.getter(info)
}
