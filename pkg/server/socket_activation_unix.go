//go:build !windows

package server

import (
	"net"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/rs/zerolog/log"
)

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
