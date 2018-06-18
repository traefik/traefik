package pipelining

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPipelining(t *testing.T) {
	testCases := []struct {
		desc                   string
		HTTPMethod             string
		implementCloseNotifier bool
	}{
		{
			desc:                   "should not implement CloseNotifier with GET method",
			HTTPMethod:             http.MethodGet,
			implementCloseNotifier: false,
		},
		{
			desc:                   "should implement CloseNotifier with PUT method",
			HTTPMethod:             http.MethodPut,
			implementCloseNotifier: true,
		},
		{
			desc:                   "should implement CloseNotifier with POST method",
			HTTPMethod:             http.MethodPost,
			implementCloseNotifier: true,
		},
		{
			desc:                   "should  not implement CloseNotifier with GET method",
			HTTPMethod:             http.MethodHead,
			implementCloseNotifier: false,
		},
		{
			desc:                   "should  not implement CloseNotifier with PROPFIND method",
			HTTPMethod:             "PROPFIND",
			implementCloseNotifier: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, ok := w.(http.CloseNotifier)
				assert.Equal(t, test.implementCloseNotifier, ok)
				w.WriteHeader(http.StatusOK)
			})
			handler := NewPipelining(nextHandler)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(test.HTTPMethod, "http://localhost", nil)

			handler.ServeHTTP(recorder, req)
		})
	}
}
