package dnspod

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestDomains_domainAction(t *testing.T) {
	var pathTests = []struct {
		input    string
		expected string
	}{
		{"Create", "Domain.Create"},
		{"", "Domain.List"},
	}

	for _, pt := range pathTests {
		actual := domainAction(pt.input)
		if actual != pt.expected {
			t.Errorf("domainAction(%+v): expected %s, actual %s", pt.input, pt.expected, actual)
		}
	}
}

func TestDomainsService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Domain.List", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{
			"status": {"code":"1","message":""},
			"domains": [
				{
					"id": 2238269,
					"status": "enable"

				},
				{
					"id": 10360095,
					"status": "enable"

				}
			]}`)
	})

	domains, _, err := client.Domains.List()

	if err != nil {
		t.Errorf("Domains.List returned error: %v", err)
	}

	want := []Domain{{ID: 2238269, Status: "enable"}, {ID: 10360095, Status: "enable"}}
	if !reflect.DeepEqual(domains, want) {
		t.Errorf("Domains.List returned %+v, want %+v", domains, want)
	}
}

func TestDomainsService_Create(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Domain.Create", func(w http.ResponseWriter, r *http.Request) {
		want := make(map[string]interface{})
		want["domain"] = map[string]interface{}{"name": "example.com"}

		testMethod(t, r, "POST")
		// testRequestJSON(t, r, want)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"status": {"code":"1","message":""},"domain":{"id":1, "name":"example.com"}}`)
	})

	domainValues := Domain{Name: "example.com"}
	domain, _, err := client.Domains.Create(domainValues)

	if err != nil {
		t.Errorf("Domains.Create returned error: %v", err)
	}

	want := Domain{ID: 1, Name: "example.com"}
	if !reflect.DeepEqual(domain, want) {
		t.Fatalf("Domains.Create returned %+v, want %+v", domain, want)
	}
}

func TestDomainsService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Domain.Info", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		fmt.Fprint(w, `{"status": {"code":"1","message":""},"domain": {"id":1, "name":"example.com"}}`)
	})

	domain, _, err := client.Domains.Get(1)

	if err != nil {
		t.Errorf("Domains.Get returned error: %v", err)
	}

	want := Domain{ID: 1, Name: "example.com"}
	if !reflect.DeepEqual(domain, want) {
		t.Fatalf("Domains.Get returned %+v, want %+v", domain, want)
	}
}

func TestDomainsService_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Domain.Remove", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"status": {"code":"1","message":""}}`)
	})

	_, err := client.Domains.Delete(1)

	if err != nil {
		t.Errorf("Domains.Delete returned error: %v", err)
	}
}
