package render

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GreetingP struct {
	One string `json:"one"`
	Two string `json:"two"`
}

func TestJSONPBasic(t *testing.T) {
	render := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSONP(w, 299, "helloCallback", GreetingP{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentJSONP+"; charset=UTF-8")
	expect(t, res.Body.String(), "helloCallback({\"one\":\"hello\",\"two\":\"world\"});")
}

func TestJSONPRenderIndented(t *testing.T) {
	render := New(Options{
		IndentJSON: true,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSONP(w, http.StatusOK, "helloCallback", GreetingP{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentJSONP+"; charset=UTF-8")
	expect(t, res.Body.String(), "helloCallback({\n  \"one\": \"hello\",\n  \"two\": \"world\"\n});\n")
}

func TestJSONPWithError(t *testing.T) {
	render := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSONP(w, 299, "helloCallback", math.NaN())
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expect(t, res.Code, 500)
}
