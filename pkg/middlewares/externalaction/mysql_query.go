package externalaction

import (
	"context"
	"database/sql"
	"errors"

	"github.com/traefik/traefik/v3/pkg/middlewares/runtimeconfig"
)

type mysqlPreparer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// mysqlQuery is compiled from a validated table and constant SQL at store creation.
// Execution only accepts bound values; request data cannot change statement text.
type mysqlQuery string

func (q mysqlQuery) prepare(ctx context.Context, db mysqlPreparer) (*sql.Stmt, error) {
	if q == "" {
		return nil, errors.New("MySQL query is not initialized")
	}
	return db.PrepareContext(ctx, string(q))
}

func (q mysqlQuery) exec(ctx context.Context, db mysqlPreparer, args ...any) (sql.Result, error) {
	statement, err := q.prepare(ctx, db)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	return statement.ExecContext(ctx, args...)
}

func (q mysqlQuery) scan(ctx context.Context, db mysqlPreparer, args []any, dest ...any) error {
	statement, err := q.prepare(ctx, db)
	if err != nil {
		return err
	}
	defer statement.Close()
	return statement.QueryRowContext(ctx, args...).Scan(dest...)
}

type mysqlBindingQueries struct {
	insertCredential mysqlQuery
	selectCredential mysqlQuery
	selectBinding    mysqlQuery
	insertBinding    mysqlQuery
	updateBinding    mysqlQuery
	resolveBinding   mysqlQuery
	cleanup          mysqlQuery
}

func compileMySQLBindingQueries(table string) (mysqlBindingQueries, error) {
	quoted, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return mysqlBindingQueries{}, err
	}
	return mysqlBindingQueries{
		insertCredential: mysqlQuery("INSERT INTO " + quoted + " (`key_hash`,`binding_kind`,`binding_key`,`credential_set`,`identity_hash`,`expires_at`) VALUES (?,'credential',?,?,?,NULL) ON DUPLICATE KEY UPDATE `updated_at`=`updated_at`"),
		selectCredential: mysqlQuery("SELECT `identity_hash` FROM " + quoted + " WHERE `key_hash`=? FOR UPDATE"),
		selectBinding:    mysqlQuery("SELECT `credential_set`,`identity_hash`,`expires_at`<=UTC_TIMESTAMP(6) FROM " + quoted + " WHERE `key_hash`=? FOR UPDATE"),
		insertBinding:    mysqlQuery("INSERT INTO " + quoted + " (`key_hash`,`binding_kind`,`binding_key`,`credential_set`,`identity_hash`,`expires_at`) VALUES (?,'lifecycle',?,?,?,DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND))"),
		updateBinding:    mysqlQuery("UPDATE " + quoted + " SET `binding_kind`='lifecycle',`binding_key`=?,`credential_set`=?,`identity_hash`=?,`expires_at`=DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND),`updated_at`=UTC_TIMESTAMP(6) WHERE `key_hash`=?"),
		resolveBinding:   mysqlQuery("SELECT `credential_set`,`identity_hash` FROM " + quoted + " WHERE `key_hash`=? AND `binding_kind`='lifecycle' AND `expires_at`>UTC_TIMESTAMP(6)"),
		cleanup:          mysqlQuery("DELETE FROM " + quoted + " WHERE `binding_kind`='lifecycle' AND `expires_at`<=UTC_TIMESTAMP(6) LIMIT 1000"),
	}, nil
}

type mysqlDedupQueries struct {
	selectClaim mysqlQuery
	insertClaim mysqlQuery
	updateClaim mysqlQuery
	complete    mysqlQuery
	release     mysqlQuery
	cleanup     mysqlQuery
}

func compileMySQLDedupQueries(table string) (mysqlDedupQueries, error) {
	quoted, err := runtimeconfig.QuoteMySQLIdentifier(table)
	if err != nil {
		return mysqlDedupQueries{}, err
	}
	return mysqlDedupQueries{
		selectClaim: mysqlQuery("SELECT `state`,`lease_until`<=UTC_TIMESTAMP(6),`expires_at`<=UTC_TIMESTAMP(6) FROM " + quoted + " WHERE `key_hash`=? FOR UPDATE"),
		insertClaim: mysqlQuery("INSERT INTO " + quoted + " (`key_hash`,`rule_name`,`state`,`claim_id`,`lease_until`,`expires_at`) VALUES (?,?,'claimed',?,DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND),DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND))"),
		updateClaim: mysqlQuery("UPDATE " + quoted + " SET `rule_name`=?,`state`='claimed',`claim_id`=?,`lease_until`=DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND),`expires_at`=DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND),`updated_at`=UTC_TIMESTAMP(6) WHERE `key_hash`=?"),
		complete:    mysqlQuery("UPDATE " + quoted + " SET `state`='completed',`lease_until`=UTC_TIMESTAMP(6),`expires_at`=DATE_ADD(UTC_TIMESTAMP(6), INTERVAL ? MICROSECOND),`updated_at`=UTC_TIMESTAMP(6) WHERE `key_hash`=? AND `claim_id`=? AND `state`='claimed'"),
		release:     mysqlQuery("DELETE FROM " + quoted + " WHERE `key_hash`=? AND `claim_id`=? AND `state`='claimed'"),
		cleanup:     mysqlQuery("DELETE FROM " + quoted + " WHERE `expires_at`<=UTC_TIMESTAMP(6) LIMIT 1000"),
	}, nil
}
