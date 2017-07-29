package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestDomainServices_domainServicesPath(t *testing.T) {
	if want, got := "/1010/domains/example.com/services", domainServicesPath("1010", "example.com", ""); want != got {
		t.Errorf("domainServicesPath(%v, %v, ) = %v, want %v", "1010", "example.com", got, want)
	}

	if want, got := "/1010/domains/example.com/services/1", domainServicesPath("1010", "example.com", "1"); want != got {
		t.Errorf("domainServicesPath(%v, %v, 1) = %v, want %v", "1010", "example.com", got, want)
	}
}

func TestServicesService_AppliedServices(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/services", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/appliedServices/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)
		testQuery(t, r, url.Values{})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	servicesResponse, err := client.Services.AppliedServices("1010", "example.com", nil)
	if err != nil {
		t.Fatalf("DomainServices.AppliedServices() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 1}), servicesResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("DomainServices.AppliedServices() pagination expected to be %v, got %v", want, got)
	}

	services := servicesResponse.Data
	if want, got := 1, len(services); want != got {
		t.Errorf("DomainServices.AppliedServices() expected to return %v services, got %v", want, got)
	}

	if want, got := 1, services[0].ID; want != got {
		t.Fatalf("DomainServices.AppliedServices() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "wordpress", services[0].SID; want != got {
		t.Fatalf("DomainServices.AppliedServices() returned ShortName expected to be `%v`, got `%v`", want, got)
	}
}

func TestServicesService_ApplyService(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/services/service1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/applyService/success.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	settings := DomainServiceSettings{Settings: map[string]string{"app": "foo"}}

	_, err := client.Services.ApplyService("1010", "service1", "example.com", settings)
	if err != nil {
		t.Fatalf("DomainServices.ApplyService() returned error: %v", err)
	}
}

func TestServicesService_UnapplyService(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/services/service1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/unapplyService/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Services.UnapplyService("1010", "service1", "example.com")
	if err != nil {
		t.Fatalf("DomainServices.UnapplyService() returned error: %v", err)
	}
}
