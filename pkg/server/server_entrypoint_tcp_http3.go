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
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tsnet"
)

type http3server struct {
	*http3.Server

	http3conn net.PacketConn

	lock   sync.RWMutex
	getter func(info *tls.ClientHelloInfo) (*tls.Config, error)
}

func newHTTP3Server(ctx context.Context, name string, config *static.EntryPoint, httpsServer *httpServer, ts *tsnet.Server) (*http3server, error) {
	var conn net.PacketConn
	var err error

	if config.HTTP3 == nil {
		return nil, nil
	}

	if config.HTTP3.AdvertisedPort < 0 {
		return nil, errors.New("advertised port must be greater than or equal to zero")
	}

	addr := config.GetAddress()

	if ts != nil {
		// ListenPacket for a tsnet.Server requires an explicit IP
		// Since the IP is determined by the Tailscale service, we want to make sure that we don't have an IP specified in the address
		// We then connect to the Tailscale network and get the IP from there
		if len(addr) == 0 || addr[0] != ':' {
			return nil, errors.New("endpoint's address is not valid for a tsnet endpoint: must include the port only")
		}

		var status *ipnstate.Status
		status, err = ts.Up(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to the Tailscale network: %w", err)
		}

		var found bool
		for _, ip := range status.TailscaleIPs {
			if !ip.Is4() {
				continue
			}
			addr = ip.String() + addr
			found = true
			break
		}
		if !found {
			return nil, errors.New("did not find an IPv4 address for Tailscale")
		}

		conn, err = ts.ListenPacket("udp4", addr)
		if err != nil {
			return nil, fmt.Errorf("starting tsnet listener: %w", err)
		}
	} else {
		// if we have predefined connections from socket activation
		if socketActivation.isEnabled() {
			conn, err = socketActivation.getConn(name)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("name", name).Msg("Unable to use socket activation for entrypoint")
			}
		}

		if conn == nil {
			listenConfig := newListenConfig(config)
			conn, err = listenConfig.ListenPacket(ctx, "udp", addr)
			if err != nil {
				return nil, fmt.Errorf("starting listener: %w", err)
			}
		}
	}

	h3 := &http3server{
		http3conn: conn,
		getter: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			return nil, errors.New("no tls config")
		},
	}

	h3.Server = &http3.Server{
		Addr:      addr,
		Port:      config.HTTP3.AdvertisedPort,
		Handler:   httpsServer.Server.(*http.Server).Handler,
		TLSConfig: &tls.Config{GetConfigForClient: h3.getGetConfigForClient},
		QUICConfig: &quic.Config{
			Allow0RTT: false,
		},
	}

	previousHandler := httpsServer.Server.(*http.Server).Handler

	httpsServer.Server.(*http.Server).Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := h3.Server.SetQUICHeaders(rw.Header()); err != nil {
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
