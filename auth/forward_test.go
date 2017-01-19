package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"github.com/containous/traefik/types"
)

func TestForwarder(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Get("emailField") == "" {
			t.Errorf("Missing forward request parameters. Informed just: %v", r.URL.Query())
		}

		if r.URL.Query().Get("idField") == "" {
			t.Errorf("Missing forward request parameters. Informed just: %v", r.URL.Query())
		}

		if r.Header.Get("tokenField") == "" {
			t.Errorf("Missing forward request parameter from header that was suppose to be transformed. Informed just: %v", r.Header)
		}

		if r.Header.Get("some_header") == "" {
			t.Errorf("Missing forward request parameters from header that was supposed to be forwarded as is. Informed just: %v", r.Header)
		}
		cookie, err := r.Cookie("test")
		if err != nil {
			t.Errorf("Error getting forward request cookie test. Error: %s", err)
		}

		if cookie.Value != "data" {
			t.Errorf("The forward cookie is invalid: %v", cookie)
		}

		cookie, err = r.Cookie("authTokenCookie")
		if err != nil {
			t.Errorf("Error getting forward transformation request cookie test. Error: %s", err)
		}

		if cookie.Value != "abc123" {
			t.Errorf("The forward transformed cookie is invalid: %v", cookie)
		}

		fmt.Println("Chamou o servidor")
		fmt.Fprintln(w, "{ \"user\" : { \"id\" : 100, \"name\": \"John Lennon\", \"accounts\": [\"first\", \"second\"] }}")

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

		if r.Header.Get("X-User-Accounts") == "" {
			t.Errorf("Missing replay header X-User-Accounts. Headers: %v", r.Header)
		}
		var accounts []string
		err := json.Unmarshal([]byte(r.Header.Get("X-User-Accounts")), &accounts)
		if err != nil {
			t.Errorf("Couldn't Unmarshal accounts got an error [ %v ] for input [ %s ]", err, r.Header.Get("X-User-Accounts"))
		}
		if len(accounts) != 2 {
			t.Errorf("got an invalid amount of accounts %d while expecing 2 obj: %v", len(accounts), accounts)
		}

		nextCalled = true
	}

	req := httptest.NewRequest("GET", "http://example.com/foo?email=john@beatles.com&id=xyz", nil)
	req.Header.Set("token", "abc")
	req.Header.Set("some_header", "some_data")
	cookie := &http.Cookie{
		Name:   "test",
		Value:  "data",
		Path:   "/test",
		Domain: "www.example.com",
	}
	req.AddCookie(cookie)

	cookie = &http.Cookie{
		Name:   "auth_token",
		Value:  "abc123",
		Domain: ".example.com",
	}
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	forward := types.Forward{}
	forward.Address = ts.URL
	forward.ForwardAllHeaders = true
	forward.RequestParameters = map[string]*types.ForwardRequestParameter{
		"email": {
			Name: "email",
			As:   "emailField",
			In:   "parameter",
		},
		"token": {
			Name: "token",
			As:   "tokenField",
			In:   "header",
		},
		"id": {
			Name: "id",
			As:   "idField",
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
		"accounts": {
			Path: "user.accounts",
			As:   "X-User-Accounts",
			In:   "header",
		},
	}

	forward.RequestCookies = map[string]*types.ForwardRequestCookie{
		"auth_token": {
			Name: "auth_token",
			As:   "authTokenCookie",
		},
	}

	Forward(&forward, w, req, next)

	if !nextCalled {
		t.Error("Next not called")
	}

}
