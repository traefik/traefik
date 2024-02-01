//go:build !windows

package server

import (
	"net"

	"github.com/coreos/go-systemd/activation"
	"github.com/rs/zerolog/log"
)

func populateSocketActivationListeners() {
	listenersWithName, _ := activation.ListenersWithNames()
	if len(listenersWithName) > 0 {
		listeners = make(map[string]net.Listener)
		for name, lns := range listenersWithName {
			if len(lns) == 1 {
				listeners[name] = lns[0]
				continue
			}
			log.Error().Str("listenersName", name).Msg("socket activation listeners must have one and only one listener per name")
		}
	}
}
