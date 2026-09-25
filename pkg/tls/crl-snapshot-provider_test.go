package tls

import (
	"math/big"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenSnapshotProvider_NewEntry_Success(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1), snap.crlSerial())
}

func TestOpenSnapshotProvider_NewEntry_LoadFailure(t *testing.T) {
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	assert.Error(t, err)
}

func TestOpenSnapshotProvider_FreshSnapshot_NoReload(t *testing.T) {
	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry("dp1")
	fresh := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(42), modTime: time.Now(), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(fresh)
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be triggered for a fresh snapshot")
		return nil, nil
	})

	snap, err := provider.getVerifiedSnapshot(store, "dp1", ca.cert)
	require.NoError(t, err)
	assert.Same(t, crlSnapshot(fresh), snap)
}

func TestOpenSnapshotProvider_StaleSnapshot_Refreshes(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 2, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	store := &CRLStore{crlReloadInterval: time.Millisecond}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2), snap.crlSerial())
}

func TestOpenSnapshotProvider_ConcurrentRefresh_ReturnsStale(t *testing.T) {
	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Millisecond}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry("dp1")
	stale := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(stale)

	// Simulate an in-flight refresh held by another goroutine.
	entry.refreshMu.Lock()
	defer entry.refreshMu.Unlock()

	snap, err := provider.getVerifiedSnapshot(store, "dp1", ca.cert)
	require.NoError(t, err)
	assert.Same(t, crlSnapshot(stale), snap)
}

func TestOpenSnapshotProvider_NewEntry_BackoffSkipsSubsequentRetries(t *testing.T) {
	var reqCount int32
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: time.Minute}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	// Within the backoff window: must not hit the server again.
	_, err = provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry not attempted")
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))
}

func TestOpenSnapshotProvider_NewEntry_BackoffExpired_RetrySucceeds(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 5, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	var fail int32 = 1
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&fail) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(der)
	})

	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: 10 * time.Millisecond}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)

	time.Sleep(20 * time.Millisecond) // let the backoff window expire
	atomic.StoreInt32(&fail, 0)

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(5), snap.crlSerial())
}

func TestOpenSnapshotProvider_NewEntry_NoBackoffConfigured_AlwaysRetries(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 3, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	var fail int32 = 1
	var reqCount int32
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		if atomic.LoadInt32(&fail) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(der)
	})

	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: 0}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	atomic.StoreInt32(&fail, 0)

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3), snap.crlSerial())
	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCount),
		"with backoff disabled every call should attempt a reload")
}

func TestOpenSnapshotProvider_StaleSnapshot_BackoffActive_ReturnsStaleWithoutReload(t *testing.T) {
	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Millisecond, crlErrorBackoff: time.Minute}
	provider := &openSnaphotProvider{}

	entry := store.getOrCreateEntry("dp1")
	stale := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(stale)
	entry.recordFailure()
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be attempted while backoff is active")
		return nil, nil
	})

	snap, err := provider.getVerifiedSnapshot(store, "dp1", ca.cert)
	require.NoError(t, err)
	assert.Same(t, crlSnapshot(stale), snap)
}

func TestOpenSnapshotProvider_StaleSnapshot_BackoffExpired_AttemptsReload(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 7, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	store := &CRLStore{crlReloadInterval: time.Millisecond, crlErrorBackoff: time.Minute}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})
	// Simulate a previous failure whose backoff window has already elapsed.
	entry.lastFailureUnixNano.Store(time.Now().Add(-time.Hour).UnixNano())

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(7), snap.crlSerial())
}

func TestFailedCloseProvider_FreshSnapshot_FastPath(t *testing.T) {
	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry("dp1")
	fresh := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now(), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(fresh)
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be triggered for a fresh snapshot")
		return nil, nil
	})

	snap, err := provider.getVerifiedSnapshot(store, "dp1", ca.cert)
	require.NoError(t, err)
	assert.Same(t, crlSnapshot(fresh), snap)
}

func TestFailedCloseProvider_StaleSnapshot_BlockingRefreshSuccess(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 2, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	store := &CRLStore{crlReloadInterval: time.Millisecond}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2), snap.crlSerial())
}

func TestFailedCloseProvider_RefreshFailure_FailsClosedEvenWithStaleSnapshot(t *testing.T) {
	ca := newTestCA(t)
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	store := &CRLStore{crlReloadInterval: time.Millisecond}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	assert.Error(t, err) // fail-closed: stale data must never be returned silently
}

func TestFailedCloseProvider_NoSnapshot_LoadFailure(t *testing.T) {
	ca := newTestCA(t)
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	assert.Error(t, err)
}

func TestFailedCloseProvider_NewEntry_BackoffSkipsSubsequentRetries(t *testing.T) {
	var reqCount int32
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: time.Minute}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	_, err = provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry not attempted")
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))
}

func TestFailedCloseProvider_StaleSnapshot_BackoffActive_FailsClosedWithoutReload(t *testing.T) {
	ca := newTestCA(t)
	store := &CRLStore{crlReloadInterval: time.Millisecond, crlErrorBackoff: time.Minute}
	provider := &failedCloseSnapshotProvider{}

	entry := store.getOrCreateEntry("dp1")
	stale := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(stale)
	entry.recordFailure()
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be attempted while backoff is active")
		return nil, nil
	})

	_, err := provider.getVerifiedSnapshot(store, "dp1", ca.cert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry not attempted")
}

func TestFailedCloseProvider_StaleSnapshot_BackoffExpired_AttemptsReload(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 9, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(der) })

	store := &CRLStore{crlReloadInterval: time.Millisecond, crlErrorBackoff: time.Minute}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})
	entry.lastFailureUnixNano.Store(time.Now().Add(-time.Hour).UnixNano())

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(9), snap.crlSerial())
}

func TestFailedCloseProvider_SuccessfulReload_ClearsBackoff(t *testing.T) {
	ca := newTestCA(t)
	der := newTestCRLDER(t, ca, 11, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))

	var fail int32 = 1
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&fail) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(der)
	})

	store := &CRLStore{crlReloadInterval: time.Millisecond, crlErrorBackoff: 10 * time.Millisecond}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	entry := store.getOrCreateEntry(srv.URL)
	entry.loader = &crlHTTPLoader{
		distributionPoint: srv.URL,
		clientHolder:      provider.clientHolder,
	}
	entry.storeSnapshot(&dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now().Add(-time.Hour), nextUpdate: time.Now().Add(time.Hour),
	}})

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.Error(t, err)

	skip, _ := entry.inBackoff(store.crlErrorBackoff)
	assert.True(t, skip, "backoff should be active right after a failure")

	time.Sleep(20 * time.Millisecond)
	atomic.StoreInt32(&fail, 0)

	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca.cert)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(11), snap.crlSerial())

	skip, _ = entry.inBackoff(store.crlErrorBackoff)
	assert.False(t, skip, "backoff should be cleared after a successful reload")
}

// --- Issuer rotation: openSnaphotProvider ---

func TestOpenSnapshotProvider_IssuerRotation_TriggersReload(t *testing.T) {
	ca1 := newTestCA(t)
	ca2 := newTestCA(t)

	var reqCount int32
	activeCA := ca1
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		der := newTestCRLDER(t, activeCA, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
		_, _ = w.Write(der)
	})

	// Long reload interval: any further reload can only be explained by the
	// issuer rotation, not by time-based staleness.
	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &openSnaphotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca1.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	// Same issuer, fresh snapshot: no new request expected.
	_, err = provider.getVerifiedSnapshot(store, srv.URL, ca1.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	// Issuer rotation: must force a reload even though the snapshot is still
	// fresh with regard to crlReloadInterval.
	activeCA = ca2
	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca2.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCount))
	assert.Equal(t, big.NewInt(1), snap.crlSerial())

	entry, ok := store.getEntry(srv.URL)
	require.True(t, ok)
	assert.True(t, entry.issuer().Equal(ca2.cert))
}

func TestOpenSnapshotProvider_IssuerRotation_BackoffActive_ReturnsStale(t *testing.T) {
	ca1 := newTestCA(t)
	ca2 := newTestCA(t)

	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: time.Minute}
	provider := &openSnaphotProvider{}

	entry := store.getOrCreateEntry("dp1")
	stale := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now(), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(stale)
	entry.updateIssuer(ca1.cert)
	entry.recordFailure()
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be attempted while backoff is active")
		return nil, nil
	})

	// Issuer rotates while the distribution point is in backoff: the stale
	// snapshot is still served instead of failing the handshake outright.
	snap, err := provider.getVerifiedSnapshot(store, "dp1", ca2.cert)
	require.NoError(t, err)
	assert.Same(t, crlSnapshot(stale), snap)
}

// --- Issuer rotation: failedCloseSnapshotProvider ---

func TestFailedCloseProvider_IssuerRotation_TriggersBlockingReload(t *testing.T) {
	ca1 := newTestCA(t)
	ca2 := newTestCA(t)

	var reqCount int32
	activeCA := ca1
	srv := newHTTPTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		der := newTestCRLDER(t, activeCA, 1, nil, time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
		_, _ = w.Write(der)
	})

	store := &CRLStore{crlReloadInterval: time.Hour}
	provider := &failedCloseSnapshotProvider{clientHolder: newCRLHTTPClientHolder(http.DefaultTransport.(*http.Transport).Clone(), time.Second, 1024)}

	_, err := provider.getVerifiedSnapshot(store, srv.URL, ca1.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	// Fresh snapshot, same issuer: fast path, no new request.
	_, err = provider.getVerifiedSnapshot(store, srv.URL, ca1.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&reqCount))

	// Issuer rotation forces a blocking reload even though the cached
	// snapshot is still within crlReloadInterval.
	activeCA = ca2
	snap, err := provider.getVerifiedSnapshot(store, srv.URL, ca2.cert)
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&reqCount))
	assert.Equal(t, big.NewInt(1), snap.crlSerial())
}

func TestFailedCloseProvider_IssuerRotation_BackoffActive_FailsClosed(t *testing.T) {
	ca1 := newTestCA(t)
	ca2 := newTestCA(t)

	store := &CRLStore{crlReloadInterval: time.Hour, crlErrorBackoff: time.Minute}
	provider := &failedCloseSnapshotProvider{}

	entry := store.getOrCreateEntry("dp1")
	stale := &dynamicCrlSnapshot{crlSnapshotCommon{
		number: big.NewInt(1), modTime: time.Now(), nextUpdate: time.Now().Add(time.Hour),
	}}
	entry.storeSnapshot(stale)
	entry.updateIssuer(ca1.cert)
	entry.recordFailure()
	entry.loader = loaderFunc(func(e *crlEntry) (crlSnapshot, error) {
		t.Fatal("reload should not be attempted while backoff is active")
		return nil, nil
	})

	// Fail-closed semantics are preserved under rotation: a stale snapshot
	// is never returned silently, even if it was valid before the rotation.
	_, err := provider.getVerifiedSnapshot(store, "dp1", ca2.cert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry not attempted")
}
