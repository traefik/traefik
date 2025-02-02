package server

import (
	"errors"
	"net"
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
