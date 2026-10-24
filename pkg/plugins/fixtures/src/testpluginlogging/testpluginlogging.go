package testpluginlogging

import (
	"context"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

type Config struct{}

func CreateConfig() *Config {
	return &Config{}
}

func New(ctx context.Context, next http.Handler, _ *Config, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, "testpluginlogging")
	logger.Info().Str("phase", "initialization").Msg("plugin ready")
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		logger.Debug().Msg("plugin debug")
		logger.Info().Str("path", req.URL.Path).Int("count", 1).Msg("plugin request")
		logger.Warn().Msg("plugin warning")
		logger.Error().Msg("plugin error")
		fmt.Println("legacy output")
		next.ServeHTTP(rw, req)
	}), nil
}
