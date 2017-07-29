package dnsimple

import (
	"io"
	"net/http"
	"testing"
)

func TestVanityNameServers_vanityNameServerPath(t *testing.T) {
	if want, got := "/1010/vanity/example.com", vanityNameServerPath("1010", "example.com"); want != got {
		t.Errorf("vanity_name_serverPath(%v,  ) = %v, want %v", "1010", got, want)
	}
}

func TestVanityNameServersService_EnableVanityNameServers(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/vanity/example.com", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/enableVanityNameServers/success.http")

		testMethod(t, r, "PUT")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	vanityNameServerResponse, err := client.VanityNameServers.EnableVanityNameServers("1010", "example.com")
	if err != nil {
		t.Fatalf("VanityNameServers.EnableVanityNameServers() returned error: %v", err)
	}

	delegation := vanityNameServerResponse.Data[0].Name
	wantSingle := "ns1.example.com"

	if delegation != wantSingle {
		t.Fatalf("VanityNameServers.EnableVanityNameServers() returned %+v, want %+v", delegation, wantSingle)
	}
}

func TestVanityNameServersService_DisableVanityNameServers(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/vanity/example.com", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/disableVanityNameServers/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.VanityNameServers.DisableVanityNameServers("1010", "example.com")
	if err != nil {
		t.Fatalf("VanityNameServers.DisableVanityNameServers() returned error: %v", err)
	}
}
