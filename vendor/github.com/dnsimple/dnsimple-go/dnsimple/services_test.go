package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestServicePath(t *testing.T) {
	if want, got := "/services", servicePath(""); want != got {
		t.Errorf("servicePath(%v) = %v, want %v", "", got, want)
	}

	if want, got := "/services/name", servicePath("name"); want != got {
		t.Errorf("servicePath(%v) = %v, want %v", "name", got, want)
	}
}

func TestServicesService_List(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/services", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listServices/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)
		testQuery(t, r, url.Values{})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	servicesResponse, err := client.Services.ListServices(nil)
	if err != nil {
		t.Fatalf("Services.ListServices() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 2}), servicesResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Services.ListServices() pagination expected to be %v, got %v", want, got)
	}

	services := servicesResponse.Data
	if want, got := 2, len(services); want != got {
		t.Errorf("Services.ListServices() expected to return %v services, got %v", want, got)
	}

	if want, got := 1, services[0].ID; want != got {
		t.Fatalf("Services.ListServices() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "Service 1", services[0].Name; want != got {
		t.Fatalf("Services.ListServices() returned Name expected to be `%v`, got `%v`", want, got)
	}
}

func TestServicesService_List_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/services", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listServices/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Services.ListServices(&ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Services.ListServices() returned error: %v", err)
	}
}

func TestServicesService_Get(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/services/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getService/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	serviceID := "1"

	serviceResponse, err := client.Services.GetService(serviceID)
	if err != nil {
		t.Fatalf("Services.GetService() returned error: %v", err)
	}

	service := serviceResponse.Data
	wantSingle := &Service{
		ID:               1,
		SID:              "service1",
		Name:             "Service 1",
		Description:      "First service example.",
		SetupDescription: "",
		RequiresSetup:    true,
		DefaultSubdomain: "",
		CreatedAt:        "2014-02-14T19:15:19Z",
		UpdatedAt:        "2016-03-04T09:23:27Z",
		Settings: []ServiceSetting{
			{
				Name:        "username",
				Label:       "Service 1 Account Username",
				Append:      ".service1.com",
				Description: "Your Service 1 username is used to connect services to your account.",
				Example:     "username",
				Password:    false,
			},
		},
	}

	if !reflect.DeepEqual(service, wantSingle) {
		t.Fatalf("Services.GetService() returned %+v, want %+v", service, wantSingle)
	}
}
