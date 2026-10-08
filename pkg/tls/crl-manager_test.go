package tls

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	ptypes "github.com/traefik/paerser/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) *CRLManager {
	t.Helper()
	m, err := NewCRLManager(context.Background(), CRLManagerConfig{
		ReloadInterval:     time.Minute,
		DefaultHTTPTimeout: time.Second,
	})
	require.NoError(t, err)
	return m
}

func TestNewCRLManager_NoFileCRLs_CreatesEmptyGlobalStore(t *testing.T) {
	m := newTestManager(t)
	require.NotNil(t, m.global)
}

func TestNewCRLManager_WithFileCRLs_Success(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	path := filepath.Join(t.TempDir(), "test.crl")
	require.NoError(t, os.WriteFile(path, der, 0o600))

	m, err := NewCRLManager(context.Background(), CRLManagerConfig{
		FileCRLs:       map[string]string{"http://dp1": path},
		ReloadInterval: time.Minute,
	})
	require.NoError(t, err)

	entry, ok := m.global.getEntry("http://dp1")
	require.True(t, ok)
	require.NotNil(t, entry.snapshot())
}

func TestNewCRLManager_WithFileCRLs_InvalidPath_ReturnsError(t *testing.T) {
	_, err := NewCRLManager(context.Background(), CRLManagerConfig{
		FileCRLs: map[string]string{"http://dp1": "/does/not/exist"},
	})
	assert.Error(t, err)
}

func TestCRLManager_GetEnforcer_NOOPMode(t *testing.T) {
	m := newTestManager(t)
	enforcer, err := m.GetEnforcer("options1", CRL{Mode: CRLNOOP})
	require.NoError(t, err)
	_, isNOOP := enforcer.(*crlEnforcerNOOP)
	assert.True(t, isNOOP)
}

func TestCRLManager_GetEnforcer_EmptyMode_DefaultsToNOOP(t *testing.T) {
	m := newTestManager(t)
	enforcer, err := m.GetEnforcer("options1", CRL{})
	require.NoError(t, err)
	_, isNOOP := enforcer.(*crlEnforcerNOOP)
	assert.True(t, isNOOP)
}

func TestCRLManager_GetEnforcer_UnknownExpirationStrategy_ReturnsError(t *testing.T) {
	m := newTestManager(t)
	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{ExpirationStrategy: "bogus"}})
	assert.Error(t, err)
}

func TestCRLManager_GetEnforcer_CreatesStateOnFirstCall(t *testing.T) {
	m := newTestManager(t)
	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax})
	require.NoError(t, err)

	m.mu.RLock()
	_, ok := m.states["options1"]
	m.mu.RUnlock()
	assert.True(t, ok)
}

func TestCRLManager_GetEnforcer_CachesEnforcerWhenConfigUnchanged(t *testing.T) {
	m := newTestManager(t)
	cfg := CRL{Mode: CRLLax}

	e1, err := m.GetEnforcer("options1", cfg)
	require.NoError(t, err)
	e2, err := m.GetEnforcer("options1", cfg)
	require.NoError(t, err)

	assert.Same(t, e1, e2)
}

func TestCRLManager_GetEnforcer_RebuildsEnforcerOnPolicyChange_StorePreserved(t *testing.T) {
	m := newTestManager(t)

	e1, err := m.GetEnforcer("options1", CRL{Mode: CRLLax})
	require.NoError(t, err)

	m.mu.RLock()
	store := m.states["options1"].store
	m.mu.RUnlock()

	e2, err := m.GetEnforcer("options1", CRL{Mode: CRLStrict})
	require.NoError(t, err)

	assert.NotSame(t, e1, e2)

	m.mu.RLock()
	storeAfter := m.states["options1"].store
	m.mu.RUnlock()

	// The cache (store) must survive a pure policy change.
	assert.Same(t, store, storeAfter)
}

func TestCRLManager_GetEnforcer_ClientUnchanged_WhenTimeoutUnchanged(t *testing.T) {
	m := newTestManager(t)
	timeout := ptypes.Duration(time.Second)

	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{Timeout: timeout}})
	require.NoError(t, err)

	m.mu.RLock()
	client1 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	// Only the mode changes; Timeout is identical.
	_, err = m.GetEnforcer("options1", CRL{Mode: CRLStrict, HTTP: CRLHTTP{Timeout: timeout}})
	require.NoError(t, err)

	m.mu.RLock()
	client2 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	assert.Same(t, client1, client2)
}

func TestCRLManager_GetEnforcer_ClientSwapped_WhenTimeoutChanges(t *testing.T) {
	m := newTestManager(t)

	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{Timeout: ptypes.Duration(time.Second)}})
	require.NoError(t, err)

	m.mu.RLock()
	client1 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	_, err = m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{Timeout: ptypes.Duration(5 * time.Second)}})
	require.NoError(t, err)

	m.mu.RLock()
	client2 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	assert.NotSame(t, client1, client2)
	assert.Equal(t, 5*time.Second, client2.client.Timeout)
}

func TestCRLManager_Prune_RemovesInactiveStates(t *testing.T) {
	m := newTestManager(t)

	_, err := m.GetEnforcer("keep", CRL{Mode: CRLLax})
	require.NoError(t, err)
	_, err = m.GetEnforcer("remove", CRL{Mode: CRLLax})
	require.NoError(t, err)

	m.Prune(map[string]Options{"keep": {}})

	m.mu.RLock()
	_, keepOK := m.states["keep"]
	_, removeOK := m.states["remove"]
	m.mu.RUnlock()

	assert.True(t, keepOK)
	assert.False(t, removeOK)
}

func TestHashCRLConfig_Deterministic(t *testing.T) {
	cfg := CRL{Mode: CRLLax}

	h1, err := hashCRLConfig(cfg)
	require.NoError(t, err)
	h2, err := hashCRLConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, h1, h2)

	h3, err := hashCRLConfig(CRL{Mode: CRLStrict})
	require.NoError(t, err)
	assert.NotEqual(t, h1, h3)
}

func TestHashClientConfig_Deterministic(t *testing.T) {
	h1 := hashClientConfig(CRLHTTP{Timeout: ptypes.Duration(time.Second)})
	h2 := hashClientConfig(CRLHTTP{Timeout: ptypes.Duration(time.Second)})
	assert.Equal(t, h1, h2)

	h3 := hashClientConfig(CRLHTTP{Timeout: ptypes.Duration(2 * time.Second)})
	assert.NotEqual(t, h1, h3)
}

func TestCRLManager_GetEnforcer_UpdatesReloadIntervalInPlace(t *testing.T) {
	m := newTestManager(t)

	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax, ReloadInterval: ptypes.Duration(time.Minute)})
	require.NoError(t, err)

	m.mu.RLock()
	store := m.states["options1"].store
	enforcerBefore := m.states["options1"].enforcer
	clientBefore := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	assert.Equal(t, time.Minute, *store.crlReloadInterval.Load())

	// Seed the store with an entry to make sure the cache survives the change.
	store.getOrCreateEntry("dp1").storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{number: big.NewInt(1)}})

	_, err = m.GetEnforcer("options1", CRL{Mode: CRLLax, ReloadInterval: ptypes.Duration(5 * time.Minute)})
	require.NoError(t, err)

	m.mu.RLock()
	storeAfter := m.states["options1"].store
	enforcerAfter := m.states["options1"].enforcer
	clientAfter := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	// Same store instance (cache preserved)...
	assert.Same(t, store, storeAfter)
	// ...but with the new reload interval applied.
	assert.Equal(t, 5*time.Minute, *storeAfter.crlReloadInterval.Load())
	// HTTP client untouched since only ReloadInterval changed.
	assert.Same(t, clientBefore, clientAfter)
	// Enforcer is rebuilt (policy fingerprint changed) but that's cheap and expected.
	assert.NotSame(t, enforcerBefore, enforcerAfter)

	entry, ok := storeAfter.getEntry("dp1")
	require.True(t, ok)
	assert.NotNil(t, entry.snapshot(), "cached CRL entries must survive a reloadInterval-only change")
}

func TestCRLManager_GetEnforcer_UpdatesErrorBackoffInPlace(t *testing.T) {
	m := newTestManager(t)

	_, err := m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{ErrorBackoff: ptypes.Duration(10 * time.Second)}})
	require.NoError(t, err)

	m.mu.RLock()
	store := m.states["options1"].store
	m.mu.RUnlock()
	assert.Equal(t, 10*time.Second, *store.crlErrorBackoff.Load())

	_, err = m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{ErrorBackoff: ptypes.Duration(time.Minute)}})
	require.NoError(t, err)

	assert.Equal(t, time.Minute, *store.crlErrorBackoff.Load())
}

func TestCRLManager_GetEnforcer_ErrorBackoffFallsBackToDefaultWhenUnset(t *testing.T) {
	m, err := NewCRLManager(context.Background(), CRLManagerConfig{
		ReloadInterval:      time.Minute,
		DefaultHTTPTimeout:  time.Second,
		DefaultErrorBackoff: 42 * time.Second,
	})
	require.NoError(t, err)

	_, err = m.GetEnforcer("options1", CRL{Mode: CRLLax, HTTP: CRLHTTP{ErrorBackoff: ptypes.Duration(10 * time.Second)}})
	require.NoError(t, err)

	m.mu.RLock()
	store := m.states["options1"].store
	m.mu.RUnlock()
	assert.Equal(t, 10*time.Second, *store.crlErrorBackoff.Load())

	// Explicit value removed: must fall back to the manager-level default again,
	// not keep the previously applied value indefinitely.
	_, err = m.GetEnforcer("options1", CRL{Mode: CRLLax})
	require.NoError(t, err)
	assert.Equal(t, 42*time.Second, *store.crlErrorBackoff.Load())
}

func TestCRLManager_GetEnforcer_ReloadIntervalChange_DoesNotSwapClient(t *testing.T) {
	m := newTestManager(t)

	_, err := m.GetEnforcer("options1", CRL{
		Mode:           CRLLax,
		ReloadInterval: ptypes.Duration(time.Minute),
		HTTP:           CRLHTTP{Timeout: ptypes.Duration(time.Second)},
	})
	require.NoError(t, err)

	m.mu.RLock()
	client1 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	_, err = m.GetEnforcer("options1", CRL{
		Mode:           CRLLax,
		ReloadInterval: ptypes.Duration(10 * time.Minute),
		HTTP:           CRLHTTP{Timeout: ptypes.Duration(time.Second)},
	})
	require.NoError(t, err)

	m.mu.RLock()
	client2 := m.states["options1"].clientHolder.get()
	m.mu.RUnlock()

	assert.Same(t, client1, client2, "changing only ReloadInterval must not trigger an HTTP client swap")
}
