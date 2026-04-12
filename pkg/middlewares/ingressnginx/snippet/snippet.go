package snippet

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/observability/tracing"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"github.com/vulcand/oxy/v2/forward"
	"github.com/vulcand/oxy/v2/utils"
	"go.opentelemetry.io/otel/trace"
)

const typeName = "Snippet"

// Snippet is a middleware allowing to parse and interpret NGINX snippets containing directives.
// It executes directives in NGINX phase order:
// INPUT_FILTER → SERVER_REWRITE → FIND_CONFIG → REWRITE → ACCESS → CONTENT → HEADER_FILTER.
type Snippet struct {
	next http.Handler
	name string

	// Collectable actions from server-snippet
	serverActions *SnippetActions

	// Collectable actions from configuration-snippet (location context)
	locationActions *SnippetActions

	// Collectable actions from auth-snippet
	authActions *SnippetActions

	// Forward auth configuration
	address             string
	method              string
	authResponseHeaders []string
	authSigninURL       string
	client              http.Client
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

	if config.ServerSnippet == "" && config.ConfigurationSnippet == "" && config.Auth == nil {
		return nil, errors.New("at least one of serverSnippet, configurationSnippet or auth option must be provided")
	}

	if config.Auth != nil && config.Auth.Address == "" {
		return nil, errors.New("address is required in auth configuration")
	}

	parserOptions := []parser.Option{
		parser.WithSkipComments(),
		parser.WithCustomDirectives("more_set_headers", "more_set_input_headers", "more_clear_headers", "more_clear_input_headers", "proxy_hide_header"),
	}

	var serverActions *SnippetActions
	if config.ServerSnippet != "" {
		// Parse the snippet, note that we are wrapping the server snippet in a server block to ensure that it is parsed in the correct context.
		p := parser.NewStringParser(fmt.Sprintf("server{%s}", config.ServerSnippet), parserOptions...)

		conf, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing server-snippet: %w", parseErr)
		}

		serverActions, err = BuildSnippetActions(conf.GetDirectives()[0].GetBlock())
		if err != nil {
			return nil, fmt.Errorf("building actions from server-snippet: %w", err)
		}
	}

	var locationActions *SnippetActions
	if config.ConfigurationSnippet != "" {
		// Parse the snippet, note that we are wrapping the configuration snippet in a location block to ensure that it is parsed in the correct context.
		p := parser.NewStringParser(fmt.Sprintf("location / {%s}", config.ConfigurationSnippet), parserOptions...)

		conf, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing configuration-snippet: %w", parseErr)
		}

		locationActions, err = BuildSnippetActions(conf.GetDirectives()[0].GetBlock())
		if err != nil {
			return nil, fmt.Errorf("building actions from configuration-snippet: %w", err)
		}
	}

	var authActions *SnippetActions
	if config.Auth != nil && config.Auth.Snippet != "" {
		// Parse the snippet, note that we are wrapping the snippet in a location block to ensure that it is parsed in the correct context.
		p := parser.NewStringParser(fmt.Sprintf("location = /_external-auth {%s}", config.Auth.Snippet), parserOptions...)

		conf, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parsing auth-snippet: %w", parseErr)
		}

		authActions, err = BuildSnippetActions(conf.GetDirectives()[0].GetBlock())
		if err != nil {
			return nil, fmt.Errorf("building actions from auth-snippet: %w", err)
		}
	}

	m := &Snippet{
		next:            next,
		name:            name,
		serverActions:   serverActions,
		locationActions: locationActions,
		authActions:     authActions,
	}

	if config.Auth != nil {
		m.address = config.Auth.Address
		m.method = config.Auth.Method
		m.authResponseHeaders = config.Auth.AuthResponseHeaders
		m.authSigninURL = config.Auth.AuthSigninURL

		m.client = http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 30 * time.Second,
		}
	}
	return m, nil
}

func (s *Snippet) GetTracingInformation() (string, string) {
	return s.name, typeName
}

func (s *Snippet) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := &actionContext{
		vars: make(map[string]string),
		req:  req,
	}

	// Collectors for server and location phases
	serverCollector := &PhaseCollector{}
	locationCollector := &PhaseCollector{}

	// ========================================================================
	// PHASE: SERVER_REWRITE_PHASE
	// Collects phases from server-snippet.
	// Rewrite directives (set, rewrite, return, if) execute immediately.
	// Other directives are collected for later phase execution.
	// ========================================================================
	skipLocationSnippet := false
	if s.serverActions != nil {
		terminated, skipToAccess, err := s.serverActions.Collect(rw, req, ctx, serverCollector)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if terminated {
			// Use wrapped response writer for proper status-code filtering on add_header
			wrappedRW := s.wrapResponseWriterWithCollectors(rw, ctx, serverCollector, locationCollector)
			WriteResponse(wrappedRW, req, ctx)
			return
		}
		// rewrite...break skips location snippet processing
		skipLocationSnippet = skipToAccess
	}

	// ========================================================================
	// PHASE: REWRITE_PHASE (location context)
	// Collects phases from configuration-snippet.
	// Rewrite directives execute immediately, others are collected.
	// ========================================================================
	if !skipLocationSnippet && s.locationActions != nil {
		ctx.stopCurrentBlock = false
		ctx.stopAllDirectives = false

		terminated, _, err := s.locationActions.Collect(rw, req, ctx, locationCollector)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if terminated {
			// Use wrapped response writer for proper status-code filtering on add_header
			wrappedRW := s.wrapResponseWriterWithCollectors(rw, ctx, serverCollector, locationCollector)
			WriteResponse(wrappedRW, req, ctx)
			return
		}
	}

	// ========================================================================
	// PHASE: INPUT_HEADER_FILTER (request header manipulation)
	// Additive: both server and location directives apply
	// ========================================================================
	for _, act := range serverCollector.InputFilter {
		if _, err := act(rw, req, ctx); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	for _, act := range locationCollector.InputFilter {
		if _, err := act(rw, req, ctx); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// ========================================================================
	// PHASE: ACCESS_PHASE - allow/deny
	// Override: location rules replace server rules entirely
	// ========================================================================
	accessActions := serverCollector.Access
	if locationCollector.HasAccess() {
		accessActions = locationCollector.Access
	}

	for _, act := range accessActions {
		terminated, err := act(rw, req, ctx)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if terminated {
			WriteResponse(rw, req, ctx)
			return
		}
	}

	// ========================================================================
	// PHASE: ACCESS_PHASE - auth_request (forward auth)
	// ========================================================================
	if s.address != "" {
		authTerminated, err := s.executeForwardAuth(rw, req, ctx)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if authTerminated {
			WriteResponse(rw, req, ctx)
			return
		}
	}

	// ========================================================================
	// PHASE: CONTENT_PHASE
	// Directives: proxy_set_header, proxy_method
	// Override: location rules replace server rules entirely
	// ========================================================================
	contentActions := serverCollector.Content
	if locationCollector.HasContent() {
		contentActions = locationCollector.Content
	}

	for _, act := range contentActions {
		if _, err := act(rw, req, ctx); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// ========================================================================
	// PHASE: HEADER_FILTER (response headers)
	// Wrap response writer to apply header modifications when response is written.
	// add_header: Override behavior - location replaces server
	// more_*, proxy_hide_header, expires: Additive - both apply
	// ========================================================================
	wrappedRW := s.wrapResponseWriterWithCollectors(rw, ctx, serverCollector, locationCollector)

	// ========================================================================
	// PHASE: CONTENT_PHASE - proxy_pass equivalent
	// Forward request to upstream (next handler)
	// ========================================================================
	s.next.ServeHTTP(wrappedRW, req)
}

// wrapResponseWriterWithCollectors wraps the response writer to intercept WriteHeader
// and apply HEADER_FILTER phase directives from the collectors.
func (s *Snippet) wrapResponseWriterWithCollectors(rw http.ResponseWriter, ctx *actionContext, serverCollector, locationCollector *PhaseCollector) *snippetResponseWriter {
	wrappedRW := &snippetResponseWriter{ResponseWriter: rw}

	// Register header filter callback
	wrappedRW.onWriteHeader = append(wrappedRW.onWriteHeader, func(code int, h http.Header) {
		// Create a minimal response writer that operates on the header map
		headerRW := &headerOnlyWriter{header: h}

		// add_header: Override behavior - location replaces server
		var overrideEntries []addHeaderEntry
		if locationCollector.HasHeaderOverride() {
			overrideEntries = locationCollector.HeaderFilterOverride
		} else {
			overrideEntries = serverCollector.HeaderFilterOverride
		}

		for _, entry := range overrideEntries {
			// Status code filtering: without "always", only add for specific status codes
			if !entry.always && !isAddHeaderStatusCode(code) {
				continue
			}
			resolvedVal := ingressnginx.ReplaceVariables(entry.value, ctx.req, nil, ctx.vars)
			h.Add(entry.key, resolvedVal)
		}

		// more_*, proxy_hide_header, expires: Additive behavior - both apply
		for _, act := range serverCollector.HeaderFilterAdditive {
			_, _ = act(headerRW, ctx.req, ctx)
		}
		for _, act := range locationCollector.HeaderFilterAdditive {
			_, _ = act(headerRW, ctx.req, ctx)
		}
	})

	return wrappedRW
}

// isAddHeaderStatusCode returns true if the status code allows add_header without "always".
func isAddHeaderStatusCode(code int) bool {
	return slices.Contains(addHeaderStatusCodes, code)
}

// executeForwardAuth executes the forward auth subrequest.
// Returns true if the request was terminated (auth failed).
func (s *Snippet) executeForwardAuth(rw http.ResponseWriter, req *http.Request, ctx *actionContext) (bool, error) {
	address := ingressnginx.ReplaceVariables(s.address, req, nil, ctx.vars)

	method := http.MethodGet
	if s.method != "" {
		method = s.method
	}

	forwardReq, err := http.NewRequestWithContext(req.Context(), method, address, nil)
	if err != nil {
		log.Debug().Err(err).Msgf("Error calling %s", address)
		observability.SetStatusErrorf(req.Context(), "Error calling %s. Cause %s", address, err)
		rw.WriteHeader(http.StatusInternalServerError)
		return true, nil
	}

	// Copy headers from original request
	writeHeader(req, forwardReq)

	// Apply auth-snippet directives to the auth request
	if s.authActions != nil {
		// Collect phases from auth-snippet
		authCollector := &PhaseCollector{}

		// Collect and execute rewrite-phase directives (set, if)
		// Other directives are collected into authCollector
		terminated, _, err := s.authActions.Collect(rw, forwardReq, ctx, authCollector)
		if err != nil {
			return false, err
		}
		if terminated {
			return true, nil
		}

		// Execute input filter for auth request
		for _, act := range authCollector.InputFilter {
			if _, err := act(rw, forwardReq, ctx); err != nil {
				return false, err
			}
		}

		// Execute content phase for auth request (proxy_set_header)
		for _, act := range authCollector.Content {
			if _, err := act(rw, forwardReq, ctx); err != nil {
				return false, err
			}
		}
	}

	var forwardSpan trace.Span
	var tracer *tracing.Tracer
	if tracer = tracing.TracerFromContext(forwardReq.Context()); tracer != nil && observability.TracingEnabled(forwardReq.Context()) {
		var tracingCtx context.Context
		tracingCtx, forwardSpan = tracer.Start(forwardReq.Context(), "AuthRequest", trace.WithSpanKind(trace.SpanKindClient))
		defer forwardSpan.End()

		forwardReq = forwardReq.WithContext(tracingCtx)

		tracing.InjectContextIntoCarrier(forwardReq)
		tracer.CaptureClientRequest(forwardSpan, forwardReq)
	}

	forwardResponse, forwardErr := s.client.Do(forwardReq)
	if forwardErr != nil {
		log.Error().Err(forwardErr).Msgf("Error calling %s", s.address)
		observability.SetStatusErrorf(forwardReq.Context(), "Error calling %s. Cause: %s", s.address, forwardErr)

		statusCode := http.StatusInternalServerError
		if errors.Is(forwardErr, context.Canceled) {
			statusCode = httputil.StatusClientClosedRequest
		}

		rw.WriteHeader(statusCode)
		return true, nil
	}
	defer forwardResponse.Body.Close()

	// Ending the forward request span as soon as the response is handled.
	// If any errors happen earlier, this span will be close by the defer instruction.
	if forwardSpan != nil {
		forwardSpan.End()
	}

	// If auth server returns 401 and AuthSigninURL is configured, redirect to signin URL.
	if s.authSigninURL != "" && forwardResponse.StatusCode == http.StatusUnauthorized {
		signinURL := s.authSigninURL
		// If the signin URL doesn't contain "rd=" parameter,
		// add it with the original request URL to match the NGINX behavior.
		if !strings.Contains(signinURL, "rd=") {
			suffix := "rd=$scheme://$best_http_host$escaped_request_uri"
			if !strings.Contains(signinURL, "?") {
				signinURL += "?" + suffix
			} else {
				signinURL += "&" + suffix
			}
		}

		// Use original request for variable replacement (scheme, host, uri)
		signinURL = ingressnginx.ReplaceVariables(signinURL, req, nil, nil)

		log.Debug().Msgf("Redirecting to signin URL: %s", signinURL)
		tracer.CaptureResponse(forwardSpan, forwardResponse.Header, http.StatusFound, trace.SpanKindClient)
		http.Redirect(rw, req, signinURL, http.StatusFound)
		return true, nil
	}

	// Pass the forward response's body and selected headers if it
	// didn't return a response within the range of [200, 300).
	if forwardResponse.StatusCode < http.StatusOK || forwardResponse.StatusCode >= http.StatusMultipleChoices {
		log.Debug().Msgf("Remote error %s. StatusCode: %d", s.address, forwardResponse.StatusCode)

		utils.CopyHeaders(rw.Header(), forwardResponse.Header)
		utils.RemoveHeaders(rw.Header(), hopHeaders...)

		redirectURL, err := forwardResponse.Location()
		if err != nil {
			if !errors.Is(err, http.ErrNoLocation) {
				log.Debug().Err(err).Msgf("Error reading response location header %s", s.address)
				observability.SetStatusErrorf(forwardReq.Context(), "Error reading response location header %s. Cause: %s", s.address, err)

				rw.WriteHeader(http.StatusInternalServerError)
				return true, nil
			}
		} else if redirectURL.String() != "" {
			// Set the location in our response if one was sent back.
			rw.Header().Set("Location", redirectURL.String())
		}

		tracer.CaptureResponse(forwardSpan, forwardResponse.Header, forwardResponse.StatusCode, trace.SpanKindClient)
		rw.WriteHeader(forwardResponse.StatusCode)

		if _, err = io.Copy(rw, forwardResponse.Body); err != nil {
			log.Error().Err(err).Send()
		}
		return true, nil
	}

	// Copy auth response headers to the original request for downstream processing
	for _, headerName := range s.authResponseHeaders {
		headerKey := http.CanonicalHeaderKey(headerName)
		req.Header.Del(headerKey)
		if len(forwardResponse.Header[headerKey]) > 0 {
			req.Header[headerKey] = append([]string(nil), forwardResponse.Header[headerKey]...)
		}
	}

	tracer.CaptureResponse(forwardSpan, forwardResponse.Header, forwardResponse.StatusCode, trace.SpanKindClient)

	ctx.authResponseHeaders = forwardResponse.Header
	return false, nil
}

// headerOnlyWriter is a minimal http.ResponseWriter that only provides header access.
// Used for header filter phase where we only need to modify headers.
type headerOnlyWriter struct {
	header http.Header
}

func (h *headerOnlyWriter) Header() http.Header {
	return h.header
}

func (h *headerOnlyWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (h *headerOnlyWriter) WriteHeader(int) {}

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

// WriteResponse writes the final response based on the action context.
// For redirect status codes (301, 302, 303, 307, 308) with a URL, it performs an HTTP redirect.
// For other status codes, it writes the status code and optional body text.
func WriteResponse(rw http.ResponseWriter, req *http.Request, ctx *actionContext) {
	if ctx.statusCode == 0 {
		return
	}

	if ctx.redirectURL != "" {
		http.Redirect(rw, req, ctx.redirectURL, ctx.statusCode)
		return
	}

	rw.WriteHeader(ctx.statusCode)
	_, _ = rw.Write([]byte(ctx.body))
}

func writeHeader(req, forwardReq *http.Request) {
	utils.CopyHeaders(forwardReq.Header, req.Header)

	RemoveConnectionHeaders(forwardReq)
	utils.RemoveHeaders(forwardReq.Header, hopHeaders...)

	if _, ok := req.Header[userAgentHeader]; !ok {
		// If the incoming request doesn't have a User-Agent header set,
		// don't send the default Go HTTP client User-Agent for the forwarded request.
		forwardReq.Header.Set(userAgentHeader, "")
	}

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		forwardReq.Header.Set(forward.XForwardedFor, clientIP)
	}

	forwardReq.Header.Del(xForwardedMethod)
	if req.Method != "" {
		forwardReq.Header.Set(xForwardedMethod, req.Method)
	}

	forwardReq.Header.Set(forward.XForwardedProto, "http")
	if req.TLS != nil {
		forwardReq.Header.Set(forward.XForwardedProto, "https")
	}

	forwardReq.Header.Del(forward.XForwardedHost)
	if req.Host != "" {
		forwardReq.Header.Set(forward.XForwardedHost, req.Host)
	}

	forwardReq.Header.Del(xForwardedURI)
	if req.URL.RequestURI() != "" {
		forwardReq.Header.Set(xForwardedURI, req.URL.RequestURI())
	}
}
