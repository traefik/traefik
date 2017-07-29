package render

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTMLBad(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "nope", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 500)
	expect(t, res.Body.String(), "html/template: \"nope\" is undefined\n")
}

func TestHTMLBadDisableHTTPErrorRendering(t *testing.T) {
	render := New(Options{
		Directory:                 "fixtures/basic",
		DisableHTTPErrorRendering: true,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "nope", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Body.String(), "")
}

func TestHTMLBasic(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestHTMLXHTML(t *testing.T) {
	render := New(Options{
		Directory:       "fixtures/basic",
		HTMLContentType: ContentXHTML,
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentXHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestHTMLExtensions(t *testing.T) {
	render := New(Options{
		Directory:  "fixtures/basic",
		Extensions: []string{".tmpl", ".html"},
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hypertext", nil)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "Hypertext!\n")
}

func TestHTMLFuncs(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/custom_funcs",
		Funcs: []template.FuncMap{
			{
				"myCustomFunc": func() string {
					return "My custom function"
				},
			},
		},
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "index", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Body.String(), "My custom function\n")
}

func TestRenderLayout(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "layout",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "content", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Body.String(), "head\n<h1>gophers</h1>\n\nfoot\n")
}

func TestHTMLLayoutCurrent(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "current_layout",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "content", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Body.String(), "content head\n<h1>gophers</h1>\n\ncontent foot\n")
}

func TestHTMLNested(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "admin/index", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Admin gophers</h1>\n")
}

func TestHTMLBadPath(t *testing.T) {
	render := New(Options{
		Directory: "../../../../../../../../../../../../../../../../fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNotNil(t, err)
	expect(t, res.Code, 500)
}

func TestHTMLDelimiters(t *testing.T) {
	render := New(Options{
		Delims:    Delims{"{[{", "}]}"},
		Directory: "fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "delims", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>")
}

func TestHTMLDefaultCharset(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "hello", "gophers")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")

	// ContentLength should be deferred to the ResponseWriter and not Render
	expect(t, res.Header().Get(ContentLength), "")
	expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
}

func TestHTMLOverrideLayout(t *testing.T) {
	render := New(Options{
		Directory: "fixtures/basic",
		Layout:    "layout",
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "content", "gophers", HTMLOptions{
			Layout: "another_layout",
		})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "another head\n<h1>gophers</h1>\n\nanother foot\n")
}

func TestHTMLNoRace(t *testing.T) {
	// This test used to fail if run with -race
	render := New(Options{
		Directory: "fixtures/basic",
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := render.HTML(w, http.StatusOK, "hello", "gophers")
		expectNil(t, err)
	})

	done := make(chan bool)
	doreq := func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/foo", nil)

		h.ServeHTTP(res, req)

		expect(t, res.Code, 200)
		expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
		// ContentLength should be deferred to the ResponseWriter and not Render
		expect(t, res.Header().Get(ContentLength), "")
		expect(t, res.Body.String(), "<h1>Hello gophers</h1>\n")
		done <- true
	}
	// Run two requests to check there is no race condition
	go doreq()
	go doreq()
	<-done
	<-done
}

func TestHTMLLoadFromAssets(t *testing.T) {
	render := New(Options{
		Asset: func(file string) ([]byte, error) {
			switch file {
			case "templates/test.tmpl":
				return []byte("<h1>gophers</h1>\n"), nil
			case "templates/layout.tmpl":
				return []byte("head\n{{ yield }}\nfoot\n"), nil
			default:
				return nil, errors.New("file not found: " + file)
			}
		},
		AssetNames: func() []string {
			return []string{"templates/test.tmpl", "templates/layout.tmpl"}
		},
	})

	var err error
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = render.HTML(w, http.StatusOK, "test", "gophers", HTMLOptions{
			Layout: "layout",
		})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	expectNil(t, err)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "head\n<h1>gophers</h1>\n\nfoot\n")
}

func TestCompileTemplatesFromDir(t *testing.T) {
	baseDir := "fixtures/template-dir-test"
	fname0Rel := "0"
	fname1Rel := "subdir/1"
	fnameShouldParsedRel := "dedicated.tmpl/notbad"
	dirShouldNotParsedRel := "dedicated"

	r := New(Options{
		Directory:  baseDir,
		Extensions: []string{".tmpl", ".html"},
	})
	r.compileTemplatesFromDir()

	expect(t, r.TemplateLookup(fname1Rel) != nil, true)
	expect(t, r.TemplateLookup(fname0Rel) != nil, true)
	expect(t, r.TemplateLookup(fnameShouldParsedRel) != nil, true)
	expect(t, r.TemplateLookup(dirShouldNotParsedRel) == nil, true)
}
