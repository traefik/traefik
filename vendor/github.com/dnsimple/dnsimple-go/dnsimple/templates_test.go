package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestTemplatePath(t *testing.T) {
	if want, got := "/1010/templates", templatePath("1010", ""); want != got {
		t.Errorf("templatePath(%v) = %v, want %v", "", got, want)
	}

	if want, got := "/1010/templates/1", templatePath("1010", "1"); want != got {
		t.Errorf("templatePath(%v) = %v, want %v", "1", got, want)
	}
}

func TestTemplatesService_List(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTemplates/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)
		testQuery(t, r, url.Values{})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templatesResponse, err := client.Templates.ListTemplates("1010", nil)
	if err != nil {
		t.Fatalf("Templates.ListTemplates() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 2}), templatesResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Templates.ListTemplates() pagination expected to be %v, got %v", want, got)
	}

	templates := templatesResponse.Data
	if want, got := 2, len(templates); want != got {
		t.Errorf("Templates.ListTemplates() expected to return %v templates, got %v", want, got)
	}

	if want, got := 1, templates[0].ID; want != got {
		t.Fatalf("Templates.ListTemplates() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "Alpha", templates[0].Name; want != got {
		t.Fatalf("Templates.ListTemplates() returned Name expected to be `%v`, got `%v`", want, got)
	}
}

func TestTemplatesService_List_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listTemplates/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Templates.ListTemplates("1010", &ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Templates.ListTemplates() returned error: %v", err)
	}
}

func TestTemplatesService_Create(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createTemplate/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"name": "Beta"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"
	templateAttributes := Template{Name: "Beta"}

	templateResponse, err := client.Templates.CreateTemplate(accountID, templateAttributes)
	if err != nil {
		t.Fatalf("Templates.CreateTemplate() returned error: %v", err)
	}

	template := templateResponse.Data
	if want, got := 1, template.ID; want != got {
		t.Fatalf("Templates.CreateTemplate() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "Beta", template.Name; want != got {
		t.Fatalf("Templates.CreateTemplate() returned Label expected to be `%v`, got `%v`", want, got)
	}
}

func TestTemplatesService_Get(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getTemplate/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templateResponse, err := client.Templates.GetTemplate("1010", "1")
	if err != nil {
		t.Fatalf("Templates.GetTemplate() returned error: %v", err)
	}

	template := templateResponse.Data
	wantSingle := &Template{
		ID:          1,
		SID:         "alpha",
		AccountID:   1010,
		Name:        "Alpha",
		Description: "An alpha template.",
		CreatedAt:   "2016-03-22T11:08:58Z",
		UpdatedAt:   "2016-03-22T11:08:58Z"}

	if !reflect.DeepEqual(template, wantSingle) {
		t.Fatalf("Templates.GetTemplate() returned %+v, want %+v", template, wantSingle)
	}
}

func TestTemplatesService_UpdateTemplate(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/updateTemplate/success.http")

		testMethod(t, r, "PATCH")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	templateAttributes := Template{Name: "Alpha"}
	templateResponse, err := client.Templates.UpdateTemplate("1010", "1", templateAttributes)
	if err != nil {
		t.Fatalf("Templates.UpdateTemplate() returned error: %v", err)
	}

	template := templateResponse.Data
	wantSingle := &Template{
		ID:          1,
		SID:         "alpha",
		AccountID:   1010,
		Name:        "Alpha",
		Description: "An alpha template.",
		CreatedAt:   "2016-03-22T11:08:58Z",
		UpdatedAt:   "2016-03-22T11:08:58Z"}

	if !reflect.DeepEqual(template, wantSingle) {
		t.Fatalf("Templates.UpdateTemplate() returned %+v, want %+v", template, wantSingle)
	}
}

func TestTemplatesService_DeleteTemplate(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/templates/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteTemplate/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Templates.DeleteTemplate("1010", "1")
	if err != nil {
		t.Fatalf("Templates.DeleteTemplate() returned error: %v", err)
	}
}
