package externalaction

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The driver rejects direct execution so these tests also check that values
// cross the prepared-statement protocol boundary, not a client interpolator.
type queryRecorder struct {
	query        string
	values       []driver.NamedValue
	closed       int
	rows         [][]driver.Value
	prepareErr   error
	executionErr error
}

func (r *queryRecorder) Connect(context.Context) (driver.Conn, error) {
	return &recordingConnection{r}, nil
}
func (r *queryRecorder) Driver() driver.Driver { return recordingDriver{r} }

type recordingDriver struct{ recorder *queryRecorder }

func (d recordingDriver) Open(string) (driver.Conn, error) {
	return &recordingConnection{d.recorder}, nil
}

type recordingConnection struct{ recorder *queryRecorder }

func (c *recordingConnection) Prepare(query string) (driver.Stmt, error) {
	c.recorder.query = query
	if c.recorder.prepareErr != nil {
		return nil, c.recorder.prepareErr
	}
	return &recordingStatement{c.recorder}, nil
}

func (c *recordingConnection) Close() error              { return nil }
func (c *recordingConnection) Begin() (driver.Tx, error) { return recordingTransaction{}, nil }

func (c *recordingConnection) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("direct execution is forbidden")
}

func (c *recordingConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("direct execution is forbidden")
}

type recordingTransaction struct{}

func (recordingTransaction) Commit() error   { return nil }
func (recordingTransaction) Rollback() error { return nil }

type recordingStatement struct{ recorder *queryRecorder }

func (s *recordingStatement) Close() error  { s.recorder.closed++; return nil }
func (s *recordingStatement) NumInput() int { return strings.Count(s.recorder.query, "?") }
func (s *recordingStatement) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("context required")
}

func (s *recordingStatement) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("context required")
}

func (s *recordingStatement) ExecContext(_ context.Context, values []driver.NamedValue) (driver.Result, error) {
	s.recorder.values = append([]driver.NamedValue(nil), values...)
	return driver.RowsAffected(1), s.recorder.executionErr
}

func (s *recordingStatement) QueryContext(_ context.Context, values []driver.NamedValue) (driver.Rows, error) {
	s.recorder.values = append([]driver.NamedValue(nil), values...)
	if s.recorder.executionErr != nil {
		return nil, s.recorder.executionErr
	}
	return &recordingRows{rows: s.recorder.rows}, nil
}

type recordingRows struct{ rows [][]driver.Value }

func (r *recordingRows) Columns() []string { return []string{"value"} }
func (r *recordingRows) Close() error      { return nil }
func (r *recordingRows) Next(values []driver.Value) error {
	if len(r.rows) == 0 {
		return io.EOF
	}
	copy(values, r.rows[0])
	r.rows = r.rows[1:]
	return nil
}

func TestMySQLQueriesUseBoundValues(t *testing.T) {
	binding, err := compileMySQLBindingQueries("lifecycle_events")
	require.NoError(t, err)
	dedup, err := compileMySQLDedupQueries("dedup_events")
	require.NoError(t, err)
	queries := map[string]mysqlQuery{
		"insertCredential": binding.insertCredential, "selectCredential": binding.selectCredential,
		"selectBinding": binding.selectBinding, "insertBinding": binding.insertBinding,
		"updateBinding": binding.updateBinding, "resolveBinding": binding.resolveBinding,
		"cleanupBindings": binding.cleanup, "selectClaim": dedup.selectClaim,
		"insertClaim": dedup.insertClaim, "updateClaim": dedup.updateClaim,
		"complete": dedup.complete, "release": dedup.release, "cleanupClaims": dedup.cleanup,
	}
	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			recorder := &queryRecorder{rows: [][]driver.Value{{"literal"}}}
			db := sql.OpenDB(recorder)
			defer db.Close()
			tx, err := db.BeginTx(t.Context(), nil)
			require.NoError(t, err)
			defer func() { _ = tx.Rollback() }()
			const payload = "flow'); DROP TABLE lifecycle_events; -- \\ ' \x00"
			values := make([]any, strings.Count(string(query), "?"))
			for index := range values {
				values[index] = payload
			}
			if strings.HasPrefix(string(query), "SELECT") {
				var result string
				require.NoError(t, query.scan(t.Context(), tx, values, &result))
				require.Equal(t, "literal", result)
			} else {
				_, err = query.exec(t.Context(), tx, values...)
				require.NoError(t, err)
			}
			require.Equal(t, string(query), recorder.query)
			require.NotContains(t, recorder.query, payload)
			require.Len(t, recorder.values, len(values))
			for index, value := range recorder.values {
				require.Equal(t, index+1, value.Ordinal)
				require.Equal(t, payload, value.Value)
			}
			require.Equal(t, 1, recorder.closed)
		})
	}
}

func TestMySQLQueryClosesStatementsOnFailure(t *testing.T) {
	for _, scan := range []bool{false, true} {
		t.Run(map[bool]string{false: "exec", true: "scan"}[scan], func(t *testing.T) {
			sentinel := errors.New("database unavailable")
			recorder := &queryRecorder{executionErr: sentinel}
			db := sql.OpenDB(recorder)
			defer db.Close()
			query := mysqlQuery("SELECT ?")
			var err error
			if scan {
				var value string
				err = query.scan(t.Context(), db, []any{"value"}, &value)
			} else {
				_, err = query.exec(t.Context(), db, "value")
			}
			require.ErrorIs(t, err, sentinel)
			require.Equal(t, 1, recorder.closed)
		})
	}
	recorder := &queryRecorder{}
	db := sql.OpenDB(recorder)
	defer db.Close()
	var value string
	require.ErrorIs(t, mysqlQuery("SELECT ?").scan(t.Context(), db, []any{"value"}, &value), sql.ErrNoRows)
	require.Equal(t, 1, recorder.closed)
	recorder.prepareErr = errors.New("prepare failed")
	_, err := mysqlQuery("SELECT ?").exec(t.Context(), db, "value")
	require.ErrorIs(t, err, recorder.prepareErr)
	require.Equal(t, 1, recorder.closed)
	_, err = mysqlQuery("").exec(t.Context(), nil)
	require.ErrorContains(t, err, "not initialized")
}
