package externalaction

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	phaseAfter    = "after"
	failureIgnore = "ignore"
	schemeHTTP    = "http"
)

const typeName = "ExternalAction"

var errQueueFull = errors.New("external action queue is full")

const (
	maxQueueCapacity = 65536
	maxQueueWorkers  = 128
	maxBodyBytes     = 16 * 1024 * 1024
	maxResponseBytes = 16 * 1024 * 1024
	maxEnqueueWait   = time.Second
)

type compiledRule struct {
	config                dynamic.ExternalActionRule
	path                  *regexp.Regexp
	headers               map[string]*regexp.Regexp
	query                 map[string]*regexp.Regexp
	bodyJSONPaths         map[string]*regexp.Regexp
	responseBodyJSONPaths map[string]*regexp.Regexp
	forEachJSONPaths      map[string]*regexp.Regexp
	methods               map[string]struct{}
	statuses              types.HTTPCodeRanges
	target                target
	deduplication         deduplicator
	deduplicationKey      *compiledTemplate
	binding               *compiledBinding
}

type compiledConfig struct {
	config      dynamic.ExternalAction
	credentials map[string]*compiledAuth
	pollers     []*compiledPoller
	rules       []*compiledRule
}

type middleware struct {
	next       http.Handler
	name       string
	config     *compiledConfig
	dispatcher *dispatcher
	logger     zerolog.Logger
}

// New creates a general external action middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ExternalAction, name string) (http.Handler, error) {
	return NewWithEnrichment(ctx, next, config, name, nil)
}

// NewWithEnrichment resolves event enrichment independently of the HTTP chain so pollers use it too.
func NewWithEnrichment(ctx context.Context, next http.Handler, config dynamic.ExternalAction, name string, definitions map[string]dynamic.QueryEnrichment) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if len(config.Enrichment) > 8 {
		return nil, errors.New("at most 8 query enrichment references may be configured")
	}
	var enrichments []*compiledEnrichment
	for _, reference := range config.Enrichment {
		definition, ok := definitions[reference]
		if !ok {
			return nil, fmt.Errorf("query enrichment %q is not configured", reference)
		}
		enrichment, err := getEnrichment(ctx, reference, definition)
		if err != nil {
			return nil, fmt.Errorf("query enrichment %q: %w", reference, err)
		}
		enrichments = append(enrichments, enrichment)
	}
	encoded, err := json.Marshal(definitions)
	if err != nil {
		return nil, err
	}
	config.ConfigurationID = name
	compiled, err := compileConfigWithContext(ctx, config)
	if err != nil {
		return nil, err
	}
	for _, rule := range compiled.rules {
		switch destination := rule.target.(type) {
		case *httpTarget:
			destination.enrichments = enrichments
		case *execTarget:
			destination.enrichments = enrichments
		}
	}
	raw, err := json.Marshal(compiled.config)
	if err != nil {
		return nil, fmt.Errorf("encoding external action configuration: %w", err)
	}
	configHash := sha256.Sum256(append(raw, encoded...))
	dispatch := getDispatcher(ctx, compiled.config, configHash, *logger)
	value := &middleware{
		next:       next,
		name:       name,
		config:     compiled,
		dispatcher: dispatch,
		logger:     *logger,
	}
	startPollers(ctx, value, configHash)
	return value, nil
}

func (m *middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var candidates []*compiledRule
	var maxBody int64
	var maxResponseBody int64
	for _, rule := range m.config.rules {
		if rule.matchesRequest(req) {
			candidates = append(candidates, rule)
			needsRequestBody := !strings.EqualFold(rule.config.Body.ReadMode, "none") ||
				len(rule.bodyJSONPaths) > 0 ||
				rule.config.ForEach != nil && strings.EqualFold(rule.config.ForEach.Source, "requestBody")
			if needsRequestBody {
				maxBody = max(maxBody, bodyLimit(rule.config.Body.MaxBytes))
			}
			if phase(rule.config.Phase) == phaseAfter && strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") {
				maxResponseBody = max(maxResponseBody, responseBodyLimit(rule.config.ResponseBody.MaxBytes))
			}
		}
	}
	if len(candidates) == 0 {
		m.next.ServeHTTP(rw, req)
		return
	}

	rawBody, overflow, err := readAndRestoreBody(req, maxBody)
	if err != nil {
		http.Error(rw, "could not read request body", http.StatusBadRequest)
		return
	}

	matched := make([]*compiledRule, 0, len(candidates))
	for _, rule := range candidates {
		// An overflowing candidate remains selected so its configured
		// skip/reject policy is applied below. A partial body must never be
		// treated as a successful JSON-path match.
		if overflowFor(rule, rawBody, overflow) || rule.matchesBody(req.Header.Get("Content-Type"), rawBody) {
			matched = append(matched, rule)
		}
	}
	if len(matched) == 0 {
		m.next.ServeHTTP(rw, req)
		return
	}

	started := time.Now().UTC()
	beforeData := newActionData(req, rawBody, 0, nil, started, time.Time{})
	for _, rule := range matched {
		if phase(rule.config.Phase) != "before" {
			continue
		}
		if overflowFor(rule, rawBody, overflow) {
			if strings.EqualFold(rule.config.Body.Overflow, "reject") {
				http.Error(rw, "request body is too large for external action", http.StatusRequestEntityTooLarge)
				return
			}
			m.logger.Warn().Str("rule", rule.config.Name).Msg("External action skipped because the request body exceeded its configured limit")
			if m.config.selectsFirst() {
				break
			}
			continue
		}
		ruleData := beforeData.withFormats(rule.config.Body.Format, rule.config.ResponseBody.Format)
		dispatched := false
		for _, data := range rule.expandPrepared(ruleData) {
			bindingMatches, err := rule.binding.matches(data)
			if err != nil {
				m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action credential identity match failed")
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					http.Error(rw, "external action failed", http.StatusBadGateway)
					return
				}
				continue
			}
			if !bindingMatches {
				continue
			}
			dispatched = true
			if err := m.dispatch(rule, data); err != nil {
				if !errors.Is(err, errQueueFull) {
					m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action failed")
				}
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					http.Error(rw, "external action failed", http.StatusBadGateway)
					return
				}
			}
		}
		if m.config.selectsFirst() && dispatched {
			break
		}
	}

	recorder := newResponseRecorder(rw, maxResponseBody)
	m.next.ServeHTTP(recorder, req)
	completed := time.Now().UTC()
	status := recorder.Status()
	if status == 0 {
		status = http.StatusOK
	}

	capturedResponseBody := recorder.Body()
	responseCaptureOverflow := recorder.Overflow()
	if !responseCaptureOverflow {
		decoded, overflow, err := decodeCapturedResponseBody(recorder.Header().Get("Content-Encoding"), capturedResponseBody, maxResponseBody)
		if err != nil {
			m.logger.Warn().Err(err).Str("contentEncoding", recorder.Header().Get("Content-Encoding")).Msg("External action skipped because the captured response body could not be decoded")
			responseCaptureOverflow = true
			capturedResponseBody = nil
		} else {
			capturedResponseBody = decoded
			responseCaptureOverflow = overflow
		}
	}

	// After-actions must not replace the already-written business response.
	_ = m.dispatchAfter(matched, req, rawBody, overflow, status, recorder.Header(), capturedResponseBody, recorder.Size(), responseCaptureOverflow, started, completed)
}

func (m *middleware) dispatchAfter(matched []*compiledRule, req *http.Request, rawBody []byte, requestBodyOverflow bool, status int, responseHeaders http.Header, capturedResponseBody []byte, responseSize int64, responseCaptureOverflow bool, started, completed time.Time) error {
	var failures observationFailures
	afterData := newActionData(req, rawBody, status, responseHeaders, started, completed)
	afterData.Response.Bytes = responseSize
	afterData.Response.RawBody = string(capturedResponseBody)
	afterData.Response.Overflow = responseCaptureOverflow
	for _, rule := range matched {
		if phase(rule.config.Phase) != phaseAfter || !rule.matchesStatus(status) {
			continue
		}
		if overflowFor(rule, rawBody, requestBodyOverflow) {
			m.logger.Warn().Str("rule", rule.config.Name).Msg("External action skipped because the request body exceeded its configured limit")
			if m.config.selectsFirst() {
				break
			}
			continue
		}
		if strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") &&
			(afterData.Response.Overflow || responseOverflowFor(rule, int64(len(capturedResponseBody)))) {
			m.logger.Warn().Str("rule", rule.config.Name).Msg("External action skipped because the response body exceeded its configured limit")
			if m.config.selectsFirst() {
				break
			}
			continue
		}
		data := afterData.withFormats(rule.config.Body.Format, rule.config.ResponseBody.Format)
		if !rule.matchesResponseBody(data.Response.Body) {
			continue
		}
		dispatched := false
		for _, expanded := range rule.expandPrepared(data) {
			if err := req.Context().Err(); err != nil {
				return errors.Join(failures.err(), err)
			}
			bindingMatches, err := rule.binding.matches(expanded)
			if err != nil {
				m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action credential identity match failed")
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					failures.add(fmt.Errorf("rule %q credential identity match: %w", rule.config.Name, err))
				}
				continue
			}
			if !bindingMatches {
				continue
			}
			dispatched = true
			if err := m.dispatchPrepared(rule, expanded); err != nil {
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					failures.add(fmt.Errorf("rule %q: %w", rule.config.Name, err))
				}
				if !errors.Is(err, errQueueFull) {
					m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action failed")
				}
			}
		}
		if m.config.selectsFirst() && dispatched {
			break
		}
	}
	return failures.err()
}

func (m *middleware) observeResponse(req *http.Request, rawBody []byte, status int, responseHeaders http.Header, responseBody []byte, started, completed time.Time) error {
	matched := make([]*compiledRule, 0, len(m.config.rules))
	for _, rule := range m.config.rules {
		if phase(rule.config.Phase) != phaseAfter || !rule.matchesRequest(req) || !rule.matchesBody(req.Header.Get("Content-Type"), rawBody) || !rule.matchesStatus(status) {
			continue
		}
		matched = append(matched, rule)
	}
	if len(matched) == 0 {
		return nil
	}
	return m.dispatchAfter(matched, req, rawBody, false, status, responseHeaders, responseBody, int64(len(responseBody)), false, started, completed)
}

func (m *middleware) observeRequest(req *http.Request, rawBody []byte) error {
	var failures observationFailures
	data := newActionData(req, rawBody, 0, nil, time.Now().UTC(), time.Time{})
	for _, rule := range m.config.rules {
		if phase(rule.config.Phase) != "before" || !rule.matchesRequest(req) || !rule.matchesBody(req.Header.Get("Content-Type"), rawBody) {
			continue
		}
		ruleData := data.withFormats(rule.config.Body.Format, rule.config.ResponseBody.Format)
		dispatched := false
		for _, expanded := range rule.expandPrepared(ruleData) {
			if err := req.Context().Err(); err != nil {
				return errors.Join(failures.err(), err)
			}
			bindingMatches, err := rule.binding.matches(expanded)
			if err != nil {
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					failures.add(fmt.Errorf("rule %q credential identity match: %w", rule.config.Name, err))
				}
				m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action credential identity match failed")
				continue
			}
			if !bindingMatches {
				continue
			}
			dispatched = true
			if err := m.dispatchPrepared(rule, expanded); err != nil {
				if failurePolicy(rule.config.FailurePolicy) != failureIgnore {
					failures.add(fmt.Errorf("rule %q: %w", rule.config.Name, err))
				}
				if !errors.Is(err, errQueueFull) {
					m.logger.Error().Err(err).Str("rule", rule.config.Name).Msg("External action failed")
				}
			}
		}
		if m.config.selectsFirst() && dispatched {
			break
		}
	}
	return failures.err()
}

// Keep only the first failure so a large unsuccessful scan cannot accumulate errors.
type observationFailures struct {
	count int
	first error
}

func (f *observationFailures) add(err error) {
	if err == nil {
		return
	}
	f.count++
	if f.first == nil {
		f.first = err
	}
}

func (f *observationFailures) err() error {
	if f.first == nil {
		return nil
	}
	return fmt.Errorf("%d observation failures (first): %w", f.count, f.first)
}

func (c *compiledConfig) selectsFirst() bool {
	return strings.EqualFold(c.config.SelectionMode, "first")
}

func (m *middleware) dispatch(rule *compiledRule, data *actionData) error {
	data = data.withFormats(rule.config.Body.Format, rule.config.ResponseBody.Format)
	return m.dispatchPrepared(rule, data)
}

func (m *middleware) dispatchPrepared(rule *compiledRule, data *actionData) error {
	prepared := *data
	if strings.EqualFold(rule.config.Body.ReadMode, "none") {
		prepared.Body = nil
		prepared.RawBody = ""
	}
	if !strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") {
		prepared.Response.Body = nil
		prepared.Response.RawBody = ""
	}
	if mode(rule.config.Mode) == "async" {
		pinned, err := rule.pinCredential(&prepared)
		if err != nil {
			return err
		}
		task := actionTask{rule: rule, data: pinned.clone()}
		if !m.dispatcher.enqueue(task) {
			m.dispatcher.logDrop(rule.config.Name)
			return errQueueFull
		}
		return nil
	}

	timeout := time.Duration(rule.config.Timeout)
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(prepared.Context, timeout)
	defer cancel()
	return runRule(ctx, rule, &prepared)
}

func (r *compiledRule) pinCredential(data *actionData) (*actionData, error) {
	if r.binding != nil {
		pinned, err := r.binding.pin(data)
		if err != nil {
			return nil, err
		}
		if pinned != data {
			return pinned, nil
		}
	}
	httpTarget, ok := r.target.(*httpTarget)
	if !ok {
		return data, nil
	}
	return httpTarget.pinCredential(data)
}

func (r *compiledRule) matchesRequest(req *http.Request) bool {
	if len(r.methods) > 0 {
		if _, ok := r.methods[req.Method]; !ok {
			return false
		}
	}
	if len(r.config.Match.Hosts) > 0 {
		found := false
		for _, host := range r.config.Match.Hosts {
			if strings.EqualFold(host, req.Host) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if r.path != nil && !r.path.MatchString(req.URL.Path) {
		return false
	}
	for name, expression := range r.headers {
		if !expression.MatchString(req.Header.Get(name)) {
			return false
		}
	}
	for name, expression := range r.query {
		if !expression.MatchString(req.URL.Query().Get(name)) {
			return false
		}
	}
	return true
}

func (r *compiledRule) matchesStatus(status int) bool {
	return len(r.statuses) == 0 || r.statuses.Contains(status)
}

func (r *compiledRule) matchesBody(_ string, raw []byte) bool {
	if len(r.bodyJSONPaths) == 0 {
		return true
	}

	body, err := decodeJSON(raw)
	if err != nil {
		return false
	}
	for path, expression := range r.bodyJSONPaths {
		if !expression.MatchString(matchValue(selectJSON(body, path))) {
			return false
		}
	}
	return true
}

func (r *compiledRule) matchesResponseBody(body any) bool {
	return matchesJSONPaths(body, r.responseBodyJSONPaths)
}

func matchesJSONPaths(body any, expressions map[string]*regexp.Regexp) bool {
	if len(expressions) == 0 {
		return true
	}
	if body == nil {
		return false
	}
	for path, expression := range expressions {
		if !expression.MatchString(matchValue(selectJSON(body, path))) {
			return false
		}
	}
	return true
}

func (r *compiledRule) expandPrepared(data *actionData) []*actionData {
	if r.config.ForEach == nil {
		return []*actionData{data}
	}

	var source any
	switch {
	case strings.EqualFold(r.config.ForEach.Source, "requestBody"):
		source = data.Body
	case strings.EqualFold(r.config.ForEach.Source, "responseBody"):
		source = data.Response.Body
	}
	items, ok := selectJSON(source, r.config.ForEach.JSONPath).([]any)
	if !ok {
		return nil
	}

	result := make([]*actionData, 0, len(items))
	for index, item := range items {
		if !matchesJSONPaths(item, r.forEachJSONPaths) {
			continue
		}
		value := *data
		value.Item = item
		value.ItemIndex = index
		result = append(result, &value)
	}
	return result
}

func matchValue(value any) string {
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
