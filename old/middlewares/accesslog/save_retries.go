package accesslog

import (
	"net/http"
)

// SaveRetries is an implementation of RetryListener that stores RetryAttempts in the LogDataTable.
type SaveRetries struct{}

// Retried implements the RetryListener interface and will be called for each retry that happens.
func (s *SaveRetries) Retried(req *http.Request, attempt int) {
	// it is the request attempt x, but the retry attempt is x-1
	if attempt > 0 {
		attempt--
	}

	table := GetLogDataTable(req)
	table.Core[RetryAttempts] = attempt
}
