package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetry(t *testing.T) {
	testCases := []struct {
		failAtCalls    []int
		attempts       int
		responseStatus int
		listener       *countingRetryListener
		retriedCount   int
	}{
		{
			failAtCalls:    []int{1, 2},
			attempts:       3,
			responseStatus: http.StatusOK,
			listener:       &countingRetryListener{},
			retriedCount:   2,
		},
		{
			failAtCalls:    []int{1, 2},
			attempts:       2,
			responseStatus: http.StatusBadGateway,
			listener:       &countingRetryListener{},
			retriedCount:   1,
		},
	}

	for _, tc := range testCases {
		// bind tc locally
		tc := tc
		tcName := fmt.Sprintf("FailAtCalls(%v) RetryAttempts(%v)", tc.failAtCalls, tc.attempts)

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			var httpHandler http.Handler
			httpHandler = &networkFailingHTTPHandler{failAtCalls: tc.failAtCalls}
			httpHandler = NewRetry(tc.attempts, httpHandler, tc.listener)

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:3000/ok", ioutil.NopCloser(nil))
			if err != nil {
				t.Fatalf("could not create request: %+v", err)
			}

			httpHandler.ServeHTTP(recorder, req)

			if tc.responseStatus != recorder.Code {
				t.Errorf("wrong status code %d, want %d", recorder.Code, tc.responseStatus)
			}
			if tc.retriedCount != tc.listener.timesCalled {
				t.Errorf("RetryListener called %d times, want %d times", tc.listener.timesCalled, tc.retriedCount)
			}
		})
	}
}

// networkFailingHTTPHandler is an http.Handler implementation you can use to test retries.
type networkFailingHTTPHandler struct {
	failAtCalls []int
	callNumber  int
}

func (handler *networkFailingHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.callNumber++

	for _, failAtCall := range handler.failAtCalls {
		if handler.callNumber == failAtCall {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// countingRetryListener is a RetryListener implementation to count the times the Retried fn is called.
type countingRetryListener struct {
	timesCalled int
}

func (l *countingRetryListener) Retried(attempt int) {
	l.timesCalled++
}
