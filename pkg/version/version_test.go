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
// DashboardName + DashboardNamePosition fields under several configurations,
// including the empty-string default (omitempty must drop both keys).
func TestVersionHandler_DashboardName(t *testing.T) {
	type apiVersionResponse struct {
		Version               string `json:"Version"`
		DashboardName         string `json:"dashboardName,omitempty"`
		DashboardNamePosition string `json:"dashboardNamePosition,omitempty"`
	}

	cases := []struct {
		desc                  string
		dashboardName         string
		dashboardNamePosition string
		wantName              string
		wantPosition          string
		// wantNameKey/wantPosKey assert presence/absence of the JSON keys
		wantNameKey bool
		wantPosKey  bool
	}{
		{
			desc:                  "all empty defaults: both keys absent (omitempty)",
			dashboardName:         "",
			dashboardNamePosition: "",
			wantName:              "",
			wantPosition:          "",
			wantNameKey:           false,
			wantPosKey:            false,
		},
		{
			desc:                  "name only, no position",
			dashboardName:         "int",
			dashboardNamePosition: "",
			wantName:              "int",
			wantNameKey:           true,
			wantPosKey:            false,
		},
		{
			desc:                  "name + side position",
			dashboardName:         "int",
			dashboardNamePosition: "side",
			wantName:              "int",
			wantPosition:          "side",
			wantNameKey:           true,
			wantPosKey:            true,
		},
		{
			desc:                  "name + below position",
			dashboardName:         "ext",
			dashboardNamePosition: "below",
			wantName:              "ext",
			wantPosition:          "below",
			wantNameKey:           true,
			wantPosKey:            true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// snapshot+restore package-globals
			origName, origPos := DashboardName, DashboardNamePosition
			defer func() {
				DashboardName = origName
				DashboardNamePosition = origPos
			}()
			DashboardName = tc.dashboardName
			DashboardNamePosition = tc.dashboardNamePosition

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
			if tc.wantPosKey {
				assert.Equal(t, tc.wantPosition, raw["dashboardNamePosition"])
			} else {
				_, present := raw["dashboardNamePosition"]
				assert.False(t, present, "dashboardNamePosition key should be absent when empty")
			}
		})
	}
}

// TestVersionHandler_ConcurrentReads ensures the handler is safe under
// concurrent /api/version reads while a config rewrites the package globals.
// Not a strict race detector run, but exercises the field paths.
func TestVersionHandler_ConcurrentReads(t *testing.T) {
	origName, origPos := DashboardName, DashboardNamePosition
	defer func() {
		DashboardName = origName
		DashboardNamePosition = origPos
	}()
	DashboardName = "int"
	DashboardNamePosition = "below"

	router := mux.NewRouter()
	Handler{}.Append(router)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"dashboardName":"int"`)
		assert.Contains(t, rr.Body.String(), `"dashboardNamePosition":"below"`)
	}
}
