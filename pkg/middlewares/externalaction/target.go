package externalaction

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
	"github.com/traefik/traefik/v3/pkg/types"
)

type target interface {
	Run(ctx context.Context, data *actionData) error
}

type compiledTemplate struct {
	value     *template.Template
	literal   string
	isLiteral bool
}

func newTemplate(name, value string) (*compiledTemplate, error) {
	functions := sprig.TxtFuncMap()
	delete(functions, "env")
	delete(functions, "expandenv")
	functions["jsonPath"] = selectJSON
	functions["header"] = func(data *actionData, name string) string {
		return data.Request.Headers.Get(name)
	}
	functions["query"] = func(data *actionData, name string) string {
		return data.Request.Query.Get(name)
	}
	functions["cookie"] = func(data *actionData, name string) string {
		return data.Request.Cookies[name]
	}
	// authParameter exposes the already configured per-target authentication
	// parameters to URL/Header/Body templates. This keeps project metadata in
	// the same rule-local source of truth as the outbound authentication
	// headers instead of hard-coding it in a business payload template.
	functions["authParameter"] = func(data *actionData, name string) string {
		return data.AuthParameters[name]
	}
	functions["parseInt"] = func(value any) (int64, error) {
		return strconv.ParseInt(strings.TrimSpace(fmt.Sprint(value)), 10, 64)
	}
	functions["scalarString"] = func(value any) (string, error) {
		switch typed := value.(type) {
		case string:
			return typed, nil
		case json.Number:
			return typed.String(), nil
		case float64:
			return strconv.FormatFloat(typed, 'f', -1, 64), nil
		case float32:
			return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
		case int:
			return strconv.Itoa(typed), nil
		case int8, int16, int32, int64:
			return fmt.Sprintf("%d", typed), nil
		case uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", typed), nil
		default:
			return "", fmt.Errorf("value of type %T is not a scalar string or number", value)
		}
	}
	functions["required"] = func(message string, value any) (any, error) {
		if isEmpty(value) {
			return nil, errors.New(message)
		}
		return value, nil
	}

	parsed, err := template.New(name).Option("missingkey=error").Funcs(functions).Parse(value)
	if err != nil {
		return nil, err
	}
	return &compiledTemplate{value: parsed}, nil
}

func (t *compiledTemplate) render(data *actionData) (string, error) {
	if t == nil {
		return "", nil
	}
	if t.isLiteral {
		return t.literal, nil
	}
	var buffer bytes.Buffer
	if err := t.value.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func newLiteralTemplate(value string) *compiledTemplate {
	return &compiledTemplate{literal: value, isLiteral: true}
}

type httpTarget struct {
	client          *http.Client
	url             *compiledTemplate
	method          string
	headers         map[string]*compiledTemplate
	auth            *compiledAuth
	credentialSet   string
	credentialSets  map[string]*compiledAuth
	body            *compiledTemplate
	response        *compiledHTTPResponse
	contentType     string
	maxResponseSize int64
}

type compiledHTTPResponse struct {
	statuses        types.HTTPCodeRanges
	successJSONPath string
	successValues   map[string]struct{}
	bodyRegex       *regexp.Regexp
	messageJSONPath string
}

type compiledAuth struct {
	kind                string
	username            string
	password            string
	token               string
	header              string
	scheme              string
	parameters          map[string]*compiledTemplate
	passwordFile        string
	tokenFile           string
	parameterFiles      map[string]string
	immutableParameters []string
	reloadInterval      time.Duration

	mu             sync.Mutex
	material       authMaterial
	materialLoaded bool
	nextReload     time.Time
	lastWarning    [32]byte
	lastWarningAt  time.Time
}

type authMaterial struct {
	password   string
	token      string
	parameters map[string]string
	identity   [32]byte
}

// compiledCredentialSetsCache preserves the last valid credential material
// across Kubernetes dynamic-configuration rebuilds. A projected Secret update
// can trigger a provider refresh even though the Middleware CR itself did not
// change. Without this process-wide cache, a rebuilt Middleware could accept a
// new immutable project identity before the lifecycle store rejects it, and it
// would no longer have the old token needed to close already-bound work.
//
// The key includes the semantic name and the complete credential-set
// configuration (but never the file contents). Different paths or auth
// settings therefore do not share material. Entries intentionally live for the
// process lifetime so queued tasks and lifecycle completions retain their last
// valid credentials while an old set is being drained.
var compiledCredentialSetsCache sync.Map

type boundAuth struct {
	kind       string
	username   string
	password   string
	token      string
	header     string
	scheme     string
	parameters map[string]string
	secrets    []string
}

func newHTTPTarget(rule dynamic.ExternalActionRule, configuredCredentialSets ...map[string]*compiledAuth) (target, error) {
	credentialSets := make(map[string]*compiledAuth)
	if len(configuredCredentialSets) > 0 {
		credentialSets = configuredCredentialSets[0]
	}
	urlTemplate, err := newTemplate("url", rule.HTTP.URL)
	if err != nil {
		return nil, fmt.Errorf("compiling URL template: %w", err)
	}
	headers, err := compileTemplates("header", rule.HTTP.Headers)
	if err != nil {
		return nil, err
	}
	body, err := newTemplate("body", rule.Body.Template)
	if err != nil {
		return nil, fmt.Errorf("compiling body template: %w", err)
	}
	var credentialSet string
	authConfig := rule.HTTP.Auth
	if authConfig != nil {
		credentialSet = strings.TrimSpace(authConfig.CredentialSet)
		if credentialSet != "" {
			if _, ok := credentialSets[credentialSet]; !ok {
				return nil, fmt.Errorf("credential set %q is not configured", credentialSet)
			}
			if hasInlineAuth(*authConfig) {
				return nil, errors.New("credentialSet cannot be combined with inline authentication fields")
			}
			authConfig = nil
		}
	}
	auth, err := compileAuth(authConfig)
	if err != nil {
		return nil, err
	}
	transport, err := buildTransport(rule.HTTP.TLS)
	if err != nil {
		return nil, err
	}
	response, err := compileHTTPResponse(rule.HTTP.ResponseValidation)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(rule.HTTP.Method)
	if method == "" {
		method = http.MethodPost
	}
	maxResponse := rule.HTTP.MaxResponseBytes
	if maxResponse <= 0 {
		maxResponse = 64 * 1024
	}
	return &httpTarget{
		client: &http.Client{
			Transport: transport,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		url:             urlTemplate,
		method:          method,
		headers:         headers,
		auth:            auth,
		credentialSet:   credentialSet,
		credentialSets:  credentialSets,
		body:            body,
		response:        response,
		contentType:     rule.Body.ContentType,
		maxResponseSize: maxResponse,
	}, nil
}

func (h *httpTarget) Run(ctx context.Context, data *actionData) error {
	auth := h.auth
	credentialSet := data.credentialSet
	if credentialSet == "" {
		credentialSet = h.credentialSet
	}
	if credentialSet != "" {
		var ok bool
		auth, ok = h.credentialSets[credentialSet]
		if !ok {
			return fmt.Errorf("credential set %q is not configured", credentialSet)
		}
	}

	renderData, bound, identity, warning, err := auth.bind(data)
	if err != nil {
		return err
	}
	if warning != nil {
		log.Ctx(ctx).Warn().Err(warning).Str("credentialSet", credentialSet).Msg("Using the last valid external action credentials")
	}
	if data.expectedCredentialIdentity != ([32]byte{}) && identity != data.expectedCredentialIdentity {
		return fmt.Errorf("credential set %q immutable identity does not match the lifecycle binding", credentialSet)
	}
	targetURL, err := h.url.render(renderData)
	if err != nil {
		return fmt.Errorf("rendering URL: %w", err)
	}
	body, err := h.body.render(renderData)
	if err != nil {
		return fmt.Errorf("rendering body: %w", err)
	}
	if strings.Contains(h.contentType, "json") && !json.Valid([]byte(body)) {
		return errors.New("rendered external action body is not valid JSON")
	}
	request, err := http.NewRequestWithContext(ctx, h.method, targetURL, strings.NewReader(body))
	if err != nil {
		return err
	}
	for name, value := range h.headers {
		rendered, err := value.render(renderData)
		if err != nil {
			return fmt.Errorf("rendering header %q: %w", name, err)
		}
		request.Header.Set(name, rendered)
	}
	if h.contentType != "" {
		request.Header.Set("Content-Type", h.contentType)
	}
	if err := bound.apply(request); err != nil {
		return err
	}

	response, err := h.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, h.maxResponseSize+1))
	if err != nil {
		return err
	}
	if int64(len(responseBody)) > h.maxResponseSize {
		return errors.New("external action response exceeded maxResponseBytes")
	}
	if h.response != nil && len(h.response.statuses) > 0 && !h.response.statuses.Contains(response.StatusCode) {
		message := h.response.errorMessage(responseBody)
		if message != "" {
			return bound.redactError(fmt.Errorf("external action returned status %d: %s", response.StatusCode, message))
		}
		return bound.redactError(fmt.Errorf("external action returned status %d", response.StatusCode))
	}
	if (h.response == nil || len(h.response.statuses) == 0) && (response.StatusCode < 200 || response.StatusCode >= 300) {
		return bound.redactError(fmt.Errorf("external action returned status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody))))
	}
	if h.response != nil {
		if err := h.response.validate(responseBody); err != nil {
			return bound.redactError(err)
		}
	}
	return nil
}

func (h *httpTarget) pinCredential(data *actionData) (*actionData, error) {
	credentialSet := data.credentialSet
	if credentialSet == "" {
		credentialSet = h.credentialSet
	}
	if credentialSet == "" {
		return data, nil
	}
	auth := h.credentialSets[credentialSet]
	if auth == nil {
		return nil, fmt.Errorf("credential set %q is not configured", credentialSet)
	}
	identity, err := auth.identity()
	if err != nil {
		return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, err)
	}
	result := *data
	result.credentialSet = credentialSet
	result.expectedCredentialIdentity = identity
	return &result, nil
}

func compileHTTPResponse(config *dynamic.ExternalActionHTTPResponse) (*compiledHTTPResponse, error) {
	if config == nil {
		return nil, nil
	}
	compiled := &compiledHTTPResponse{
		successJSONPath: strings.TrimSpace(config.SuccessJSONPath),
		successValues:   make(map[string]struct{}, len(config.SuccessValues)),
		messageJSONPath: strings.TrimSpace(config.MessageJSONPath),
	}
	if len(config.Statuses) > 0 {
		statuses, err := types.NewHTTPCodeRanges(config.Statuses)
		if err != nil {
			return nil, fmt.Errorf("parsing HTTP response validation statuses: %w", err)
		}
		compiled.statuses = statuses
	}
	for _, value := range config.SuccessValues {
		compiled.successValues[value] = struct{}{}
	}
	if config.BodyRegex != "" {
		expression, err := regexp.Compile(config.BodyRegex)
		if err != nil {
			return nil, fmt.Errorf("compiling HTTP response validation bodyRegex: %w", err)
		}
		compiled.bodyRegex = expression
	}
	return compiled, nil
}

func (r *compiledHTTPResponse) validate(body []byte) error {
	if r.bodyRegex != nil && !r.bodyRegex.Match(body) {
		message := r.errorMessage(body)
		if message != "" {
			return fmt.Errorf("external action response body did not match configured success pattern: %s", message)
		}
		return errors.New("external action response body did not match configured success pattern")
	}
	if r.successJSONPath == "" {
		return nil
	}

	payload, err := decodeJSON(body)
	if err != nil {
		return fmt.Errorf("external action response is not valid JSON: %w", err)
	}
	actualValue := selectJSON(payload, r.successJSONPath)
	actual := responseValueString(actualValue)
	if _, ok := r.successValues[actual]; ok {
		return nil
	}

	message := responseMessage(payload, r.messageJSONPath)
	if message != "" {
		return fmt.Errorf("external action response validation failed at %s (value %q): %s", r.successJSONPath, actual, message)
	}
	return fmt.Errorf("external action response validation failed at %s (value %q)", r.successJSONPath, actual)
}

func (r *compiledHTTPResponse) errorMessage(body []byte) string {
	if r.messageJSONPath != "" {
		payload, err := decodeJSON(body)
		if err == nil {
			if message := responseMessage(payload, r.messageJSONPath); message != "" {
				return message
			}
		}
	}
	return sanitizeResponseMessage(string(body))
}

func decodeJSON(body []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return payload, nil
}

func responseMessage(payload any, path string) string {
	if path == "" {
		return ""
	}
	value := selectJSON(payload, path)
	if value == nil {
		return ""
	}
	return sanitizeResponseMessage(responseValueString(value))
}

func responseValueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return "<missing>"
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}

func sanitizeResponseMessage(value string) string {
	value = strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f {
			return ' '
		}
		return character
	}, strings.TrimSpace(value))
	const maxMessageRunes = 1024
	runes := []rune(value)
	if len(runes) > maxMessageRunes {
		return string(runes[:maxMessageRunes]) + "..."
	}
	return value
}

func compileAuth(config *dynamic.ExternalActionAuth) (*compiledAuth, error) {
	if config == nil {
		return nil, nil
	}
	value := &compiledAuth{
		kind:           strings.ToLower(config.Type),
		username:       config.Username,
		password:       config.Password,
		token:          config.Token,
		header:         config.Header,
		scheme:         config.Scheme,
		parameters:     make(map[string]*compiledTemplate),
		passwordFile:   config.PasswordFile,
		tokenFile:      config.TokenFile,
		parameterFiles: cloneMap(config.ParameterFiles),
	}
	if !oneOf(value.kind, "", "none", "basic", "bearer", "header", "headers") {
		return nil, fmt.Errorf("unsupported auth type %q", config.Type)
	}
	for name, parameter := range config.Parameters {
		parsed, err := newTemplate("auth-"+name, parameter)
		if err != nil {
			return nil, fmt.Errorf("compiling auth parameter %q: %w", name, err)
		}
		value.parameters[name] = parsed
	}
	if _, warning, err := value.refresh(true); err != nil {
		return nil, err
	} else if warning != nil {
		return nil, warning
	}
	return value, nil
}

func compileCredentialSets(config map[string]dynamic.ExternalActionCredentialSet) (map[string]*compiledAuth, error) {
	result := make(map[string]*compiledAuth, len(config))
	for name, credentialConfig := range config {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("credential set name cannot be empty")
		}
		if credentialConfig.Auth.CredentialSet != "" {
			return nil, fmt.Errorf("credential set %q cannot reference another credential set", name)
		}
		cacheKey, err := credentialSetCacheKey(name, credentialConfig)
		if err != nil {
			return nil, fmt.Errorf("fingerprinting credential set %q: %w", name, err)
		}
		auth, err := compileReloadableAuth(credentialConfig)
		if err != nil {
			return nil, fmt.Errorf("compiling credential set %q: %w", name, err)
		}
		actual, loaded := compiledCredentialSetsCache.LoadOrStore(cacheKey, auth)
		cached := actual.(*compiledAuth)
		if loaded {
			if warning, err := cached.update(auth); err != nil {
				return nil, fmt.Errorf("updating credential set %q: %w", name, err)
			} else if warning != nil {
				// Immutable rotations are deliberately non-fatal: the last
				// valid snapshot remains able to close already-bound work.
				log.Warn().Err(warning).Str("credentialSet", name).Msg("Credential set runtime update rejected")
			}
		}
		result[name] = cached
	}
	return result, nil
}

func credentialSetCacheKey(name string, config dynamic.ExternalActionCredentialSet) ([32]byte, error) {
	if config.CacheIdentity != "" {
		return sha256.Sum256([]byte(name + "\x00" + config.CacheIdentity)), nil
	}
	encoded, err := json.Marshal(struct {
		Name   string                              `json:"name"`
		Config dynamic.ExternalActionCredentialSet `json:"config"`
	}{
		Name:   name,
		Config: config,
	})
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(encoded), nil
}

func (a *compiledAuth) update(candidate *compiledAuth) (error, error) {
	material, err := candidate.loadMaterial()
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.materialLoaded && a.material.identity != ([32]byte{}) && material.identity != a.material.identity {
		return errors.New("reload rejected because immutable credential parameters changed; create a new credential set name"), nil
	}

	a.kind = candidate.kind
	a.username = candidate.username
	a.password = candidate.password
	a.token = candidate.token
	a.header = candidate.header
	a.scheme = candidate.scheme
	a.parameters = candidate.parameters
	a.passwordFile = candidate.passwordFile
	a.tokenFile = candidate.tokenFile
	a.parameterFiles = candidate.parameterFiles
	a.immutableParameters = candidate.immutableParameters
	a.reloadInterval = candidate.reloadInterval
	a.material = material
	a.materialLoaded = true
	a.nextReload = time.Now().Add(a.reloadInterval)
	return nil, nil
}

func compileReloadableAuth(config dynamic.ExternalActionCredentialSet) (*compiledAuth, error) {
	authConfig := config.Auth
	value := &compiledAuth{
		kind:                strings.ToLower(authConfig.Type),
		username:            authConfig.Username,
		password:            authConfig.Password,
		token:               authConfig.Token,
		header:              authConfig.Header,
		scheme:              authConfig.Scheme,
		parameters:          make(map[string]*compiledTemplate),
		passwordFile:        authConfig.PasswordFile,
		tokenFile:           authConfig.TokenFile,
		parameterFiles:      cloneMap(authConfig.ParameterFiles),
		immutableParameters: append([]string(nil), config.ImmutableParameters...),
		reloadInterval:      time.Duration(config.ReloadInterval),
	}
	if !oneOf(value.kind, "", "none", "basic", "bearer", "header", "headers") {
		return nil, fmt.Errorf("unsupported auth type %q", authConfig.Type)
	}
	if value.reloadInterval < 0 {
		return nil, errors.New("reloadInterval must not be negative")
	}
	if value.reloadInterval == 0 {
		value.reloadInterval = 5 * time.Second
	}
	slices.Sort(value.immutableParameters)
	value.immutableParameters = slices.Compact(value.immutableParameters)
	for name, parameter := range authConfig.Parameters {
		if !strings.Contains(parameter, "{{") {
			value.parameters[name] = newLiteralTemplate(parameter)
			continue
		}
		parsed, err := newTemplate("auth-"+name, parameter)
		if err != nil {
			return nil, fmt.Errorf("compiling auth parameter %q: %w", name, err)
		}
		value.parameters[name] = parsed
	}
	for _, name := range value.immutableParameters {
		parameter, hasStatic := authConfig.Parameters[name]
		_, hasFile := authConfig.ParameterFiles[name]
		if !hasStatic && !hasFile {
			return nil, fmt.Errorf("immutable parameter %q is not configured", name)
		}
		if hasStatic && !hasFile && strings.Contains(parameter, "{{") {
			return nil, fmt.Errorf("immutable parameter %q must be a literal or file value", name)
		}
	}
	return value, nil
}

func (a *compiledAuth) identity() ([32]byte, error) {
	if a == nil {
		return [32]byte{}, errors.New("credential set has no authentication configuration")
	}
	material, _, err := a.refresh(false)
	if err != nil {
		return [32]byte{}, err
	}
	return material.identity, nil
}

func (a *compiledAuth) immutableValues() (map[string]string, error) {
	if a == nil {
		return nil, errors.New("credential set has no authentication configuration")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	material, _, err := a.refreshLocked(false)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(a.immutableParameters))
	for _, name := range a.immutableParameters {
		if value, ok := material.parameters[name]; ok {
			result[name] = value
			continue
		}
		templateValue := a.parameters[name]
		if templateValue == nil || !templateValue.isLiteral {
			return nil, fmt.Errorf("immutable parameter %q is not a literal", name)
		}
		result[name] = templateValue.literal
	}
	return result, nil
}

func (a *compiledAuth) immutableParameterNames() []string {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.immutableParameters...)
}

func (a *compiledAuth) bind(data *actionData) (*actionData, *boundAuth, [32]byte, error, error) {
	if a == nil {
		return data, nil, [32]byte{}, nil, nil
	}
	a.mu.Lock()
	material, warning, err := a.refreshLocked(false)
	if err != nil {
		a.mu.Unlock()
		return nil, nil, [32]byte{}, warning, err
	}
	parameters := make(map[string]*compiledTemplate, len(a.parameters))
	maps.Copy(parameters, a.parameters)
	kind := a.kind
	username := a.username
	header := a.header
	scheme := a.scheme
	a.mu.Unlock()

	result := *data
	result.AuthParameters = make(map[string]string, len(parameters)+len(material.parameters))
	for name, value := range parameters {
		rendered, err := value.render(data)
		if err != nil {
			return nil, nil, [32]byte{}, warning, fmt.Errorf("rendering auth parameter %q: %w", name, err)
		}
		result.AuthParameters[name] = rendered
	}
	maps.Copy(result.AuthParameters, material.parameters)
	bound := &boundAuth{
		kind:       kind,
		username:   username,
		password:   material.password,
		token:      material.token,
		header:     header,
		scheme:     scheme,
		parameters: cloneMap(result.AuthParameters),
	}
	bound.addSecret(bound.password)
	bound.addSecret(bound.token)
	for _, value := range bound.parameters {
		bound.addSecret(value)
	}
	return &result, bound, material.identity, warning, nil
}

func (a *compiledAuth) refresh(force bool) (authMaterial, error, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.refreshLocked(force)
}

func (a *compiledAuth) refreshLocked(force bool) (authMaterial, error, error) {
	now := time.Now()
	if !force && a.materialLoaded && (a.reloadInterval <= 0 || now.Before(a.nextReload)) {
		return a.material, nil, nil
	}
	material, err := a.loadMaterial()
	if err != nil {
		if !a.materialLoaded {
			return authMaterial{}, nil, err
		}
		a.nextReload = now.Add(a.reloadInterval)
		return a.material, a.throttledWarning(err, now), nil
	}
	if a.materialLoaded && a.material.identity != ([32]byte{}) && material.identity != a.material.identity {
		err := errors.New("reload rejected because immutable credential parameters changed; create a new credential set name")
		a.nextReload = now.Add(a.reloadInterval)
		return a.material, a.throttledWarning(err, now), nil
	}
	a.material = material
	a.materialLoaded = true
	a.nextReload = now.Add(a.reloadInterval)
	return material, nil, nil
}

func (a *compiledAuth) throttledWarning(err error, now time.Time) error {
	fingerprint := sha256.Sum256([]byte(err.Error()))
	if fingerprint == a.lastWarning && now.Sub(a.lastWarningAt) < time.Minute {
		return nil
	}
	a.lastWarning = fingerprint
	a.lastWarningAt = now
	return err
}

func (a *compiledAuth) loadMaterial() (authMaterial, error) {
	files := make(map[string]string, len(a.parameterFiles)+2)
	if a.passwordFile != "" {
		files["password"] = a.passwordFile
	}
	if a.tokenFile != "" {
		files["token"] = a.tokenFile
	}
	for name, path := range a.parameterFiles {
		files["parameter:"+name] = path
	}
	loaded, err := readCredentialFiles(files)
	if err != nil {
		return authMaterial{}, err
	}
	material := authMaterial{
		password:   a.password,
		token:      a.token,
		parameters: make(map[string]string, len(a.parameterFiles)),
	}
	if value, ok := loaded["password"]; ok {
		material.password = value
	}
	if value, ok := loaded["token"]; ok {
		material.token = value
	}
	for name := range a.parameterFiles {
		material.parameters[name] = loaded["parameter:"+name]
	}

	if len(a.immutableParameters) > 0 {
		var identity strings.Builder
		for _, name := range a.immutableParameters {
			value, ok := material.parameters[name]
			if !ok {
				templateValue := a.parameters[name]
				if templateValue == nil || !templateValue.isLiteral {
					return authMaterial{}, fmt.Errorf("immutable parameter %q is not a literal", name)
				}
				value = templateValue.literal
			}
			identity.WriteString(name)
			identity.WriteByte(0)
			identity.WriteString(value)
			identity.WriteByte(0)
		}
		material.identity = sha256.Sum256([]byte(identity.String()))
	}
	return material, nil
}

func readCredentialFiles(files map[string]string) (map[string]string, error) {
	raw, err := runtimeconfig.ReadFiles(files)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(raw))
	for name, value := range raw {
		result[name] = strings.TrimSpace(string(value))
	}
	return result, nil
}

func (a *boundAuth) addSecret(value string) {
	if len(value) >= 4 {
		a.secrets = append(a.secrets, value)
	}
}

func (a *boundAuth) redactError(err error) error {
	if a == nil {
		return err
	}
	value := err.Error()
	for _, secret := range a.secrets {
		value = strings.ReplaceAll(value, secret, "<redacted>")
	}
	return errors.New(value)
}

func (a *boundAuth) apply(request *http.Request) error {
	if a == nil {
		return nil
	}
	switch a.kind {
	case "", "none":
		return nil
	case "basic":
		request.SetBasicAuth(a.username, a.password)
	case "bearer":
		scheme := a.scheme
		if scheme == "" {
			scheme = "Bearer"
		}
		request.Header.Set("Authorization", scheme+" "+a.token)
	case "header":
		name := a.header
		if name == "" {
			name = "Authorization"
		}
		request.Header.Set(name, a.token)
	case "headers":
		for name, value := range a.parameters {
			request.Header.Set(name, value)
		}
	default:
		return fmt.Errorf("unsupported auth type %q", a.kind)
	}
	return nil
}

func hasInlineAuth(config dynamic.ExternalActionAuth) bool {
	return config.Type != "" ||
		config.Username != "" ||
		config.Password != "" ||
		config.PasswordFile != "" ||
		config.Token != "" ||
		config.TokenFile != "" ||
		config.Header != "" ||
		config.Scheme != "" ||
		len(config.Parameters) > 0 ||
		len(config.ParameterFiles) > 0
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return reflected.Len() == 0
	case reflect.Bool:
		return !reflected.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflected.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflected.Float() == 0
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Pointer:
		return reflected.IsNil()
	default:
		return false
	}
}

func buildTransport(config *dynamic.ExternalActionTLS) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if config == nil {
		return transport, nil
	}
	tlsConfig := &tls.Config{
		ServerName:         config.ServerName,
		InsecureSkipVerify: config.InsecureSkipVerify,
	}
	if config.CAFile != "" {
		value, err := runtimeconfig.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA file: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if !pool.AppendCertsFromPEM(value) {
			return nil, errors.New("CA file does not contain a certificate")
		}
		tlsConfig.RootCAs = pool
	}
	if config.CertFile != "" || config.KeyFile != "" {
		if config.CertFile == "" || config.KeyFile == "" {
			return nil, errors.New("both certFile and keyFile are required for mTLS")
		}
		material, err := runtimeconfig.ReadFiles(map[string]string{"cert": config.CertFile, "key": config.KeyFile})
		if err != nil {
			return nil, err
		}
		certificate, err := tls.X509KeyPair(material["cert"], material["key"])
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}

type execTarget struct {
	command string
	args    []*compiledTemplate
	env     map[string]*compiledTemplate
	stdin   *compiledTemplate
}

func newExecTarget(rule dynamic.ExternalActionRule) (target, error) {
	if !filepath.IsAbs(rule.Exec.Command) {
		return nil, errors.New("exec command must be an absolute path")
	}
	if strings.ContainsAny(rule.Exec.Command, "\r\n\x00") {
		return nil, errors.New("exec command contains invalid characters")
	}
	value := &execTarget{
		command: rule.Exec.Command,
		env:     make(map[string]*compiledTemplate),
	}
	for index, argument := range rule.Exec.Args {
		parsed, err := newTemplate(fmt.Sprintf("arg-%d", index), argument)
		if err != nil {
			return nil, fmt.Errorf("compiling argument %d: %w", index, err)
		}
		value.args = append(value.args, parsed)
	}
	for name, variable := range rule.Exec.Env {
		parsed, err := newTemplate("env-"+name, variable)
		if err != nil {
			return nil, fmt.Errorf("compiling environment variable %q: %w", name, err)
		}
		value.env[name] = parsed
	}
	stdin, err := newTemplate("stdin", rule.Exec.Stdin)
	if err != nil {
		return nil, fmt.Errorf("compiling stdin: %w", err)
	}
	value.stdin = stdin
	return value, nil
}

func (e *execTarget) Run(ctx context.Context, data *actionData) error {
	arguments := make([]string, len(e.args))
	for index, argument := range e.args {
		value, err := argument.render(data)
		if err != nil {
			return fmt.Errorf("rendering argument %d: %w", index, err)
		}
		arguments[index] = value
	}
	command := exec.CommandContext(ctx, e.command, arguments...)
	command.Env = os.Environ()
	for name, variable := range e.env {
		value, err := variable.render(data)
		if err != nil {
			return fmt.Errorf("rendering environment variable %q: %w", name, err)
		}
		command.Env = append(command.Env, name+"="+value)
	}
	stdin, err := e.stdin.render(data)
	if err != nil {
		return fmt.Errorf("rendering stdin: %w", err)
	}
	command.Stdin = strings.NewReader(stdin)

	stdout := &limitedBuffer{remaining: 64 * 1024}
	command.Stdout = stdout
	command.Stderr = stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf("command failed: %w: %s", err, strings.TrimSpace(stdout.String()))
	}
	return nil
}

func compileTemplates(prefix string, values map[string]string) (map[string]*compiledTemplate, error) {
	result := make(map[string]*compiledTemplate, len(values))
	for name, value := range values {
		parsed, err := newTemplate(prefix+"-"+name, value)
		if err != nil {
			return nil, fmt.Errorf("compiling %s %q: %w", prefix, name, err)
		}
		result[name] = parsed
	}
	return result, nil
}

func selectJSON(value any, path string) any {
	tokens, ok := parseJSONPath(path)
	if !ok {
		return nil
	}
	if len(tokens) == 0 {
		return value
	}

	current := value
	for _, token := range tokens {
		if token.index == nil {
			object, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current = object[token.key]
		} else {
			array, ok := current.([]any)
			if !ok || *token.index < 0 || *token.index >= len(array) {
				return nil
			}
			current = array[*token.index]
		}
	}
	return current
}

type jsonPathToken struct {
	key   string
	index *int
}

func parseJSONPath(value string) ([]jsonPathToken, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "$" {
		return nil, true
	}
	if value[0] == '$' {
		value = value[1:]
	}

	var result []jsonPathToken
	for position := 0; position < len(value); {
		switch value[position] {
		case '.':
			position++
			if position == len(value) || value[position] == '.' || value[position] == '[' {
				return nil, false
			}
		case '[':
			token, next, ok := parseJSONPathBracket(value, position)
			if !ok {
				return nil, false
			}
			result = append(result, token)
			position = next
			continue
		}

		start := position
		for position < len(value) && value[position] != '.' && value[position] != '[' {
			position++
		}
		if start == position {
			return nil, false
		}
		result = append(result, jsonPathToken{key: value[start:position]})
	}
	return result, true
}

func parseJSONPathBracket(value string, position int) (jsonPathToken, int, bool) {
	position++
	if position >= len(value) {
		return jsonPathToken{}, 0, false
	}

	if value[position] == '"' || value[position] == '\'' {
		quote := value[position]
		position++
		var key strings.Builder
		for position < len(value) {
			character := value[position]
			if character == quote {
				position++
				if position >= len(value) || value[position] != ']' {
					return jsonPathToken{}, 0, false
				}
				return jsonPathToken{key: key.String()}, position + 1, true
			}
			if character == '\\' {
				position++
				if position >= len(value) {
					return jsonPathToken{}, 0, false
				}
				escaped := value[position]
				switch escaped {
				case '\\', '"', '\'':
					key.WriteByte(escaped)
				case 'n':
					key.WriteByte('\n')
				case 'r':
					key.WriteByte('\r')
				case 't':
					key.WriteByte('\t')
				default:
					return jsonPathToken{}, 0, false
				}
				position++
				continue
			}
			key.WriteByte(character)
			position++
		}
		return jsonPathToken{}, 0, false
	}

	start := position
	for position < len(value) && value[position] >= '0' && value[position] <= '9' {
		position++
	}
	if start == position || position >= len(value) || value[position] != ']' {
		return jsonPathToken{}, 0, false
	}
	index, err := strconv.Atoi(value[start:position])
	if err != nil {
		return jsonPathToken{}, 0, false
	}
	return jsonPathToken{index: &index}, position + 1, true
}

type limitedBuffer struct {
	value     bytes.Buffer
	remaining int
	truncated bool
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	original := len(value)
	if len(value) > b.remaining {
		value = value[:max(b.remaining, 0)]
		b.truncated = true
	}
	if len(value) > 0 {
		_, _ = b.value.Write(value)
		b.remaining -= len(value)
	}
	return original, nil
}

func (b *limitedBuffer) String() string {
	if b.truncated {
		return b.value.String() + "…"
	}
	return b.value.String()
}
