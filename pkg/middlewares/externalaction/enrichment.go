package externalaction

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

const maxEnrichmentBytes = 1024 * 1024

type enrichmentContextKey struct{}

type compiledEnrichment struct {
	name          string
	input, output *compiledTemplate
	queries       []*compiledQuery
}

type compiledQuery struct {
	name         string
	when, result *compiledTemplate
	target       *httpTarget
	timeout      time.Duration
	origins      map[string]bool
	slots        chan struct{}
	ttl          time.Duration
	maxEntries   int
	maxBytes     int64
	mu           sync.Mutex
	cache        map[[32]byte]queryCacheEntry
	bytes        int64
	pagination   *dynamic.ExternalActionPollerPagination
}

type queryCacheEntry struct {
	body    []byte
	status  int
	headers http.Header
	expires time.Time
}

var enrichmentRegistry = struct {
	sync.Mutex

	values map[dispatcherKey]*compiledEnrichment
}{values: make(map[dispatcherKey]*compiledEnrichment)}

func getEnrichment(ctx context.Context, name string, config dynamic.QueryEnrichment) (*compiledEnrichment, error) {
	encoded, err := json.Marshal(struct {
		Name   string                  `json:"name"`
		Config dynamic.QueryEnrichment `json:"config"`
	}{name, config})
	if err != nil {
		return nil, err
	}
	key := dispatcherKey{done: ctx.Done(), config: sha256.Sum256(encoded)}
	enrichmentRegistry.Lock()
	defer enrichmentRegistry.Unlock()
	if existing := enrichmentRegistry.values[key]; existing != nil {
		return existing, nil
	}
	compiled, err := compileEnrichment(name, config)
	if err != nil {
		return nil, err
	}
	// Context-free callers are tests or validators; do not retain them forever.
	if ctx.Done() != nil {
		enrichmentRegistry.values[key] = compiled
		go func() {
			<-ctx.Done()
			enrichmentRegistry.Lock()
			delete(enrichmentRegistry.values, key)
			enrichmentRegistry.Unlock()
			for _, query := range compiled.queries {
				query.target.client.CloseIdleConnections()
			}
		}()
	}
	return compiled, nil
}

// NewQueryEnrichment defers lookups until ExternalAction selects a concrete event.
// A request that never triggers an action performs no remote lookups.
func NewQueryEnrichment(ctx context.Context, next http.Handler, config dynamic.QueryEnrichment, name string) (http.Handler, error) {
	compiled, err := getEnrichment(ctx, name, config)
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		values := append([]*compiledEnrichment(nil), enrichmentsFromContext(req.Context())...)
		values = append(values, compiled)
		next.ServeHTTP(rw, req.WithContext(context.WithValue(req.Context(), enrichmentContextKey{}, values)))
	}), nil
}

func enrichmentsFromContext(ctx context.Context) []*compiledEnrichment {
	values, _ := ctx.Value(enrichmentContextKey{}).([]*compiledEnrichment)
	return values
}

func compileEnrichment(name string, config dynamic.QueryEnrichment) (*compiledEnrichment, error) {
	// JSON intentionally hides RuntimeValue.Value from the dashboard. Use the
	// generated deep copy so Secret material survives without sharing mutable maps.
	resolved := *config.DeepCopy()
	credentialConfig := dynamic.ExternalAction{ConfigurationID: name, RuntimeValues: resolved.RuntimeValues, CredentialSets: resolved.CredentialSets}
	if err := resolveRuntimeConfig(&credentialConfig); err != nil {
		return nil, err
	}
	if err := runtimeconfig.Resolve(resolved.RuntimeValues, &resolved); err != nil {
		return nil, err
	}
	credentials, err := compileCredentialSets(credentialConfig.CredentialSets)
	if err != nil {
		return nil, err
	}
	if len(resolved.Queries) == 0 || len(resolved.Queries) > 16 {
		return nil, errors.New("queryEnrichment requires between 1 and 16 queries")
	}
	value := &compiledEnrichment{name: name}
	if resolved.InputTemplate != "" {
		value.input, err = newTemplate("enrichment-input", resolved.InputTemplate)
		if err != nil {
			return nil, err
		}
	}
	if resolved.OutputTemplate != "" {
		value.output, err = newTemplate("enrichment-output", resolved.OutputTemplate)
		if err != nil {
			return nil, err
		}
	}
	names := make(map[string]bool)
	for _, query := range resolved.Queries {
		if query.Name == "" || names[query.Name] {
			return nil, errors.New("query names must be nonempty and unique")
		}
		names[query.Name] = true
		compiled, err := compileQuery(query, credentials)
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", query.Name, err)
		}
		value.queries = append(value.queries, compiled)
	}
	return value, nil
}

func compileQuery(config dynamic.EnrichmentQuery, credentials map[string]*compiledAuth) (*compiledQuery, error) {
	if config.Timeout == 0 {
		config.Timeout = ptypes.Duration(2 * time.Second)
	}
	if time.Duration(config.Timeout) <= 0 || time.Duration(config.Timeout) > 30*time.Second {
		return nil, errors.New("timeout must be positive and at most 30s")
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 8
	}
	if config.MaxConcurrent < 1 || config.MaxConcurrent > 64 {
		return nil, errors.New("maxConcurrent must be between 1 and 64")
	}
	if config.CacheEntries == 0 {
		config.CacheEntries = 256
	}
	if config.CacheBytes == 0 {
		config.CacheBytes = 8 * 1024 * 1024
	}
	if config.CacheEntries < 1 || config.CacheEntries > 4096 || config.CacheBytes < 1 || config.CacheBytes > 64*1024*1024 {
		return nil, errors.New("cache limits exceed allowed bounds")
	}
	if time.Duration(config.CacheTTL) < 0 || time.Duration(config.CacheTTL) > time.Hour {
		return nil, errors.New("cacheTTL must be between 0 and 1h")
	}
	if config.HTTP.MaxResponseBytes == 0 {
		config.HTTP.MaxResponseBytes = maxEnrichmentBytes
	}
	if config.HTTP.MaxResponseBytes < 1 || config.HTTP.MaxResponseBytes > maxEnrichmentBytes {
		return nil, errors.New("maxResponseBytes must be between 1 and 1048576")
	}
	if config.HTTP.Method == "" {
		config.HTTP.Method = http.MethodGet
	}
	config.HTTP.Method = strings.ToUpper(config.HTTP.Method)
	if !oneOf(config.HTTP.Method, http.MethodGet, http.MethodPost) {
		return nil, errors.New("lookup method must be GET or POST")
	}
	if config.HTTP.URL == "" {
		return nil, errors.New("lookup URL is required")
	}
	query := &compiledQuery{
		name: config.Name, timeout: time.Duration(config.Timeout), origins: make(map[string]bool),
		slots: make(chan struct{}, config.MaxConcurrent), ttl: time.Duration(config.CacheTTL),
		maxEntries: config.CacheEntries, maxBytes: config.CacheBytes, cache: make(map[[32]byte]queryCacheEntry),
	}
	if len(config.AllowedOrigins) == 0 {
		return nil, errors.New("allowedOrigins is required")
	}
	for _, origin := range config.AllowedOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Host == "" || !oneOf(parsed.Scheme, "http", "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" {
			return nil, errors.New("allowedOrigins must contain only literal http(s) scheme and host origins")
		}
		query.origins[parsed.Scheme+"://"+parsed.Host] = true
	}
	var err error
	if config.When != "" {
		query.when, err = newTemplate("enrichment-when", config.When)
		if err != nil {
			return nil, err
		}
	}
	if config.ResultTemplate != "" {
		query.result, err = newTemplate("enrichment-result", config.ResultTemplate)
		if err != nil {
			return nil, err
		}
	}
	target, err := newHTTPTarget(dynamic.ExternalActionRule{HTTP: &config.HTTP, Body: config.Body}, credentials)
	if err != nil {
		return nil, err
	}
	query.target = target.(*httpTarget)
	if config.Pagination != nil {
		p := config.Pagination
		if config.HTTP.Method != http.MethodGet || p.PageParameter == "" || p.PageSizeParameter == "" || p.PageParameter == p.PageSizeParameter || p.ItemsJSONPath == "" || p.PageSize < 1 || p.PageSize > 1000 || p.MaxPages < 1 || p.MaxPages > 100 || p.StartPage < 0 {
			return nil, errors.New("pagination requires GET, distinct page parameters, itemsJSONPath, pageSize 1..1000 and maxPages 1..100")
		}
		query.pagination = p
	}
	// Bound headers separately from the JSON response body and cache only data.
	if transport, ok := query.target.client.Transport.(*http.Transport); ok {
		transport.MaxResponseHeaderBytes = 32 * 1024
	}
	query.target.perform = query.performLookupRequest
	return query, nil
}

// Payload is the already-rendered destination body. It remains event-local even
// when several workers expand the same source response into independent events.
func enrichPayload(ctx context.Context, data *actionData, body string, configured []*compiledEnrichment) (*actionData, string, error) {
	values := append(append([]*compiledEnrichment(nil), configured...), data.deferredEnrichments...)
	if len(values) == 0 {
		return data, body, nil
	}
	result := data.clone()
	result.Context = ctx
	result.SourceAuthParameters = cloneMap(data.AuthParameters)
	result.Enrichment = make(map[string]any)
	result.deferredEnrichments = nil
	var err error
	result.Payload, err = decodeEnrichmentJSON(body)
	if err != nil {
		return nil, "", fmt.Errorf("enrichment requires a JSON destination body: %w", err)
	}
	// Templates may use Sprig's set/merge. Isolate the input trees from sibling events.
	for _, field := range []*any{&result.Body, &result.Response.Body, &result.Item} {
		encoded, err := json.Marshal(*field)
		if err != nil {
			return nil, "", err
		}
		*field, err = decodeJSON(encoded)
		if err != nil {
			return nil, "", err
		}
	}
	seen := make(map[string]bool)
	for _, enrichment := range values {
		if seen[enrichment.name] {
			continue
		}
		seen[enrichment.name] = true
		if len(seen) > 8 {
			return nil, "", errors.New("at most 8 query enrichment middlewares may be used per event")
		}
		if err := enrichment.apply(ctx, result); err != nil {
			return nil, "", fmt.Errorf("query enrichment %q: %w", enrichment.name, err)
		}
	}
	encoded, err := json.Marshal(result.Payload)
	if err != nil {
		return nil, "", err
	}
	if len(encoded) > maxEnrichmentBytes {
		return nil, "", errors.New("enriched payload exceeds 1048576 bytes")
	}
	return result, string(encoded), nil
}

func (e *compiledEnrichment) apply(ctx context.Context, data *actionData) error {
	data.Input = nil
	if e.input != nil {
		input, err := renderEnrichmentJSON(e.input, data)
		if err != nil {
			return fmt.Errorf("input template: %w", err)
		}
		data.Input = input
	}
	for _, query := range e.queries {
		if query.when != nil {
			value, err := renderEnrichmentTemplate(query.when, data)
			if err != nil {
				return fmt.Errorf("query %q condition: %w", query.name, err)
			}
			matches, err := strconv.ParseBool(strings.TrimSpace(value))
			if err != nil {
				return fmt.Errorf("query %q condition must return true or false", query.name)
			}
			if !matches {
				continue
			}
		}
		if _, exists := data.Enrichment[query.name]; exists {
			return fmt.Errorf("query result name %q is already used", query.name)
		}
		queryContext, cancel := context.WithTimeout(ctx, query.timeout)
		value, err := query.lookup(queryContext, data)
		cancel()
		if err != nil {
			return fmt.Errorf("query %q failed: %w", query.name, err)
		}
		data.Enrichment[query.name] = value
	}
	if e.output != nil {
		payload, err := renderEnrichmentJSON(e.output, data)
		if err != nil {
			return fmt.Errorf("output template: %w", err)
		}
		data.Payload = payload
	}
	return nil
}

func (q *compiledQuery) lookup(ctx context.Context, data *actionData) (any, error) {
	maxPages, startPage := 1, 1
	if p := q.pagination; p != nil {
		maxPages = p.MaxPages
		if p.StartPage > 0 {
			startPage = p.StartPage
		}
	}
	pages := make([]any, 0)
	retained, observed := 0, 0
	for page := startPage; page < startPage+maxPages; page++ {
		queryData := *data
		queryData.LookupTime = time.Now()
		queryData.credentialSet = ""
		queryData.expectedCredentialIdentity = [32]byte{}
		if p := q.pagination; p != nil {
			queryData.lookupPageQuery = url.Values{p.PageParameter: {strconv.Itoa(page)}, p.PageSizeParameter: {strconv.Itoa(p.PageSize)}}
		}
		var mapped any
		done := q.pagination == nil
		err := q.target.run(ctx, &queryData, func(raw []byte) error {
			parsed, err := decodeJSON(raw)
			if err != nil {
				return errors.New("lookup response must contain valid JSON")
			}
			if p := q.pagination; p != nil {
				items, ok := selectJSON(parsed, p.ItemsJSONPath).([]any)
				if !ok {
					return errors.New("pagination itemsJSONPath must select an array")
				}
				observed += len(items)
				done = len(items) < p.PageSize
				if p.TotalJSONPath != "" {
					total, err := strconv.Atoi(responseValueString(selectJSON(parsed, p.TotalJSONPath)))
					if err != nil || total < 0 || observed > total {
						return errors.New("pagination total is missing or inconsistent")
					}
					done = observed == total
					if len(items) == 0 && !done {
						return errors.New("pagination ended before advertised total")
					}
				}
			}
			mapped = parsed
			if q.result != nil {
				mappingData := *data
				mappingData.Result = parsed
				mapped, err = renderEnrichmentJSON(q.result, &mappingData)
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		if q.pagination == nil {
			return mapped, nil
		}
		encoded, err := json.Marshal(mapped)
		if err != nil {
			return nil, err
		}
		retained += len(encoded)
		if retained > maxEnrichmentBytes {
			return nil, errors.New("mapped lookup pages exceed 1048576 bytes")
		}
		pages = append(pages, mapped)
		if done {
			return pages, nil
		}
	}
	return nil, errors.New("pagination reached maxPages before source was exhausted")
}

type enrichmentWriter struct{ bytes.Buffer }

func (w *enrichmentWriter) Write(p []byte) (int, error) {
	if w.Len()+len(p) > maxEnrichmentBytes {
		return 0, errors.New("enrichment template exceeds 1048576 bytes")
	}
	return w.Buffer.Write(p)
}

func renderEnrichmentJSON(t *compiledTemplate, data *actionData) (any, error) {
	value, err := renderEnrichmentTemplate(t, data)
	if err != nil {
		return nil, err
	}
	return decodeEnrichmentJSON(value)
}

func renderEnrichmentTemplate(t *compiledTemplate, data *actionData) (string, error) {
	var buffer enrichmentWriter
	if err := t.value.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func decodeEnrichmentJSON(body string) (any, error) {
	if len(body) > maxEnrichmentBytes || !json.Valid([]byte(body)) {
		return nil, errors.New("expected valid JSON within 1048576 bytes")
	}
	return decodeJSON([]byte(body))
}

func (q *compiledQuery) performLookupRequest(req *http.Request) (*http.Response, error) {
	if req.URL.User != nil || !q.origins[req.URL.Scheme+"://"+req.URL.Host] {
		return nil, errors.New("lookup destination is outside allowedOrigins")
	}
	if req.ContentLength > maxEnrichmentBytes {
		return nil, errors.New("lookup request exceeds 1048576 bytes")
	}
	select {
	case q.slots <- struct{}{}:
		defer func() { <-q.slots }()
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}
	var body []byte
	if req.GetBody != nil {
		reader, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		body, err = io.ReadAll(io.LimitReader(reader, maxEnrichmentBytes+1))
		reader.Close()
		if err != nil {
			return nil, err
		}
		if len(body) > maxEnrichmentBytes {
			return nil, errors.New("lookup request is too large")
		}
	}
	encoded, err := json.Marshal(struct {
		Method string      `json:"method"`
		URL    string      `json:"url"`
		Header http.Header `json:"header"`
		Body   []byte      `json:"body"`
	}{req.Method, req.URL.String(), req.Header, body})
	if err != nil {
		return nil, err
	}
	key := sha256.Sum256(encoded)
	q.mu.Lock()
	entry, exists := q.cache[key]
	q.mu.Unlock()
	if exists && time.Now().Before(entry.expires) {
		return cachedQueryResponse(entry, req), nil
	}
	response, err := q.target.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err = io.ReadAll(io.LimitReader(response.Body, q.target.maxResponseSize+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > q.target.maxResponseSize {
		return nil, errors.New("lookup response exceeds maxResponseBytes")
	}
	// Consumers use only the JSON body/status. Do not retain upstream cookies or
	// unbounded headers in the cache, nor forward them to the business response.
	entry = queryCacheEntry{body: body, status: response.StatusCode, expires: time.Now().Add(q.ttl)}
	valid := response.StatusCode >= 200 && response.StatusCode < 300 && json.Valid(body)
	if validator := q.target.response; validator != nil {
		valid = valid && (len(validator.statuses) == 0 || validator.statuses.Contains(response.StatusCode)) && validator.validate(body) == nil
	}
	if valid && q.ttl > 0 && int64(len(body)) <= q.maxBytes {
		q.mu.Lock()
		for existingKey, cached := range q.cache {
			if !time.Now().Before(cached.expires) || existingKey == key {
				delete(q.cache, existingKey)
				q.bytes -= int64(len(cached.body))
			}
		}
		for existingKey, cached := range q.cache {
			if len(q.cache) < q.maxEntries && q.bytes+int64(len(body)) <= q.maxBytes {
				break
			}
			delete(q.cache, existingKey)
			q.bytes -= int64(len(cached.body))
		}
		q.cache[key] = entry
		q.bytes += int64(len(body))
		q.mu.Unlock()
	}
	return cachedQueryResponse(entry, req), nil
}

func cachedQueryResponse(entry queryCacheEntry, req *http.Request) *http.Response {
	return &http.Response{StatusCode: entry.status, Header: entry.headers.Clone(), Body: io.NopCloser(bytes.NewReader(entry.body)), ContentLength: int64(len(entry.body)), Request: req}
}
