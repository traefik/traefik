package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/vulcand/oxy/utils"
)

// GetSourceExtractor returns the SourceExtractor function corresponding to the given sourceMatcher.
// It defaults to a RemoteAddrStrategy IPStrategy if need be.
func GetSourceExtractor(ctx context.Context, sourceMatcher *dynamic.SourceCriterion) (utils.SourceExtractor, error) {
	if sourceMatcher == nil ||
		sourceMatcher.IPStrategy == nil &&
			sourceMatcher.RequestHeaderName == "" && !sourceMatcher.RequestHost {
		sourceMatcher = &dynamic.SourceCriterion{
			IPStrategy: &dynamic.IPStrategy{},
		}
	}

	logger := log.FromContext(ctx)
	if sourceMatcher.IPStrategy != nil {
		strategy, err := sourceMatcher.IPStrategy.Get()
		if err != nil {
			return nil, err
		}

		logger.Debug("Using IPStrategy")
		return utils.ExtractorFunc(func(req *http.Request) (string, int64, error) {
			return strategy.GetIP(req), 1, nil
		}), nil
	}

	if sourceMatcher.RequestHeaderName != "" {
		logger.Debug("Using RequestHeaderName")
		return utils.NewExtractor(fmt.Sprintf("request.header.%s", sourceMatcher.RequestHeaderName))
	}

	if sourceMatcher.RequestHost {
		logger.Debug("Using RequestHost")
		return utils.NewExtractor("request.host")
	}

	return nil, errors.New("no SourceCriterion criterion defined")
}
