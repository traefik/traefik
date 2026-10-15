package accessgovernance

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/udp"
	"go.opentelemetry.io/otel/trace"
)

const (
	resultNotEvaluated = "not_evaluated"
)

const typeName = "AccessGovernance"

const (
	maxAccessQueueCapacity = 100000
	maxAccessQueueWorkers  = 64
	maxAccessBodyBytes     = 1024 * 1024
	maxAccessEnqueueWait   = time.Second
	maxDatabaseConnections = 256
)

// RouteMetadata identifies the selected entry point, router, and service.
type RouteMetadata struct {
	EntryPoint string
	Router     string
	Service    string
}

type metadataKey struct{}

// WithRouteMetadata attaches protocol-neutral route metadata to a middleware construction context.
func WithRouteMetadata(ctx context.Context, entryPoint, router, service string) context.Context {
	return context.WithValue(ctx, metadataKey{}, RouteMetadata{
		EntryPoint: entryPoint,
		Router:     router,
		Service:    service,
	})
}

// GetRouteMetadata returns route metadata attached to a request or construction context.
func GetRouteMetadata(ctx context.Context) RouteMetadata {
	value, _ := ctx.Value(metadataKey{}).(RouteMetadata)
	return value
}

func getRouteMetadata(ctx context.Context) RouteMetadata {
	return GetRouteMetadata(ctx)
}

type compiledRoute struct {
	config        dynamic.AccessGovernanceRoute
	path          *regexp.Regexp
	headers       map[string]*regexp.Regexp
	bodyJSONPaths map[string]*regexp.Regexp
	protocols     map[string]struct{}
	methods       map[string]struct{}
}

type compiledIdentity struct {
	config dynamic.AccessGovernanceIdentitySource
	regex  *regexp.Regexp
}

type compiledConfig struct {
	config     dynamic.AccessGovernance
	routes     []compiledRoute
	identities []compiledIdentity
}

type middleware struct {
	next       http.Handler
	name       string
	config     *compiledConfig
	dispatcher *dispatcher
	logger     zerolog.Logger
}

// New creates an HTTP access governance middleware.
func New(ctx context.Context, next http.Handler, config dynamic.AccessGovernance, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if err := resolveRuntimeConfig(&config); err != nil {
		return nil, fmt.Errorf("resolving runtime values: %w", err)
	}
	compiled, err := compileConfig(config)
	if err != nil {
		return nil, err
	}

	dispatch, err := getDispatcher(ctx, name, compiled.config, *logger)
	if err != nil {
		return nil, err
	}

	return &middleware{
		next:       next,
		name:       name,
		config:     compiled,
		dispatcher: dispatch,
		logger:     *logger,
	}, nil
}

func (m *middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	started := time.Now().UTC()
	recorder := newResponseRecorder(rw)

	var bodyCapture *boundedReadCloser
	needsRouteBody := m.config.needsBodyRoute(requestRoute(req))
	needsIdentityBody := needsBody(m.config.identities) && m.config.identity(req, nil) == ""
	if req.Body != nil && (needsRouteBody || needsIdentityBody) {
		var err error
		// Database selection must be deterministic even when the backend does
		// not consume the request body, so body-routed requests are read eagerly.
		if !needsRouteBody && strings.EqualFold(m.config.config.Capture.BodyReadMode, "lazy") {
			bodyCapture = newLazyBodyCapture(req.Body, bodyLimit(m.config.config.Capture.MaxBodyBytes))
		} else {
			bodyCapture, err = newEagerBodyCapture(req.Body, bodyLimit(m.config.config.Capture.MaxBodyBytes))
		}
		req.Body = bodyCapture
		if err != nil {
			m.logger.Warn().Err(err).Msg("Could not eagerly inspect the request body for an access governance identity")
		}
	}

	m.next.ServeHTTP(recorder, req)

	route := requestRoute(req)
	if needsRouteBody && bodyCapture != nil && !bodyCapture.overflow {
		route.Body = parseJSONBody(bodyCapture.Bytes())
	}
	database, governanceRoute := m.config.destinationFor(route)
	if database == "" {
		return
	}

	principal := m.config.identity(req, bodyCapture)
	host, port := splitAddress(req.RemoteAddr)
	completed := time.Now().UTC()
	status := recorder.Status()
	if status == 0 {
		status = http.StatusOK
	}

	requestBytes := req.ContentLength
	if requestBytes < 0 {
		requestBytes = 0
		if bodyCapture != nil {
			requestBytes = max(bodyCapture.readBytes, bodyCapture.inspectedBytes)
		}
	}

	spanContext := trace.SpanFromContext(req.Context()).SpanContext()
	var traceID, spanID string
	if spanContext.HasTraceID() {
		traceID = spanContext.TraceID().String()
	}
	if spanContext.HasSpanID() {
		spanID = spanContext.SpanID().String()
	}

	query := ""
	if m.config.config.Capture.StoreQuery {
		query = req.URL.RawQuery
	}
	userAgent := ""
	if m.config.config.Capture.StoreUserAgent {
		userAgent = req.UserAgent()
	}

	metadata := map[string]any{
		"tls":        tlsMetadata(req.TLS),
		"route_name": governanceRoute,
	}
	if bodyCapture != nil && bodyCapture.overflow {
		metadata["identity_body_truncated"] = true
	}

	dimensions := m.config.config.Dimensions
	event := httpEvent{
		EventID:        newEventID(),
		OccurredAt:     started,
		CompletedAt:    completed,
		Protocol:       protocolName(req),
		EntryPoint:     route.EntryPoint,
		Router:         route.Router,
		Service:        route.Service,
		Middleware:     m.name,
		TenantID:       dimensions.TenantID,
		PrincipalType:  principalType(principal),
		PrincipalID:    principal,
		AuthnMethod:    dimensions.AuthnMethod,
		AuthnResult:    authnResult(status, dimensions.AuthnMethod),
		PolicyID:       dimensions.PolicyID,
		AuthzResult:    authzResult(status, dimensions.PolicyID),
		BillingAccount: dimensions.BillingAccount,
		Meter:          dimensions.Meter,
		BillableUnits:  billableUnits(dimensions.Meter),
		ClientIP:       host,
		ClientPort:     port,
		Method:         req.Method,
		Scheme:         requestScheme(req),
		Host:           req.Host,
		Path:           req.URL.EscapedPath(),
		Query:          query,
		StatusCode:     status,
		DurationUS:     completed.Sub(started).Microseconds(),
		RequestBytes:   requestBytes,
		ResponseBytes:  recorder.Size(),
		TraceID:        traceID,
		SpanID:         spanID,
		RequestID:      firstHeader(req.Header, "X-Request-ID", "X-Correlation-ID"),
		UserAgent:      userAgent,
		Labels:         dimensions.Labels,
		Metadata:       metadata,
	}

	if !m.dispatcher.enqueue(task{database: database, http: &event}) {
		m.dispatcher.logDrop(database, "http")
	}
}

type tcpMiddleware struct {
	next       tcp.Handler
	name       string
	config     *compiledConfig
	dispatcher *dispatcher
	metadata   RouteMetadata
	logger     zerolog.Logger
}

// NewTCP creates a TCP access governance middleware.
func NewTCP(ctx context.Context, next tcp.Handler, config dynamic.AccessGovernance, name string) (tcp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating TCP middleware")

	if err := resolveRuntimeConfig(&config); err != nil {
		return nil, fmt.Errorf("resolving runtime values: %w", err)
	}
	compiled, err := compileConfig(config)
	if err != nil {
		return nil, err
	}

	dispatch, err := getDispatcher(ctx, name, compiled.config, *logger)
	if err != nil {
		return nil, err
	}

	return &tcpMiddleware{
		next:       next,
		name:       name,
		config:     compiled,
		dispatcher: dispatch,
		metadata:   getRouteMetadata(ctx),
		logger:     *logger,
	}, nil
}

func (m *tcpMiddleware) ServeTCP(conn tcp.WriteCloser) {
	started := time.Now().UTC()
	counted := &countingTCPConn{WriteCloser: conn}
	m.next.ServeTCP(counted)
	completed := time.Now().UTC()

	route := eventRoute{
		Protocol:   "tcp",
		EntryPoint: m.metadata.EntryPoint,
		Router:     m.metadata.Router,
		Service:    m.metadata.Service,
	}
	database, governanceRoute := m.config.destinationFor(route)
	if database == "" {
		return
	}

	event := networkEvent{
		EventID:        newEventID(),
		OccurredAt:     started,
		CompletedAt:    completed,
		Protocol:       "tcp",
		EntryPoint:     route.EntryPoint,
		Router:         route.Router,
		Service:        route.Service,
		Middleware:     m.name,
		TenantID:       m.config.config.Dimensions.TenantID,
		PrincipalType:  "network",
		AuthnResult:    resultNotEvaluated,
		AuthnMethod:    m.config.config.Dimensions.AuthnMethod,
		PolicyID:       m.config.config.Dimensions.PolicyID,
		AuthzResult:    resultNotEvaluated,
		BillingAccount: m.config.config.Dimensions.BillingAccount,
		Meter:          m.config.config.Dimensions.Meter,
		BillableUnits:  billableUnits(m.config.config.Dimensions.Meter),
		ClientAddr:     addressString(conn.RemoteAddr()),
		LocalAddr:      addressString(conn.LocalAddr()),
		DurationUS:     completed.Sub(started).Microseconds(),
		BytesReceived:  counted.read.Load(),
		BytesSent:      counted.written.Load(),
		Outcome:        "completed",
		Labels:         m.config.config.Dimensions.Labels,
		Metadata:       map[string]any{"route_name": governanceRoute},
	}

	if !m.dispatcher.enqueue(task{database: database, network: &event}) {
		m.dispatcher.logDrop(database, "tcp")
	}
}

type udpMiddleware struct {
	next       udp.Handler
	name       string
	config     *compiledConfig
	dispatcher *dispatcher
	metadata   RouteMetadata
	logger     zerolog.Logger
}

// NewUDP creates a UDP access governance middleware.
func NewUDP(ctx context.Context, next udp.Handler, config dynamic.AccessGovernance, name string) (udp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating UDP middleware")

	if err := resolveRuntimeConfig(&config); err != nil {
		return nil, fmt.Errorf("resolving runtime values: %w", err)
	}
	compiled, err := compileConfig(config)
	if err != nil {
		return nil, err
	}

	dispatch, err := getDispatcher(ctx, name, compiled.config, *logger)
	if err != nil {
		return nil, err
	}

	return &udpMiddleware{
		next:       next,
		name:       name,
		config:     compiled,
		dispatcher: dispatch,
		metadata:   getRouteMetadata(ctx),
		logger:     *logger,
	}, nil
}

func (m *udpMiddleware) ServeUDP(conn *udp.Conn) {
	started := time.Now().UTC()
	beforeReceived, beforeSent := conn.Stats()
	m.next.ServeUDP(conn)
	completed := time.Now().UTC()
	afterReceived, afterSent := conn.Stats()

	route := eventRoute{
		Protocol:   "udp",
		EntryPoint: m.metadata.EntryPoint,
		Router:     m.metadata.Router,
		Service:    m.metadata.Service,
	}
	database, governanceRoute := m.config.destinationFor(route)
	if database == "" {
		return
	}

	event := networkEvent{
		EventID:        newEventID(),
		OccurredAt:     started,
		CompletedAt:    completed,
		Protocol:       "udp",
		EntryPoint:     route.EntryPoint,
		Router:         route.Router,
		Service:        route.Service,
		Middleware:     m.name,
		TenantID:       m.config.config.Dimensions.TenantID,
		PrincipalType:  "network",
		AuthnResult:    resultNotEvaluated,
		AuthnMethod:    m.config.config.Dimensions.AuthnMethod,
		PolicyID:       m.config.config.Dimensions.PolicyID,
		AuthzResult:    resultNotEvaluated,
		BillingAccount: m.config.config.Dimensions.BillingAccount,
		Meter:          m.config.config.Dimensions.Meter,
		BillableUnits:  billableUnits(m.config.config.Dimensions.Meter),
		ClientAddr:     addressString(conn.RemoteAddr()),
		LocalAddr:      addressString(conn.LocalAddr()),
		DurationUS:     completed.Sub(started).Microseconds(),
		BytesReceived:  afterReceived - beforeReceived,
		BytesSent:      afterSent - beforeSent,
		Outcome:        "completed",
		Labels:         m.config.config.Dimensions.Labels,
		Metadata:       map[string]any{"route_name": governanceRoute},
	}

	if !m.dispatcher.enqueue(task{database: database, network: &event}) {
		m.dispatcher.logDrop(database, "udp")
	}
}

func compileConfig(config dynamic.AccessGovernance) (*compiledConfig, error) {
	if len(config.Databases) == 0 {
		return nil, errors.New("access governance requires at least one database")
	}
	if config.Queue.Capacity > maxAccessQueueCapacity {
		return nil, fmt.Errorf("access governance queue capacity must not exceed %d", maxAccessQueueCapacity)
	}
	if config.Queue.Workers > maxAccessQueueWorkers {
		return nil, fmt.Errorf("access governance queue workers must not exceed %d", maxAccessQueueWorkers)
	}
	if config.Queue.BatchSize > 1000 {
		return nil, errors.New("access governance queue batchSize must not exceed 1000")
	}
	if time.Duration(config.Queue.EnqueueWait) > maxAccessEnqueueWait {
		return nil, fmt.Errorf("access governance queue enqueueWait must not exceed %s", maxAccessEnqueueWait)
	}
	if config.Capture.MaxBodyBytes > maxAccessBodyBytes {
		return nil, fmt.Errorf("access governance capture maxBodyBytes must not exceed %d", maxAccessBodyBytes)
	}
	if config.DefaultDatabase == "" && len(config.Databases) == 1 {
		for name := range config.Databases {
			config.DefaultDatabase = name
		}
	}
	if config.DefaultDatabase != "" {
		if _, ok := config.Databases[config.DefaultDatabase]; !ok {
			return nil, fmt.Errorf("access governance default database %q does not exist", config.DefaultDatabase)
		}
	}
	for name, database := range config.Databases {
		if database.MaxOpenConnections > maxDatabaseConnections || database.MaxIdleConnections > maxDatabaseConnections {
			return nil, fmt.Errorf("access governance database %q connection limits must not exceed %d", name, maxDatabaseConnections)
		}
		httpTable := tableName(database.HTTPTable, defaultHTTPTable)
		networkTable := tableName(database.NetworkTable, defaultNetworkTable)
		if !identifierPattern.MatchString(httpTable) || !identifierPattern.MatchString(networkTable) {
			return nil, fmt.Errorf("access governance database %q table names must match %s", name, identifierPattern)
		}
	}

	compiled := &compiledConfig{config: config}
	for _, route := range config.Routes {
		if _, ok := config.Databases[route.Database]; !ok {
			return nil, fmt.Errorf("access governance route %q references unknown database %q", route.Name, route.Database)
		}

		item := compiledRoute{
			config:        route,
			headers:       make(map[string]*regexp.Regexp),
			bodyJSONPaths: make(map[string]*regexp.Regexp),
			protocols:     stringSet(route.Protocols, true),
			methods:       stringSet(route.Methods, false),
		}
		if route.PathRegex != "" {
			expression, err := regexp.Compile(route.PathRegex)
			if err != nil {
				return nil, fmt.Errorf("compiling access governance route %q path: %w", route.Name, err)
			}
			item.path = expression
		}
		for name, expression := range route.Headers {
			value, err := regexp.Compile(expression)
			if err != nil {
				return nil, fmt.Errorf("compiling access governance route %q header %q: %w", route.Name, name, err)
			}
			item.headers[http.CanonicalHeaderKey(name)] = value
		}
		for path, expression := range route.BodyJSONPaths {
			value, err := regexp.Compile(expression)
			if err != nil {
				return nil, fmt.Errorf("compiling access governance route %q body JSON path %q: %w", route.Name, path, err)
			}
			item.bodyJSONPaths[path] = value
		}
		compiled.routes = append(compiled.routes, item)
	}

	for _, identity := range config.Identity {
		item := compiledIdentity{config: identity}
		if identity.Regex != "" {
			expression, err := regexp.Compile(identity.Regex)
			if err != nil {
				return nil, fmt.Errorf("compiling access governance identity regex: %w", err)
			}
			item.regex = expression
		}
		compiled.identities = append(compiled.identities, item)
	}

	return compiled, nil
}

func resolveRuntimeConfig(config *dynamic.AccessGovernance) error {
	connectionIDs := make(map[string]string, len(config.Databases))
	connectionFiles := make(map[string]string, len(config.Databases))
	for name, database := range config.Databases {
		identity, err := runtimeconfig.SourceIdentity(database.DSN, database.DSNFile, config.RuntimeValues)
		if err != nil {
			return fmt.Errorf("database %q: %w", name, err)
		}
		connectionIDs[name] = identity
		connectionFiles[name] = runtimeconfig.ExactReferenceFile(database.DSN, config.RuntimeValues)
	}
	if err := runtimeconfig.Resolve(config.RuntimeValues, config); err != nil {
		return err
	}
	for name, database := range config.Databases {
		database.ConnectionID = connectionIDs[name]
		if connectionFiles[name] != "" {
			database.DSN = ""
			database.DSNFile = connectionFiles[name]
		}
		config.Databases[name] = database
	}
	return nil
}

type eventRoute struct {
	Protocol   string
	EntryPoint string
	Router     string
	Service    string
	Method     string
	Host       string
	Path       string
	Header     http.Header
	Body       any
}

func requestRoute(req *http.Request) eventRoute {
	metadata := getRouteMetadata(req.Context())
	return eventRoute{
		Protocol:   protocolName(req),
		EntryPoint: metadata.EntryPoint,
		Router:     metadata.Router,
		Service:    metadata.Service,
		Method:     req.Method,
		Host:       req.Host,
		Path:       req.URL.Path,
		Header:     req.Header,
	}
}

func (c *compiledConfig) databaseFor(event eventRoute) string {
	database, _ := c.destinationFor(event)
	return database
}

func (c *compiledConfig) destinationFor(event eventRoute) (database, routeName string) {
	for _, route := range c.routes {
		if route.matches(event) {
			return route.config.Database, route.config.Name
		}
	}
	return c.config.DefaultDatabase, ""
}

func (r compiledRoute) matches(event eventRoute) bool {
	if !r.matchesMetadata(event) {
		return false
	}
	for path, expression := range r.bodyJSONPaths {
		if !expression.MatchString(routeMatchValue(selectJSON(event.Body, path))) {
			return false
		}
	}
	return true
}

func (r compiledRoute) matchesMetadata(event eventRoute) bool {
	if len(r.protocols) > 0 {
		_, exact := r.protocols[strings.ToLower(event.Protocol)]
		_, httpFamily := r.protocols["http"]
		if !exact && !(httpFamily && strings.HasPrefix(strings.ToLower(event.Protocol), "http")) {
			return false
		}
	}
	if !matchesAny(r.config.EntryPoints, event.EntryPoint) ||
		!matchesAny(r.config.Routers, event.Router) ||
		!matchesAny(r.config.Services, event.Service) ||
		!matchesAnyFold(r.config.Hosts, event.Host) {
		return false
	}
	if len(r.methods) > 0 {
		if _, ok := r.methods[strings.ToUpper(event.Method)]; !ok {
			return false
		}
	}
	if r.path != nil && !r.path.MatchString(event.Path) {
		return false
	}
	for name, expression := range r.headers {
		if !expression.MatchString(event.Header.Get(name)) {
			return false
		}
	}
	return true
}

func (c *compiledConfig) needsBodyRoute(event eventRoute) bool {
	for _, route := range c.routes {
		if len(route.bodyJSONPaths) > 0 && route.matchesMetadata(event) {
			return true
		}
	}
	return false
}

func routeMatchValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}

func (c *compiledConfig) identity(req *http.Request, body *boundedReadCloser) string {
	for _, source := range c.identities {
		value := identityValue(source.config, req, body)
		value = decodeIdentity(source.config.Decode, value)
		if source.regex != nil {
			match := source.regex.FindStringSubmatch(value)
			if len(match) == 0 {
				continue
			}
			group := source.config.Group
			if group <= 0 {
				group = 1
			}
			if group >= len(match) {
				continue
			}
			value = match[group]
		}
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func identityValue(source dynamic.AccessGovernanceIdentitySource, req *http.Request, body *boundedReadCloser) string {
	switch strings.ToLower(source.Kind) {
	case "header":
		return req.Header.Get(source.Name)
	case "cookie":
		cookie, err := req.Cookie(source.Name)
		if err == nil {
			return cookie.Value
		}
	case "query":
		return req.URL.Query().Get(source.Name)
	case "basic":
		username, _, ok := req.BasicAuth()
		if ok {
			return username
		}
	case "clientcert":
		if req.TLS != nil && len(req.TLS.PeerCertificates) > 0 {
			if source.Name == "commonName" {
				return req.TLS.PeerCertificates[0].Subject.CommonName
			}
			return req.TLS.PeerCertificates[0].Subject.String()
		}
	case "json":
		if body == nil || body.overflow {
			return ""
		}
		value := parseJSONBody(body.Bytes())
		return stringify(selectJSON(value, firstNonEmpty(source.JSONPath, source.Name)))
	case "form":
		if body == nil || body.overflow {
			return ""
		}
		values, err := url.ParseQuery(string(body.Bytes()))
		if err == nil {
			return values.Get(source.Name)
		}
	}
	return ""
}

func parseJSONBody(raw []byte) any {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil
	}
	return value
}

func selectJSON(value any, path string) any {
	path = strings.TrimPrefix(strings.TrimSpace(path), "$")
	path = strings.TrimPrefix(path, ".")
	if path == "" {
		return value
	}

	current := value
	for segment := range strings.SplitSeq(path, ".") {
		name, index, hasIndex := parsePathSegment(segment)
		if name != "" {
			object, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current = object[name]
		}
		if hasIndex {
			array, ok := current.([]any)
			if !ok || index < 0 || index >= len(array) {
				return nil
			}
			current = array[index]
		}
	}
	return current
}

func parsePathSegment(segment string) (string, int, bool) {
	open := strings.LastIndex(segment, "[")
	if open < 0 || !strings.HasSuffix(segment, "]") {
		return segment, 0, false
	}
	index, err := strconv.Atoi(segment[open+1 : len(segment)-1])
	if err != nil {
		return segment, 0, false
	}
	return segment[:open], index, true
}

func decodeIdentity(kind, value string) string {
	switch strings.ToLower(kind) {
	case "url":
		decoded, err := url.QueryUnescape(value)
		if err == nil {
			return decoded
		}
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err == nil {
			return string(decoded)
		}
	case "base64url":
		decoded, err := base64.RawURLEncoding.DecodeString(value)
		if err == nil {
			return string(decoded)
		}
	}
	return value
}

type boundedReadCloser struct {
	source         io.ReadCloser
	buffer         bytes.Buffer
	limit          int64
	readBytes      int64
	inspectedBytes int64
	overflow       bool
	capture        bool
}

func (b *boundedReadCloser) Read(data []byte) (int, error) {
	n, err := b.source.Read(data)
	b.readBytes += int64(n)
	if b.capture {
		remaining := b.limit - int64(b.buffer.Len())
		if remaining > 0 {
			write := min(int64(n), remaining)
			_, _ = b.buffer.Write(data[:write])
		}
		if b.readBytes > b.limit {
			b.overflow = true
		}
	}
	return n, err
}

func (b *boundedReadCloser) Close() error {
	return b.source.Close()
}

func (b *boundedReadCloser) Bytes() []byte {
	return b.buffer.Bytes()
}

func newLazyBodyCapture(source io.ReadCloser, limit int64) *boundedReadCloser {
	return &boundedReadCloser{
		source:  source,
		limit:   limit,
		capture: true,
	}
}

func newEagerBodyCapture(source io.ReadCloser, limit int64) (*boundedReadCloser, error) {
	value, err := io.ReadAll(io.LimitReader(source, limit+1))
	capture := &boundedReadCloser{
		limit:          limit,
		inspectedBytes: int64(len(value)),
		overflow:       int64(len(value)) > limit || err != nil,
	}
	_, _ = capture.buffer.Write(value[:min(int64(len(value)), limit)])

	if int64(len(value)) > limit || err != nil {
		capture.source = &multiReadCloser{
			Reader: io.MultiReader(bytes.NewReader(value), source),
			closer: source,
		}
		return capture, err
	}
	if closeErr := source.Close(); closeErr != nil {
		capture.source = io.NopCloser(bytes.NewReader(value))
		return capture, closeErr
	}
	capture.source = io.NopCloser(bytes.NewReader(value))
	return capture, nil
}

type multiReadCloser struct {
	io.Reader

	closer io.Closer
}

func (m *multiReadCloser) Close() error {
	return m.closer.Close()
}

type responseRecorder struct {
	http.ResponseWriter

	status int
	size   int64
}

func newResponseRecorder(rw http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: rw}
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(data)
	r.size += int64(n)
	return n, err
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Size() int64 {
	return r.size
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) Flush() {
	if value, ok := r.ResponseWriter.(http.Flusher); ok {
		value.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	value, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer %T does not support hijacking", r.ResponseWriter)
	}
	return value.Hijack()
}

func (r *responseRecorder) Push(target string, options *http.PushOptions) error {
	value, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return value.Push(target, options)
}

type countingTCPConn struct {
	tcp.WriteCloser

	read    atomic.Int64
	written atomic.Int64
}

func (c *countingTCPConn) Read(data []byte) (int, error) {
	n, err := c.WriteCloser.Read(data)
	c.read.Add(int64(n))
	return n, err
}

func (c *countingTCPConn) Write(data []byte) (int, error) {
	n, err := c.WriteCloser.Write(data)
	c.written.Add(int64(n))
	return n, err
}

func protocolName(req *http.Request) string {
	switch req.ProtoMajor {
	case 3:
		return "http3"
	case 2:
		return "http2"
	default:
		return "http"
	}
}

func requestScheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	if value := req.Header.Get("X-Forwarded-Proto"); value != "" {
		return value
	}
	return "http"
}

func authnResult(status int, method string) string {
	if status == http.StatusUnauthorized {
		return "denied"
	}
	if method != "" {
		return "allowed"
	}
	return resultNotEvaluated
}

func authzResult(status int, policy string) string {
	if status == http.StatusForbidden {
		return "denied"
	}
	if policy != "" {
		return "allowed"
	}
	return resultNotEvaluated
}

func billableUnits(meter string) float64 {
	if meter == "" {
		return 0
	}
	return 1
}

func principalType(principal string) string {
	if principal == "" {
		return "anonymous"
	}
	return "oa"
}

func newEventID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err == nil {
		return hex.EncodeToString(value)
	}
	return fmt.Sprintf("%032x", time.Now().UnixNano())
}

func splitAddress(value string) (string, string) {
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return value, ""
	}
	return host, port
}

func addressString(value net.Addr) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func tlsMetadata(state *tls.ConnectionState) map[string]any {
	if state == nil {
		return nil
	}
	result := map[string]any{
		"version":      state.Version,
		"cipher_suite": state.CipherSuite,
		"server_name":  state.ServerName,
	}
	if len(state.PeerCertificates) > 0 {
		result["client_subject"] = state.PeerCertificates[0].Subject.String()
	}
	return result
}

func firstHeader(header http.Header, names ...string) string {
	for _, name := range names {
		if value := header.Get(name); value != "" {
			return value
		}
	}
	return ""
}

func bodyLimit(configured int64) int64 {
	if configured <= 0 {
		return 64 * 1024
	}
	return configured
}

func needsBody(sources []compiledIdentity) bool {
	for _, source := range sources {
		if source.config.Kind == "json" || source.config.Kind == "form" {
			return true
		}
	}
	return false
}

func matchesAny(values []string, candidate string) bool {
	if len(values) == 0 {
		return true
	}
	return slices.Contains(values, candidate)
}

func matchesAnyFold(values []string, candidate string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
}

func stringSet(values []string, lower bool) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if lower {
			value = strings.ToLower(value)
		} else {
			value = strings.ToUpper(value)
		}
		result[value] = struct{}{}
	}
	return result
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
