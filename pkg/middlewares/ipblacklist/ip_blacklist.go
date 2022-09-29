package ipblacklist

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "IPBlackLister"
)

// ipBlackLister is a middleware that provides Checks of the Requesting IP against a set of Blacklists.
type ipBlackLister struct {
	next        http.Handler
	blackLister *ip.Checker
	strategy    ip.Strategy
	name        string
}

// New builds a new IPBlackLister given a list of CIDR-Strings to blacklist.
func New(ctx context.Context, next http.Handler, config dynamic.IPBlackList, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName))
	logger.Debug("Creating middleware")

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPBlackLister not created")
	}

	checker, err := ip.NewChecker(config.SourceRange)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDR blacklist %s: %w", config.SourceRange, err)
	}

	strategy, err := config.IPStrategy.Get()
	if err != nil {
		return nil, err
	}

	logger.Debugf("Setting up IPBlackLister with sourceRange: %s", config.SourceRange)

	return &ipBlackLister{
		strategy:    strategy,
		blackLister: checker,
		next:        next,
		name:        name,
	}, nil
}

func (bl *ipBlackLister) GetTracingInformation() (string, ext.SpanKindEnum) {
	return bl.name, tracing.SpanKindNoneEnum
}

func (bl *ipBlackLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := middlewares.GetLoggerCtx(req.Context(), bl.name, typeName)
	logger := log.FromContext(ctx)

	clientIP := bl.strategy.GetIP(req)
	err := bl.blackLister.IsAuthorized(clientIP)
	if err == nil {
		msg := fmt.Sprintf("Rejecting IP %s: %v", clientIP, err)
		logger.Debug(msg)
		tracing.SetErrorWithEvent(req, msg)
		reject(ctx, rw)
		return
	}
	logger.Debugf("Accepting IP %s", clientIP)

	bl.next.ServeHTTP(rw, req)
}

func reject(ctx context.Context, rw http.ResponseWriter) {
	statusCode := http.StatusForbidden

	rw.WriteHeader(statusCode)
	_, err := rw.Write([]byte(http.StatusText(statusCode)))
	if err != nil {
		log.FromContext(ctx).Error(err)
	}
}
