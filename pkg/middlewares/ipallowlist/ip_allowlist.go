package ipallowlist

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "IPAllowLister"
)

// ipAllowLister is a middleware that provides Checks of the Requesting IP against a set of Allowlists.
type ipAllowLister struct {
	next        http.Handler
	allowLister *ip.Checker
	strategy    ip.Strategy
	name        string
}

// New builds a new IPAllowLister given a list of CIDR-Strings to allow.
func New(ctx context.Context, next http.Handler, config dynamic.IPAllowList, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPAllowLister not created")
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
		strategy:    strategy,
		allowLister: checker,
		next:        next,
		name:        name,
	}, nil
}

func (al *ipAllowLister) GetTracingInformation() (string, ext.SpanKindEnum) {
	return al.name, tracing.SpanKindNoneEnum
}

func (al *ipAllowLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), al.name, typeName)
	ctx := logger.WithContext(req.Context())

	clientIP := al.strategy.GetIP(req)
	err := al.allowLister.IsAuthorized(clientIP)
	if err != nil {
		msg := fmt.Sprintf("Rejecting IP %s: %v", clientIP, err)
		logger.Debug().Msg(msg)
		tracing.SetErrorWithEvent(req, msg)
		reject(ctx, rw)
		return
	}
	logger.Debug().Msgf("Accepting IP %s", clientIP)

	al.next.ServeHTTP(rw, req)
}

func reject(ctx context.Context, rw http.ResponseWriter) {
	statusCode := http.StatusForbidden

	rw.WriteHeader(statusCode)
	_, err := rw.Write([]byte(http.StatusText(statusCode)))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Send()
	}
}
