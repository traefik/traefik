package auroradns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeAuroraDNSUserId = "asdf1234"
var fakeAuroraDNSKey = "key"

func TestAuroraDNSPresent(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/zones" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
			return
		}

		requestReceived = true

		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Expected method to be '%s' but got '%s'", want, got)
		}

		if got, want := r.URL.Path, "/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records"; got != want {
			t.Errorf("Expected path to be '%s' but got '%s'", want, got)
		}

		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("Expected Content-Type to be '%s' but got '%s'", want, got)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}

		if got, want := string(reqBody),
			`{"type":"TXT","name":"_acme-challenge","content":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":300}`; got != want {

			t.Errorf("Expected body data to be: `%s` but got `%s`", want, got)
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
		      "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
		      "type": "TXT",
		      "name": "_acme-challenge",
		      "ttl":  300
		    }`)
	}))

	defer mock.Close()

	auroraProvider, err := NewDNSProviderCredentials(mock.URL, fakeAuroraDNSUserId, fakeAuroraDNSKey)
	if auroraProvider == nil {
		t.Fatal("Expected non-nil AuroraDNS provider, but was nil")
	}

	if err != nil {
		t.Fatalf("Expected no error creating provider, but got: %v", err)
	}

	err = auroraProvider.Present("example.com", "", "foobar")
	if err != nil {
		t.Fatalf("Expected no error creating TXT record, but got: %v", err)
	}

	if !requestReceived {
		t.Error("Expected request to be received by mock backend, but it wasn't")
	}
}

func TestAuroraDNSCleanUp(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/zones" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{
			        "id":   "ec56a4180-65aa-42ec-a945-5fd21dec0538",
			        "type": "TXT",
			        "name": "_acme-challenge",
			        "ttl":  300
			      }`)
			return
		}

		requestReceived = true

		if got, want := r.Method, "DELETE"; got != want {
			t.Errorf("Expected method to be '%s' but got '%s'", want, got)
		}

		if got, want := r.URL.Path,
			"/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records/ec56a4180-65aa-42ec-a945-5fd21dec0538"; got != want {
			t.Errorf("Expected path to be '%s' but got '%s'", want, got)
		}

		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("Expected Content-Type to be '%s' but got '%s'", want, got)
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{}`)
	}))
	defer mock.Close()

	auroraProvider, err := NewDNSProviderCredentials(mock.URL, fakeAuroraDNSUserId, fakeAuroraDNSKey)
	if auroraProvider == nil {
		t.Fatal("Expected non-nil AuroraDNS provider, but was nil")
	}

	if err != nil {
		t.Fatalf("Expected no error creating provider, but got: %v", err)
	}

	err = auroraProvider.Present("example.com", "", "foobar")
	if err != nil {
		t.Fatalf("Expected no error creating TXT record, but got: %v", err)
	}

	err = auroraProvider.CleanUp("example.com", "", "foobar")
	if err != nil {
		t.Fatalf("Expected no error removing TXT record, but got: %v", err)
	}

	if !requestReceived {
		t.Error("Expected request to be received by mock backend, but it wasn't")
	}
}
