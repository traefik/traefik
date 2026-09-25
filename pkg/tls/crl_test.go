package tls

import (
	"context"
	"crypto/x509"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrlSnapshotCommon_ContainsSerial(t *testing.T) {
	snap := &crlSnapshotCommon{
		revokedSerials: map[string]revocationInfo{"42": {revokedAt: time.Now()}},
	}

	assert.True(t, snap.containsSerial(big.NewInt(42)))
	assert.False(t, snap.containsSerial(big.NewInt(7)))
}

func TestCrlSnapshotCommon_Accessors(t *testing.T) {
	now := time.Now()
	snap := &crlSnapshotCommon{
		number:         big.NewInt(5),
		revokedSerials: map[string]revocationInfo{"1": {}, "2": {}},
		nextUpdate:     now,
	}

	assert.Equal(t, big.NewInt(5), snap.crlSerial())
	assert.Equal(t, now, snap.nextUpdateTime())
	assert.Equal(t, 2, snap.serialCount())
}

func TestStaticCrlSnapshot_NeverNeedsRefresh(t *testing.T) {
	snap := &staticCrlSnapshot{crlSnapshotCommon{nextUpdate: time.Now().Add(-time.Hour)}}
	assert.False(t, snap.needsRefresh(time.Second))
	assert.False(t, snap.needsRefresh(0))
}

func TestDynamicCrlSnapshot_NeedsRefresh(t *testing.T) {
	tests := []struct {
		name     string
		snap     *dynamicCrlSnapshot
		interval time.Duration
		want     bool
	}{
		{
			name: "fresh, nextUpdate in the future",
			snap: &dynamicCrlSnapshot{crlSnapshotCommon{
				modTime: time.Now(), nextUpdate: time.Now().Add(time.Hour),
			}},
			interval: time.Hour, want: false,
		},
		{
			name: "reload interval elapsed",
			snap: &dynamicCrlSnapshot{crlSnapshotCommon{
				modTime: time.Now().Add(-2 * time.Hour), nextUpdate: time.Now().Add(time.Hour),
			}},
			interval: time.Hour, want: true,
		},
		{
			name: "interval zero is ignored",
			snap: &dynamicCrlSnapshot{crlSnapshotCommon{
				modTime: time.Now().Add(-2 * time.Hour), nextUpdate: time.Now().Add(time.Hour),
			}},
			interval: 0, want: false,
		},
		{
			name: "nextUpdate passed",
			snap: &dynamicCrlSnapshot{crlSnapshotCommon{
				modTime: time.Now(), nextUpdate: time.Now().Add(-time.Minute),
			}},
			interval: time.Hour, want: true,
		},
		{
			name: "zero nextUpdate is ignored",
			snap: &dynamicCrlSnapshot{crlSnapshotCommon{
				modTime: time.Now(), nextUpdate: time.Time{},
			}},
			interval: time.Hour, want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.snap.needsRefresh(tt.interval))
		})
	}
}

func TestCrlEntry_SnapshotRoundTrip(t *testing.T) {
	entry := &crlEntry{}
	assert.Nil(t, entry.snapshot())

	snap := &staticCrlSnapshot{crlSnapshotCommon{number: big.NewInt(1)}}
	entry.storeSnapshot(snap)

	assert.Same(t, crlSnapshot(snap), entry.snapshot())
}

func TestCrlEntry_Reload_Success(t *testing.T) {
	entry := &crlEntry{loader: crlNOOPLoader{}}
	require.NoError(t, entry.reload("dp"))
	require.NotNil(t, entry.snapshot())
}

type errorLoader struct{}

func (errorLoader) Load(entry *crlEntry) (crlSnapshot, error) {
	return nil, fmt.Errorf("boom")
}

func TestCrlEntry_Reload_Error(t *testing.T) {
	entry := &crlEntry{loader: errorLoader{}}
	err := entry.reload("dp")
	require.Error(t, err)
	assert.Nil(t, entry.snapshot())
}

func TestCrlEntry_InBackoff_NoBackoffConfigured(t *testing.T) {
	entry := &crlEntry{}
	entry.recordFailure()

	skip, retryAt := entry.inBackoff(0)
	assert.False(t, skip)
	assert.True(t, retryAt.IsZero())
}

func TestCrlEntry_InBackoff_NoFailureRecorded(t *testing.T) {
	entry := &crlEntry{}

	skip, retryAt := entry.inBackoff(time.Minute)
	assert.False(t, skip)
	assert.True(t, retryAt.IsZero())
}

func TestCrlEntry_InBackoff_RecentFailure_SkipsRetry(t *testing.T) {
	entry := &crlEntry{}
	entry.recordFailure()

	skip, retryAt := entry.inBackoff(time.Minute)
	assert.True(t, skip)
	assert.True(t, retryAt.After(time.Now()))
}

func TestCrlEntry_InBackoff_ExpiredFailure_AllowsRetry(t *testing.T) {
	entry := &crlEntry{}
	entry.lastFailureUnixNano.Store(time.Now().Add(-time.Hour).UnixNano())

	skip, retryAt := entry.inBackoff(time.Minute)
	assert.False(t, skip)
	assert.True(t, retryAt.Before(time.Now()))
}

func TestCrlEntry_ClearFailure_ResetsBackoff(t *testing.T) {
	entry := &crlEntry{}
	entry.recordFailure()
	entry.clearFailure()

	skip, _ := entry.inBackoff(time.Minute)
	assert.False(t, skip)
}

func TestCrlEntry_Issuer_DefaultNil(t *testing.T) {
	entry := &crlEntry{}
	assert.Nil(t, entry.issuer())
}

func TestCrlEntry_UpdateIssuer_DetectsChange(t *testing.T) {
	ca1 := newTestCA(t)
	ca2 := newTestCA(t)

	entry := &crlEntry{}

	// First call: no previous issuer recorded, must not be reported as a change.
	assert.False(t, entry.updateIssuer(ca1.cert))
	assert.True(t, entry.issuer().Equal(ca1.cert))

	// Same issuer again: no change.
	assert.False(t, entry.updateIssuer(ca1.cert))

	// Different issuer: change detected.
	assert.True(t, entry.updateIssuer(ca2.cert))
	assert.True(t, entry.issuer().Equal(ca2.cert))

	// Stabilizes: no further change reported once propagated.
	assert.False(t, entry.updateIssuer(ca2.cert))
}

func TestCRLStore_GetOrCreateEntry_ReturnsSameInstance(t *testing.T) {
	store := &CRLStore{}
	e1 := store.getOrCreateEntry("dp1")
	e2 := store.getOrCreateEntry("dp1")
	assert.Same(t, e1, e2)

	e3 := store.getOrCreateEntry("dp2")
	assert.NotSame(t, e1, e3)
}

func TestCRLStore_GetEntry(t *testing.T) {
	store := &CRLStore{}
	_, ok := store.getEntry("missing")
	assert.False(t, ok)

	store.getOrCreateEntry("dp1")
	entry, ok := store.getEntry("dp1")
	assert.True(t, ok)
	assert.NotNil(t, entry)
}

func TestCRLStore_LoadFile_Success(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	path := filepath.Join(t.TempDir(), "test.crl")
	require.NoError(t, os.WriteFile(path, der, 0o600))

	store := &CRLStore{}
	require.NoError(t, store.loadFile("dp1", path))

	entry, ok := store.getEntry("dp1")
	require.True(t, ok)
	require.NotNil(t, entry.snapshot())
	assert.Equal(t, big.NewInt(1), entry.snapshot().crlSerial())
}

func TestCRLStore_LoadFile_MissingFile(t *testing.T) {
	store := &CRLStore{}
	assert.Error(t, store.loadFile("dp1", "/nonexistent/path.crl"))
}

func TestCRLStore_LoadFiles_StopsOnFirstError(t *testing.T) {
	store := &CRLStore{}
	assert.Error(t, store.loadFiles(map[string]string{"dp1": "/nonexistent"}))
}

func TestCRLStore_Reload_RefreshesAllEntries(t *testing.T) {
	store := &CRLStore{}
	entry := store.getOrCreateEntry("dp1")
	entry.loader = crlNOOPLoader{}

	store.reload()

	assert.NotNil(t, entry.snapshot())
}

func TestCRLStore_WatchEntries_StopsOnContextCancel(t *testing.T) {
	store := &CRLStore{crlReloadInterval: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		store.WatchEntries(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("WatchEntries did not stop after context cancellation")
	}
}

func TestCRLStore_WatchEntries_PeriodicReload(t *testing.T) {
	store := &CRLStore{crlReloadInterval: 10 * time.Millisecond}
	entry := store.getOrCreateEntry("dp1")

	var mu sync.Mutex
	count := 0
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		mu.Lock()
		count++
		mu.Unlock()
		return &staticCrlSnapshot{}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go store.WatchEntries(ctx)

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return count >= 2
	}, time.Second, 5*time.Millisecond)
}

func TestCRLStore_WatchFiles_ReloadsOnChange(t *testing.T) {
	ca := newTestCA(t)
	der1 := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	path := filepath.Join(t.TempDir(), "test.crl")
	require.NoError(t, os.WriteFile(path, der1, 0o600))

	store := &CRLStore{}
	require.NoError(t, store.loadFile("dp1", path))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher, err := store.WatchFiles(ctx, map[string]string{"dp1": path})
	require.NoError(t, err)
	go watcher(ctx)

	der2 := newTestCRLDER(t, ca, 2, []x509.RevocationListEntry{
		{SerialNumber: big.NewInt(99), RevocationTime: time.Now()},
	}, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	require.NoError(t, os.WriteFile(path, der2, 0o600))

	require.Eventually(t, func() bool {
		entry, _ := store.getEntry("dp1")
		snap := entry.snapshot()
		return snap != nil && snap.crlSerial().Cmp(big.NewInt(2)) == 0
	}, 2*time.Second, 20*time.Millisecond)
}

func TestParseCRL_DER(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	crl, err := parseCRL(der)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1), crl.Number)
}

func TestParseCRL_PEM(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	crl, err := parseCRL(pemEncodeCRL(der))
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1), crl.Number)
}

func TestParseCRL_InvalidData(t *testing.T) {
	_, err := parseCRL([]byte("not a crl"))
	assert.Error(t, err)
}

func TestCRLStore_GetEntry_MissingKey_DoesNotPanic(t *testing.T) {
	store := &CRLStore{}

	var (
		entry *crlEntry
		ok    bool
	)

	assert.NotPanics(t, func() {
		entry, ok = store.getEntry("unknown-dp")
	})
	assert.False(t, ok)
	assert.Nil(t, entry)
}

func TestCRLStore_GetEntry_MissingKey_OnEmptyStore(t *testing.T) {
	// sync.Map never initialized via getOrCreateEntry: edge case of a brand-new store.
	store := &CRLStore{}

	entry, ok := store.getEntry("dp1")
	assert.False(t, ok)
	assert.Nil(t, entry)
}
