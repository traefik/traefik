package accesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatFieldHandler_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc           string
		existingValue  interface{}
		newValue       string
		expectedResult string
	}{
		{
			desc:           "first router - no existing value",
			existingValue:  nil,
			newValue:       "router1",
			expectedResult: "router1",
		},
		{
			desc:           "second router - concatenate with existing string",
			existingValue:  "router1",
			newValue:       "router2",
			expectedResult: "router1 -> router2",
		},
		{
			desc:           "third router - concatenate with existing chain",
			existingValue:  "router1 -> router2",
			newValue:       "router3",
			expectedResult: "router1 -> router2 -> router3",
		},
		{
			desc:           "empty existing value - treat as first",
			existingValue:  "    ",
			newValue:       "router1",
			expectedResult: "router1",
		},
		{
			desc:           "non-string existing value - replace with new value",
			existingValue:  123,
			newValue:       "router1",
			expectedResult: "router1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			logData := &LogData{
				Core: CoreLogData{},
			}
			if test.existingValue != nil {
				logData.Core[RouterName] = test.existingValue
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), DataTableKey, logData))

			handler := NewConcatFieldHandler(nextHandler, RouterName, test.newValue)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, test.expectedResult, logData.Core[RouterName])
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestConcatFieldHandler_ServeHTTP_NoLogData(t *testing.T) {
	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := NewConcatFieldHandler(nextHandler, RouterName, "router1")

	// Create request without LogData in context.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify next handler was called and no panic occurred.
	assert.True(t, nextHandlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}
