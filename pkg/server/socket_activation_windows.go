//go:build windows

package server

func populateSocketActivationListeners() *SocketActivation {
	return &SocketActivation{enabled: false}
}
