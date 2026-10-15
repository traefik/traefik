package externalaction

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

const (
	defaultBindingTTL       = 180 * 24 * time.Hour
	defaultBindingTable     = "traefik_external_action_binding"
	maxMemoryBindingEntries = 1_000_000
)

var errLifecycleBindingNotFound = errors.New("lifecycle binding was not found or has expired")

type compiledBinding struct {
	config      dynamic.ExternalActionBinding
	key         *compiledTemplate
	store       lifecycleBindingStore
	credentials map[string]*compiledAuth
	allowed     map[string]struct{}
	match       map[string]*compiledTemplate
}

type lifecycleBindingStore interface {
	Bind(ctx context.Context, key, credentialSet string, identity [32]byte, ttl time.Duration) error
	Resolve(ctx context.Context, key string) (string, [32]byte, error)
}

func newCompiledBinding(ctx context.Context, config dynamic.ExternalActionBinding, credentials map[string]*compiledAuth) (*compiledBinding, error) {
	key, err := newTemplate("lifecycle-binding-key", config.KeyTemplate)
	if err != nil {
		return nil, fmt.Errorf("compiling keyTemplate: %w", err)
	}
	store, err := getLifecycleBindingStore(ctx, config)
	if err != nil {
		return nil, err
	}
	value := &compiledBinding{
		config:      config,
		key:         key,
		store:       store,
		credentials: credentials,
		allowed:     make(map[string]struct{}, len(config.AllowedCredentialSets)),
		match:       make(map[string]*compiledTemplate, len(config.Match.Templates)),
	}
	for _, name := range config.AllowedCredentialSets {
		if _, exists := value.allowed[name]; exists {
			return nil, fmt.Errorf("allowed credential set %q is duplicated", name)
		}
		value.allowed[name] = struct{}{}
	}
	createsBinding := strings.EqualFold(config.Mode, "create") || strings.EqualFold(config.Mode, "resolveOrCreate")
	if createsBinding {
		candidateNames := []string{config.CredentialSet}
		if len(config.AllowedCredentialSets) > 0 {
			candidateNames = config.AllowedCredentialSets
		}
		for _, credentialName := range candidateNames {
			credential := credentials[credentialName]
			if credential == nil {
				return nil, fmt.Errorf("credential set %q is not configured", credentialName)
			}
			immutableParameters := credential.immutableParameterNames()
			if len(immutableParameters) == 0 {
				return nil, fmt.Errorf("credential set %q must define immutableParameters for lifecycle binding", credentialName)
			}
			immutable := make(map[string]struct{}, len(immutableParameters))
			for _, name := range immutableParameters {
				immutable[name] = struct{}{}
			}
			for name := range config.Match.Headers {
				if _, ok := immutable[name]; !ok {
					return nil, fmt.Errorf("binding match parameter %q is not immutable in credential set %q", name, credentialName)
				}
			}
			for name := range config.Match.BodyJSONPaths {
				if _, ok := immutable[name]; !ok {
					return nil, fmt.Errorf("binding match parameter %q is not immutable in credential set %q", name, credentialName)
				}
			}
			for name := range config.Match.ResponseBodyJSONPaths {
				if _, ok := immutable[name]; !ok {
					return nil, fmt.Errorf("binding match parameter %q is not immutable in credential set %q", name, credentialName)
				}
			}
			for name := range config.Match.Templates {
				if _, ok := immutable[name]; !ok {
					return nil, fmt.Errorf("binding match parameter %q is not immutable in credential set %q", name, credentialName)
				}
			}
		}
		for name, header := range config.Match.Headers {
			if strings.TrimSpace(header) == "" {
				return nil, fmt.Errorf("binding match header for parameter %q is empty", name)
			}
		}
		for name, path := range config.Match.BodyJSONPaths {
			if _, ok := parseJSONPath(path); !ok {
				return nil, fmt.Errorf("binding request body JSONPath for parameter %q is invalid", name)
			}
		}
		for name, path := range config.Match.ResponseBodyJSONPaths {
			if _, ok := parseJSONPath(path); !ok {
				return nil, fmt.Errorf("binding response body JSONPath for parameter %q is invalid", name)
			}
		}
		for name, source := range config.Match.Templates {
			templateValue, err := newTemplate("binding-identity-"+name, source)
			if err != nil {
				return nil, fmt.Errorf("compiling binding match template for parameter %q: %w", name, err)
			}
			value.match[name] = templateValue
		}
	}
	if strings.EqualFold(config.Mode, "resolve") || strings.EqualFold(config.Mode, "resolveOrCreate") {
		for name := range value.allowed {
			credential := credentials[name]
			if credential == nil {
				return nil, fmt.Errorf("allowed credential set %q is not configured", name)
			}
			if len(credential.immutableParameterNames()) == 0 {
				return nil, fmt.Errorf("credential set %q must define immutableParameters for lifecycle binding", name)
			}
		}
	}
	return value, nil
}

func (b *compiledBinding) matches(data *actionData) (bool, error) {
	if b == nil || !(strings.EqualFold(b.config.Mode, "create") || strings.EqualFold(b.config.Mode, "resolveOrCreate")) {
		return true, nil
	}
	if len(b.config.Match.Headers) == 0 && len(b.config.Match.BodyJSONPaths) == 0 && len(b.config.Match.ResponseBodyJSONPaths) == 0 && len(b.match) == 0 {
		return true, nil
	}
	if len(b.config.AllowedCredentialSets) == 0 {
		return b.matchesCredential(data, b.config.CredentialSet)
	}

	selected := ""
	for _, credentialSet := range b.config.AllowedCredentialSets {
		matches, err := b.matchesCredential(data, credentialSet)
		if err != nil {
			return false, err
		}
		if !matches {
			continue
		}
		if selected != "" {
			return false, fmt.Errorf("source identity matches multiple allowed credential sets %q and %q", selected, credentialSet)
		}
		selected = credentialSet
	}
	if selected == "" {
		return false, nil
	}
	data.credentialSet = selected
	return true, nil
}

func (b *compiledBinding) matchesCredential(data *actionData, credentialSet string) (bool, error) {
	credential := b.credentials[credentialSet]
	expected, err := credential.immutableValues()
	if err != nil {
		return false, fmt.Errorf("loading credential set %q identity: %w", credentialSet, err)
	}
	for name, header := range b.config.Match.Headers {
		if data.Request.Headers.Get(header) != expected[name] {
			return false, nil
		}
	}
	for name, path := range b.config.Match.BodyJSONPaths {
		if matchValue(selectJSON(data.Body, path)) != expected[name] {
			return false, nil
		}
	}
	for name, path := range b.config.Match.ResponseBodyJSONPaths {
		if matchValue(selectJSON(data.Response.Body, path)) != expected[name] {
			return false, nil
		}
	}
	for name, templateValue := range b.match {
		actual, err := templateValue.render(data)
		if err != nil {
			return false, fmt.Errorf("rendering binding identity parameter %q: %w", name, err)
		}
		if strings.TrimSpace(actual) != expected[name] {
			return false, nil
		}
	}
	return true, nil
}

func (b *compiledBinding) pin(data *actionData) (*actionData, error) {
	if b == nil || !(strings.EqualFold(b.config.Mode, "create") || strings.EqualFold(b.config.Mode, "resolveOrCreate")) {
		return data, nil
	}
	credentialSet := b.config.CredentialSet
	if data.credentialSet != "" {
		credentialSet = data.credentialSet
	}
	credential := b.credentials[credentialSet]
	if credential == nil {
		return nil, fmt.Errorf("selected credential set %q is not configured", credentialSet)
	}
	identity, err := credential.identity()
	if err != nil {
		return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, err)
	}
	result := *data
	result.credentialSet = credentialSet
	result.expectedCredentialIdentity = identity
	return &result, nil
}

func (b *compiledBinding) prepare(ctx context.Context, data *actionData) (*actionData, error) {
	if b == nil {
		return data, nil
	}
	key, err := b.key.render(data)
	if err != nil {
		return nil, fmt.Errorf("rendering lifecycle binding key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errors.New("rendered lifecycle binding key is empty")
	}

	result := *data
	switch strings.ToLower(b.config.Mode) {
	case "create":
		credentialSet := b.config.CredentialSet
		if data.credentialSet != "" {
			credentialSet = data.credentialSet
		}
		credential := b.credentials[credentialSet]
		if credential == nil {
			return nil, fmt.Errorf("selected credential set %q is not configured", credentialSet)
		}
		identity := data.expectedCredentialIdentity
		if identity == ([32]byte{}) {
			identity, err = credential.identity()
			if err != nil {
				return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, err)
			}
		}
		if err := b.store.Bind(ctx, key, credentialSet, identity, bindingTTL(time.Duration(b.config.TTL))); err != nil {
			return nil, fmt.Errorf("creating lifecycle binding: %w", err)
		}
		result.credentialSet = credentialSet
		result.expectedCredentialIdentity = identity

	case "resolve":
		credentialSet, identity, err := b.store.Resolve(ctx, key)
		if err != nil {
			if errors.Is(err, errLifecycleBindingNotFound) && strings.EqualFold(b.config.OnMissing, "skip") {
				return nil, errSkipRule
			}
			return nil, fmt.Errorf("resolving lifecycle binding: %w", err)
		}
		if len(b.allowed) > 0 {
			if _, ok := b.allowed[credentialSet]; !ok {
				return nil, fmt.Errorf("resolved credential set %q is not allowed by this rule", credentialSet)
			}
		}
		credential := b.credentials[credentialSet]
		if credential == nil {
			return nil, fmt.Errorf("resolved credential set %q is not configured", credentialSet)
		}
		currentIdentity, err := credential.identity()
		if err != nil {
			return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, err)
		}
		if currentIdentity != identity {
			return nil, fmt.Errorf("credential set %q immutable identity does not match its persisted lifecycle binding", credentialSet)
		}
		result.credentialSet = credentialSet
		result.expectedCredentialIdentity = identity

	case "resolveorcreate":
		credentialSet, identity, resolveErr := b.store.Resolve(ctx, key)
		if resolveErr == nil {
			if data.credentialSet != "" && data.credentialSet != credentialSet {
				return nil, fmt.Errorf("resolved credential set %q does not match source identity credential set %q", credentialSet, data.credentialSet)
			}
			if _, ok := b.allowed[credentialSet]; !ok {
				return nil, fmt.Errorf("resolved credential set %q is not allowed by this rule", credentialSet)
			}
			credential := b.credentials[credentialSet]
			if credential == nil {
				return nil, fmt.Errorf("resolved credential set %q is not configured", credentialSet)
			}
			currentIdentity, identityErr := credential.identity()
			if identityErr != nil {
				return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, identityErr)
			}
			if currentIdentity != identity {
				return nil, fmt.Errorf("credential set %q immutable identity does not match its persisted lifecycle binding", credentialSet)
			}
			result.credentialSet = credentialSet
			result.expectedCredentialIdentity = identity
			break
		}
		if !errors.Is(resolveErr, errLifecycleBindingNotFound) {
			return nil, fmt.Errorf("resolving lifecycle binding: %w", resolveErr)
		}

		credentialSet = data.credentialSet
		credential := b.credentials[credentialSet]
		if credential == nil {
			return nil, fmt.Errorf("selected credential set %q is not configured", credentialSet)
		}
		identity = data.expectedCredentialIdentity
		if identity == ([32]byte{}) {
			identity, err = credential.identity()
			if err != nil {
				return nil, fmt.Errorf("loading credential set %q: %w", credentialSet, err)
			}
		}
		if err := b.store.Bind(ctx, key, credentialSet, identity, bindingTTL(time.Duration(b.config.TTL))); err != nil {
			return nil, fmt.Errorf("reconciling lifecycle binding: %w", err)
		}
		result.credentialSet = credentialSet
		result.expectedCredentialIdentity = identity
		result.LifecycleBinding.Created = true

	default:
		return nil, fmt.Errorf("unsupported lifecycle binding mode %q", b.config.Mode)
	}
	return &result, nil
}

func bindingTTL(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultBindingTTL
	}
	return value
}

type bindingBackendConfig struct {
	Memory *dynamic.ExternalActionBindingMemory `json:"memory,omitempty"`
	MySQL  *dynamic.ExternalActionBindingMySQL  `json:"mysql,omitempty"`
}

type bindingStoreKey struct {
	done   <-chan struct{}
	config [32]byte
}

var bindingStoreRegistry = struct {
	sync.Mutex

	values map[bindingStoreKey]lifecycleBindingStore
}{
	values: make(map[bindingStoreKey]lifecycleBindingStore),
}

func getLifecycleBindingStore(ctx context.Context, config dynamic.ExternalActionBinding) (lifecycleBindingStore, error) {
	backend := bindingBackendConfig{Memory: config.Memory, MySQL: config.MySQL}
	raw, err := json.Marshal(backend)
	if err != nil {
		return nil, fmt.Errorf("encoding binding backend configuration: %w", err)
	}
	key := bindingStoreKey{done: ctx.Done(), config: sha256.Sum256(raw)}

	bindingStoreRegistry.Lock()
	defer bindingStoreRegistry.Unlock()
	if existing := bindingStoreRegistry.values[key]; existing != nil {
		return existing, nil
	}

	var value lifecycleBindingStore
	switch {
	case config.Memory != nil:
		maxEntries := config.Memory.MaxEntries
		if maxEntries <= 0 {
			maxEntries = 100_000
		}
		if maxEntries > maxMemoryBindingEntries {
			return nil, fmt.Errorf("memory maxEntries must not exceed %d", maxMemoryBindingEntries)
		}
		value = &memoryLifecycleBindingStore{
			bindings:   make(map[[32]byte]memoryLifecycleBinding),
			identities: make(map[string][32]byte),
			maxEntries: maxEntries,
		}
	case config.MySQL != nil:
		mysqlValue, err := newMySQLLifecycleBindingStore(ctx, *config.MySQL)
		if err != nil {
			return nil, err
		}
		value = mysqlValue
	default:
		return nil, errors.New("lifecycle binding backend is required")
	}

	bindingStoreRegistry.values[key] = value
	if ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			if closer, ok := value.(interface{ close() }); ok {
				closer.close()
			}
			bindingStoreRegistry.Lock()
			delete(bindingStoreRegistry.values, key)
			bindingStoreRegistry.Unlock()
		}()
	}
	return value, nil
}

type memoryLifecycleBinding struct {
	credentialSet string
	identity      [32]byte
	expiresAt     time.Time
}

type memoryLifecycleBindingStore struct {
	mu          sync.Mutex
	bindings    map[[32]byte]memoryLifecycleBinding
	identities  map[string][32]byte
	maxEntries  int
	lastCleanup time.Time
}

func (m *memoryLifecycleBindingStore) Bind(_ context.Context, key, credentialSet string, identity [32]byte, ttl time.Duration) error {
	now := time.Now()
	hash := lifecycleBindingHash(key)

	m.mu.Lock()
	defer m.mu.Unlock()
	if known, ok := m.identities[credentialSet]; ok && known != identity {
		return fmt.Errorf("credential set %q was already registered with another immutable identity", credentialSet)
	}
	m.identities[credentialSet] = identity

	if now.Sub(m.lastCleanup) >= time.Minute || len(m.bindings) >= m.maxEntries {
		for candidate, entry := range m.bindings {
			if !entry.expiresAt.After(now) {
				delete(m.bindings, candidate)
			}
		}
		m.lastCleanup = now
	}
	if existing, ok := m.bindings[hash]; ok && existing.expiresAt.After(now) {
		if existing.credentialSet != credentialSet || existing.identity != identity {
			return errors.New("lifecycle key is already bound to another credential set")
		}
		existing.expiresAt = now.Add(ttl)
		m.bindings[hash] = existing
		return nil
	}
	if len(m.bindings) >= m.maxEntries {
		return errors.New("memory lifecycle binding capacity is full")
	}
	m.bindings[hash] = memoryLifecycleBinding{
		credentialSet: credentialSet,
		identity:      identity,
		expiresAt:     now.Add(ttl),
	}
	return nil
}

func (m *memoryLifecycleBindingStore) Resolve(_ context.Context, key string) (string, [32]byte, error) {
	now := time.Now()
	hash := lifecycleBindingHash(key)
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.bindings[hash]
	if !ok || !entry.expiresAt.After(now) {
		delete(m.bindings, hash)
		return "", [32]byte{}, errLifecycleBindingNotFound
	}
	return entry.credentialSet, entry.identity, nil
}

type mysqlLifecycleBindingStore struct {
	config  dynamic.ExternalActionBindingMySQL
	queries mysqlBindingQueries
	manager *runtimeconfig.MySQLManager
	cleanup atomic.Int64
}

func newMySQLLifecycleBindingStore(ctx context.Context, config dynamic.ExternalActionBindingMySQL) (*mysqlLifecycleBindingStore, error) {
	if strings.TrimSpace(config.DSN) == "" && strings.TrimSpace(config.DSNFile) == "" {
		return nil, errors.New("mysql dsn or dsnFile is required")
	}
	table := strings.TrimSpace(config.Table)
	if table == "" {
		table = defaultBindingTable
	}
	if !deduplicationTablePattern.MatchString(table) {
		return nil, fmt.Errorf("mysql table must match %s", deduplicationTablePattern)
	}
	queries, err := compileMySQLBindingQueries(table)
	if err != nil {
		return nil, err
	}
	connectionID := config.ConnectionID
	if connectionID == "" {
		var err error
		connectionID, err = runtimeconfig.SourceIdentity(config.DSN, config.DSNFile, nil)
		if err != nil {
			return nil, err
		}
	}
	managerKey := runtimeconfig.StableKey("external-action-binding", connectionID, table, strconv.FormatBool(config.AutoCreate))
	var initializer runtimeconfig.MySQLInitializer
	if config.AutoCreate {
		initializer = func(schemaContext context.Context, db *sql.DB) error {
			return ensureBindingSchema(schemaContext, db, table)
		}
	}
	manager, err := runtimeconfig.GetMySQLManager(ctx, managerKey, runtimeconfig.MySQLSettings{
		DSN:                config.DSN,
		DSNFile:            config.DSNFile,
		MaxOpenConnections: 4,
		MaxIdleConnections: 2,
		ConnectionMaxAge:   30 * time.Minute,
	}, initializer)
	if err != nil {
		return nil, err
	}
	return &mysqlLifecycleBindingStore{config: config, queries: queries, manager: manager}, nil
}

func (m *mysqlLifecycleBindingStore) Bind(ctx context.Context, key, credentialSet string, identity [32]byte, ttl time.Duration) error {
	if m.queries.insertCredential == "" {
		return errors.New("MySQL binding queries are not initialized")
	}
	db, release, err := m.manager.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	m.cleanupExpired(ctx, db)

	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Rollback is harmless after Commit and releases the transaction on every error path.
	defer func() { _ = transaction.Rollback() }()

	identityKey := credentialIdentityHash(credentialSet)
	_, err = m.queries.insertCredential.exec(ctx, transaction,
		identityKey[:], credentialSet, credentialSet, identity[:])
	if err != nil {
		return err
	}
	var knownIdentity []byte
	err = m.queries.selectCredential.scan(ctx, transaction, []any{identityKey[:]}, &knownIdentity)
	if err != nil {
		return err
	}
	if !equalIdentity(knownIdentity, identity) {
		return fmt.Errorf("credential set %q was already registered with another immutable identity", credentialSet)
	}

	hash := lifecycleBindingHash(key)
	var existingSet string
	var existingIdentity []byte
	var expired bool
	err = m.queries.selectBinding.scan(ctx, transaction, []any{hash[:]}, &existingSet, &existingIdentity, &expired)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = m.queries.insertBinding.exec(ctx, transaction,
			hash[:], bindingKeyPreview(key), credentialSet, identity[:], durationMicroseconds(ttl))
	case err != nil:
		return err
	case !expired && (existingSet != credentialSet || !equalIdentity(existingIdentity, identity)):
		return errors.New("lifecycle key is already bound to another credential set")
	default:
		_, err = m.queries.updateBinding.exec(ctx, transaction,
			bindingKeyPreview(key), credentialSet, identity[:], durationMicroseconds(ttl), hash[:])
	}
	if err != nil {
		return err
	}
	return transaction.Commit()
}

func (m *mysqlLifecycleBindingStore) Resolve(ctx context.Context, key string) (string, [32]byte, error) {
	if m.queries.resolveBinding == "" {
		return "", [32]byte{}, errors.New("MySQL binding queries are not initialized")
	}
	db, release, err := m.manager.Acquire(ctx)
	if err != nil {
		return "", [32]byte{}, err
	}
	defer release()
	hash := lifecycleBindingHash(key)
	var credentialSet string
	var identityBytes []byte
	err = m.queries.resolveBinding.scan(ctx, db, []any{hash[:]}, &credentialSet, &identityBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return "", [32]byte{}, errLifecycleBindingNotFound
	}
	if err != nil {
		return "", [32]byte{}, err
	}
	if len(identityBytes) != sha256.Size {
		return "", [32]byte{}, errors.New("lifecycle binding contains an invalid identity hash")
	}
	var identity [32]byte
	copy(identity[:], identityBytes)
	return credentialSet, identity, nil
}

func ensureBindingSchema(ctx context.Context, db *sql.DB, table string) error {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return err
	}
	query := "CREATE TABLE IF NOT EXISTS " + quotedTable + " (" +
		"`key_hash` BINARY(32) NOT NULL," +
		"`binding_kind` VARCHAR(16) NOT NULL," +
		"`binding_key` VARCHAR(512) NOT NULL," +
		"`credential_set` VARCHAR(255) NOT NULL," +
		"`identity_hash` BINARY(32) NOT NULL," +
		"`expires_at` DATETIME(6) NULL," +
		"`created_at` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)," +
		"`updated_at` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)," +
		"PRIMARY KEY (`key_hash`)," +
		"KEY `idx_credential_set` (`credential_set`)," +
		"KEY `idx_expires_at` (`expires_at`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
	_, err = db.ExecContext(ctx, query)
	return err
}

func (m *mysqlLifecycleBindingStore) cleanupExpired(ctx context.Context, db *sql.DB) {
	if m.queries.cleanup == "" {
		return
	}
	now := time.Now().Unix()
	last := m.cleanup.Load()
	if last != 0 && now-last < int64(time.Minute/time.Second) {
		return
	}
	if !m.cleanup.CompareAndSwap(last, now) {
		return
	}
	_, _ = m.queries.cleanup.exec(ctx, db)
}

func (m *mysqlLifecycleBindingStore) close() {}

func lifecycleBindingHash(key string) [32]byte {
	return sha256.Sum256([]byte("lifecycle\x00" + key))
}

func credentialIdentityHash(credentialSet string) [32]byte {
	return sha256.Sum256([]byte("credential-set\x00" + credentialSet))
}

func equalIdentity(value []byte, expected [32]byte) bool {
	return len(value) == len(expected) && string(value) == string(expected[:])
}

func bindingKeyPreview(value string) string {
	runes := []rune(value)
	if len(runes) > 256 {
		runes = runes[:256]
	}
	return string(runes)
}
