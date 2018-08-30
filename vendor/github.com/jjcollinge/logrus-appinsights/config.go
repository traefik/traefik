package logrus_appinsights

import "time"

// Config for Application Insights settings
type Config struct {
	InstrumentationKey string
	EndpointUrl        string
	MaxBatchSize       int
	MaxBatchInterval   time.Duration
}
