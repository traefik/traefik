package appinsights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

func (items TelemetryBufferItems) serialize() []byte {
	var result bytes.Buffer
	encoder := json.NewEncoder(&result)

	for _, item := range items {
		end := result.Len()
		if err := encoder.Encode(prepare(item)); err != nil {
			diagnosticsWriter.Write(fmt.Sprintf("Telemetry item failed to serialize: %s", err.Error()))
			result.Truncate(end)
		}
	}

	return result.Bytes()
}

func prepare(item Telemetry) *envelope {
	data := &data{
		BaseType: item.baseTypeName() + "Data",
		BaseData: item.baseData(),
	}

	context := item.Context()

	envelope := &envelope{
		Name: "Microsoft.ApplicationInsights." + item.baseTypeName(),
		Time: item.Timestamp().Format(time.RFC3339),
		IKey: context.InstrumentationKey(),
		Data: data,
	}

	if tcontext, ok := context.(*telemetryContext); ok {
		envelope.Tags = tcontext.tags
	}

	return envelope
}
