// +build !go1.6

package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRenderBlock(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/blocks",
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

func TestRenderBlockRequireBlocksOff(t *testing.T) {
	render := New(Options{
		Directory:     "fixtures/blocks",
		Layout:        "layout",
		RequireBlocks: false,
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

func TestRenderBlockRequireBlocksOn(t *testing.T) {
	render := New(Options{
		Directory:     "fixtures/blocks",
		Layout:        "layout",
		RequireBlocks: true,
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

	expect(t, res.Body.String(), "template: layout:1:3: executing \"layout\" at <block \"before\">: error calling block: html/template: \"before-content-partial\" is undefined\n")
}
