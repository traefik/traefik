package corazawaf

import (
	"context"
	"net/http"

	coreruleset "github.com/corazawaf/coraza-coreruleset"
	"github.com/corazawaf/coraza/v3"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "CorazaWAF"

// CorazaWAF is a web application firewall (WAF) middleware.
// it will help you to block the possible malicious requests.
type CorazaWAF struct {
	next http.Handler
}

func newErrorCb(logger *zerolog.Logger) func(types.MatchedRule) {
	return func(mr types.MatchedRule) {
		logMsg := mr.ErrorLog()
		switch mr.Rule().Severity() {
		case types.RuleSeverityEmergency,
			types.RuleSeverityAlert,
			types.RuleSeverityCritical,
			types.RuleSeverityError:
			logger.Error().Msg(logMsg)
		case types.RuleSeverityWarning:
			logger.Warn().Msg(logMsg)
		case types.RuleSeverityNotice:
			logger.Info().Msg(logMsg)
		case types.RuleSeverityInfo:
			logger.Info().Msg(logMsg)
		case types.RuleSeverityDebug:
			logger.Debug().Msg(logMsg)
		}
	}
}

// NewCorazaWAF constructs a new coraza waf instance from supplied frontend coraza struct.
func NewCorazaWAF(ctx context.Context, next http.Handler, cfg dynamic.CorazaWAF, name string) (*CorazaWAF, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	corazaCfg := coraza.NewWAFConfig().
		WithDirectives(cfg.Directives).
		WithErrorCallback(newErrorCb(logger))

	if cfg.CRSEnabled {
		corazaCfg = corazaCfg.WithRootFS(coreruleset.FS)
	}

	waf, err := coraza.NewWAF(corazaCfg)
	if err != nil {
		return nil, err
	}
	return &CorazaWAF{
		next: txhttp.WrapHandler(waf, next),
	}, nil
}

func (c *CorazaWAF) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c.next.ServeHTTP(rw, req)
}
