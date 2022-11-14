package ipallowlist

import (
	"context"
	"errors"
	"fmt"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

const (
	typeName = "IPAllowListerTCP"
)

// ipAllowLister is a middleware that provides Checks of the Requesting IP against a set of Allowlists.
type ipAllowLister struct {
	next        tcp.Handler
	allowLister *ip.Checker
	name        string
}

// New builds a new TCP IPAllowLister given a list of CIDR-Strings to allow.
func New(ctx context.Context, next tcp.Handler, config dynamic.TCPIPAllowList, name string) (tcp.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPAllowLister not created")
	}

	checker, err := ip.NewChecker(config.SourceRange)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDRs %s: %w", config.SourceRange, err)
	}

	logger.Debugf("Setting up IPAllowLister with sourceRange: %s", config.SourceRange)

	return &ipAllowLister{
		allowLister: checker,
		next:        next,
		name:        name,
	}, nil
}

func (al *ipAllowLister) ServeTCP(conn tcp.WriteCloser) {
	ctx := middlewares.GetLoggerCtx(context.Background(), al.name, typeName)
	logger := log.FromContext(ctx)

	addr := conn.RemoteAddr().String()

	err := al.allowLister.IsAuthorized(addr)
	if err != nil {
		logger.Errorf("Connection from %s rejected: %v", addr, err)
		conn.Close()
		return
	}

	logger.Debugf("Connection from %s accepted", addr)

	al.next.ServeTCP(conn)
}
