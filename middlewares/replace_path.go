package middlewares

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	"github.com/containous/traefik/log"
)

// ReplacePath is a middleware used to replace the path of a URL request
type ReplacePath struct {
	Handler http.Handler
	Path    string
}

// ReplacedPathHeader is the default header to set the old path to
const ReplacedPathHeader = "X-Replaced-Path"

func (s *ReplacePath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Add(ReplacedPathHeader, r.URL.Path)
	r.URL.Path = s.ReplacePath(r.URL.Path)
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}

// ReplacePath returns a path replaced with path template.
func (s *ReplacePath) ReplacePath(source string) string {
	f := template.FuncMap{
		"Replace": strings.Replace,
	}
	t, err := template.New("replace_path").Funcs(f).Parse(s.Path)
	if err != nil {
		log.Error("parsing: ", err)
	} else {
		var buffer bytes.Buffer
		if err = t.Execute(&buffer, source); err != nil {
			log.Error("execution: ", err)
		} else {
			r := buffer.String()
			log.Debugf("ReplacePath: %s -> %s", source, r)
			return r
		}
	}
	panic(err)
}
