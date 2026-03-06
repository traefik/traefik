//go:build windows
// +build windows

package winsvc

import (
	"github.com/rs/zerolog/log"
	wsvc "golang.org/x/sys/windows/svc"
)

type serviceWindows struct{}

func init() {
	isService, err := wsvc.IsWindowsService()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to determine if running as a Windows Service")
		return
	}
	if !isService {
		log.Info().Msg("Running in CMD mode, skipping Windows Service setup.")
		return
	}
	go func() {
		log.Info().Msg("Starting Traefik as a Windows Service")
		err := wsvc.Run("traefik", serviceWindows{})
		if err != nil {
			log.Error().Err(err).Msg("Failed to run Windows Service")
		}
	}()
}

func (serviceWindows) Execute(args []string, r <-chan wsvc.ChangeRequest, s chan<- wsvc.Status) (svcSpecificEC bool, exitCode uint32) {
	const accCommands = wsvc.AcceptStop | wsvc.AcceptShutdown
	s <- wsvc.Status{State: wsvc.Running, Accepts: accCommands}
	for {
		c := <-r
		switch c.Cmd {
		case wsvc.Interrogate:
			s <- c.CurrentStatus
		case wsvc.Stop, wsvc.Shutdown:
			s <- wsvc.Status{State: wsvc.StopPending}
			close(ChanExit) // Close channel safely
			return false, 0
		}
	}
}
