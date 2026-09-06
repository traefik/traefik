package udp

import (
	"context"
	"crypto/tls"
	"net"
	"slices"
	"time"

	"github.com/pion/dtls/v3"
	"github.com/rs/zerolog/log"
)

const dtlsHandshakeTimeout = 10 * time.Second

type DTLSHandler struct {
	Next           Handler
	Options        []dtls.ServerOption
	GetCertificate func(info *tls.ClientHelloInfo) (*tls.Certificate, error)
	LocalAddr      net.Addr
}

func (h *DTLSHandler) ServeUDP(conn WriteCloser) {
	pConn := newPacketConnAdapter(conn, h.LocalAddr)

	opts := slices.Clone(h.Options)
	if h.GetCertificate != nil {
		certificate := dtls.WithGetCertificate(func(chi *dtls.ClientHelloInfo) (*tls.Certificate, error) {
			return h.GetCertificate(&tls.ClientHelloInfo{
				ServerName: chi.ServerName,
				Conn:       pConn,
			})
		})

		opts = append(opts, certificate)
	}

	dtlsConn, err := dtls.ServerWithOptions(pConn, conn.RemoteAddr(), opts...)
	if err != nil {
		log.Error().Err(err).Msg("Error creating DTLS server")
		conn.Close()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dtlsHandshakeTimeout)
	defer cancel()

	if err := dtlsConn.HandshakeContext(ctx); err != nil {
		log.Error().Err(err).Msg("Error during DTLS handshake")
		dtlsConn.Close()
		return
	}

	h.Next.ServeUDP(dtlsConn)
}

// packetConnAdapter adapts a single-peer udp.WriteCloser session to the
// net.PacketConn interface pion/dtls Server/Client functions require.
type packetConnAdapter struct {
	WriteCloser

	localAddr net.Addr
}

func newPacketConnAdapter(conn WriteCloser, localAddr net.Addr) packetConnAdapter {
	return packetConnAdapter{
		WriteCloser: conn,
		localAddr:   localAddr,
	}
}

func (a packetConnAdapter) ReadFrom(p []byte) (int, net.Addr, error) {
	n, err := a.Read(p)

	return n, a.RemoteAddr(), err
}

// WriteTo addr is ignored here because this adapter is scoped to exactly one peer.
func (a packetConnAdapter) WriteTo(p []byte, _ net.Addr) (int, error) {
	return a.Write(p)
}

func (a packetConnAdapter) LocalAddr() net.Addr {
	return a.localAddr
}

// SetDeadline SetReadDeadline and SetWriteDeadline are no-ops: pion/dtls
// manages its own retransmission timing internally and was confirmed, by
// spiking against v3.1.8, to never call these on the underlying PacketConn
// during the handshake or steady-state reads.
func (a packetConnAdapter) SetDeadline(t time.Time) error {
	return nil
}

func (a packetConnAdapter) SetReadDeadline(t time.Time) error {
	return nil
}

func (a packetConnAdapter) SetWriteDeadline(t time.Time) error {
	return nil
}
