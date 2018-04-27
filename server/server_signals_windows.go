// +build windows

package server

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/containous/traefik/log"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateEvent = kernel32.NewProc("CreateEventW")
	procSetEvent    = kernel32.NewProc("SetEvent")
)

const windowsEventName = "SomeMutexNameGUID"

var windowsEventHandle uintptr

func (s *Server) configureSignals() {
	handle, err := createEvent(windowsEventName)

	if err != nil {
		panic("server already running?")
	}

	windowsEventHandle = handle
}

func (s *Server) listenSignals() {

	for {
		syscall.WaitForSingleObject(syscall.Handle(windowsEventHandle), syscall.INFINITE)

		signalFileRotation(s)
	}
}

func createEvent(name string) (uintptr, error) {
	paramEventName := uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name)))

	ret, _, err := procCreateEvent.Call(
		0, // default security attributes
		0, // manual-reset event
		0, // initial state
		paramEventName)

	switch int(err.(syscall.Errno)) {
	case 0:
		return ret, nil
	default:
		return ret, err
	}
}

func setEvent(handle uintptr) error {
	_, _, err := procSetEvent.Call(handle)

	switch int(err.(syscall.Errno)) {
	case 0:
		return nil
	default:
		return err
	}
}

func signalFileRotation(s *Server) {
	log.Infof("Closing and re-opening log files for rotation")
	fmt.Println("Closing and re-opening log files for rotation")

	if s.accessLoggerMiddleware != nil {
		if err := s.accessLoggerMiddleware.Rotate(); err != nil {
			log.Errorf("Error rotating access log: %s", err)
		}
	}

	if err := log.RotateFile(); err != nil {
		log.Errorf("Error rotating traefik log: %s", err)
	}
}
