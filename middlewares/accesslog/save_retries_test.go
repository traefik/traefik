package accesslog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaveRetries(t *testing.T) {
	tests := []struct {
		requestAttempt         int
		wantRetryAttemptsInLog int
	}{
		{
			requestAttempt:         0,
			wantRetryAttemptsInLog: 0,
		},
		{
			requestAttempt:         1,
			wantRetryAttemptsInLog: 0,
		},
		{
			requestAttempt:         3,
			wantRetryAttemptsInLog: 2,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(fmt.Sprintf("%d retries", test.requestAttempt), func(t *testing.T) {
			t.Parallel()
			saveRetries := &SaveRetries{}

			logDataTable := &LogData{Core: make(CoreLogData)}
			req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
			reqWithDataTable := req.WithContext(context.WithValue(req.Context(), DataTableKey, logDataTable))

			saveRetries.Retried(reqWithDataTable, test.requestAttempt)

			if logDataTable.Core[RetryAttempts] != test.wantRetryAttemptsInLog {
				t.Errorf("got %v in logDataTable, want %v", logDataTable.Core[RetryAttempts], test.wantRetryAttemptsInLog)
			}
		})
	}
}
