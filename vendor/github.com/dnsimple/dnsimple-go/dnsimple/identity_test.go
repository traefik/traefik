package dnsimple

import (
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestAuthService_Whoami(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/whoami", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/whoami/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	whoamiResponse, err := client.Identity.Whoami()
	if err != nil {
		t.Fatalf("Auth.Whoami() returned error: %v", err)
	}

	whoami := whoamiResponse.Data
	want := &WhoamiData{Account: &Account{ID: 1, Email: "example-account@example.com"}}
	if !reflect.DeepEqual(whoami, want) {
		t.Errorf("Auth.Whoami() returned %+v, want %+v", whoami, want)
	}
}
