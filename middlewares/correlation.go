package middlewares

import (
	"github.com/containous/traefik/types"
	"gitlab.com/JanMa/correlation"
)

// NewCorrelation constructs a new Correlation instance with supplied options
func NewCorrelation(headers *types.Headers) *correlation.Correlation {
	if headers == nil || !headers.HasCorrelationHeadersDefined() {
		return nil
	}

	cType := correlation.UUID
	switch headers.CorrelationIDType {
	case "CUID":
		cType = correlation.CUID
	case "Random":
		cType = correlation.Random
	case "Custom":
		cType = correlation.Custom
	case "Time":
		cType = correlation.Time
	}

	opt := correlation.Options{
		CorrelationHeaderName:   headers.CorrelationHeaderName,
		CorrelationIDType:       cType,
		CorrelationCustomString: headers.CorrelationCustomString,
	}

	return correlation.New(opt)

}
