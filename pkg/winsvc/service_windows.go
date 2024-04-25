//go:build windows
// +build windows

package winsvc

import (
	wsvc "golang.org/x/sys/windows/svc"
)

type serviceWindows struct{}

func init() {
	isService, err := wsvc.IsWindowsService()
	if err != nil {
		panic(err)
	}
	if !isService {
		return
	}
	go wsvc.Run("", serviceWindows{})
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
			ChanExit <- 1
			return false, 0
		}
	}
}
