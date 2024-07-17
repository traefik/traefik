//go:build !windows

package server

import (
	"net"

	"github.com/coreos/go-systemd/v22/activation"
	"github.com/rs/zerolog/log"
)

func populateSocketActivationListeners() {
	listenersWithName, _ := activation.ListenersWithNames()

	socketActivationListeners = make(map[string]net.Listener)
	for name, lns := range listenersWithName {
		if len(lns) != 1 {
			log.Error().Str("listenersName", name).Msg("Socket activation listeners must have one and only one listener per name")
			continue
		}

		socketActivationListeners[name] = lns[0]
	}
}
