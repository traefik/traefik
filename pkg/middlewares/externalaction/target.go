package externalaction

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

type target interface {
	Run(ctx context.Context, data *actionData) error
}

type httpTarget struct {
	enrichments     []*compiledEnrichment
	perform         func(*http.Request) (*http.Response, error)
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
	return h.run(ctx, data, nil)
}

func (h *httpTarget) run(ctx context.Context, data *actionData, capture func([]byte) error) (err error) {
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
	// Template, transport and mapping failures can echo credentials as well as
	// upstream error responses. Redact at the complete authenticated-call boundary.
	defer func() { err = bound.redactError(err) }()
	if warning != nil {
		log.Ctx(ctx).Warn().Err(warning).Str("credentialSet", credentialSet).Msg("Using the last valid external action credentials")
	}
	if data.expectedCredentialIdentity != ([32]byte{}) && identity != data.expectedCredentialIdentity {
		return fmt.Errorf("credential set %q immutable identity does not match the lifecycle binding", credentialSet)
	}
	render := (*compiledTemplate).render
	if capture != nil {
		render = renderEnrichmentTemplate
	}
	body, err := render(h.body, renderData)
	if err != nil {
		return fmt.Errorf("rendering body: %w", err)
	}
	if strings.Contains(h.contentType, "json") && !json.Valid([]byte(body)) {
		return errors.New("rendered external action body is not valid JSON")
	}
	// Queries run only for an already selected event, with its pinned credentials.
	// Lookup targets never recurse into the request's deferred enrichments.
	if capture == nil {
		renderData, body, err = enrichPayload(ctx, renderData, body, h.enrichments)
		if err != nil {
			return bound.redactError(err)
		}
	}
	targetURL, err := render(h.url, renderData)
	if err != nil {
		return bound.redactError(fmt.Errorf("rendering URL: %w", err))
	}
	request, err := http.NewRequestWithContext(ctx, h.method, targetURL, strings.NewReader(body))
	if err != nil {
		return err
	}
	renderData.OutboundBody = body
	if capture != nil && len(renderData.lookupPageQuery) != 0 {
		query := request.URL.Query()
		maps.Copy(query, renderData.lookupPageQuery)
		request.URL.RawQuery = query.Encode()
	}
	for name, value := range h.headers {
		rendered, err := render(value, renderData)
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

	perform := h.perform
	if perform == nil {
		perform = h.client.Do
	}
	response, err := perform(request)
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
	if capture != nil {
		return capture(responseBody)
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
