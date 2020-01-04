package h2push

import (
	"context"
	"strings"
	"errors"
	"net/http"
	"regexp"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
)

const (
	typeName = "H2Push"
)

var (
	linkRegex = regexp.MustCompile(`(?m)<([^>]+)>;\s+rel=(\w+);\s+as=(\w+)`)
	absoluteURLRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z\d+\-.]*:`)
)

// H2Push is a middleware used to push resources with HTTP2.
type h2push struct {
	name string
	next http.Handler
	files []dynamic.H2PushFile
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.H2Push, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	return &h2push{
		name: name,
		next: next,
		files: config.Files,
	}, nil
}

func (h *h2push) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	pusher, isPushable := rw.(http.Pusher)

	if isPushable && h.files != nil {
		for _, file := range h.files {
			if file.Match != "" {
				matched, err := regexp.MatchString(file.Match, req.URL.Path)

				if err != nil {
					log.FromContext(middlewares.GetLoggerCtx(req.Context(), h.name, typeName)).
						Errorf("Invalid Regex pattern: %v", file.Match)

					continue
				}

				if !matched {
					continue
				}
			}

			pusher.Push(normalizePath(file.URL), nil)
		}
	}
	
	h.next.ServeHTTP(rw, req)

	if isPushable {
		h.pushLinks(pusher, rw.Header()["Link"])
	}
}

func (h *h2push) pushLinks(p http.Pusher, linkHeaders []string) error {
	for _, link := range linkHeaders {
		fname, rel, _, err := parseLink(link)
		if err != nil {
			return err
		}

		if rel != "preload" {
			continue
		}

		fname = normalizePath(fname);

		p.Push(fname, nil)
	}

	return nil
}

func parseLink(link string) (fileName string, rel string, kind string, err error) {
	groups := linkRegex.FindStringSubmatch(link)

	if len(groups) != 4 {
		err = errors.New("invalid link header")
		return
	}

	return groups[1], groups[2], groups[3], nil;
}

func normalizePath(path string) (absolutePath string) {
	if !strings.HasPrefix(path, "/") && !absoluteURLRegex.MatchString(path) {
		return "/" + path
	}

	return path
}