package externalaction

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

const (
	stateClaimed = "claimed"
)

const (
	defaultDeduplicationTTL   = 30 * 24 * time.Hour
	defaultDeduplicationLease = time.Minute
	defaultDeduplicationTable = "traefik_external_action_dedup"
	maxMemoryDedupEntries     = 1_000_000
)

var deduplicationTablePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

var errSkipRule = errors.New("external action rule skipped")

type deduplicator interface {
	Claim(ctx context.Context, rule, key string, ttl, lease time.Duration) (string, bool, error)
	Complete(ctx context.Context, key, claimID string, ttl time.Duration) error
	Release(ctx context.Context, key, claimID string) error
}

func newDeduplicator(ctx context.Context, config dynamic.ExternalActionDeduplication) (deduplicator, error) {
	switch {
	case config.Memory != nil:
		maxEntries := config.Memory.MaxEntries
		if maxEntries <= 0 {
			maxEntries = 100_000
		}
		if maxEntries > maxMemoryDedupEntries {
			return nil, fmt.Errorf("memory maxEntries must not exceed %d", maxMemoryDedupEntries)
		}
		return &memoryDeduplicator{
			entries:    make(map[[32]byte]memoryDeduplicationEntry),
			maxEntries: maxEntries,
		}, nil
	case config.MySQL != nil:
		return getMySQLDeduplicator(ctx, *config.MySQL)
	default:
		return nil, errors.New("deduplication backend is required")
	}
}

func runRule(ctx context.Context, rule *compiledRule, data *actionData) error {
	if rule.deduplication == nil {
		err := executeRule(ctx, rule, data)
		if errors.Is(err, errSkipRule) {
			return nil
		}
		return err
	}

	key, err := rule.deduplicationKey.render(data)
	if err != nil {
		return fmt.Errorf("rendering deduplication key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("rendered deduplication key is empty")
	}

	ttl := time.Duration(rule.config.Deduplication.TTL)
	if ttl <= 0 {
		ttl = defaultDeduplicationTTL
	}
	lease := time.Duration(rule.config.Deduplication.Lease)
	if lease <= 0 {
		lease = defaultDeduplicationLease
	}

	claimID, claimed, err := rule.deduplication.Claim(ctx, rule.config.Name, key, ttl, lease)
	if err != nil {
		return fmt.Errorf("claiming deduplication key: %w", err)
	}
	if !claimed {
		return nil
	}

	if err := executeRule(ctx, rule, data); err != nil {
		releaseContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		releaseErr := rule.deduplication.Release(releaseContext, key, claimID)
		cancel()
		if releaseErr != nil {
			return errors.Join(err, fmt.Errorf("releasing deduplication key: %w", releaseErr))
		}
		if errors.Is(err, errSkipRule) {
			return nil
		}
		return err
	}

	completeContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	err = rule.deduplication.Complete(completeContext, key, claimID, ttl)
	cancel()
	if err != nil {
		return fmt.Errorf("completing deduplication key: %w", err)
	}
	return nil
}

func executeRule(ctx context.Context, rule *compiledRule, data *actionData) error {
	prepared, err := rule.binding.prepare(ctx, data)
	if err != nil {
		return err
	}
	return rule.target.Run(ctx, prepared)
}

type memoryDeduplicationEntry struct {
	claimID   string
	state     string
	leaseTill time.Time
	expiresAt time.Time
}

type memoryDeduplicator struct {
	mu          sync.Mutex
	entries     map[[32]byte]memoryDeduplicationEntry
	maxEntries  int
	lastCleanup time.Time
	nextExpiry  time.Time
}

func (m *memoryDeduplicator) Claim(_ context.Context, _ string, key string, ttl, lease time.Duration) (string, bool, error) {
	now := time.Now()
	hash := sha256.Sum256([]byte(key))

	m.mu.Lock()
	defer m.mu.Unlock()

	// At capacity, scan only when an entry could have expired. Repeated claims
	// against a full live map must not turn each request into a full-map scan.
	if now.Sub(m.lastCleanup) >= time.Minute || (len(m.entries) >= m.maxEntries && !m.nextExpiry.After(now)) {
		m.nextExpiry = time.Time{}
		for candidate, entry := range m.entries {
			if !entry.expiresAt.After(now) {
				delete(m.entries, candidate)
				continue
			}
			m.trackExpiry(entry.expiresAt)
		}
		m.lastCleanup = now
	}
	if entry, ok := m.entries[hash]; ok {
		if entry.state == "completed" && entry.expiresAt.After(now) {
			return "", false, nil
		}
		if entry.state == stateClaimed && entry.leaseTill.After(now) {
			return "", false, nil
		}
	}
	if _, ok := m.entries[hash]; !ok && len(m.entries) >= m.maxEntries {
		return "", false, errors.New("memory deduplication capacity is full")
	}

	claimID, err := newClaimID()
	if err != nil {
		return "", false, err
	}
	m.entries[hash] = memoryDeduplicationEntry{
		claimID:   claimID,
		state:     stateClaimed,
		leaseTill: now.Add(lease),
		expiresAt: now.Add(ttl),
	}
	m.trackExpiry(now.Add(ttl))
	return claimID, true, nil
}

func (m *memoryDeduplicator) Complete(_ context.Context, key, claimID string, ttl time.Duration) error {
	hash := sha256.Sum256([]byte(key))
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[hash]
	if !ok || entry.claimID != claimID || entry.state != stateClaimed {
		return errors.New("deduplication claim no longer belongs to this action")
	}
	entry.state = "completed"
	entry.leaseTill = now
	entry.expiresAt = now.Add(ttl)
	m.entries[hash] = entry
	m.trackExpiry(entry.expiresAt)
	return nil
}

func (m *memoryDeduplicator) Release(_ context.Context, key, claimID string) error {
	hash := sha256.Sum256([]byte(key))

	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.entries[hash]; ok && entry.claimID == claimID && entry.state == stateClaimed {
		delete(m.entries, hash)
	}
	return nil
}

// Extensions and removals may leave an earlier lower bound, costing at most
// one extra cleanup; shortening an expiry must immediately lower the bound.
func (m *memoryDeduplicator) trackExpiry(expiresAt time.Time) {
	if m.nextExpiry.IsZero() || expiresAt.Before(m.nextExpiry) {
		m.nextExpiry = expiresAt
	}
}

type mysqlDeduplicatorKey struct {
	done   <-chan struct{}
	config [32]byte
}

var mysqlDeduplicatorRegistry = struct {
	sync.Mutex

	values map[mysqlDeduplicatorKey]*mysqlDeduplicator
}{
	values: make(map[mysqlDeduplicatorKey]*mysqlDeduplicator),
}

type mysqlDeduplicator struct {
	config  dynamic.ExternalActionDedupMySQL
	queries mysqlDedupQueries
	manager *runtimeconfig.MySQLManager
	cleanup atomic.Int64
}

func getMySQLDeduplicator(ctx context.Context, config dynamic.ExternalActionDedupMySQL) (*mysqlDeduplicator, error) {
	if strings.TrimSpace(config.DSN) == "" && strings.TrimSpace(config.DSNFile) == "" {
		return nil, errors.New("mysql dsn or dsnFile is required")
	}
	table := strings.TrimSpace(config.Table)
	if table == "" {
		table = defaultDeduplicationTable
	}
	if !deduplicationTablePattern.MatchString(table) {
		return nil, fmt.Errorf("mysql table must match %s", deduplicationTablePattern)
	}

	queries, err := compileMySQLDedupQueries(table)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encoding deduplication configuration: %w", err)
	}
	key := mysqlDeduplicatorKey{done: ctx.Done(), config: sha256.Sum256(raw)}
	mysqlDeduplicatorRegistry.Lock()
	defer mysqlDeduplicatorRegistry.Unlock()
	if existing := mysqlDeduplicatorRegistry.values[key]; existing != nil {
		return existing, nil
	}

	connectionID := config.ConnectionID
	if connectionID == "" {
		var err error
		connectionID, err = runtimeconfig.SourceIdentity(config.DSN, config.DSNFile, nil)
		if err != nil {
			return nil, err
		}
	}
	managerKey := runtimeconfig.StableKey("external-action-deduplication", connectionID, table, strconv.FormatBool(config.AutoCreate))
	var initializer runtimeconfig.MySQLInitializer
	if config.AutoCreate {
		initializer = func(schemaContext context.Context, db *sql.DB) error {
			return ensureDeduplicationSchema(schemaContext, db, table)
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
	value := &mysqlDeduplicator{config: config, queries: queries, manager: manager}
	mysqlDeduplicatorRegistry.values[key] = value
	if ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			value.close()
			mysqlDeduplicatorRegistry.Lock()
			delete(mysqlDeduplicatorRegistry.values, key)
			mysqlDeduplicatorRegistry.Unlock()
		}()
	}
	return value, nil
}

func (m *mysqlDeduplicator) Claim(ctx context.Context, rule, key string, ttl, lease time.Duration) (string, bool, error) {
	if m.queries.selectClaim == "" {
		return "", false, errors.New("MySQL deduplication queries are not initialized")
	}
	db, release, err := m.manager.Acquire(ctx)
	if err != nil {
		return "", false, err
	}
	defer release()
	m.cleanupExpired(ctx, db)

	claimID, err := newClaimID()
	if err != nil {
		return "", false, err
	}
	hash := sha256.Sum256([]byte(key))

	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	// Rollback is harmless after Commit and releases the transaction on every error path.
	defer func() { _ = transaction.Rollback() }()

	var state string
	var leaseExpired, expired bool
	err = m.queries.selectClaim.scan(ctx, transaction, []any{hash[:]}, &state, &leaseExpired, &expired)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = m.queries.insertClaim.exec(ctx, transaction,
			hash[:], rule, claimID, durationMicroseconds(lease), durationMicroseconds(ttl))
		if err != nil {
			var mysqlError *mysqlDriver.MySQLError
			if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
				return "", false, nil
			}
			return "", false, err
		}
	case err != nil:
		return "", false, err
	case state == "completed" && !expired:
		return "", false, nil
	case state == stateClaimed && !leaseExpired && !expired:
		return "", false, nil
	default:
		_, err = m.queries.updateClaim.exec(ctx, transaction,
			rule, claimID, durationMicroseconds(lease), durationMicroseconds(ttl), hash[:])
		if err != nil {
			return "", false, err
		}
	}
	if err := transaction.Commit(); err != nil {
		return "", false, err
	}
	return claimID, true, nil
}

func (m *mysqlDeduplicator) Complete(ctx context.Context, key, claimID string, ttl time.Duration) error {
	if m.queries.complete == "" {
		return errors.New("MySQL deduplication queries are not initialized")
	}
	db, release, err := m.manager.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	hash := sha256.Sum256([]byte(key))
	result, err := m.queries.complete.exec(ctx, db,
		durationMicroseconds(ttl), hash[:], claimID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("deduplication claim no longer belongs to this action")
	}
	return nil
}

func (m *mysqlDeduplicator) Release(ctx context.Context, key, claimID string) error {
	if m.queries.release == "" {
		return errors.New("MySQL deduplication queries are not initialized")
	}
	db, release, err := m.manager.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release()
	hash := sha256.Sum256([]byte(key))
	_, err = m.queries.release.exec(ctx, db,
		hash[:], claimID)
	return err
}

func ensureDeduplicationSchema(ctx context.Context, db *sql.DB, table string) error {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return err
	}
	query := "CREATE TABLE IF NOT EXISTS " + quotedTable + " (" +
		"`key_hash` BINARY(32) NOT NULL," +
		"`rule_name` VARCHAR(255) NOT NULL," +
		"`state` VARCHAR(16) NOT NULL," +
		"`claim_id` CHAR(32) NOT NULL," +
		"`lease_until` DATETIME(6) NOT NULL," +
		"`expires_at` DATETIME(6) NOT NULL," +
		"`created_at` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)," +
		"`updated_at` TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)," +
		"PRIMARY KEY (`key_hash`)," +
		"KEY `idx_expires_at` (`expires_at`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
	_, err = db.ExecContext(ctx, query)
	return err
}

func (m *mysqlDeduplicator) cleanupExpired(ctx context.Context, db *sql.DB) {
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

func (m *mysqlDeduplicator) close() {}

func newClaimID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func durationMicroseconds(value time.Duration) int64 {
	return max(value.Microseconds(), 1)
}
