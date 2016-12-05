package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/types"
)

func TestForwarder(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Get("emailField") == "" {
			t.Errorf("Missing forward request parameters. Informed just: %v", r.URL.Query())
		}

		fmt.Println("Chamou o servidor")
		fmt.Fprintln(w, "{ \"user\" : { \"id\" : 100, \"name\": \"John Lennon\" }}")

	}))
	defer ts.Close()

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-User-Id") != "100" {
			t.Errorf("Missing replay header X-User-Id. Headers: %v", r.Header)
		}

		if r.URL.Query().Get("name") != "John Lennon" {
			t.Errorf("Missing replay parameter name. Parameters: %v", r.URL.Query())
		}

		nextCalled = true
	}

	req := httptest.NewRequest("GET", "http://example.com/foo?email=john@beatles.com", nil)
	w := httptest.NewRecorder()

	forward := types.Forward{}
	forward.Address = ts.URL
	forward.RequestParameters = map[string]*types.ForwardRequestParameter{
		"email": {
			Name: "email",
			As:   "emailField",
		},
	}
	forward.ResponseReplayFields = map[string]*types.ResponseReplayField{
		"user": {
			Path: "user.id",
			As:   "X-User-Id",
			In:   "header",
		},
		"name": {
			Path: "user.name",
			As:   "name",
			In:   "parameter",
		},
	}

	Forward(&forward, w, req, next)

	if !nextCalled {
		t.Error("Next not called")
	}

}
