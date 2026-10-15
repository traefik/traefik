package runtimeconfig

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	// Register the driver used by reloadable MySQL connection pools.
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

const (
	defaultMySQLRefreshInterval = 5 * time.Second
	mysqlManagerIdleRetention   = time.Minute
)

// MySQLSettings controls one reloadable MySQL connection pool.
type MySQLSettings struct {
	DSN                string
	DSNFile            string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnectionMaxAge   time.Duration
	RefreshInterval    time.Duration
}

// MySQLInitializer validates or initializes a candidate database before it is
// made visible to requests.
type MySQLInitializer func(context.Context, *sql.DB) error

type mysqlPool struct {
	db       *sql.DB
	inflight sync.WaitGroup
}

// MySQLManager maintains a validated pool and atomically replaces it when the
// underlying DSN changes. Existing requests retain a lease on the old pool.
type MySQLManager struct {
	key string

	mu             sync.Mutex
	current        *mysqlPool
	activeRevision string
	desired        MySQLSettings
	initializer    MySQLInitializer
	lastErr        error
	nextTry        time.Time
	closed         bool

	reconcileMu sync.Mutex
	wake        chan struct{}
	stop        chan struct{}
	done        chan struct{}
}

type mysqlRegistryEntry struct {
	manager *MySQLManager
	refs    int
	timer   *time.Timer
}

var mysqlRegistry = struct {
	sync.Mutex

	values map[string]*mysqlRegistryEntry
}{
	values: make(map[string]*mysqlRegistryEntry),
}

// StableKey builds a non-secret registry key.
func StableKey(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

// GetMySQLManager returns a process-level manager. The manager is retained
// briefly after its last middleware reference so a dynamic configuration
// rebuild can reuse the active pool without interrupting requests.
func GetMySQLManager(ctx context.Context, key string, settings MySQLSettings, initializer MySQLInitializer) (*MySQLManager, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("mysql manager key is required")
	}
	if strings.TrimSpace(settings.DSN) == "" && strings.TrimSpace(settings.DSNFile) == "" {
		return nil, errors.New("mysql dsn or dsnFile is required")
	}
	if strings.TrimSpace(settings.DSN) != "" && strings.TrimSpace(settings.DSNFile) != "" {
		return nil, errors.New("mysql dsn and dsnFile are mutually exclusive")
	}

	mysqlRegistry.Lock()
	entry := mysqlRegistry.values[key]
	if entry == nil {
		manager := &MySQLManager{
			key:  key,
			wake: make(chan struct{}, 1),
			stop: make(chan struct{}),
			done: make(chan struct{}),
		}
		entry = &mysqlRegistryEntry{manager: manager}
		mysqlRegistry.values[key] = entry
		go manager.maintain()
	}
	if entry.timer != nil {
		entry.timer.Stop()
		entry.timer = nil
	}
	entry.refs++
	manager := entry.manager
	mysqlRegistry.Unlock()

	manager.Configure(settings, initializer)
	if ctx.Done() != nil {
		go func() {
			<-ctx.Done()
			releaseMySQLManager(key, manager)
		}()
	}
	return manager, nil
}

// Configure publishes a new desired configuration. Reconciliation is
// asynchronous when an old pool is healthy, so a bad rotation cannot interrupt
// traffic. Initial acquisition performs a synchronous reconciliation.
func (m *MySQLManager) Configure(settings MySQLSettings, initializer MySQLInitializer) {
	settings = normalizeMySQLSettings(settings)
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.desired = settings
	m.initializer = initializer
	m.nextTry = time.Time{}
	m.mu.Unlock()
	m.signal()
}

// Acquire leases the current pool. The returned release function must be
// called once the operation finishes.
func (m *MySQLManager) Acquire(ctx context.Context) (*sql.DB, func(), error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, nil, errors.New("mysql connection manager is closed")
	}
	if m.current != nil {
		pool := m.current
		pool.inflight.Add(1)
		m.mu.Unlock()
		return pool.db, pool.inflight.Done, nil
	}
	m.mu.Unlock()

	m.reconcile(ctx, true)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil {
		if m.lastErr != nil {
			return nil, nil, m.lastErr
		}
		return nil, nil, errors.New("mysql connection is not ready")
	}
	pool := m.current
	pool.inflight.Add(1)
	return pool.db, pool.inflight.Done, nil
}

func (m *MySQLManager) maintain() {
	defer close(m.done)
	ticker := time.NewTicker(defaultMySQLRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-m.wake:
			m.reconcile(context.Background(), false)
		case <-ticker.C:
			m.reconcile(context.Background(), false)
		}
	}
}

func (m *MySQLManager) reconcile(parent context.Context, force bool) {
	m.reconcileMu.Lock()
	defer m.reconcileMu.Unlock()

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	if !force && time.Now().Before(m.nextTry) {
		m.mu.Unlock()
		return
	}
	settings := m.desired
	initializer := m.initializer
	activeRevision := m.activeRevision
	m.mu.Unlock()

	dsn, revision, err := mysqlMaterial(settings)
	if err != nil {
		m.failed(err)
		return
	}
	if revision == activeRevision {
		return
	}

	candidate, err := openMySQLCandidate(parent, dsn, settings, initializer)
	if err != nil {
		m.failed(err)
		return
	}

	next := &mysqlPool{db: candidate}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		_ = candidate.Close()
		return
	}
	// A newer Configure may have arrived while this candidate was opening.
	currentDSN, currentRevision, materialErr := mysqlMaterial(m.desired)
	if materialErr != nil || currentRevision != revision || currentDSN != dsn {
		m.mu.Unlock()
		_ = candidate.Close()
		m.signal()
		return
	}
	previous := m.current
	m.current = next
	m.activeRevision = revision
	m.lastErr = nil
	m.nextTry = time.Time{}
	m.mu.Unlock()

	if previous != nil {
		go retireMySQLPool(previous)
	}
	log.Info().Str("manager", m.key).Msg("MySQL connection pool activated")
}

func (m *MySQLManager) failed(err error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.lastErr = err
	m.nextTry = time.Now().Add(defaultMySQLRefreshInterval)
	hasCurrent := m.current != nil
	m.mu.Unlock()

	event := log.Warn().Err(err).Str("manager", m.key)
	if hasCurrent {
		event.Msg("Could not activate new MySQL connection; keeping the current pool")
	} else {
		event.Msg("MySQL connection is not ready; maintenance will retry")
	}
}

func (m *MySQLManager) signal() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *MySQLManager) close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	current := m.current
	m.current = nil
	close(m.stop)
	m.mu.Unlock()
	<-m.done
	if current != nil {
		retireMySQLPool(current)
	}
}

func releaseMySQLManager(key string, manager *MySQLManager) {
	mysqlRegistry.Lock()
	entry := mysqlRegistry.values[key]
	if entry == nil || entry.manager != manager {
		mysqlRegistry.Unlock()
		return
	}
	entry.refs--
	if entry.refs > 0 {
		mysqlRegistry.Unlock()
		return
	}
	entry.timer = time.AfterFunc(mysqlManagerIdleRetention, func() {
		mysqlRegistry.Lock()
		current := mysqlRegistry.values[key]
		if current != entry || current.refs != 0 {
			mysqlRegistry.Unlock()
			return
		}
		delete(mysqlRegistry.values, key)
		mysqlRegistry.Unlock()
		manager.close()
	})
	mysqlRegistry.Unlock()
}

func retireMySQLPool(pool *mysqlPool) {
	pool.inflight.Wait()
	if err := pool.db.Close(); err != nil {
		log.Warn().Err(err).Msg("Could not close retired MySQL connection pool")
	}
}

func normalizeMySQLSettings(settings MySQLSettings) MySQLSettings {
	settings.DSN = strings.TrimSpace(settings.DSN)
	settings.DSNFile = strings.TrimSpace(settings.DSNFile)
	if settings.MaxOpenConnections <= 0 {
		settings.MaxOpenConnections = 8
	}
	if settings.MaxIdleConnections < 0 {
		settings.MaxIdleConnections = 0
	} else if settings.MaxIdleConnections == 0 {
		settings.MaxIdleConnections = min(4, settings.MaxOpenConnections)
	}
	if settings.ConnectionMaxAge <= 0 {
		settings.ConnectionMaxAge = 30 * time.Minute
	}
	if settings.RefreshInterval <= 0 {
		settings.RefreshInterval = defaultMySQLRefreshInterval
	}
	return settings
}

func mysqlMaterial(settings MySQLSettings) (string, string, error) {
	dsn := settings.DSN
	if settings.DSNFile != "" {
		material, err := ReadFile(settings.DSNFile)
		if err != nil {
			return "", "", fmt.Errorf("reading mysql dsnFile: %w", err)
		}
		dsn = strings.TrimSpace(string(material))
	}
	if dsn == "" {
		return "", "", errors.New("mysql dsn is empty")
	}
	sum := sha256.Sum256([]byte(dsn))
	return dsn, hex.EncodeToString(sum[:]), nil
}

func openMySQLCandidate(parent context.Context, dsn string, settings MySQLSettings, initializer MySQLInitializer) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	fail := func(err error) (*sql.DB, error) {
		_ = db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(settings.MaxOpenConnections)
	db.SetMaxIdleConns(min(settings.MaxIdleConnections, settings.MaxOpenConnections))
	db.SetConnMaxLifetime(settings.ConnectionMaxAge)

	pingContext, cancel := context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
	err = db.PingContext(pingContext)
	cancel()
	if err != nil {
		return fail(fmt.Errorf("connecting to mysql: %w", err))
	}
	if initializer != nil {
		initializeContext, cancel := context.WithTimeout(context.WithoutCancel(parent), 30*time.Second)
		err = initializer(initializeContext, db)
		cancel()
		if err != nil {
			return fail(fmt.Errorf("initializing mysql connection: %w", err))
		}
	}
	return db, nil
}
