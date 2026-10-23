package externalaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
	"github.com/traefik/traefik/v3/pkg/types"
)

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
	foreachResponse := rule.ForEach != nil && strings.EqualFold(rule.ForEach.Source, "responseBody")
	requiresResponseBody := len(rule.Match.ResponseBodyJSONPaths) > 0 || foreachResponse
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
