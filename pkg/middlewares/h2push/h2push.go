package h2push

import (
	"context"
	"strings"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
)

const (
	typeName = "H2Push"
)

// H2Push is a middleware used to push resources with HTTP2.
type h2push struct {
	next http.Handler
	path string

	linkRegex *regexp.Regexp
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.H2Push, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	fmt.Println("Path: " + config.PushPath)

	return &h2push{
		next: next,
		path: config.PushPath,
		linkRegex: regexp.MustCompile(`(?m)<([^>]+)>;\s+rel=(\w+);\s+as=(\w+)`),
	}, nil
}

func (h *h2push) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.next.ServeHTTP(rw, req)

	if pusher, ok := rw.(http.Pusher); ok {
		h.pushLinks(pusher, rw.Header()["Link"])

		pusher.Push(h.path, nil)
	}
}

func (h *h2push) pushLinks(p http.Pusher, links []string) error {
	for _, link := range links {
		fname, rel, kind, err := h.parseLink(link)
		if err != nil {
			return err
		}

		if rel != "preload" {
			continue
		}

		fmt.Printf("Link file name: %v, kind: %v\n", fname, kind);

		if !strings.HasPrefix(fname, "/") {
			fname = "/" + fname
		}
		
		p.Push(fname, nil)
	}

	return nil
}

func (h *h2push) parseLink(link string) (fileName string, rel string, kind string, err error) {
	groups := h.linkRegex.FindStringSubmatch(link)

	if len(groups) != 4 {
		return "", "", "", errors.New("invalid Link header");
	}

	return groups[1], groups[2], groups[3], nil;
}
