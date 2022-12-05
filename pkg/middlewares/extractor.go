package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/vulcand/oxy/v2/utils"
)

// GetSourceExtractor returns the SourceExtractor function corresponding to the given sourceMatcher.
// It defaults to a RemoteAddrStrategy IPStrategy if need be.
// It returns an error if more than one source criterion is provided.
func GetSourceExtractor(ctx context.Context, sourceMatcher *dynamic.SourceCriterion) (utils.SourceExtractor, error) {
	if sourceMatcher != nil {
		if sourceMatcher.IPStrategy != nil && sourceMatcher.RequestHeaderName != "" {
			return nil, errors.New("iPStrategy and RequestHeaderName are mutually exclusive")
		}
		if sourceMatcher.IPStrategy != nil && sourceMatcher.RequestHost {
			return nil, errors.New("iPStrategy and RequestHost are mutually exclusive")
		}
		if sourceMatcher.RequestHeaderName != "" && sourceMatcher.RequestHost {
			return nil, errors.New("requestHost and RequestHeaderName are mutually exclusive")
		}
	}

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
