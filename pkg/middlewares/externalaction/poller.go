package externalaction

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

const (
	minimumPollInterval = time.Second
	maximumPollInterval = 24 * time.Hour
	maximumPollTimeout  = 5 * time.Minute
	maximumPollPages    = 1000
	maximumPollPageSize = 10000
)

type compiledPoller struct {
	config          dynamic.ExternalActionPoller
	client          *http.Client
	url             *compiledTemplate
	headers         map[string]*compiledTemplate
	body            *compiledTemplate
	auth            *compiledAuth
	response        *compiledHTTPResponse
	coordination    deduplicator
	coordinationKey *compiledTemplate
}

type pollerRegistryKey struct {
	done   <-chan struct{}
	config [32]byte
}

var pollerRegistry = struct {
	sync.Mutex

	values map[pollerRegistryKey]struct{}
}{
	values: make(map[pollerRegistryKey]struct{}),
}

func newCompiledPoller(ctx context.Context, config dynamic.ExternalActionPoller, credentials map[string]*compiledAuth) (*compiledPoller, error) {
	if err := validatePoller(config); err != nil {
		return nil, err
	}

	urlTemplate, err := newTemplate("poller-url", config.Request.URL)
	if err != nil {
		return nil, fmt.Errorf("compiling URL: %w", err)
	}
	headers, err := compileTemplates("poller-header", config.Request.Headers)
	if err != nil {
		return nil, err
	}
	body, err := newTemplate("poller-body", config.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("compiling body: %w", err)
	}

	var auth *compiledAuth
	if config.Request.Auth != nil && strings.TrimSpace(config.Request.Auth.CredentialSet) != "" {
		if hasInlineAuth(*config.Request.Auth) {
			return nil, errors.New("request credentialSet cannot be combined with inline authentication fields")
		}
		auth = credentials[config.Request.Auth.CredentialSet]
		if auth == nil {
			return nil, fmt.Errorf("request credential set %q is not configured", config.Request.Auth.CredentialSet)
		}
	} else {
		auth, err = compileAuth(config.Request.Auth)
		if err != nil {
			return nil, fmt.Errorf("compiling request authentication: %w", err)
		}
	}

	transport, err := buildTransport(config.Request.TLS)
	if err != nil {
		return nil, fmt.Errorf("building request transport: %w", err)
	}
	response, err := compileHTTPResponse(config.Request.ResponseValidation)
	if err != nil {
		return nil, fmt.Errorf("compiling response validation: %w", err)
	}
	coordination, err := newDeduplicator(ctx, *config.Coordination)
	if err != nil {
		return nil, fmt.Errorf("creating coordination store: %w", err)
	}
	coordinationKey, err := newTemplate("poller-coordination-key", config.Coordination.KeyTemplate)
	if err != nil {
		return nil, fmt.Errorf("compiling coordination key: %w", err)
	}

	return &compiledPoller{
		config: config,
		client: &http.Client{
			Transport: transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		url:             urlTemplate,
		headers:         headers,
		body:            body,
		auth:            auth,
		response:        response,
		coordination:    coordination,
		coordinationKey: coordinationKey,
	}, nil
}

func validatePoller(config dynamic.ExternalActionPoller) error {
	interval := time.Duration(config.Interval)
	if interval < minimumPollInterval || interval > maximumPollInterval {
		return fmt.Errorf("interval must be between %s and %s", minimumPollInterval, maximumPollInterval)
	}
	if delay := time.Duration(config.InitialDelay); delay < 0 || delay > maximumPollInterval {
		return fmt.Errorf("initialDelay must be between 0 and %s", maximumPollInterval)
	}
	if timeout := time.Duration(config.Timeout); timeout < 0 || timeout > maximumPollTimeout {
		return fmt.Errorf("timeout must be between 0 and %s", maximumPollTimeout)
	}
	if strings.TrimSpace(config.Request.URL) == "" {
		return errors.New("request URL is required")
	}
	if config.Request.MaxResponseBytes < 0 || config.Request.MaxResponseBytes > maxResponseBytes {
		return fmt.Errorf("request maxResponseBytes must not exceed %d", maxResponseBytes)
	}
	if strings.TrimSpace(config.Observation.Path) == "" || config.Observation.Path[0] != '/' {
		return errors.New("observation path must be absolute")
	}
	if config.Observation.Event != nil {
		if strings.TrimSpace(config.Observation.Event.Source) == "" {
			return errors.New("observation event source is required")
		}
		if strings.TrimSpace(config.Observation.Event.Principal) == "" {
			return errors.New("observation event principal is required")
		}
		if path := strings.TrimSpace(config.Observation.Event.Path); path != "" && path[0] != '/' {
			return errors.New("observation event path must be absolute")
		}
	}
	if config.Coordination == nil {
		return errors.New("coordination is required to prevent duplicate polling across replicas")
	}
	if strings.TrimSpace(config.Coordination.KeyTemplate) == "" {
		return errors.New("coordination keyTemplate is required")
	}
	if (config.Coordination.Memory == nil) == (config.Coordination.MySQL == nil) {
		return errors.New("coordination must define exactly one memory or mysql backend")
	}
	ttl := time.Duration(config.Coordination.TTL)
	lease := time.Duration(config.Coordination.Lease)
	if ttl <= 0 || lease <= 0 || ttl < lease {
		return errors.New("coordination ttl and lease must be positive and ttl must be greater than or equal to lease")
	}
	if config.Pagination != nil {
		pagination := config.Pagination
		if strings.TrimSpace(pagination.PageParameter) == "" || strings.TrimSpace(pagination.PageSizeParameter) == "" {
			return errors.New("pagination pageParameter and pageSizeParameter are required")
		}
		if pagination.PageSize <= 0 || pagination.PageSize > maximumPollPageSize {
			return fmt.Errorf("pagination pageSize must be between 1 and %d", maximumPollPageSize)
		}
		if pagination.MaxPages <= 0 || pagination.MaxPages > maximumPollPages {
			return fmt.Errorf("pagination maxPages must be between 1 and %d", maximumPollPages)
		}
		if strings.TrimSpace(pagination.ItemsJSONPath) == "" {
			return errors.New("pagination itemsJSONPath is required")
		}
	}
	return nil
}

func startPollers(ctx context.Context, middleware *middleware, configHash [32]byte) {
	if len(middleware.config.pollers) == 0 {
		return
	}
	key := pollerRegistryKey{done: ctx.Done(), config: configHash}

	pollerRegistry.Lock()
	if _, exists := pollerRegistry.values[key]; exists {
		pollerRegistry.Unlock()
		return
	}
	pollerRegistry.values[key] = struct{}{}
	pollerRegistry.Unlock()

	for _, poller := range middleware.config.pollers {
		go poller.run(ctx, middleware, middleware.logger)
	}
	if ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			pollerRegistry.Lock()
			delete(pollerRegistry.values, key)
			pollerRegistry.Unlock()
		}()
	}
}

func (p *compiledPoller) run(ctx context.Context, middleware *middleware, logger zerolog.Logger) {
	if delay := time.Duration(p.config.InitialDelay); delay > 0 {
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		}
	}

	p.runAndLog(ctx, middleware, logger)
	ticker := time.NewTicker(time.Duration(p.config.Interval))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.runAndLog(ctx, middleware, logger)
		case <-ctx.Done():
			return
		}
	}
}

func (p *compiledPoller) runAndLog(ctx context.Context, middleware *middleware, logger zerolog.Logger) {
	if err := p.runCoordinated(ctx, middleware); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error().Err(err).Str("poller", p.config.Name).Msg("External action poller failed")
	}
}

func (p *compiledPoller) runCoordinated(ctx context.Context, middleware *middleware) error {
	data, err := p.observationData(ctx)
	if err != nil {
		return err
	}
	key, err := p.coordinationKey.render(data)
	if err != nil {
		return fmt.Errorf("rendering coordination key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("rendered coordination key is empty")
	}
	ttl := time.Duration(p.config.Coordination.TTL)
	lease := time.Duration(p.config.Coordination.Lease)
	claimID, claimed, err := p.coordination.Claim(ctx, "poller:"+p.config.Name, key, ttl, lease)
	if err != nil || !claimed {
		return err
	}

	if err := p.poll(ctx, middleware, data); err != nil {
		releaseContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		releaseErr := p.coordination.Release(releaseContext, key, claimID)
		cancel()
		if releaseErr != nil {
			return errors.Join(err, fmt.Errorf("releasing coordination key: %w", releaseErr))
		}
		return err
	}
	completeContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	err = p.coordination.Complete(completeContext, key, claimID, ttl)
	cancel()
	if err != nil {
		return fmt.Errorf("completing coordination key: %w", err)
	}
	return nil
}

func (p *compiledPoller) observationData(ctx context.Context) (*actionData, error) {
	req, err := p.observationRequest(ctx)
	if err != nil {
		return nil, err
	}
	return newActionData(req, nil, 0, nil, time.Time{}, time.Time{}), nil
}

func (p *compiledPoller) observationRequest(ctx context.Context) (*http.Request, error) {
	method := strings.ToUpper(strings.TrimSpace(p.config.Observation.Method))
	if method == "" {
		method = http.MethodGet
	}
	scheme := strings.TrimSpace(p.config.Observation.Scheme)
	if scheme == "" {
		scheme = schemeHTTP
	}
	query := make(url.Values, len(p.config.Observation.Query))
	for name, value := range p.config.Observation.Query {
		query.Set(name, value)
	}
	requestURL := &url.URL{
		Scheme:   scheme,
		Host:     p.config.Observation.Host,
		Path:     p.config.Observation.Path,
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Host = p.config.Observation.Host
	for name, value := range p.config.Observation.Headers {
		req.Header.Set(name, value)
	}
	return req, nil
}

func (p *compiledPoller) poll(ctx context.Context, middleware *middleware, renderData *actionData) error {
	if p.config.Pagination == nil {
		result, err := p.pollPage(ctx, middleware, renderData, 0, 0)
		return errors.Join(result.observationErr, err)
	}

	pagination := p.config.Pagination
	page := pagination.StartPage
	if page <= 0 {
		page = 1
	}
	var failures observationFailures
	for range pagination.MaxPages {
		if err := ctx.Err(); err != nil {
			return errors.Join(failures.err(), err)
		}
		result, err := p.pollPage(ctx, middleware, renderData, page, pagination.PageSize)
		failures.add(result.observationErr)
		if err != nil {
			return errors.Join(failures.err(), err)
		}
		if result.done {
			return failures.err()
		}
		page++
	}
	return errors.Join(failures.err(), fmt.Errorf("pagination reached maxPages=%d before the source was exhausted", pagination.MaxPages))
}

type pollPageResult struct {
	done           bool
	observationErr error
}

func (p *compiledPoller) pollPage(ctx context.Context, middleware *middleware, renderData *actionData, page, pageSize int) (pollPageResult, error) {
	authenticated, bound, _, warning, err := p.auth.bind(renderData)
	if warning != nil {
		// The last valid reloadable material remains in use; the warning is
		// surfaced without including the credential value.
		middleware.logger.Warn().Err(warning).Str("poller", p.config.Name).Msg("External action poller credential reload warning")
	}
	if err != nil {
		return pollPageResult{}, fmt.Errorf("binding request authentication: %w", err)
	}
	targetURL, err := p.url.render(authenticated)
	if err != nil {
		return pollPageResult{}, bound.redactError(fmt.Errorf("rendering URL: %w", err))
	}
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return pollPageResult{}, bound.redactError(fmt.Errorf("parsing URL: %w", err))
	}
	validScheme := parsedURL.Scheme == schemeHTTP || parsedURL.Scheme == "https"
	if !validScheme || parsedURL.Host == "" {
		return pollPageResult{}, errors.New("rendered poller URL must use http or https with a host")
	}
	if p.config.Pagination != nil {
		query := parsedURL.Query()
		query.Set(p.config.Pagination.PageParameter, strconv.Itoa(page))
		query.Set(p.config.Pagination.PageSizeParameter, strconv.Itoa(pageSize))
		parsedURL.RawQuery = query.Encode()
	}
	body, err := p.body.render(authenticated)
	if err != nil {
		return pollPageResult{}, bound.redactError(fmt.Errorf("rendering body: %w", err))
	}
	timeout := time.Duration(p.config.Timeout)
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	method := strings.ToUpper(strings.TrimSpace(p.config.Request.Method))
	if method == "" {
		method = http.MethodGet
	}
	request, err := http.NewRequestWithContext(requestContext, method, parsedURL.String(), strings.NewReader(body))
	if err != nil {
		return pollPageResult{}, bound.redactError(err)
	}
	for name, value := range p.headers {
		rendered, err := value.render(authenticated)
		if err != nil {
			return pollPageResult{}, bound.redactError(fmt.Errorf("rendering header %q: %w", name, err))
		}
		request.Header.Set(name, rendered)
	}
	if p.config.Request.ContentType != "" {
		request.Header.Set("Content-Type", p.config.Request.ContentType)
	}
	if bound != nil {
		if err := bound.apply(request); err != nil {
			return pollPageResult{}, bound.redactError(err)
		}
	}
	started := time.Now().UTC()
	response, err := p.client.Do(request)
	if err != nil {
		return pollPageResult{}, bound.redactError(err)
	}
	defer response.Body.Close()
	limit := p.config.Request.MaxResponseBytes
	if limit <= 0 {
		limit = 4 * 1024 * 1024
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return pollPageResult{}, bound.redactError(err)
	}
	if int64(len(responseBody)) > limit {
		return pollPageResult{}, fmt.Errorf("response exceeded maxResponseBytes=%d", limit)
	}
	completed := time.Now().UTC()
	if p.response != nil && len(p.response.statuses) > 0 && !p.response.statuses.Contains(response.StatusCode) {
		message := p.response.errorMessage(responseBody)
		if message != "" {
			return pollPageResult{}, bound.redactError(fmt.Errorf("source returned status %d: %s", response.StatusCode, message))
		}
		return pollPageResult{}, fmt.Errorf("source returned status %d", response.StatusCode)
	}
	if (p.response == nil || len(p.response.statuses) == 0) && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return pollPageResult{}, fmt.Errorf("source returned status %d", response.StatusCode)
	}
	if p.response != nil {
		if err := p.response.validate(responseBody); err != nil {
			return pollPageResult{}, bound.redactError(err)
		}
	}

	result := pollPageResult{}
	if p.config.Observation.Event == nil {
		observation, err := p.observationRequest(ctx)
		if err != nil {
			return pollPageResult{}, err
		}
		result.observationErr = middleware.observeResponse(observation, nil, response.StatusCode, response.Header.Clone(), responseBody, started, completed)
	} else {
		eventRequest, payload, err := p.responseEventRequest(ctx, response.StatusCode, responseBody, completed)
		if err != nil {
			return pollPageResult{}, err
		}
		result.observationErr = middleware.observeRequest(eventRequest, payload)
	}

	if p.config.Pagination == nil {
		result.done = true
		return result, ctx.Err()
	}
	payload, err := decodeJSON(responseBody)
	if err != nil {
		return result, fmt.Errorf("paginated response is not valid JSON: %w", err)
	}
	items, ok := selectJSON(payload, p.config.Pagination.ItemsJSONPath).([]any)
	if !ok {
		return result, fmt.Errorf("pagination itemsJSONPath %q did not select an array", p.config.Pagination.ItemsJSONPath)
	}
	if len(items) == 0 || len(items) < pageSize {
		result.done = true
		return result, ctx.Err()
	}
	if path := strings.TrimSpace(p.config.Pagination.TotalJSONPath); path != "" {
		total, ok := integerValue(selectJSON(payload, path))
		if !ok {
			return result, fmt.Errorf("pagination totalJSONPath %q did not select an integer", path)
		}
		startPage := p.config.Pagination.StartPage
		if startPage <= 0 {
			startPage = 1
		}
		observed := (page - startPage + 1) * pageSize
		if observed >= total {
			result.done = true
			return result, ctx.Err()
		}
	}
	return result, ctx.Err()
}

func (p *compiledPoller) responseEventRequest(ctx context.Context, status int, responseBody []byte, observedAt time.Time) (*http.Request, []byte, error) {
	var responsePayload any
	decoder := json.NewDecoder(strings.NewReader(string(responseBody)))
	decoder.UseNumber()
	if err := decoder.Decode(&responsePayload); err != nil {
		return nil, nil, fmt.Errorf("event response body is not valid JSON: %w", err)
	}
	original, err := p.observationRequest(ctx)
	if err != nil {
		return nil, nil, err
	}
	eventConfig := p.config.Observation.Event
	eventType := strings.TrimSpace(eventConfig.Type)
	if eventType == "" {
		eventType = "http.response.observed"
	}
	identity := sha256.Sum256([]byte(strings.Join([]string{
		p.config.Name,
		original.Method,
		original.URL.Path,
		original.URL.RawQuery,
		string(responseBody),
	}, "\x00")))
	payload, err := json.Marshal(map[string]any{
		"specversion": "1.0",
		"id":          hex.EncodeToString(identity[:]),
		"source":      eventConfig.Source,
		"type":        eventType,
		"subject":     original.URL.Path,
		"time":        observedAt,
		"data": map[string]any{
			"principal": eventConfig.Principal,
			"request": map[string]any{
				"method":    original.Method,
				"path":      original.URL.Path,
				"raw_query": original.URL.RawQuery,
				"host":      original.Host,
			},
			"response":      map[string]any{"status": status},
			"response_body": responsePayload,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	eventPath := strings.TrimSpace(eventConfig.Path)
	if eventPath == "" {
		eventPath = "/_governance/v1/events"
	}
	eventHost := strings.TrimSpace(eventConfig.Host)
	if eventHost == "" {
		eventHost = original.Host
	}
	requestURL := &url.URL{Scheme: schemeHTTP, Host: eventHost, Path: eventPath}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), strings.NewReader(string(payload)))
	if err != nil {
		return nil, nil, err
	}
	request.Host = eventHost
	request.Header.Set("Content-Type", "application/cloudevents+json")
	request.Header.Set("X-Governance-Principal", eventConfig.Principal)
	return request, payload, nil
}

func integerValue(value any) (int, bool) {
	switch typed := value.(type) {
	case json.Number:
		parsed, err := strconv.ParseInt(typed.String(), 10, 64)
		return int(parsed), err == nil && parsed >= 0
	case float64:
		return int(typed), typed >= 0 && typed == float64(int(typed))
	case int:
		return typed, typed >= 0
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil && parsed >= 0
	default:
		return 0, false
	}
}
