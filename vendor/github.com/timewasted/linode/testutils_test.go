package linode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type (
	LinodeResponse struct {
		Action string          `json:"ACTION"`
		Data   interface{}     `json:"DATA"`
		Errors []ResponseError `json:"ERRORARRAY"`
	}
	MockResponse struct {
		StatusCode int
		Response   interface{}
		Errors     []ResponseError
	}
	MockResponseMap map[string]MockResponse
)

func newMockServer(t *testing.T, responses MockResponseMap) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure that we support the requested action.
		action := r.URL.Query().Get("api_action")
		resp, ok := responses[action]
		if !ok {
			msg := fmt.Sprintf("Unsupported mock action: %s", action)
			require.FailNow(t, msg)
		}
		if resp.StatusCode == 0 {
			resp.StatusCode = http.StatusOK
		}

		// Build the response that the server will return.
		linodeResponse := LinodeResponse{
			Action: action,
			Data:   resp.Response,
			Errors: resp.Errors,
		}
		rawResponse, err := json.Marshal(linodeResponse)
		if err != nil {
			msg := fmt.Sprintf("Failed to JSON encode response: %v", err)
			require.FailNow(t, msg)
		}

		// Send the response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(rawResponse)
	}))

	time.Sleep(100 * time.Millisecond)
	return srv
}
