package rewrite

import (
	"io"
	"io/ioutil"
	"net/http"
	"text/template"
)

// data represents template data that is available to use in templates.
type data struct {
	Request *http.Request
}

// Apply reads a template string from the provided reader, applies variables
// from the provided request object to it and writes the result into
// the provided writer.
//
// Template is standard Go's http://golang.org/pkg/text/template/.
func Apply(in io.Reader, out io.Writer, request *http.Request) error {
	body, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	return ApplyString(string(body), out, request)
}

// ApplyString applies variables from the provided request object to the provided
// template string and writes the result into the provided writer.
//
// Template is standard Go's http://golang.org/pkg/text/template/.
func ApplyString(in string, out io.Writer, request *http.Request) error {
	t, err := template.New("t").Parse(in)
	if err != nil {
		return err
	}

	if err = t.Execute(out, data{request}); err != nil {
		return err
	}

	return nil
}
