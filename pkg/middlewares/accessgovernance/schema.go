package accessgovernance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

const (
	columnOccurredAt     = "occurred_at"
	columnRouterName     = "router_name"
	columnEventID        = "event_id"
	columnServiceName    = "service_name"
	columnBillingAccount = "billing_account"
	columnMeter          = "meter"
	sqlRequiredTimestamp = "DATETIME(6) NOT NULL"
	sqlString128         = "VARCHAR(128) NOT NULL DEFAULT ''"
	sqlString255         = "VARCHAR(255) NOT NULL DEFAULT ''"
	sqlEvaluationResult  = "VARCHAR(32) NOT NULL DEFAULT 'not_evaluated'"
	sqlUnsignedCounter   = "BIGINT UNSIGNED NOT NULL DEFAULT 0"
	sqlOptionalJSON      = "JSON NULL"
	sqlOptionalTimestamp = "DATETIME(6) NULL"
)

var httpColumns = map[string]string{
	columnEventID:        "CHAR(32) NOT NULL",
	columnOccurredAt:     sqlRequiredTimestamp,
	"completed_at":       sqlRequiredTimestamp,
	"event_type":         "VARCHAR(32) NOT NULL DEFAULT 'access'",
	"protocol":           "VARCHAR(16) NOT NULL",
	"entry_point":        sqlString128,
	columnRouterName:     sqlString255,
	columnServiceName:    sqlString255,
	"middleware_name":    sqlString255,
	"tenant_id":          sqlString128,
	"principal_type":     "VARCHAR(32) NOT NULL DEFAULT ''",
	"principal_id":       sqlString255,
	"authn_method":       "VARCHAR(64) NOT NULL DEFAULT ''",
	"authn_result":       sqlEvaluationResult,
	"policy_id":          sqlString255,
	"authz_result":       sqlEvaluationResult,
	columnBillingAccount: sqlString128,
	columnMeter:          sqlString128,
	"billable_units":     "DECIMAL(20,6) NOT NULL DEFAULT 0",
	"client_ip":          "VARCHAR(64) NOT NULL DEFAULT ''",
	"client_port":        "VARCHAR(16) NOT NULL DEFAULT ''",
	"request_method":     "VARCHAR(16) NOT NULL DEFAULT ''",
	"request_scheme":     "VARCHAR(16) NOT NULL DEFAULT ''",
	"request_host":       sqlString255,
	"request_path":       "VARCHAR(2048) NOT NULL DEFAULT ''",
	"request_query":      "TEXT NULL",
	"status_code":        "SMALLINT UNSIGNED NOT NULL DEFAULT 0",
	"duration_us":        sqlUnsignedCounter,
	"request_bytes":      sqlUnsignedCounter,
	"response_bytes":     sqlUnsignedCounter,
	"trace_id":           "CHAR(32) NOT NULL DEFAULT ''",
	"span_id":            "CHAR(16) NOT NULL DEFAULT ''",
	"request_id":         sqlString128,
	"user_agent":         "VARCHAR(1024) NOT NULL DEFAULT ''",
	"labels":             sqlOptionalJSON,
	"metadata":           sqlOptionalJSON,
	"usage_window_start": sqlOptionalTimestamp,
	"usage_window_end":   sqlOptionalTimestamp,
	"created_at":         "TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)",
}

var networkColumns = map[string]string{
	columnEventID:        "CHAR(32) NOT NULL",
	columnOccurredAt:     sqlRequiredTimestamp,
	"completed_at":       sqlRequiredTimestamp,
	"event_type":         "VARCHAR(32) NOT NULL DEFAULT 'connection'",
	"protocol":           "VARCHAR(16) NOT NULL",
	"entry_point":        sqlString128,
	columnRouterName:     sqlString255,
	columnServiceName:    sqlString255,
	"middleware_name":    sqlString255,
	"tenant_id":          sqlString128,
	"principal_type":     "VARCHAR(32) NOT NULL DEFAULT ''",
	"principal_id":       sqlString255,
	"authn_method":       "VARCHAR(64) NOT NULL DEFAULT ''",
	"authn_result":       sqlEvaluationResult,
	"policy_id":          sqlString255,
	"authz_result":       sqlEvaluationResult,
	columnBillingAccount: sqlString128,
	columnMeter:          sqlString128,
	"billable_units":     "DECIMAL(20,6) NOT NULL DEFAULT 0",
	"client_addr":        sqlString255,
	"local_addr":         sqlString255,
	"duration_us":        sqlUnsignedCounter,
	"bytes_received":     sqlUnsignedCounter,
	"bytes_sent":         sqlUnsignedCounter,
	"outcome":            "VARCHAR(32) NOT NULL DEFAULT ''",
	"labels":             sqlOptionalJSON,
	"metadata":           sqlOptionalJSON,
	"usage_window_start": sqlOptionalTimestamp,
	"usage_window_end":   sqlOptionalTimestamp,
	"created_at":         "TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)",
}

var httpIndexes = map[string][]string{
	"uq_event_id":            {columnEventID},
	"idx_occurred_at":        {columnOccurredAt},
	"idx_principal_occurred": {"principal_id", columnOccurredAt},
	"idx_router_occurred":    {columnRouterName, columnOccurredAt},
	"idx_service_occurred":   {columnServiceName, columnOccurredAt},
	"idx_billing_occurred":   {columnBillingAccount, columnMeter, columnOccurredAt},
	"idx_trace_id":           {"trace_id"},
}

var networkIndexes = map[string][]string{
	"uq_event_id":          {columnEventID},
	"idx_occurred_at":      {columnOccurredAt},
	"idx_router_occurred":  {columnRouterName, columnOccurredAt},
	"idx_service_occurred": {columnServiceName, columnOccurredAt},
	"idx_client_occurred":  {"client_addr", columnOccurredAt},
	"idx_billing_occurred": {columnBillingAccount, columnMeter, columnOccurredAt},
}

func ensureSchema(ctx context.Context, db *sql.DB, httpTable, networkTable string) error {
	if !identifierPattern.MatchString(httpTable) || !identifierPattern.MatchString(networkTable) {
		return fmt.Errorf("access governance table names must match %s", identifierPattern)
	}
	if err := ensureTable(ctx, db, httpTable, httpColumns, httpIndexes); err != nil {
		return fmt.Errorf("ensuring HTTP table: %w", err)
	}
	if err := ensureTable(ctx, db, networkTable, networkColumns, networkIndexes); err != nil {
		return fmt.Errorf("ensuring network table: %w", err)
	}
	return nil
}

func ensureTable(ctx context.Context, db *sql.DB, table string, columns map[string]string, indexes map[string][]string) error {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return err
	}
	var definitions []string
	definitions = append(definitions, "`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT")
	for _, name := range sortedKeys(columns) {
		definitions = append(definitions, "`"+name+"` "+columns[name])
	}
	definitions = append(definitions, "PRIMARY KEY (`id`)")
	for _, name := range sortedKeys(indexes) {
		kind := "KEY"
		if strings.HasPrefix(name, "uq_") {
			kind = "UNIQUE KEY"
		}
		definitions = append(definitions, kind+" `"+name+"` ("+quotedColumns(indexes[name])+")")
	}
	query := "CREATE TABLE IF NOT EXISTS " + quotedTable + " (" + strings.Join(definitions, ",") + ") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}

	existingColumns, err := readColumns(ctx, db, table)
	if err != nil {
		return err
	}
	for _, name := range sortedKeys(columns) {
		if _, ok := existingColumns[name]; ok {
			continue
		}
		if _, err := db.ExecContext(ctx, "ALTER TABLE "+quotedTable+" ADD COLUMN `"+name+"` "+columns[name]); err != nil {
			return fmt.Errorf("adding column %s: %w", name, err)
		}
	}

	existingIndexes, err := readIndexes(ctx, db, table)
	if err != nil {
		return err
	}
	for _, name := range sortedKeys(indexes) {
		if _, ok := existingIndexes[name]; ok {
			continue
		}
		kind := "INDEX"
		if strings.HasPrefix(name, "uq_") {
			kind = "UNIQUE INDEX"
		}
		if _, err := db.ExecContext(ctx, "ALTER TABLE "+quotedTable+" ADD "+kind+" `"+name+"` ("+quotedColumns(indexes[name])+")"); err != nil {
			return fmt.Errorf("adding index %s: %w", name, err)
		}
	}
	return nil
}

func readColumns(ctx context.Context, db *sql.DB, table string) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result[name] = struct{}{}
	}
	return result, rows.Err()
}

func readIndexes(ctx context.Context, db *sql.DB, table string) (map[string]struct{}, error) {
	quotedTable, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "SHOW INDEX FROM "+quotedTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	keyIndex := slices.Index(columns, "Key_name")
	if keyIndex < 0 {
		return nil, errors.New("SHOW INDEX did not return Key_name")
	}

	result := make(map[string]struct{})
	values := make([]any, len(columns))
	raw := make([]sql.RawBytes, len(columns))
	for i := range values {
		values[i] = &raw[i]
	}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		result[string(raw[keyIndex])] = struct{}{}
	}
	return result, rows.Err()
}

func sortedKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}

func quotedColumns(values []string) string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = "`" + value + "`"
	}
	return strings.Join(result, ",")
}
