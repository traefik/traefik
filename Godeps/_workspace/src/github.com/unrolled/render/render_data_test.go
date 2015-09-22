package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataBinaryBasic(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.Data(w, 299, []byte("hello there"))
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentBinary)
	expect(t, res.Body.String(), "hello there")
}

func TestDataCustomMimeType(t *testing.T) {
	render := New(Options{
	// nothing here to configure
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, "image/jpeg")
		render.Data(w, http.StatusOK, []byte("..jpeg data.."))
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), "image/jpeg")
	expect(t, res.Body.String(), "..jpeg data..")
}
