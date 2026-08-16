package logs

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/samyfodil/wazy/imports/http_handler"
)

// compile-time check to ensure ConsoleLogger implements http_handler.Logger.
var _ http_handler.Logger = WasmLogger{}

// WasmLogger is a convenience which writes anything above LogLevelInfo to os.Stdout.
type WasmLogger struct {
	logger *zerolog.Logger
}

func NewWasmLogger(logger *zerolog.Logger) *WasmLogger {
	return &WasmLogger{
		logger: logger,
	}
}

// IsEnabled implements the same method as documented on http_handler.Logger.
func (w WasmLogger) IsEnabled(level http_handler.LogLevel) bool {
	return true
}

// Log implements the same method as documented on http_handler.Logger.
func (w WasmLogger) Log(_ context.Context, level http_handler.LogLevel, message string) {
	w.logger.WithLevel(zerolog.Level(level + 1)).Msg(message)
}
