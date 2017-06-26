package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		name               string
		err                error
		wantHTTPStatus     int
		wantNetErrRecorded bool
	}{
		{
			name:               "net.Error",
			err:                net.UnknownNetworkError("any network error"),
			wantHTTPStatus:     http.StatusBadGateway,
			wantNetErrRecorded: true,
		},
		{
			name:               "net.Error with Timeout",
			err:                &timeoutError{},
			wantHTTPStatus:     http.StatusGatewayTimeout,
			wantNetErrRecorded: true,
		},
		{
			name:               "io.EOF",
			err:                io.EOF,
			wantHTTPStatus:     http.StatusBadGateway,
			wantNetErrRecorded: true,
		},
		{
			name:               "custom error",
			err:                errors.New("any error"),
			wantHTTPStatus:     http.StatusInternalServerError,
			wantNetErrRecorded: false,
		},
		{
			name:               "nil error",
			err:                nil,
			wantHTTPStatus:     http.StatusInternalServerError,
			wantNetErrRecorded: false,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()

			errorRecorder := &netErrorRecorder{}
			req := httptest.NewRequest(http.MethodGet, "http://localhost:3000/any", nil)

			recordingErrorHandler := NewRecordingErrorHandler(errorRecorder)
			recordingErrorHandler.ServeHTTP(recorder, req, test.err)

			if recorder.Code != test.wantHTTPStatus {
				t.Errorf("got HTTP status code %v, wanted %v", recorder.Code, test.wantHTTPStatus)
			}
			if errorRecorder.netErrorWasRecorded != test.wantNetErrRecorded {
				t.Errorf("net error recording wrong, got %v wanted %v", errorRecorder.netErrorWasRecorded, test.wantNetErrRecorded)
			}
		})
	}
}

type netErrorRecorder struct {
	netErrorWasRecorded bool
}

func (recorder *netErrorRecorder) Record(ctx context.Context) {
	recorder.netErrorWasRecorded = true
}
