package main

import (
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslogtcp"
	"github.com/traefik/traefik/v3/pkg/types"
)

// setupTCPAccessLog creates a TCP access log handler from the config, or returns nil if not enabled.
func setupTCPAccessLog(conf *types.AccessLog) *accesslogtcp.Handler {
	if conf == nil {
		return nil
	}
	handler, err := accesslogtcp.NewHandler(conf)
	if err != nil {
		return nil
	}
	return handler
}
