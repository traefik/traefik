package accessgovernance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

const (
	defaultHTTPTable    = "traefik_http_access_events"
	defaultNetworkTable = "traefik_network_access_events"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

type httpEvent struct {
	EventID        string
	OccurredAt     time.Time
	CompletedAt    time.Time
	Protocol       string
	EntryPoint     string
	Router         string
	Service        string
	Middleware     string
	TenantID       string
	PrincipalType  string
	PrincipalID    string
	AuthnMethod    string
	AuthnResult    string
	PolicyID       string
	AuthzResult    string
	BillingAccount string
	Meter          string
	BillableUnits  float64
	ClientIP       string
	ClientPort     string
	Method         string
	Scheme         string
	Host           string
	Path           string
	Query          string
	StatusCode     int
	DurationUS     int64
	RequestBytes   int64
	ResponseBytes  int64
	TraceID        string
	SpanID         string
	RequestID      string
	UserAgent      string
	Labels         map[string]string
	Metadata       map[string]any
}

type networkEvent struct {
	EventID        string
	OccurredAt     time.Time
	CompletedAt    time.Time
	Protocol       string
	EntryPoint     string
	Router         string
	Service        string
	Middleware     string
	TenantID       string
	PrincipalType  string
	PrincipalID    string
	AuthnMethod    string
	AuthnResult    string
	PolicyID       string
	AuthzResult    string
	BillingAccount string
	Meter          string
	BillableUnits  float64
	ClientAddr     string
	LocalAddr      string
	DurationUS     int64
	BytesReceived  int64
	BytesSent      int64
	Outcome        string
	Labels         map[string]string
	Metadata       map[string]any
}

type task struct {
	database string
	http     *httpEvent
	network  *networkEvent
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
	ctx       context.Context
	cancel    context.CancelFunc
	config    dynamic.AccessGovernance
	queue     chan task
	done      chan struct{}
	wait      time.Duration
	batchSize int
	flush     time.Duration
	logger    zerolog.Logger

	databases   map[string]*databaseTarget
	dropped     atomic.Uint64
	lastDropLog atomic.Int64
	wg          sync.WaitGroup
	stopMu      sync.RWMutex
	stopped     bool
}

type databaseTarget struct {
	config  dynamic.AccessGovernanceDatabase
	manager *runtimeconfig.MySQLManager
}

func getDispatcher(ctx context.Context, name string, config dynamic.AccessGovernance, logger zerolog.Logger) (*dispatcher, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encoding access governance configuration: %w", err)
	}

	key := dispatcherKey{done: ctx.Done(), config: sha256.Sum256(raw)}
	dispatcherRegistry.Lock()
	defer dispatcherRegistry.Unlock()
	if existing := dispatcherRegistry.values[key]; existing != nil {
		return existing, nil
	}

	capacity := config.Queue.Capacity
	if capacity <= 0 {
		capacity = 4096
	}
	workers := config.Queue.Workers
	if workers <= 0 {
		workers = 2
	}
	batchSize := config.Queue.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	} else if batchSize > 1000 {
		batchSize = 1000
	}
	flush := time.Duration(config.Queue.FlushInterval)
	if flush <= 0 {
		flush = time.Second
	}

	dispatcherContext, cancel := context.WithCancel(ctx)
	value := &dispatcher{
		ctx:       dispatcherContext,
		cancel:    cancel,
		config:    config,
		queue:     make(chan task, capacity),
		done:      make(chan struct{}),
		wait:      time.Duration(config.Queue.EnqueueWait),
		batchSize: batchSize,
		flush:     flush,
		logger:    logger.With().Str("component", name+"-writer").Logger(),
		databases: make(map[string]*databaseTarget),
	}
	for databaseName, databaseConfig := range config.Databases {
		httpTable := tableName(databaseConfig.HTTPTable, defaultHTTPTable)
		networkTable := tableName(databaseConfig.NetworkTable, defaultNetworkTable)
		managerKey := runtimeconfig.StableKey(
			"access-governance",
			databaseConfig.ConnectionID,
			httpTable,
			networkTable,
			strconv.Itoa(databaseConfig.MaxOpenConnections),
			strconv.Itoa(databaseConfig.MaxIdleConnections),
			time.Duration(databaseConfig.ConnectionMaxAge).String(),
			strconv.FormatBool(config.Schema.AutoCreate),
		)
		var initializer runtimeconfig.MySQLInitializer
		if config.Schema.AutoCreate {
			initializer = func(schemaContext context.Context, db *sql.DB) error {
				return ensureSchema(schemaContext, db, httpTable, networkTable)
			}
		}
		manager, err := runtimeconfig.GetMySQLManager(dispatcherContext, managerKey, runtimeconfig.MySQLSettings{
			DSN:                databaseConfig.DSN,
			DSNFile:            databaseConfig.DSNFile,
			MaxOpenConnections: databaseConfig.MaxOpenConnections,
			MaxIdleConnections: databaseConfig.MaxIdleConnections,
			ConnectionMaxAge:   time.Duration(databaseConfig.ConnectionMaxAge),
		}, initializer)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("database %q: %w", databaseName, err)
		}
		value.databases[databaseName] = &databaseTarget{
			config:  databaseConfig,
			manager: manager,
		}
	}
	for range workers {
		value.wg.Add(1)
		go value.runWorker()
	}
	go func() {
		<-dispatcherContext.Done()
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
	return value, nil
}

func (d *dispatcher) logDrop(database, protocol string) {
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
		Str("database", database).
		Str("protocol", protocol).
		Uint64("dropped", total).
		Msg("Access governance events dropped because the bounded queue is full")
}

func (d *dispatcher) enqueue(value task) bool {
	d.stopMu.RLock()
	defer d.stopMu.RUnlock()
	if d.stopped {
		return false
	}

	if d.wait <= 0 {
		select {
		case d.queue <- value:
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
		return true
	case <-d.done:
		return false
	case <-timer.C:
		return false
	}
}

func (d *dispatcher) runWorker() {
	defer d.wg.Done()
	ticker := time.NewTicker(d.flush)
	defer ticker.Stop()

	batches := make(map[string][]task)
	flush := func() {
		for database, values := range batches {
			if err := d.writeBatch(database, values); err != nil {
				d.logger.Error().Err(err).Str("database", database).Int("events", len(values)).Msg("Could not persist access governance events")
			}
			delete(batches, database)
		}
	}

	for {
		select {
		case value := <-d.queue:
			batches[value.database] = append(batches[value.database], value)
			if len(batches[value.database]) >= d.batchSize {
				values := batches[value.database]
				delete(batches, value.database)
				if err := d.writeBatch(value.database, values); err != nil {
					d.logger.Error().Err(err).Str("database", value.database).Int("events", len(values)).Msg("Could not persist access governance events")
				}
			}
		case <-ticker.C:
			flush()
		case <-d.done:
			for {
				select {
				case value := <-d.queue:
					batches[value.database] = append(batches[value.database], value)
				default:
					flush()
					return
				}
			}
		}
	}
}

func (d *dispatcher) writeBatch(databaseName string, values []task) error {
	target := d.databases[databaseName]
	if target == nil {
		return fmt.Errorf("database %q does not exist", databaseName)
	}

	db, release, err := target.manager.Acquire(d.ctx)
	if err != nil {
		return err
	}
	defer release()

	var httpValues []httpEvent
	var networkValues []networkEvent
	for _, value := range values {
		if value.http != nil {
			httpValues = append(httpValues, *value.http)
		}
		if value.network != nil {
			networkValues = append(networkValues, *value.network)
		}
	}

	// Dynamic configuration replacement cancels the middleware context before
	// this queue is drained. Preserve its values while giving the final writes
	// a strict deadline so runtime rotations do not lose accepted audit events.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(d.ctx), 10*time.Second)
	defer cancel()
	if len(httpValues) > 0 {
		if err := insertHTTP(ctx, db, tableName(target.config.HTTPTable, defaultHTTPTable), httpValues); err != nil {
			return err
		}
	}
	if len(networkValues) > 0 {
		if err := insertNetwork(ctx, db, tableName(target.config.NetworkTable, defaultNetworkTable), networkValues); err != nil {
			return err
		}
	}
	return nil
}

func insertHTTP(ctx context.Context, db *sql.DB, table string, values []httpEvent) error {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return err
	}
	const columnCount = 38
	placeholders := make([]string, len(values))
	arguments := make([]any, 0, len(values)*columnCount)
	for i, event := range values {
		placeholders[i] = "(" + strings.TrimSuffix(strings.Repeat("?,", columnCount), ",") + ")"
		labels, metadata, err := marshalEventMetadata(event.Labels, event.Metadata)
		if err != nil {
			return err
		}
		arguments = append(arguments,
			event.EventID, event.OccurredAt, event.CompletedAt, "access", event.Protocol,
			event.EntryPoint, event.Router, event.Service, event.Middleware,
			event.TenantID, event.PrincipalType, event.PrincipalID,
			event.AuthnMethod, event.AuthnResult, event.PolicyID, event.AuthzResult,
			event.BillingAccount, event.Meter, event.BillableUnits,
			event.ClientIP, event.ClientPort, event.Method, event.Scheme, event.Host, event.Path, event.Query,
			event.StatusCode, event.DurationUS, event.RequestBytes, event.ResponseBytes,
			event.TraceID, event.SpanID, event.RequestID, event.UserAgent,
			string(labels), string(metadata), event.OccurredAt, event.CompletedAt,
		)
	}

	query := "INSERT INTO " + quotedTable + " (" +
		"event_id,occurred_at,completed_at,event_type,protocol,entry_point,router_name,service_name,middleware_name," +
		"tenant_id,principal_type,principal_id,authn_method,authn_result,policy_id,authz_result," +
		"billing_account,meter,billable_units,client_ip,client_port,request_method,request_scheme,request_host,request_path,request_query," +
		"status_code,duration_us,request_bytes,response_bytes,trace_id,span_id,request_id,user_agent,labels,metadata,usage_window_start,usage_window_end" +
		") VALUES " + strings.Join(placeholders, ",")
	_, err = db.ExecContext(ctx, query, arguments...)
	return err
}

func insertNetwork(ctx context.Context, db *sql.DB, table string, values []networkEvent) error {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return err
	}
	const columnCount = 29
	placeholders := make([]string, len(values))
	arguments := make([]any, 0, len(values)*columnCount)
	for i, event := range values {
		placeholders[i] = "(" + strings.TrimSuffix(strings.Repeat("?,", columnCount), ",") + ")"
		labels, metadata, err := marshalEventMetadata(event.Labels, event.Metadata)
		if err != nil {
			return err
		}
		arguments = append(arguments,
			event.EventID, event.OccurredAt, event.CompletedAt, "connection", event.Protocol,
			event.EntryPoint, event.Router, event.Service, event.Middleware,
			event.TenantID, event.PrincipalType, event.PrincipalID,
			event.AuthnMethod, event.AuthnResult, event.PolicyID, event.AuthzResult,
			event.BillingAccount, event.Meter, event.BillableUnits,
			event.ClientAddr, event.LocalAddr, event.DurationUS, event.BytesReceived, event.BytesSent,
			event.Outcome, string(labels), string(metadata), event.OccurredAt, event.CompletedAt,
		)
	}

	query := "INSERT INTO " + quotedTable + " (" +
		"event_id,occurred_at,completed_at,event_type,protocol,entry_point,router_name,service_name,middleware_name," +
		"tenant_id,principal_type,principal_id,authn_method,authn_result,policy_id,authz_result," +
		"billing_account,meter,billable_units,client_addr,local_addr,duration_us,bytes_received,bytes_sent,outcome,labels,metadata,usage_window_start,usage_window_end" +
		") VALUES " + strings.Join(placeholders, ",")
	_, err = db.ExecContext(ctx, query, arguments...)
	return err
}

func marshalEventMetadata(labels map[string]string, metadata map[string]any) ([]byte, []byte, error) {
	encodedLabels, err := json.Marshal(labels)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding event labels: %w", err)
	}
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding event metadata: %w", err)
	}
	return encodedLabels, encodedMetadata, nil
}

func tableName(configured, fallback string) string {
	if configured == "" {
		return fallback
	}
	return configured
}
