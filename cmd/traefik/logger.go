package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	// hide the first logs before the setup of the logger.
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func setupLogger(ctx context.Context, staticConfiguration *static.Configuration) (func() error, error) {
	// Validate that the experimental flag is set up at this point,
	// rather than validating the static configuration before the setupLogger call.
	// This ensures that validation messages are not logged using an un-configured logger.
	if staticConfiguration.Log != nil && staticConfiguration.Log.OTLP != nil &&
		(staticConfiguration.Experimental == nil || !staticConfiguration.Experimental.OTLPLogs) {
		return nil, errors.New("the experimental OTLPLogs feature must be enabled to use OTLP logging")
	}

	// configure log format
	w, rotateFn, err := getLogWriter(staticConfiguration)
	if err != nil {
		return nil, err
	}

	// configure log level
	logLevel := getLogLevel(staticConfiguration)
	zerolog.SetGlobalLevel(logLevel)

	// create logger
	logger := zerolog.New(w).With().Timestamp()
	if logLevel <= zerolog.DebugLevel {
		logger = logger.Caller()
	}

	log.Logger = logger.Logger().Level(logLevel)

	if staticConfiguration.Log != nil && staticConfiguration.Log.OTLP != nil {
		var err error
		log.Logger, err = logs.SetupOTelLogger(ctx, log.Logger, staticConfiguration.Log.OTLP)
		if err != nil {
			return nil, fmt.Errorf("setting up OpenTelemetry logger: %w", err)
		}
	}

	zerolog.DefaultContextLogger = &log.Logger

	// Global logrus replacement (related to lib like go-rancher-metadata, docker, etc.)
	logrus.StandardLogger().Out = logs.NoLevel(log.Logger, zerolog.DebugLevel)

	// configure default standard log.
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)
	stdlog.SetOutput(logs.NoLevel(log.Logger, zerolog.DebugLevel))

	return rotateFn, nil
}

func getLogWriter(staticConfiguration *static.Configuration) (io.Writer, func() error, error) {
	var w io.Writer = os.Stdout
	var rotateFn func() error
	if staticConfiguration.Log != nil && len(staticConfiguration.Log.FilePath) > 0 {
		f, err := os.OpenFile(staticConfiguration.Log.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("opening log file %q: %w", staticConfiguration.Log.FilePath, err)
		}

		if err := f.Close(); err != nil {
			return nil, nil, fmt.Errorf("closing log file %q: %w", staticConfiguration.Log.FilePath, err)
		}

		lj := &lumberjack.Logger{
			Filename:   staticConfiguration.Log.FilePath,
			MaxSize:    staticConfiguration.Log.MaxSize,
			MaxBackups: staticConfiguration.Log.MaxBackups,
			MaxAge:     staticConfiguration.Log.MaxAge,
			Compress:   staticConfiguration.Log.Compress,
		}
		w = lj
		rotateFn = lj.Rotate
	}

	if staticConfiguration.Log == nil || staticConfiguration.Log.Format != "json" {
		w = zerolog.ConsoleWriter{
			Out:        w,
			TimeFormat: time.RFC3339,
			NoColor:    staticConfiguration.Log != nil && (staticConfiguration.Log.NoColor || len(staticConfiguration.Log.FilePath) > 0),
		}
	}

	return w, rotateFn, nil
}

func getLogLevel(staticConfiguration *static.Configuration) zerolog.Level {
	levelStr := "error"
	if staticConfiguration.Log != nil && staticConfiguration.Log.Level != "" {
		levelStr = strings.ToLower(staticConfiguration.Log.Level)
	}

	logLevel, err := zerolog.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		log.Error().Err(err).
			Str("logLevel", levelStr).
			Msg("Unspecified or invalid log level, setting the level to default (ERROR)...")

		logLevel = zerolog.ErrorLevel
	}

	return logLevel
}
