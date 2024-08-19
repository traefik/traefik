//go:build !windows

package server

import (
	"net"

	"github.com/coreos/go-systemd/activation"
	"github.com/rs/zerolog/log"
)

var (
	socketActivationListeners   map[string]net.Listener
	socketActivationPacketConns map[string]net.PacketConn
)

func init() {
	// Populates pre-defined TCP and UDP listeners provided by systemd socket activation.
	populateSocketActivationListeners()
}

func populateSocketActivationListeners() {
	// We use Files api due to activation not providing method for get PacketConn with names
	files := activation.Files(true)
	socketActivationListeners = make(map[string]net.Listener)
	socketActivationPacketConns = make(map[string]net.PacketConn)

	for _, f := range files {
		if listener, err := net.FileListener(f); err == nil {
			_, ok := socketActivationListeners[f.Name()]
			if !ok {
				socketActivationListeners[f.Name()] = listener
			} else {
				log.Error().Str("listenersName", f.Name()).Msg("Socket activation TCP listeners must have one and only one listener per name")
			}
			f.Close()
		} else if pc, err := net.FilePacketConn(f); err == nil {
			_, ok := socketActivationPacketConns[f.Name()]
			if !ok {
				socketActivationPacketConns[f.Name()] = pc
			} else {
				log.Error().Str("listenersName", f.Name()).Msg("Socket activation UDP listeners must have one and only one listener per name")
			}
			f.Close()
		}
	}
}
