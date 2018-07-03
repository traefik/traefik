package tracer

import (
	"fmt"
	"log"
	"strconv"
)

var errorPrefix = fmt.Sprintf("Datadog Tracer Error (%s): ", tracerVersion)

type traceEncodingError struct{ context error }

func (e *traceEncodingError) Error() string {
	return fmt.Sprintf("error encoding trace: %s", e.context)
}

type spanBufferFullError struct{}

func (e *spanBufferFullError) Error() string {
	return fmt.Sprintf("trace span cap (%d) reached, dropping trace", traceMaxSize)
}

type dataLossError struct {
	count   int   // number of items lost
	context error // any context error, if available
}

func (e *dataLossError) Error() string {
	return fmt.Sprintf("lost traces (count: %d), error: %v", e.count, e.context)
}

type errorSummary struct {
	Count   int
	Example string
}

func aggregateErrors(errChan <-chan error) map[string]errorSummary {
	errs := make(map[string]errorSummary, len(errChan))
	for {
		select {
		case err := <-errChan:
			if err == nil {
				break
			}
			key := fmt.Sprintf("%T", err)
			summary := errs[key]
			summary.Count++
			summary.Example = err.Error()
			errs[key] = summary
		default: // stop when there's no more data
			return errs
		}
	}
}

// logErrors logs the errors, preventing log file flooding, when there
// are many messages, it caps them and shows a quick summary.
// As of today it only logs using standard golang log package, but
// later we could send those stats to agent // TODO(ufoot).
func logErrors(errChan <-chan error) {
	errs := aggregateErrors(errChan)
	for _, v := range errs {
		var repeat string
		if v.Count > 1 {
			repeat = " (repeated " + strconv.Itoa(v.Count) + " times)"
		}
		log.Println(errorPrefix + v.Example + repeat)
	}
}
