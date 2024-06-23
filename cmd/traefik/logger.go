package main

import (
	"io"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/logs"
)

func init() {
	// hide the first logs before the setup of the logger.
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func setupLogger(staticConfiguration *static.Configuration) {
	// configure log format
	w := getLogWriter(staticConfiguration)

	// configure log level
	logLevel := getLogLevel(staticConfiguration)

	// create logger
	logCtx := zerolog.New(w).With().Timestamp()
	if logLevel <= zerolog.DebugLevel {
		logCtx = logCtx.Caller()
	}

	log.Logger = logCtx.Logger().Level(logLevel)
	zerolog.DefaultContextLogger = &log.Logger
	zerolog.SetGlobalLevel(logLevel)

	// Global logrus replacement (related to lib like go-rancher-metadata, docker, etc.)
	logrus.StandardLogger().Out = logs.NoLevel(log.Logger, zerolog.DebugLevel)

	// configure default standard log.
	stdlog.SetFlags(stdlog.Lshortfile | stdlog.LstdFlags)
	stdlog.SetOutput(logs.NoLevel(log.Logger, zerolog.DebugLevel))
}

func getLogWriter(staticConfiguration *static.Configuration) io.Writer {
	var w io.Writer = os.Stderr

	if staticConfiguration.Log != nil && len(staticConfiguration.Log.FilePath) > 0 {
		_, _ = os.OpenFile(staticConfiguration.Log.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		w = &lumberjack.Logger{
			Filename:   staticConfiguration.Log.FilePath,
			MaxSize:    staticConfiguration.Log.MaxSize,
			MaxBackups: staticConfiguration.Log.MaxBackups,
			MaxAge:     staticConfiguration.Log.MaxAge,
			Compress:   true,
		}
	}

	if staticConfiguration.Log == nil || staticConfiguration.Log.Format != "json" {
		w = zerolog.ConsoleWriter{
			Out:        w,
			TimeFormat: time.RFC3339,
			NoColor:    staticConfiguration.Log != nil && (staticConfiguration.Log.NoColor || len(staticConfiguration.Log.FilePath) > 0),
		}
	}

	return w
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
