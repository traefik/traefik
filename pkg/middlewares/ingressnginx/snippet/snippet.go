package snippet

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/tufanbarisyildirim/gonginx/parser"
)

const typeName = "Snippet"

// Snippet is a middleware allowing to parse and interpret NGINX snippets containing directives.
type Snippet struct {
	next                 http.Handler
	name                 string
	serverActions        *actions
	configurationActions *actions
}

// New creates a new Snippet middleware instance.
// It parses the provided snippets and builds the corresponding actions.
func New(ctx context.Context, next http.Handler, config *dynamic.Snippet, name string) (h http.Handler, err error) {
	// Here we are adding a recover block as the snippet parsing can panic.
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("snippet parsing recover: %v", recErr)
		}
	}()

	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if config.ServerSnippet == "" && config.ConfigurationSnippet == "" {
		return nil, errors.New("at least one of serverSnippet or configurationSnippet option must be provided")
	}

	parserOptions := []parser.Option{
		parser.WithSkipComments(),
		parser.WithCustomDirectives("more_set_headers", "more_set_input_headers", "more_clear_headers", "more_clear_input_headers", "proxy_hide_header"),
	}

	var serverActions *actions
	if config.ServerSnippet != "" {
		// Parse the snippet, note that we are wrapping the server snippet in a server block to ensure that it is parsed in the correct context.
		p := parser.NewStringParser(fmt.Sprintf("server{%s}", config.ServerSnippet), parserOptions...)

		conf, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing server-snippet: %w", parseErr)
		}

		serverActions, err = buildActions(conf.GetDirectives()[0].GetBlock())
		if err != nil {
			return nil, fmt.Errorf("building actions from server-snippet: %w", err)
		}
	}

	var (
		buildErr             error
		configurationActions *actions
	)
	if config.ConfigurationSnippet != "" {
		// Parse the snippet, note that we are wrapping the configuration snippet in a location block to ensure that it is parsed in the correct context.
		p := parser.NewStringParser(fmt.Sprintf("location / {%s}", config.ConfigurationSnippet), parserOptions...)

		conf, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing configuration-snippet: %w", parseErr)
		}

		configurationActions, buildErr = buildActions(conf.GetDirectives()[0].GetBlock())
		if buildErr != nil {
			return nil, fmt.Errorf("building actions from configuration-snippet: %w", buildErr)
		}
	}

	return &Snippet{
		next:                 next,
		name:                 name,
		serverActions:        serverActions,
		configurationActions: configurationActions,
	}, nil
}

func (s *Snippet) GetTracingInformation() (string, string) {
	return s.name, typeName
}

func (s *Snippet) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrappedRW := &snippetResponseWriter{ResponseWriter: rw}

	ctx := &actionContext{
		vars:                    make(map[string]string),
		nonMergeablePostActions: make(map[string][]action),
	}

	stop, err := s.serverActions.Execute(wrappedRW, req, ctx)
	if err != nil {
		http.Error(wrappedRW, err.Error(), http.StatusInternalServerError)
		return
	}
	if stop {
		if err = executePostActions(wrappedRW, req, ctx); err != nil {
			http.Error(wrappedRW, err.Error(), http.StatusInternalServerError)
			return
		}

		writeResponse(wrappedRW, req, ctx)
		return
	}

	// In NGINX, proxy_set_header directives in the server snippet are ignored
	// because the generated location block always contains proxy_set_header directives that override them.
	ctx.nonMergeablePostActions["proxy_set_header"] = nil

	// rewrite...break in the server snippet stops all directive processing,
	// but post-actions (headers, etc.) and upstream forwarding still proceed.
	if !ctx.stopAllDirectives {
		ctx.stopCurrentBlock = false

		stop, err = s.configurationActions.Execute(wrappedRW, req, ctx)
		if err != nil {
			http.Error(wrappedRW, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = executePostActions(wrappedRW, req, ctx); err != nil {
		http.Error(wrappedRW, err.Error(), http.StatusInternalServerError)
		return
	}

	if stop {
		writeResponse(wrappedRW, req, ctx)
		return
	}

	s.next.ServeHTTP(wrappedRW, req)
}

// snippetResponseWriter wraps http.ResponseWriter to intercept WriteHeader calls.
// This allows deferred response header operations (e.g., proxy_hide_header, conditional
// header setting with -s/-t flags) to be applied based on the actual response status
// code and content type set by the upstream.
type snippetResponseWriter struct {
	http.ResponseWriter

	headerWritten bool
	onWriteHeader []func(code int, h http.Header)
}

func (w *snippetResponseWriter) WriteHeader(code int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true
	for _, fn := range w.onWriteHeader {
		fn(code, w.Header())
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *snippetResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter, enabling http.ResponseController
// to discover the underlying writer's capabilities (Flusher, Hijacker, etc.).
func (w *snippetResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush implements http.Flusher.
func (w *snippetResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker.
func (w *snippetResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("not a hijacker: %T", w.ResponseWriter)
}

// writeResponse writes the final response based on the action context.
// For redirect status codes (301, 302, 303, 307, 308) with a URL, it performs an HTTP redirect.
// For other status codes, it writes the status code and optional body text.
func writeResponse(rw http.ResponseWriter, req *http.Request, ctx *actionContext) {
	if ctx.statusCode == 0 {
		return
	}

	if ctx.redirectURL != "" {
		http.Redirect(rw, req, ctx.redirectURL, ctx.statusCode)
		return
	}

	rw.WriteHeader(ctx.statusCode)
	if ctx.body != "" {
		_, _ = rw.Write([]byte(ctx.body))
	}
}

func executePostActions(rw http.ResponseWriter, req *http.Request, ctx *actionContext) error {
	for _, postActions := range ctx.nonMergeablePostActions {
		for _, postAction := range postActions {
			if _, err := postAction(rw, req, ctx); err != nil {
				return fmt.Errorf("executing non-mergeable configuration action: %w", err)
			}
		}
	}

	for _, postActions := range ctx.mergeablePostActions {
		if _, err := postActions(rw, req, ctx); err != nil {
			return fmt.Errorf("executing mergeable configuration action: %w", err)
		}
	}

	return nil
}
