package externalaction

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

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
	if a == nil || err == nil {
		return err
	}
	value := err.Error()
	for _, secret := range a.secrets {
		value = strings.ReplaceAll(value, secret, "<redacted>")
	}
	if value == err.Error() {
		return err
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
