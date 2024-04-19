package ipwhitelist

import (
	"context"
	"errors"
	"fmt"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const (
	typeName = "IPWhiteListerTCP"
)

// ipWhiteLister is a middleware that provides Checks of the Requesting IP against a set of Whitelists.
type ipWhiteLister struct {
	next        tcp.Handler
	whiteLister *ip.Checker
	name        string
}

// New builds a new TCP IPWhiteLister given a list of CIDR-Strings to whitelist.
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPIPWhiteList, name string) (tcp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPWhiteLister not created")
	}

	checker, err := ip.NewChecker(config.SourceRange)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDR whitelist %s: %w", config.SourceRange, err)
	}

	logger.Debug().Msgf("Setting up IPWhiteLister with sourceRange: %s", config.SourceRange)

	return &ipWhiteLister{
		whiteLister: checker,
		next:        next,
		name:        name,
	}, nil
}

func (wl *ipWhiteLister) ServeTCP(conn tcp.WriteCloser) {
	logger := middlewares.GetLogger(context.Background(), wl.name, typeName)

	addr := conn.RemoteAddr().String()

	err := wl.whiteLister.IsAuthorized(addr)
	if err != nil {
		logger.Error().Err(err).Msgf("Connection from %s rejected", addr)
		conn.Close()
		return
	}

	logger.Debug().Msgf("Connection from %s accepted", addr)

	wl.next.ServeTCP(conn)
}
