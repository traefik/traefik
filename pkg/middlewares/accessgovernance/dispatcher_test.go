package accessgovernance

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertBatchColumnCounts(t *testing.T) {
	recordingDriverRegisterOnce.Do(func() {
		sql.Register("access-governance-recording", testRecordingDriver)
	})

	db, err := sql.Open("access-governance-recording", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})

	require.NoError(t, insertHTTP(t.Context(), db, "http_events", []httpEvent{{}}))
	query, arguments := testRecordingDriver.snapshot()
	assert.Contains(t, query, "usage_window_start,usage_window_end")
	assert.Len(t, arguments, 38)

	require.NoError(t, insertNetwork(t.Context(), db, "network_events", []networkEvent{{}}))
	query, arguments = testRecordingDriver.snapshot()
	assert.Contains(t, query, "usage_window_start,usage_window_end")
	assert.Len(t, arguments, 29)
	assert.Len(t, arguments, strings.Count(query, "?"))
}

func TestInsertRejectsInvalidMetadata(t *testing.T) {
	t.Parallel()
	metadata := map[string]any{"invalid": make(chan struct{})}
	// Serialization must fail before any SQL is issued, rather than sending invalid JSON.
	require.ErrorContains(t, insertHTTP(t.Context(), nil, "events", []httpEvent{{Metadata: metadata}}), "encoding event metadata")
	require.ErrorContains(t, insertNetwork(t.Context(), nil, "events", []networkEvent{{Metadata: metadata}}), "encoding event metadata")
}

var (
	recordingDriverRegisterOnce sync.Once
	testRecordingDriver         = &recordingDriver{}
)

type recordingDriver struct {
	mu        sync.Mutex
	query     string
	arguments []driver.NamedValue
}

func (d *recordingDriver) Open(string) (driver.Conn, error) {
	return &recordingConnection{driver: d}, nil
}

func (d *recordingDriver) snapshot() (string, []driver.NamedValue) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.query, append([]driver.NamedValue(nil), d.arguments...)
}

type recordingConnection struct {
	driver *recordingDriver
}

func (c *recordingConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not implemented")
}

func (c *recordingConnection) Close() error {
	return nil
}

func (c *recordingConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not implemented")
}

func (c *recordingConnection) ExecContext(_ context.Context, query string, arguments []driver.NamedValue) (driver.Result, error) {
	c.driver.mu.Lock()
	defer c.driver.mu.Unlock()
	c.driver.query = query
	c.driver.arguments = append([]driver.NamedValue(nil), arguments...)
	return driver.RowsAffected(1), nil
}
