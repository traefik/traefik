package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestTldsService_ListTlds(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/tlds", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTlds/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	tldsResponse, err := client.Tlds.ListTlds(nil)
	if err != nil {
		t.Fatalf("Tlds.ListTlds() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 2, TotalPages: 98, TotalEntries: 195}), tldsResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Tlds.ListTlds() pagination expected to be %v, got %v", want, got)
	}

	tlds := tldsResponse.Data
	if want, got := 2, len(tlds); want != got {
		t.Errorf("Tlds.ListTlds() expected to return %v TLDs, got %v", want, got)
	}

	if want, got := "ac", tlds[0].Tld; want != got {
		t.Fatalf("Tlds.ListTlds() returned Tld expected to be `%v`, got `%v`", want, got)
	}

	if want, got := 1, tlds[0].MinimumRegistration; want != got {
		t.Fatalf("Tlds.ListTlds() returned MinimumRegistration expected to be `%v`, got `%v`", want, got)
	}

	if want, got := true, tlds[0].RegistrationEnabled; want != got {
		t.Fatalf("Tlds.ListTlds() returned RegistrationEnabled expected to be `%v`, got `%v`", want, got)
	}

	if want, got := true, tlds[0].RenewalEnabled; want != got {
		t.Fatalf("Tlds.ListTlds() returned RenewalEnabled expected to be `%v`, got `%v`", want, got)
	}

	if want, got := false, tlds[0].TransferEnabled; want != got {
		t.Fatalf("Tlds.ListTlds() returned TransferEnabled expected to be `%v`, got `%v`", want, got)
	}
}

func TestTldsService_ListTlds_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/tlds", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTlds/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Tlds.ListTlds(&ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Tlds.ListTlds() returned error: %v", err)
	}
}

func TestTldsService_GetTld(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/tlds/com", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getTld/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	tldResponse, err := client.Tlds.GetTld("com")
	if err != nil {
		t.Fatalf("Tlds.GetTlds() returned error: %v", err)
	}

	tld := tldResponse.Data
	if want, got := "com", tld.Tld; want != got {
		t.Fatalf("Tlds.GetTlds() returned Tld expected to be `%v`, got `%v`", want, got)
	}
	if want, got := 1, tld.MinimumRegistration; want != got {
		t.Fatalf("Tlds.GetTlds() returned Tld expected to be `%v`, got `%v`", want, got)
	}
}

func TestTldsService_GetTldExtendedAttributes(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/tlds/com/extended_attributes", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getTldExtendedAttributes/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	tldResponse, err := client.Tlds.GetTldExtendedAttributes("com")
	if err != nil {
		t.Fatalf("Tlds.GetTldExtendedAttributes() returned error: %v", err)
	}

	attributes := tldResponse.Data
	if want, got := 4, len(attributes); want != got {
		t.Errorf("Tlds.GetTldExtendedAttributes() expected to return %v TLDs, got %v", want, got)
	}

	if want, got := "uk_legal_type", attributes[0].Name; want != got {
		t.Fatalf("Tlds.GetTldExtendedAttributes() returned Tld expected to be `%v`, got `%v`", want, got)
	}
}
