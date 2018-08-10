package appinsights

import "time"

type TelemetryConfiguration struct {
	InstrumentationKey string
	EndpointUrl        string
	MaxBatchSize       int
	MaxBatchInterval   time.Duration
}

func NewTelemetryConfiguration(instrumentationKey string) *TelemetryConfiguration {
	return &TelemetryConfiguration{
		InstrumentationKey: instrumentationKey,
		EndpointUrl:        "https://dc.services.visualstudio.com/v2/track",
		MaxBatchSize:       1024,
		MaxBatchInterval:   time.Duration(10) * time.Second,
	}
}
