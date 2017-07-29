package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestTemplates_templateRecordPath(t *testing.T) {
	if want, got := "/1010/templates/1/records", templateRecordPath("1010", "1", ""); want != got {
		t.Errorf("templateRecordPath(%v, %v, ) = %v, want %v", "1010", "1", got, want)
	}

	if want, got := "/1010/templates/1/records/2", templateRecordPath("1010", "1", "2"); want != got {
		t.Errorf("templateRecordPath(%v, %v, 2) = %v, want %v", "1010", "1", got, want)
	}
}

func TestTemplatesService_ListTemplateRecords(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTemplateRecords/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)
		testQuery(t, r, url.Values{})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templatesRecordsResponse, err := client.Templates.ListTemplateRecords("1010", "1", nil)
	if err != nil {
		t.Fatalf("Templates.ListTemplateRecords() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 2}), templatesRecordsResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Templates.ListTemplateRecords() pagination expected to be %v, got %v", want, got)
	}

	templates := templatesRecordsResponse.Data
	if want, got := 2, len(templates); want != got {
		t.Errorf("Templates.ListTemplateRecords() expected to return %v templates, got %v", want, got)
	}

	if want, got := 296, templates[0].ID; want != got {
		t.Fatalf("Templates.ListTemplateRecords() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "192.168.1.1", templates[0].Content; want != got {
		t.Fatalf("Templates.ListTemplateRecords() returned Content expected to be `%v`, got `%v`", want, got)
	}
}

func TestTemplatesService_ListTemplateRecords_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTemplateRecords/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Templates.ListTemplateRecords("1010", "1", &ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Templates.ListTemplateRecords() returned error: %v", err)
	}
}

func TestTemplatesService_CreateTemplateRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createTemplateRecord/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"name": "Beta"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templateRecordAttributes := TemplateRecord{Name: "Beta"}

	templateRecordResponse, err := client.Templates.CreateTemplateRecord("1010", "1", templateRecordAttributes)
	if err != nil {
		t.Fatalf("Templates.CreateTemplateRecord() returned error: %v", err)
	}

	templateRecord := templateRecordResponse.Data
	if want, got := 300, templateRecord.ID; want != got {
		t.Fatalf("Templates.CreateTemplateRecord() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "mx.example.com", templateRecord.Content; want != got {
		t.Fatalf("Templates.CreateTemplateRecord() returned Content expected to be `%v`, got `%v`", want, got)
	}
}

func TestTemplatesService_GetTemplateRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1/records/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getTemplateRecord/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templateRecordResponse, err := client.Templates.GetTemplateRecord("1010", "1", "2")
	if err != nil {
		t.Fatalf("Templates.GetTemplateRecord() returned error: %v", err)
	}

	templateRecord := templateRecordResponse.Data
	wantSingle := &TemplateRecord{
		ID:         301,
		TemplateID: 268,
		Name:       "",
		Content:    "mx.example.com",
		TTL:        600,
		Priority:   10,
		Type:       "MX",
		CreatedAt:  "2016-05-03T08:03:26Z",
		UpdatedAt:  "2016-05-03T08:03:26Z"}

	if !reflect.DeepEqual(templateRecord, wantSingle) {
		t.Fatalf("Templates.GetTemplateRecord() returned %+v, want %+v", templateRecord, wantSingle)
	}
}

func TestTemplatesService_DeleteTemplateRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1/records/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteTemplateRecord/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Templates.DeleteTemplateRecord("1010", "1", "2")
	if err != nil {
		t.Fatalf("Templates.DeleteTemplateRecord() returned error: %v", err)
	}
}
