package tcp

import (
	"sync"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

type TcpConfig struct {
	rtLock  sync.RWMutex
	configs map[string]*dynamic.ServersTransportTCP
}

// Instatiates a new tcpconfig.
func NewTcpConfig() *TcpConfig {
	return &TcpConfig{
		configs: make(map[string]*dynamic.ServersTransportTCP),
	}
}

// Update updates the tcp configurations.
func (t *TcpConfig) Update(newConfigs map[string]*dynamic.ServersTransportTCP) {
	t.rtLock.Lock()
	defer t.rtLock.Unlock()
}
