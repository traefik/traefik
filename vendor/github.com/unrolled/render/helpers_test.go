package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderPartial(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/partials",
		Layout:    "layout",
	})

	var renErr error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renErr = render.HTML(w, http.StatusOK, "content", "gophers")
	})

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatalf("couldn't create a request. err = %s", err)
	}
	h.ServeHTTP(res, req)

	expectNil(t, renErr)
	expect(t, res.Body.String(), "before gophers\n<h1>during</h1>\nafter gophers\n")
}

func TestRenderPartialRequirePartialsOff(t *testing.T) {
	render := New(Options{
		Directory:       "fixtures/partials",
		Layout:          "layout",
		RequirePartials: false,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "content-partial", "gophers")
	})

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatalf("couldn't create a request. err = %s", err)
	}
	h.ServeHTTP(res, req)

	expect(t, res.Body.String(), "\n<h1>during</h1>\nafter gophers\n")
}

func TestRenderPartialRequirePartialsOn(t *testing.T) {
	render := New(Options{
		Directory:       "fixtures/partials",
		Layout:          "layout",
		RequirePartials: true,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.HTML(w, http.StatusOK, "content-partial", "gophers")
	})

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatalf("couldn't create a request. err = %s", err)
	}
	h.ServeHTTP(res, req)

	expect(t, res.Body.String(), "template: layout:1:3: executing \"layout\" at <partial \"before\">: error calling partial: html/template: \"before-content-partial\" is undefined\n")
}
