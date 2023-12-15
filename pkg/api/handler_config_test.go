package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// MockHandler is a mock for the Handler struct.
type MockHandler struct {
	Handler
}

// putLogLevel is a mock implementation for testing.
func (h *MockHandler) putLogLevel(rw http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	levelStr := vars["logLevel"]

	// Parse and set the log level.
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.SetLevel(level)
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("Log level set to " + levelStr))
}

func TestPutLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectedCode  int
		expectedBody  string
	}{
		{"Valid LogLevel Debug", "debug", http.StatusOK, "Log level set to debug"},
		{"Valid LogLevel Error", "error", http.StatusOK, "Log level set to error"},
		{"Invalid LogLevel", "invalid", http.StatusBadRequest, "not a valid logrus Level: \"invalid\""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("PUT", "/api/config/loglevel/"+tc.logLevel, nil)
			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			mockHandler := &MockHandler{}

			router.HandleFunc("/api/config/loglevel/{logLevel}", mockHandler.putLogLevel)
			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.Equal(t, tc.expectedBody, rr.Body.String())
		})
	}
}
