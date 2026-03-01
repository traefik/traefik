package ipallowlist

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/ip"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
)

const (
	typeName = "IPAllowLister"
)

// ipAllowLister is a middleware that provides Checks of the Requesting IP against a set of Allowlists.
type ipAllowLister struct {
	next             http.Handler
	allowLister      *ip.Checker
	strategy         ip.Strategy
	name             string
	rejectStatusCode int
}

// New builds a new IPAllowLister given a list of CIDR-Strings to allow.
func New(ctx context.Context, next http.Handler, config dynamic.IPAllowList, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPAllowLister not created")
	}

	rejectStatusCode := config.RejectStatusCode
	// If RejectStatusCode is not given, default to Forbidden (403).
	if rejectStatusCode == 0 {
		rejectStatusCode = http.StatusForbidden
	} else if http.StatusText(rejectStatusCode) == "" {
		return nil, fmt.Errorf("invalid HTTP status code %d", rejectStatusCode)
	}

	checker, err := ip.NewChecker(config.SourceRange)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDRs %s: %w", config.SourceRange, err)
	}

	strategy, err := config.IPStrategy.Get()
	if err != nil {
		return nil, err
	}

	logger.Debug().Msgf("Setting up IPAllowLister with sourceRange: %s", config.SourceRange)

	return &ipAllowLister{
		strategy:         strategy,
		allowLister:      checker,
		next:             next,
		name:             name,
		rejectStatusCode: rejectStatusCode,
	}, nil
}

func (al *ipAllowLister) GetTracingInformation() (string, string) {
	return al.name, typeName
}

func (al *ipAllowLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), al.name, typeName)
	ctx := logger.WithContext(req.Context())

	clientIP := al.strategy.GetIP(req)
	err := al.allowLister.IsAuthorized(clientIP)
	if err != nil {
		logger.Debug().Msgf("Rejecting IP %s: %v", clientIP, err)
		observability.SetStatusErrorf(req.Context(), "Rejecting IP %s: %v", clientIP, err)
		reject(ctx, al.rejectStatusCode, rw)
		return
	}
	logger.Debug().Msgf("Accepting IP %s", clientIP)

	al.next.ServeHTTP(rw, req)
}

func reject(ctx context.Context, statusCode int, rw http.ResponseWriter) {
	rw.WriteHeader(statusCode)
	_, err := rw.Write([]byte(http.StatusText(statusCode)))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Send()
	}
}
