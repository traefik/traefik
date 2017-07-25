package render

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Greeting struct {
	One string `json:"one"`
	Two string `json:"two"`
}

func TestJSONBasic(t *testing.T) {
	render := New()

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 299, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), "{\"one\":\"hello\",\"two\":\"world\"}")
}

func TestJSONPrefix(t *testing.T) {
	prefix := ")]}',\n"
	render := New(Options{
		PrefixJSON: []byte(prefix),
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 300, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+"{\"one\":\"hello\",\"two\":\"world\"}")
}

func TestJSONIndented(t *testing.T) {
	render := New(Options{
		IndentJSON: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, http.StatusOK, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, http.StatusOK)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), "{\n  \"one\": \"hello\",\n  \"two\": \"world\"\n}\n")
}

func TestJSONConsumeIndented(t *testing.T) {
	render := New(Options{
		IndentJSON: true,
	})

	var renErr error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renErr = render.JSON(w, http.StatusOK, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	var output Greeting
	err := json.Unmarshal(res.Body.Bytes(), &output)
	expectNil(t, err)
	expectNil(t, renErr)
	expect(t, output.One, "hello")
	expect(t, output.Two, "world")
}

func TestJSONWithError(t *testing.T) {
	render := New(Options{}, Options{}, Options{}, Options{})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 299, math.NaN())
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 500)
}

func TestJSONWithOutUnEscapeHTML(t *testing.T) {
	render := New(Options{
		UnEscapeHTML: false,
	})
	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, http.StatusOK, Greeting{"<span>test&test</span>", "<div>test&test</div>"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Body.String(), `{"one":"\u003cspan\u003etest\u0026test\u003c/span\u003e","two":"\u003cdiv\u003etest\u0026test\u003c/div\u003e"}`)
}

func TestJSONWithUnEscapeHTML(t *testing.T) {
	render := New(Options{
		UnEscapeHTML: true,
	})
	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, http.StatusOK, Greeting{"<span>test&test</span>", "<div>test&test</div>"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Body.String(), "{\"one\":\"<span>test&test</span>\",\"two\":\"<div>test&test</div>\"}")
}

func TestJSONStream(t *testing.T) {
	render := New(Options{
		StreamingJSON: true,
	})
	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 299, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 299)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), "{\"one\":\"hello\",\"two\":\"world\"}\n")
}

func TestJSONStreamPrefix(t *testing.T) {
	prefix := ")]}',\n"
	render := New(Options{
		PrefixJSON:    []byte(prefix),
		StreamingJSON: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 300, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+"{\"one\":\"hello\",\"two\":\"world\"}\n")
}

func TestJSONStreamWithError(t *testing.T) {
	render := New(Options{
		StreamingJSON: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 299, math.NaN())
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 299)

	// Because this is streaming, we can not catch the error.
	expect(t, res.Body.String(), "json: unsupported value: NaN\n")
	// Also the header will be incorrect.
	expect(t, res.Header().Get(ContentType), "text/plain; charset=utf-8")
}

func TestJSONCharset(t *testing.T) {
	render := New(Options{
		Charset: "foobar",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.JSON(w, 300, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=foobar")
	expect(t, res.Body.String(), "{\"one\":\"hello\",\"two\":\"world\"}")
}
