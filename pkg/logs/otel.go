package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/types"
	otellog "go.opentelemetry.io/otel/log"
)

// SetupOTelLogger sets up the OpenTelemetry logger.
func SetupOTelLogger(ctx context.Context, logger zerolog.Logger, config *types.OTelLog) (zerolog.Logger, error) {
	if config == nil {
		return logger, nil
	}

	provider, err := config.NewLoggerProvider(ctx)
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("setting up OpenTelemetry logger provider: %w", err)
	}

	return logger.Hook(&otelLoggerHook{logger: provider.Logger("traefik")}), nil
}

// otelLoggerHook is a zerolog hook that forwards logs to OpenTelemetry.
type otelLoggerHook struct {
	logger otellog.Logger
}

// Run forwards the log message to OpenTelemetry.
func (h *otelLoggerHook) Run(e *zerolog.Event, level zerolog.Level, message string) {
	if level == zerolog.Disabled {
		return
	}

	// Discard the event to avoid double logging.
	e.Discard()

	var record otellog.Record
	record.SetTimestamp(time.Now().UTC())
	record.SetSeverity(otelLogSeverity(level))
	record.SetBody(otellog.StringValue(message))

	// See https://github.com/rs/zerolog/issues/493.
	// This is a workaround to get the log fields from the event.
	// At the moment there's no way to get the log fields from the event, so we use reflection to get the buffer and parse it.
	logData := make(map[string]any)
	eventBuffer := fmt.Sprintf("%s}", reflect.ValueOf(e).Elem().FieldByName("buf"))
	if err := json.Unmarshal([]byte(eventBuffer), &logData); err != nil {
		record.AddAttributes(otellog.String("parsing_error", fmt.Sprintf("parsing log fields: %s", err)))
		h.logger.Emit(e.GetCtx(), record)
		return
	}

	recordAttributes := make([]otellog.KeyValue, 0, len(logData))
	for k, v := range logData {
		if k == "level" {
			continue
		}
		if k == "time" {
			eventTimestamp, ok := v.(string)
			if !ok {
				continue
			}
			t, err := time.Parse(time.RFC3339, eventTimestamp)
			if err == nil {
				record.SetTimestamp(t)
				continue
			}
		}
		var attributeValue otellog.Value
		switch v := v.(type) {
		case string:
			attributeValue = otellog.StringValue(v)
		case int:
			attributeValue = otellog.IntValue(v)
		case int64:
			attributeValue = otellog.Int64Value(v)
		case float64:
			attributeValue = otellog.Float64Value(v)
		case bool:
			attributeValue = otellog.BoolValue(v)
		case []byte:
			attributeValue = otellog.BytesValue(v)
		default:
			attributeValue = otellog.StringValue(fmt.Sprintf("%v", v))
		}
		recordAttributes = append(recordAttributes, otellog.KeyValue{
			Key:   k,
			Value: attributeValue,
		})
	}
	record.AddAttributes(recordAttributes...)

	h.logger.Emit(e.GetCtx(), record)
}

func otelLogSeverity(level zerolog.Level) otellog.Severity {
	switch level {
	case zerolog.TraceLevel:
		return otellog.SeverityTrace
	case zerolog.DebugLevel:
		return otellog.SeverityDebug
	case zerolog.InfoLevel:
		return otellog.SeverityInfo
	case zerolog.WarnLevel:
		return otellog.SeverityWarn
	case zerolog.ErrorLevel:
		return otellog.SeverityError
	case zerolog.FatalLevel:
		return otellog.SeverityFatal
	case zerolog.PanicLevel:
		return otellog.SeverityFatal4
	default:
		return otellog.SeverityUndefined
	}
}
