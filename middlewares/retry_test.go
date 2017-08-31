package middlewares

import (
	"context"
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

			var httpHandler http.Handler = &networkFailingHTTPHandler{failAtCalls: tc.failAtCalls, netErrorRecorder: &DefaultNetErrorRecorder{}}
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

func TestDefaultNetErrorRecorderSuccess(t *testing.T) {
	boolNetErrorOccurred := false
	recorder := DefaultNetErrorRecorder{}
	recorder.Record(context.WithValue(context.Background(), defaultNetErrCtxKey, &boolNetErrorOccurred))
	if !boolNetErrorOccurred {
		t.Errorf("got %v after recording net error, wanted %v", boolNetErrorOccurred, true)
	}
}

func TestDefaultNetErrorRecorderInvalidValueType(t *testing.T) {
	stringNetErrorOccured := "nonsense"
	recorder := DefaultNetErrorRecorder{}
	recorder.Record(context.WithValue(context.Background(), defaultNetErrCtxKey, &stringNetErrorOccured))
	if stringNetErrorOccured != "nonsense" {
		t.Errorf("got %v after recording net error, wanted %v", stringNetErrorOccured, "nonsense")
	}
}

func TestDefaultNetErrorRecorderNilValue(t *testing.T) {
	nilNetErrorOccured := interface{}(nil)
	recorder := DefaultNetErrorRecorder{}
	recorder.Record(context.WithValue(context.Background(), defaultNetErrCtxKey, &nilNetErrorOccured))
	if nilNetErrorOccured != interface{}(nil) {
		t.Errorf("got %v after recording net error, wanted %v", nilNetErrorOccured, interface{}(nil))
	}
}

func TestRetryListeners(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	retryListeners := RetryListeners{&countingRetryListener{}, &countingRetryListener{}}

	retryListeners.Retried(req, 1)
	retryListeners.Retried(req, 1)

	for _, retryListener := range retryListeners {
		listener := retryListener.(*countingRetryListener)
		if listener.timesCalled != 2 {
			t.Errorf("retry listener was called %d times, want %d", listener.timesCalled, 2)
		}
	}
}

// networkFailingHTTPHandler is an http.Handler implementation you can use to test retries.
type networkFailingHTTPHandler struct {
	netErrorRecorder NetErrorRecorder
	failAtCalls      []int
	callNumber       int
}

func (handler *networkFailingHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.callNumber++

	for _, failAtCall := range handler.failAtCalls {
		if handler.callNumber == failAtCall {
			handler.netErrorRecorder.Record(r.Context())

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

func (l *countingRetryListener) Retried(req *http.Request, attempt int) {
	l.timesCalled++
}
