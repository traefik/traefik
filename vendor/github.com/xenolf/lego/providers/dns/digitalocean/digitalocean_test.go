package digitalocean

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeDigitalOceanAuth = "asdf1234"

func TestDigitalOceanPresent(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Expected method to be '%s' but got '%s'", want, got)
		}
		if got, want := r.URL.Path, "/v2/domains/example.com/records"; got != want {
			t.Errorf("Expected path to be '%s' but got '%s'", want, got)
		}
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("Expected Content-Type to be '%s' but got '%s'", want, got)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer asdf1234"; got != want {
			t.Errorf("Expected Authorization to be '%s' but got '%s'", want, got)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		if got, want := string(reqBody), `{"type":"TXT","name":"_acme-challenge.example.com.","data":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"}`; got != want {
			t.Errorf("Expected body data to be: `%s` but got `%s`", want, got)
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
			"domain_record": {
				"id": 1234567,
				"type": "TXT",
				"name": "_acme-challenge",
				"data": "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
				"priority": null,
				"port": null,
				"weight": null
			}
		}`)
	}))
	defer mock.Close()
	digitalOceanBaseURL = mock.URL

	doprov, err := NewDNSProviderCredentials(fakeDigitalOceanAuth)
	if doprov == nil {
		t.Fatal("Expected non-nil DigitalOcean provider, but was nil")
	}
	if err != nil {
		t.Fatalf("Expected no error creating provider, but got: %v", err)
	}

	err = doprov.Present("example.com", "", "foobar")
	if err != nil {
		t.Fatalf("Expected no error creating TXT record, but got: %v", err)
	}
	if !requestReceived {
		t.Error("Expected request to be received by mock backend, but it wasn't")
	}
}

func TestDigitalOceanCleanUp(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		if got, want := r.Method, "DELETE"; got != want {
			t.Errorf("Expected method to be '%s' but got '%s'", want, got)
		}
		if got, want := r.URL.Path, "/v2/domains/example.com/records/1234567"; got != want {
			t.Errorf("Expected path to be '%s' but got '%s'", want, got)
		}
		// NOTE: Even though the body is empty, DigitalOcean API docs still show setting this Content-Type...
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("Expected Content-Type to be '%s' but got '%s'", want, got)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer asdf1234"; got != want {
			t.Errorf("Expected Authorization to be '%s' but got '%s'", want, got)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer mock.Close()
	digitalOceanBaseURL = mock.URL

	doprov, err := NewDNSProviderCredentials(fakeDigitalOceanAuth)
	if doprov == nil {
		t.Fatal("Expected non-nil DigitalOcean provider, but was nil")
	}
	if err != nil {
		t.Fatalf("Expected no error creating provider, but got: %v", err)
	}

	doprov.recordIDsMu.Lock()
	doprov.recordIDs["_acme-challenge.example.com."] = 1234567
	doprov.recordIDsMu.Unlock()

	err = doprov.CleanUp("example.com", "", "")
	if err != nil {
		t.Fatalf("Expected no error removing TXT record, but got: %v", err)
	}
	if !requestReceived {
		t.Error("Expected request to be received by mock backend, but it wasn't")
	}
}
