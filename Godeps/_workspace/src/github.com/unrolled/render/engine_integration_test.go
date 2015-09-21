// +build integration

package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eknkc/amber"
)

// go test -tags=integration

type Amber struct {
	Head
	Template *template.Template
}

func (a Amber) Render(w http.ResponseWriter, v interface{}) error {
	a.Head.Write(w)
	return a.Template.Execute(w, v)
}

func TestRenderAmberTemplate(t *testing.T) {
	dir := "fixtures/amber/"
	render := New(Options{})

	templates, err := amber.CompileDir(dir, amber.DefaultDirOptions, amber.DefaultOptions)
	if err != nil {
		t.Errorf("Could not compile Amber templates at " + dir)
	}

	a := Amber{
		Head: Head{
			ContentType: ContentHTML,
			Status:      http.StatusOK,
		},
		Template: templates["example"],
	}

	v := struct {
		VarOne string
		VarTwo string
	}{
		"Contact",
		"Content!",
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.Render(w, a, v)
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)
	h.ServeHTTP(res, req)

	body := res.Body.String()

	checkCompile := strings.Index(body, `<div id="header">`) != -1
	checkVarOne := strings.Index(body, `<li>Contact</li>`) != -1
	checkVarTwo := strings.Index(body, `<li>Content!</li>`) != -1

	expect(t, checkCompile && checkVarOne && checkVarTwo, true)
	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML)
}
