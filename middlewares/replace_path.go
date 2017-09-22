package middlewares

import (
	"net/http"
	"regexp"
	"strings"

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
	originPath := r.URL.Path
	r.Header.Add(ReplacedPathHeader, originPath)
	r.URL.Path = s.ReplacePath(originPath)
	log.Debug(originPath, " $> ", r.URL.Path)
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}

// ReplacePath returns a path replaced with path template.
func (s *ReplacePath) ReplacePath(source string) string {
	if sp := strings.SplitN(s.Path, "$>", 2); len(sp) > 1 {
		return regexp.MustCompile(strings.TrimSpace(sp[0])).ReplaceAllString(source, strings.TrimSpace(sp[1]))
	}
	return s.Path
}
