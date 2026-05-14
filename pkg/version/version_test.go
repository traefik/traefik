package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionHandler_DashboardName exercises the /api/version JSON envelope's
// DashboardName field, including the empty-string default (omitempty must drop
// the key).
func TestVersionHandler_DashboardName(t *testing.T) {
	cases := []struct {
		desc          string
		dashboardName string
		wantName      string
		// wantNameKey asserts presence/absence of the JSON key
		wantNameKey bool
	}{
		{
			desc:          "empty default: key absent (omitempty)",
			dashboardName: "",
			wantName:      "",
			wantNameKey:   false,
		},
		{
			desc:          "name set",
			dashboardName: "int",
			wantName:      "int",
			wantNameKey:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// snapshot+restore package-global
			origName := DashboardName
			defer func() {
				DashboardName = origName
			}()
			DashboardName = tc.dashboardName

			router := mux.NewRouter()
			Handler{}.Append(router)

			req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			// raw JSON parse to assert key presence vs absence
			var raw map[string]any
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &raw))
			if tc.wantNameKey {
				assert.Equal(t, tc.wantName, raw["dashboardName"])
			} else {
				_, present := raw["dashboardName"]
				assert.False(t, present, "dashboardName key should be absent when empty")
			}
		})
	}
}

// TestVersionHandler_ConcurrentReads ensures the handler is safe under
// concurrent /api/version reads while a config rewrites the package global.
// Not a strict race detector run, but exercises the field path.
func TestVersionHandler_ConcurrentReads(t *testing.T) {
	origName := DashboardName
	defer func() {
		DashboardName = origName
	}()
	DashboardName = "int"

	router := mux.NewRouter()
	Handler{}.Append(router)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"dashboardName":"int"`)
	}
}
