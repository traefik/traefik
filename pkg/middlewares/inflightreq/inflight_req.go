package inflightreq

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/vulcand/oxy/v2/connlimit"
)

const (
	typeName = "InFlightReq"
)

type inFlightReq struct {
	handler http.Handler
	name    string
}

// New creates a max request middleware.
// If no source criterion is provided in the config, it defaults to RequestHost.
func New(ctx context.Context, next http.Handler, config dynamic.InFlightReq, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	ctxLog := logger.WithContext(ctx)

	if config.SourceCriterion == nil ||
		config.SourceCriterion.IPStrategy == nil &&
			config.SourceCriterion.RequestHeaderName == "" && !config.SourceCriterion.RequestHost {
		config.SourceCriterion = &dynamic.SourceCriterion{
			RequestHost: true,
		}
	}

	sourceMatcher, err := middlewares.GetSourceExtractor(ctxLog, config.SourceCriterion)
	if err != nil {
		return nil, fmt.Errorf("error creating requests limiter: %w", err)
	}

	handler, err := connlimit.New(next, sourceMatcher, config.Amount,
		connlimit.Logger(logs.NewOxyWrapper(*logger)),
		connlimit.Verbose(logger.GetLevel() == zerolog.TraceLevel))
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %w", err)
	}

	return &inFlightReq{handler: handler, name: name}, nil
}

func (i *inFlightReq) GetTracingInformation() (string, string) {
	return i.name, typeName
}

func (i *inFlightReq) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	i.handler.ServeHTTP(rw, req)
}
