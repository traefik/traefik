//go:build !windows

package server

import (
	"errors"
	"net"

	"github.com/coreos/go-systemd/activation"
	"github.com/rs/zerolog/log"
)

type SocketActivation struct {
	enabled   bool
	listeners map[string]net.Listener
	conns     map[string]net.PacketConn
}

func (s *SocketActivation) isEnabled() bool {
	return s.enabled
}

func (s *SocketActivation) getListener(name string) (net.Listener, error) {
	listener, ok := s.listeners[name]
	if !ok {
		return nil, errors.New("unable to find socket activation TCP listener for entryPoint")
	}

	return listener, nil
}

func (s *SocketActivation) getConn(name string) (net.PacketConn, error) {
	conn, ok := s.conns[name]
	if !ok {
		return nil, errors.New("unable to find socket activation UDP listener for entryPoint")
	}

	return conn, nil
}

var socketActivation *SocketActivation

func init() {
	// Populates pre-defined TCP and UDP listeners provided by systemd socket activation.
	socketActivation = populateSocketActivationListeners()
}

func populateSocketActivationListeners() *SocketActivation {
	// We use Files api due to activation not providing method for get PacketConn with names
	files := activation.Files(true)
	sa := &SocketActivation{enabled: false}
	sa.listeners = make(map[string]net.Listener)
	sa.conns = make(map[string]net.PacketConn)

	if len(files) > 0 {
		sa.enabled = true

		for _, f := range files {
			if lc, err := net.FileListener(f); err == nil {
				_, ok := sa.listeners[f.Name()]
				if ok {
					log.Error().Str("listenersName", f.Name()).Msg("Socket activation TCP listeners must have one and only one listener per name")
				} else {
					sa.listeners[f.Name()] = lc
				}
				f.Close()
			} else if pc, err := net.FilePacketConn(f); err == nil {
				_, ok := sa.conns[f.Name()]
				if ok {
					log.Error().Str("listenersName", f.Name()).Msg("Socket activation UDP listeners must have one and only one listener per name")
				} else {
					sa.conns[f.Name()] = pc
				}
				f.Close()
			}
		}
	}

	return sa
}
