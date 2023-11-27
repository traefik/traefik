package plugins

import (
	"context"

	"github.com/http-wasm/http-wasm-host-go/api"
	"github.com/rs/zerolog"
)

func initWasmLogger(logger *zerolog.Logger) *WasmLogger {
	return &WasmLogger{
		logger: logger,
	}
}

// compile-time check to ensure ConsoleLogger implements api.Logger.
var _ api.Logger = WasmLogger{}

// WasmLogger is a convenience which writes anything above LogLevelInfo to os.Stdout.
type WasmLogger struct {
	logger *zerolog.Logger
}

// IsEnabled implements the same method as documented on api.Logger.
func (w WasmLogger) IsEnabled(level api.LogLevel) bool {
	return true
}

// Log implements the same method as documented on api.Logger.
func (w WasmLogger) Log(_ context.Context, level api.LogLevel, message string) {
	w.logger.WithLevel(zerolog.Level(level + 1)).Msg(message)
}
