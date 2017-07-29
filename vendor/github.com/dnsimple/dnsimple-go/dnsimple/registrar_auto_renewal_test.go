package dnsimple

import (
	"net/http"
	"testing"
)

func TestRegistrarService_EnableDomainAutoRenewal(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/registrar/domains/example.com/auto_renewal", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/enableDomainAutoRenewal/success.http")

		testMethod(t, r, "PUT")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
	})

	accountID := "1010"

	_, err := client.Registrar.EnableDomainAutoRenewal(accountID, "example.com")
	if err != nil {
		t.Fatalf("Registrars.EnableDomainAutoRenewal() returned error: %v", err)
	}
}

func TestRegistrarService_DisableDomainAutoRenewal(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/registrar/domains/example.com/auto_renewal", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/disableDomainAutoRenewal/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
	})

	accountID := "1010"

	_, err := client.Registrar.DisableDomainAutoRenewal(accountID, "example.com")
	if err != nil {
		t.Fatalf("Registrars.DisableDomainAutoRenewal() returned error: %v", err)
	}
}
