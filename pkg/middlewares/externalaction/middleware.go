package externalaction

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
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
	if strings.EqualFold(rule.config.Body.ReadMode, "none") {
		value := *data
		value.Body = nil
		value.RawBody = ""
		data = &value
	}
	if !strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") {
		value := *data
		value.Response.Body = nil
		value.Response.RawBody = ""
		data = &value
	}
	if mode(rule.config.Mode) == "async" {
		var err error
		data, err = rule.pinCredential(data)
		if err != nil {
			return err
		}
		task := actionTask{rule: rule, data: data.clone()}
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
	ctx, cancel := context.WithTimeout(data.Context, timeout)
	defer cancel()
	return runRule(ctx, rule, data)
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

func compileConfig(config dynamic.ExternalAction) (*compiledConfig, error) {
	return compileConfigWithContext(context.Background(), config)
}

func compileConfigWithContext(ctx context.Context, config dynamic.ExternalAction) (*compiledConfig, error) {
	if err := resolveRuntimeConfig(&config); err != nil {
		return nil, fmt.Errorf("resolving runtime values: %w", err)
	}
	if len(config.Rules) == 0 {
		return nil, errors.New("external action requires at least one rule")
	}
	if !oneOf(config.SelectionMode, "", "all", "first") {
		return nil, fmt.Errorf("unsupported external action selection mode %q", config.SelectionMode)
	}
	if config.Queue.Capacity > maxQueueCapacity {
		return nil, fmt.Errorf("external action queue capacity must not exceed %d", maxQueueCapacity)
	}
	if config.Queue.Workers > maxQueueWorkers {
		return nil, fmt.Errorf("external action queue workers must not exceed %d", maxQueueWorkers)
	}
	if time.Duration(config.Queue.EnqueueWait) > maxEnqueueWait {
		return nil, fmt.Errorf("external action queue enqueueWait must not exceed %s", maxEnqueueWait)
	}

	credentials, err := compileCredentialSets(config.CredentialSets)
	if err != nil {
		return nil, err
	}
	compiled := &compiledConfig{config: config, credentials: credentials}
	for index, pollerConfig := range config.Pollers {
		if pollerConfig.Name == "" {
			pollerConfig.Name = fmt.Sprintf("poller-%d", index+1)
		}
		poller, err := newCompiledPoller(ctx, pollerConfig, credentials)
		if err != nil {
			return nil, fmt.Errorf("creating external action poller %q: %w", pollerConfig.Name, err)
		}
		compiled.pollers = append(compiled.pollers, poller)
	}
	for index, ruleConfig := range config.Rules {
		if ruleConfig.Name == "" {
			ruleConfig.Name = fmt.Sprintf("rule-%d", index+1)
		}
		if err := validateRule(ruleConfig); err != nil {
			return nil, fmt.Errorf("validating external action rule %q: %w", ruleConfig.Name, err)
		}
		if ruleConfig.HTTP == nil == (ruleConfig.Exec == nil) {
			return nil, fmt.Errorf("external action rule %q must define exactly one http or exec target", ruleConfig.Name)
		}

		rule := &compiledRule{
			config:                ruleConfig,
			headers:               make(map[string]*regexp.Regexp),
			query:                 make(map[string]*regexp.Regexp),
			bodyJSONPaths:         make(map[string]*regexp.Regexp),
			responseBodyJSONPaths: make(map[string]*regexp.Regexp),
			forEachJSONPaths:      make(map[string]*regexp.Regexp),
			methods:               make(map[string]struct{}),
		}
		for _, method := range ruleConfig.Match.Methods {
			rule.methods[strings.ToUpper(method)] = struct{}{}
		}
		if ruleConfig.Match.PathRegex != "" {
			expression, err := regexp.Compile(ruleConfig.Match.PathRegex)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q path: %w", ruleConfig.Name, err)
			}
			rule.path = expression
		}
		for name, value := range ruleConfig.Match.Headers {
			expression, err := regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q header %q: %w", ruleConfig.Name, name, err)
			}
			rule.headers[http.CanonicalHeaderKey(name)] = expression
		}
		for name, value := range ruleConfig.Match.Query {
			expression, err := regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q query %q: %w", ruleConfig.Name, name, err)
			}
			rule.query[name] = expression
		}
		for path, value := range ruleConfig.Match.BodyJSONPaths {
			expression, err := regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q body JSON path %q: %w", ruleConfig.Name, path, err)
			}
			rule.bodyJSONPaths[path] = expression
		}
		for path, value := range ruleConfig.Match.ResponseBodyJSONPaths {
			expression, err := regexp.Compile(value)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q response body JSON path %q: %w", ruleConfig.Name, path, err)
			}
			rule.responseBodyJSONPaths[path] = expression
		}
		if ruleConfig.ForEach != nil {
			for path, value := range ruleConfig.ForEach.MatchJSONPaths {
				expression, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("compiling external action rule %q forEach JSON path %q: %w", ruleConfig.Name, path, err)
				}
				rule.forEachJSONPaths[path] = expression
			}
		}
		if len(ruleConfig.Match.ResponseStatuses) > 0 {
			statuses, err := types.NewHTTPCodeRanges(ruleConfig.Match.ResponseStatuses)
			if err != nil {
				return nil, fmt.Errorf("parsing external action rule %q response statuses: %w", ruleConfig.Name, err)
			}
			rule.statuses = statuses
		}

		var err error
		switch {
		case ruleConfig.HTTP != nil:
			rule.target, err = newHTTPTarget(ruleConfig, credentials)
		case ruleConfig.Exec != nil:
			rule.target, err = newExecTarget(ruleConfig)
		}
		if err != nil {
			return nil, fmt.Errorf("creating external action rule %q: %w", ruleConfig.Name, err)
		}
		if ruleConfig.Deduplication != nil {
			rule.deduplicationKey, err = newTemplate("deduplication-key", ruleConfig.Deduplication.KeyTemplate)
			if err != nil {
				return nil, fmt.Errorf("compiling external action rule %q deduplication key: %w", ruleConfig.Name, err)
			}
			rule.deduplication, err = newDeduplicator(ctx, *ruleConfig.Deduplication)
			if err != nil {
				return nil, fmt.Errorf("creating external action rule %q deduplication: %w", ruleConfig.Name, err)
			}
		}
		if ruleConfig.Binding != nil {
			rule.binding, err = newCompiledBinding(ctx, *ruleConfig.Binding, credentials)
			if err != nil {
				return nil, fmt.Errorf("creating external action rule %q lifecycle binding: %w", ruleConfig.Name, err)
			}
		}
		compiled.rules = append(compiled.rules, rule)
	}

	return compiled, nil
}

func resolveRuntimeConfig(config *dynamic.ExternalAction) error {
	type connectionIdentity struct {
		deduplication string
		dedupFile     string
		binding       string
		bindingFile   string
	}
	identities := make([]connectionIdentity, len(config.Rules))
	pollerIdentities := make([]connectionIdentity, len(config.Pollers))
	credentialIdentities := make(map[string]string, len(config.CredentialSets))
	for name, credential := range config.CredentialSets {
		if config.ConfigurationID != "" {
			credentialIdentities[name] = runtimeconfig.StableKey(
				"external-action-credential-set",
				config.ConfigurationID,
				name,
			)
			continue
		}
		// Non-runtime compilation paths (mainly unit tests and file-provider
		// validation) have no Middleware instance name. Keep their identity
		// stable across value/revision rotations by using only source metadata.
		runtimeSources := make(map[string]dynamic.RuntimeValue, len(config.RuntimeValues))
		for alias, source := range config.RuntimeValues {
			runtimeSources[alias] = dynamic.RuntimeValue{Source: source.Source, File: source.File}
		}
		encoded, err := json.Marshal(struct {
			Name           string                              `json:"name"`
			Credential     dynamic.ExternalActionCredentialSet `json:"credential"`
			RuntimeSources map[string]dynamic.RuntimeValue     `json:"runtimeSources"`
		}{
			Name:           name,
			Credential:     credential,
			RuntimeSources: runtimeSources,
		})
		if err != nil {
			return fmt.Errorf("credential set %q runtime identity: %w", name, err)
		}
		credentialIdentities[name] = runtimeconfig.StableKey(string(encoded))
	}
	for index, rule := range config.Rules {
		if rule.Deduplication != nil && rule.Deduplication.MySQL != nil {
			identity, err := runtimeconfig.SourceIdentity(rule.Deduplication.MySQL.DSN, rule.Deduplication.MySQL.DSNFile, config.RuntimeValues)
			if err != nil {
				return fmt.Errorf("rule %q deduplication mysql: %w", rule.Name, err)
			}
			identities[index].deduplication = identity
			identities[index].dedupFile = runtimeconfig.ExactReferenceFile(rule.Deduplication.MySQL.DSN, config.RuntimeValues)
		}
		if rule.Binding != nil && rule.Binding.MySQL != nil {
			identity, err := runtimeconfig.SourceIdentity(rule.Binding.MySQL.DSN, rule.Binding.MySQL.DSNFile, config.RuntimeValues)
			if err != nil {
				return fmt.Errorf("rule %q binding mysql: %w", rule.Name, err)
			}
			identities[index].binding = identity
			identities[index].bindingFile = runtimeconfig.ExactReferenceFile(rule.Binding.MySQL.DSN, config.RuntimeValues)
		}
	}
	for index, poller := range config.Pollers {
		if poller.Coordination != nil && poller.Coordination.MySQL != nil {
			identity, err := runtimeconfig.SourceIdentity(poller.Coordination.MySQL.DSN, poller.Coordination.MySQL.DSNFile, config.RuntimeValues)
			if err != nil {
				return fmt.Errorf("poller %q coordination mysql: %w", poller.Name, err)
			}
			pollerIdentities[index].deduplication = identity
			pollerIdentities[index].dedupFile = runtimeconfig.ExactReferenceFile(poller.Coordination.MySQL.DSN, config.RuntimeValues)
		}
	}
	if err := runtimeconfig.Resolve(config.RuntimeValues, config); err != nil {
		return err
	}
	for name, credential := range config.CredentialSets {
		credential.CacheIdentity = credentialIdentities[name]
		config.CredentialSets[name] = credential
	}
	for index := range config.Rules {
		rule := &config.Rules[index]
		if rule.Deduplication != nil && rule.Deduplication.MySQL != nil {
			rule.Deduplication.MySQL.ConnectionID = identities[index].deduplication
			if identities[index].dedupFile != "" {
				rule.Deduplication.MySQL.DSN = ""
				rule.Deduplication.MySQL.DSNFile = identities[index].dedupFile
			}
		}
		if rule.Binding != nil && rule.Binding.MySQL != nil {
			rule.Binding.MySQL.ConnectionID = identities[index].binding
			if identities[index].bindingFile != "" {
				rule.Binding.MySQL.DSN = ""
				rule.Binding.MySQL.DSNFile = identities[index].bindingFile
			}
		}
	}
	for index := range config.Pollers {
		poller := &config.Pollers[index]
		if poller.Coordination != nil && poller.Coordination.MySQL != nil {
			poller.Coordination.MySQL.ConnectionID = pollerIdentities[index].deduplication
			if pollerIdentities[index].dedupFile != "" {
				poller.Coordination.MySQL.DSN = ""
				poller.Coordination.MySQL.DSNFile = pollerIdentities[index].dedupFile
			}
		}
	}
	return nil
}

func validateRule(rule dynamic.ExternalActionRule) error {
	if !oneOf(rule.Phase, "", "before", phaseAfter) {
		return fmt.Errorf("unsupported phase %q", rule.Phase)
	}
	if !oneOf(rule.Mode, "", "sync", "async") {
		return fmt.Errorf("unsupported mode %q", rule.Mode)
	}
	if !oneOf(rule.FailurePolicy, "", "fail", failureIgnore) {
		return fmt.Errorf("unsupported failure policy %q", rule.FailurePolicy)
	}
	if !oneOf(rule.Body.ReadMode, "", "eager", "none") {
		return fmt.Errorf("unsupported body read mode %q", rule.Body.ReadMode)
	}
	if !oneOf(rule.Body.Overflow, "", "skip", "reject") {
		return fmt.Errorf("unsupported body overflow policy %q", rule.Body.Overflow)
	}
	if rule.Body.MaxBytes > maxBodyBytes {
		return fmt.Errorf("body maxBytes must not exceed %d", maxBodyBytes)
	}
	if !oneOf(rule.ResponseBody.ReadMode, "", "none", "capture") {
		return fmt.Errorf("unsupported response body read mode %q", rule.ResponseBody.ReadMode)
	}
	if !oneOf(rule.ResponseBody.Overflow, "", "skip") {
		return fmt.Errorf("unsupported response body overflow policy %q", rule.ResponseBody.Overflow)
	}
	if rule.ResponseBody.MaxBytes > maxResponseBytes {
		return fmt.Errorf("response body maxBytes must not exceed %d", maxResponseBytes)
	}
	requiresResponseBody := len(rule.Match.ResponseBodyJSONPaths) > 0 ||
		rule.ForEach != nil && strings.EqualFold(rule.ForEach.Source, "responseBody")
	if requiresResponseBody && !strings.EqualFold(rule.ResponseBody.ReadMode, "capture") {
		return errors.New("responseBody readMode must be capture when response body matching or iteration is configured")
	}
	if len(rule.Match.ResponseBodyJSONPaths) > 0 && phase(rule.Phase) != phaseAfter {
		return errors.New("response body matching requires phase after")
	}
	if rule.ForEach != nil {
		if !oneOf(rule.ForEach.Source, "requestBody", "responseBody") {
			return fmt.Errorf("unsupported forEach source %q", rule.ForEach.Source)
		}
		if strings.TrimSpace(rule.ForEach.JSONPath) == "" {
			return errors.New("forEach jsonPath is required")
		}
		if strings.EqualFold(rule.ForEach.Source, "responseBody") && phase(rule.Phase) != phaseAfter {
			return errors.New("responseBody forEach requires phase after")
		}
		if strings.EqualFold(rule.ForEach.Source, "requestBody") && strings.EqualFold(rule.Body.ReadMode, "none") {
			return errors.New("requestBody forEach cannot use body readMode none")
		}
	}
	if rule.Deduplication != nil {
		if strings.TrimSpace(rule.Deduplication.KeyTemplate) == "" {
			return errors.New("deduplication keyTemplate is required")
		}
		if (rule.Deduplication.Memory == nil) == (rule.Deduplication.MySQL == nil) {
			return errors.New("deduplication must define exactly one memory or mysql backend")
		}
		if time.Duration(rule.Deduplication.TTL) < 0 || time.Duration(rule.Deduplication.Lease) < 0 {
			return errors.New("deduplication ttl and lease must not be negative")
		}
		ttl := time.Duration(rule.Deduplication.TTL)
		if ttl <= 0 {
			ttl = defaultDeduplicationTTL
		}
		lease := time.Duration(rule.Deduplication.Lease)
		if lease <= 0 {
			lease = defaultDeduplicationLease
		}
		if ttl < lease {
			return errors.New("deduplication ttl must be greater than or equal to lease")
		}
	}
	if rule.Binding != nil {
		if !oneOf(rule.Binding.Mode, "create", "resolve", "resolveOrCreate") {
			return fmt.Errorf("unsupported lifecycle binding mode %q", rule.Binding.Mode)
		}
		if !oneOf(rule.Binding.OnMissing, "", "fail", "skip") {
			return fmt.Errorf("unsupported lifecycle binding onMissing %q", rule.Binding.OnMissing)
		}
		if strings.TrimSpace(rule.Binding.KeyTemplate) == "" {
			return errors.New("lifecycle binding keyTemplate is required")
		}
		if (rule.Binding.Memory == nil) == (rule.Binding.MySQL == nil) {
			return errors.New("lifecycle binding must define exactly one memory or mysql backend")
		}
		if time.Duration(rule.Binding.TTL) < 0 {
			return errors.New("lifecycle binding ttl must not be negative")
		}
		if strings.EqualFold(rule.Binding.Mode, "create") {
			if rule.Binding.OnMissing != "" {
				return errors.New("create lifecycle binding cannot define onMissing")
			}
			hasCredentialSet := strings.TrimSpace(rule.Binding.CredentialSet) != ""
			hasAllowedCredentialSets := len(rule.Binding.AllowedCredentialSets) > 0
			if hasCredentialSet == hasAllowedCredentialSets {
				return errors.New("create lifecycle binding requires exactly one credentialSet or allowedCredentialSets")
			}
			if hasAllowedCredentialSets && len(rule.Binding.Match.Headers) == 0 && len(rule.Binding.Match.BodyJSONPaths) == 0 && len(rule.Binding.Match.ResponseBodyJSONPaths) == 0 && len(rule.Binding.Match.Templates) == 0 {
				return errors.New("create lifecycle binding with allowedCredentialSets requires credential identity match fields")
			}
		}
		if strings.EqualFold(rule.Binding.Mode, "resolve") {
			if rule.Binding.CredentialSet != "" {
				return errors.New("resolve lifecycle binding cannot define credentialSet")
			}
			if len(rule.Binding.AllowedCredentialSets) == 0 {
				return errors.New("resolve lifecycle binding requires allowedCredentialSets")
			}
			if len(rule.Binding.Match.Headers) > 0 || len(rule.Binding.Match.BodyJSONPaths) > 0 || len(rule.Binding.Match.ResponseBodyJSONPaths) > 0 || len(rule.Binding.Match.Templates) > 0 {
				return errors.New("resolve lifecycle binding cannot define credential identity match fields")
			}
		}
		if strings.EqualFold(rule.Binding.Mode, "resolveOrCreate") {
			if rule.Binding.OnMissing != "" {
				return errors.New("resolveOrCreate lifecycle binding cannot define onMissing")
			}
			if rule.Binding.CredentialSet != "" {
				return errors.New("resolveOrCreate lifecycle binding cannot define credentialSet")
			}
			if len(rule.Binding.AllowedCredentialSets) == 0 {
				return errors.New("resolveOrCreate lifecycle binding requires allowedCredentialSets")
			}
			if len(rule.Binding.Match.Headers) == 0 && len(rule.Binding.Match.BodyJSONPaths) == 0 && len(rule.Binding.Match.ResponseBodyJSONPaths) == 0 && len(rule.Binding.Match.Templates) == 0 {
				return errors.New("resolveOrCreate lifecycle binding requires credential identity match fields")
			}
		}
		if rule.Exec != nil {
			return errors.New("lifecycle binding is only supported for HTTP targets")
		}
	}
	if rule.HTTP != nil && strings.TrimSpace(rule.HTTP.URL) == "" {
		return errors.New("HTTP target URL is required")
	}
	if rule.HTTP != nil && rule.HTTP.MaxResponseBytes > maxResponseBytes {
		return fmt.Errorf("HTTP maxResponseBytes must not exceed %d", maxResponseBytes)
	}
	if rule.HTTP != nil && rule.HTTP.ResponseValidation != nil {
		response := rule.HTTP.ResponseValidation
		if response.SuccessJSONPath != "" && len(response.SuccessValues) == 0 {
			return errors.New("HTTP responseValidation successValues is required with successJSONPath")
		}
		if response.SuccessJSONPath == "" && len(response.SuccessValues) > 0 {
			return errors.New("HTTP responseValidation successJSONPath is required with successValues")
		}
		if response.SuccessJSONPath != "" && response.BodyRegex != "" {
			return errors.New("HTTP responseValidation successJSONPath and bodyRegex are mutually exclusive")
		}
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if strings.EqualFold(value, candidate) {
			return true
		}
	}
	return false
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

type actionData struct {
	Context              context.Context
	Request              actionRequest
	Body                 any
	RawBody              string
	Response             actionResponse
	Item                 any
	ItemIndex            int
	AuthParameters       map[string]string
	LifecycleBinding     actionLifecycleBinding
	Payload              any
	Input                any
	Result               any
	Enrichment           map[string]any
	SourceAuthParameters map[string]string
	LookupTime           time.Time
	OutboundBody         string
	lookupPageQuery      url.Values
	deferredEnrichments  []*compiledEnrichment

	credentialSet              string
	expectedCredentialIdentity [32]byte
}

// actionLifecycleBinding exposes reconciliation state to action templates.
// Created is true only when resolveOrCreate recovered a missing lifecycle
// binding during the current dispatch.
type actionLifecycleBinding struct {
	Created bool
}

type actionRequest struct {
	Method   string
	Scheme   string
	Host     string
	Path     string
	RawQuery string
	Headers  http.Header
	Query    url.Values
	Cookies  map[string]string
}

type actionResponse struct {
	Status      int
	Headers     http.Header
	Bytes       int64
	Body        any
	RawBody     string
	Overflow    bool
	StartedAt   time.Time
	CompletedAt time.Time
	DurationMS  int64
}

func newActionData(req *http.Request, raw []byte, status int, responseHeaders http.Header, started, completed time.Time) *actionData {
	cookies := make(map[string]string)
	for _, cookie := range req.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	data := &actionData{
		Context:             req.Context(),
		deferredEnrichments: enrichmentsFromContext(req.Context()),
		Request: actionRequest{
			Method:   req.Method,
			Scheme:   scheme(req),
			Host:     req.Host,
			Path:     req.URL.Path,
			RawQuery: req.URL.RawQuery,
			Headers:  req.Header.Clone(),
			Query:    req.URL.Query(),
			Cookies:  cookies,
		},
		RawBody: string(raw),
		Response: actionResponse{
			Status:      status,
			Headers:     responseHeaders.Clone(),
			StartedAt:   started,
			CompletedAt: completed,
		},
	}
	if !completed.IsZero() {
		data.Response.DurationMS = completed.Sub(started).Milliseconds()
	}
	data.Body = parseBody("", req.Header.Get("Content-Type"), raw)
	return data
}

func (d *actionData) clone() *actionData {
	result := *d
	result.Context = context.WithoutCancel(d.Context)
	result.Request.Headers = d.Request.Headers.Clone()
	result.Request.Query = cloneValues(d.Request.Query)
	result.Request.Cookies = cloneMap(d.Request.Cookies)
	result.Response.Headers = d.Response.Headers.Clone()
	result.AuthParameters = cloneMap(d.AuthParameters)
	return &result
}

func (d *actionData) withBodyFormat(format string) *actionData {
	result := *d
	result.Body = parseBody(format, d.Request.Headers.Get("Content-Type"), []byte(d.RawBody))
	return &result
}

func (d *actionData) withFormats(bodyFormat, responseFormat string) *actionData {
	result := d.withBodyFormat(bodyFormat)
	result.Response.Body = parseBody(responseFormat, result.Response.Headers.Get("Content-Type"), []byte(result.Response.RawBody))
	return result
}

func parseBody(format, contentType string, raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	format = strings.ToLower(format)
	switch {
	case format == "json" || format == "" && strings.Contains(contentType, "json"):
		if result, err := decodeJSON(raw); err == nil {
			return result
		}
	case format == "form" || format == "" && strings.Contains(contentType, "application/x-www-form-urlencoded"):
		if result, err := url.ParseQuery(string(raw)); err == nil {
			return result
		}
	}
	return string(raw)
}

func readAndRestoreBody(req *http.Request, limit int64) ([]byte, bool, error) {
	if req.Body == nil || limit <= 0 {
		return nil, false, nil
	}
	original := req.Body
	value, err := io.ReadAll(io.LimitReader(original, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(value)) > limit {
		req.Body = &multiReadCloser{
			Reader: io.MultiReader(bytes.NewReader(value), original),
			closer: original,
		}
		return value, true, nil
	}
	if err := original.Close(); err != nil {
		return nil, false, err
	}
	req.Body = io.NopCloser(bytes.NewReader(value))
	return value, false, nil
}

type multiReadCloser struct {
	io.Reader

	closer io.Closer
}

func (m *multiReadCloser) Close() error {
	return m.closer.Close()
}

func overflowFor(rule *compiledRule, raw []byte, globalOverflow bool) bool {
	if strings.EqualFold(rule.config.Body.ReadMode, "none") {
		return false
	}
	limit := bodyLimit(rule.config.Body.MaxBytes)
	return globalOverflow || int64(len(raw)) > limit
}

func bodyLimit(configured int64) int64 {
	if configured <= 0 {
		return 1024 * 1024
	}
	return configured
}

func responseBodyLimit(configured int64) int64 {
	if configured <= 0 {
		return 1024 * 1024
	}
	return configured
}

func responseOverflowFor(rule *compiledRule, responseSize int64) bool {
	if !strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") {
		return false
	}
	return responseSize > responseBodyLimit(rule.config.ResponseBody.MaxBytes)
}

// decodeCapturedResponseBody decodes only the middleware's private response
// copy. The original encoded bytes have already been written to the client by
// responseRecorder and are never changed. A bounded read prevents a small
// compressed response from expanding beyond the configured capture limit.
func decodeCapturedResponseBody(contentEncoding string, raw []byte, limit int64) ([]byte, bool, error) {
	if len(raw) == 0 || strings.TrimSpace(contentEncoding) == "" {
		return raw, false, nil
	}
	if limit <= 0 {
		return nil, true, nil
	}

	encodings := strings.Split(contentEncoding, ",")
	value := raw
	for _, encoding := range slices.Backward(encodings) {
		encoding := strings.ToLower(strings.TrimSpace(strings.SplitN(encoding, ";", 2)[0]))
		if encoding == "" || encoding == "identity" {
			continue
		}

		var reader io.Reader
		var closeReader func()
		switch encoding {
		case "gzip", "x-gzip":
			decoder, err := gzip.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing gzip response decoder: %w", err)
			}
			reader = decoder
			closeReader = func() { _ = decoder.Close() }
		case "deflate":
			decoder, err := zlib.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing deflate response decoder: %w", err)
			}
			reader = decoder
			closeReader = func() { _ = decoder.Close() }
		case "br":
			reader = brotli.NewReader(bytes.NewReader(value))
		case "zstd":
			decoder, err := zstd.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing zstd response decoder: %w", err)
			}
			reader = decoder
			closeReader = decoder.Close
		default:
			return nil, false, fmt.Errorf("unsupported response content encoding %q", encoding)
		}

		decoded, err := io.ReadAll(io.LimitReader(reader, limit+1))
		if closeReader != nil {
			closeReader()
		}
		if err != nil {
			return nil, false, fmt.Errorf("decoding %s response body: %w", encoding, err)
		}
		if int64(len(decoded)) > limit {
			return decoded[:limit], true, nil
		}
		value = decoded
	}
	return value, false, nil
}

func phase(value string) string {
	if strings.EqualFold(value, "before") {
		return "before"
	}
	return phaseAfter
}

func mode(value string) string {
	if strings.EqualFold(value, "sync") {
		return "sync"
	}
	return "async"
}

func failurePolicy(value string) string {
	if strings.EqualFold(value, failureIgnore) {
		return failureIgnore
	}
	return "fail"
}

func scheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	return schemeHTTP
}

type responseRecorder struct {
	http.ResponseWriter

	status    int
	size      int64
	body      bytes.Buffer
	bodyLimit int64
	overflow  bool
}

func newResponseRecorder(rw http.ResponseWriter, bodyLimit int64) *responseRecorder {
	return &responseRecorder{ResponseWriter: rw, bodyLimit: bodyLimit}
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(value []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(value)
	r.size += int64(n)
	if r.bodyLimit > 0 && n > 0 {
		remaining := r.bodyLimit - int64(r.body.Len())
		if remaining > 0 {
			captured := n
			if int64(captured) > remaining {
				captured = int(remaining)
			}
			_, _ = r.body.Write(value[:captured])
		}
		if int64(r.body.Len()) < r.size {
			r.overflow = true
		}
	}
	return n, err
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Size() int64 {
	return r.size
}

func (r *responseRecorder) Body() []byte {
	return r.body.Bytes()
}

func (r *responseRecorder) Overflow() bool {
	return r.overflow
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

type actionTask struct {
	rule *compiledRule
	data *actionData
	size int64
}

type dispatcherKey struct {
	done   <-chan struct{}
	config [32]byte
}

var dispatcherRegistry = struct {
	sync.Mutex

	values map[dispatcherKey]*dispatcher
}{
	values: make(map[dispatcherKey]*dispatcher),
}

type dispatcher struct {
	ctx            context.Context
	queue          chan actionTask
	done           chan struct{}
	wait           time.Duration
	logger         zerolog.Logger
	dropped        atomic.Uint64
	lastDropLog    atomic.Int64
	queuedBytes    atomic.Int64
	maxQueuedBytes int64
	wg             sync.WaitGroup
	stopMu         sync.RWMutex
	stopped        bool
}

func getDispatcher(ctx context.Context, config dynamic.ExternalAction, configHash [32]byte, logger zerolog.Logger) *dispatcher {
	key := dispatcherKey{done: ctx.Done(), config: configHash}

	dispatcherRegistry.Lock()
	defer dispatcherRegistry.Unlock()
	if existing := dispatcherRegistry.values[key]; existing != nil {
		return existing
	}

	capacity := config.Queue.Capacity
	if capacity <= 0 {
		capacity = 1024
	}
	workers := config.Queue.Workers
	if workers <= 0 {
		workers = 8
	}
	maxQueuedBytes := config.Queue.MaxQueuedBytes
	if maxQueuedBytes <= 0 {
		maxQueuedBytes = 64 * 1024 * 1024
	}
	value := &dispatcher{
		ctx:            ctx,
		queue:          make(chan actionTask, capacity),
		done:           make(chan struct{}),
		wait:           time.Duration(config.Queue.EnqueueWait),
		logger:         logger,
		maxQueuedBytes: maxQueuedBytes,
	}
	for range workers {
		value.wg.Add(1)
		go value.run()
	}
	go func() {
		<-ctx.Done()
		value.stopMu.Lock()
		value.stopped = true
		close(value.done)
		value.stopMu.Unlock()
		value.wg.Wait()
		dispatcherRegistry.Lock()
		delete(dispatcherRegistry.values, key)
		dispatcherRegistry.Unlock()
	}()
	dispatcherRegistry.values[key] = value
	return value
}

func (d *dispatcher) enqueue(value actionTask) bool {
	d.stopMu.RLock()
	defer d.stopMu.RUnlock()
	if d.stopped {
		return false
	}

	value.size = estimateTaskSize(value)
	if !d.reserve(value.size) {
		return false
	}
	accepted := false
	defer func() {
		if !accepted {
			d.queuedBytes.Add(-value.size)
		}
	}()

	if d.wait <= 0 {
		select {
		case d.queue <- value:
			accepted = true
			return true
		case <-d.done:
			return false
		default:
			return false
		}
	}
	timer := time.NewTimer(d.wait)
	defer timer.Stop()
	select {
	case d.queue <- value:
		accepted = true
		return true
	case <-d.done:
		return false
	case <-timer.C:
		return false
	}
}

func (d *dispatcher) reserve(size int64) bool {
	if size > d.maxQueuedBytes {
		return false
	}
	for {
		current := d.queuedBytes.Load()
		if current+size > d.maxQueuedBytes {
			return false
		}
		if d.queuedBytes.CompareAndSwap(current, current+size) {
			return true
		}
	}
}

func estimateTaskSize(task actionTask) int64 {
	if task.data == nil {
		return 1024
	}
	size := int64(3*len(task.data.RawBody) + 3*len(task.data.Response.RawBody) + len(task.data.Request.RawQuery) + len(task.data.Request.Host) + len(task.data.Request.Path) + 1024)
	for name, values := range task.data.Request.Headers {
		size += int64(len(name))
		for _, value := range values {
			size += int64(len(value))
		}
	}
	return max(size, 1024)
}

func (d *dispatcher) logDrop(rule string) {
	total := d.dropped.Add(1)
	now := time.Now().UnixNano()
	last := d.lastDropLog.Load()
	if last != 0 && time.Duration(now-last) < 10*time.Second {
		return
	}
	if !d.lastDropLog.CompareAndSwap(last, now) {
		return
	}
	d.logger.Warn().
		Str("rule", rule).
		Uint64("dropped", total).
		Msg("External actions dropped because the bounded queue is full")
}

func (d *dispatcher) run() {
	defer d.wg.Done()
	run := func(value actionTask) {
		timeout := time.Duration(value.rule.config.Timeout)
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		// A replaced Middleware cancels d.ctx before its accepted queue has
		// drained. Preserve request-scoped values but bound every final action.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(d.ctx), timeout)
		err := runRule(ctx, value.rule, value.data)
		cancel()
		d.queuedBytes.Add(-value.size)
		if err != nil {
			d.logger.Error().Err(err).Str("rule", value.rule.config.Name).Msg("Asynchronous external action failed")
		}
	}

	for {
		select {
		case value := <-d.queue:
			run(value)
		case <-d.done:
			for {
				select {
				case value := <-d.queue:
					run(value)
				default:
					return
				}
			}
		}
	}
}

func cloneValues(source url.Values) url.Values {
	result := make(url.Values, len(source))
	for key, value := range source {
		result[key] = append([]string(nil), value...)
	}
	return result
}

func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	maps.Copy(result, source)
	return result
}
