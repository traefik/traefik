package ttlmap

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withClock returns a Map whose clock is controlled by the returned pointer.
func withClock[V any](t *testing.T, capacity int) (*Map[V], *time.Time) {
	t.Helper()

	m, err := New[V](capacity)
	require.NoError(t, err)

	now := time.Unix(1000, 0)
	m.clock = func() time.Time { return now }

	return m, &now
}

func TestNew_invalidCapacity(t *testing.T) {
	_, err := New[string](0)
	assert.Error(t, err)

	_, err = New[string](-1)
	assert.Error(t, err)
}

func TestMap_setAndGet(t *testing.T) {
	m, _ := withClock[string](t, 10)

	_, ok := m.Get("missing")
	assert.False(t, ok)

	require.NoError(t, m.Set("key", "value", 10))

	value, ok := m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", value)
}

func TestMap_setInvalidTTL(t *testing.T) {
	m, _ := withClock[string](t, 10)

	assert.Error(t, m.Set("key", "value", 0))
	assert.Error(t, m.Set("key", "value", -1))
}

func TestMap_expiration(t *testing.T) {
	m, now := withClock[string](t, 10)

	require.NoError(t, m.Set("key", "value", 10))

	// Just before expiry, the entry is still there.
	*now = now.Add(9 * time.Second)
	_, ok := m.Get("key")
	assert.True(t, ok)

	// At/after expiry, it is gone and lazily removed.
	*now = now.Add(2 * time.Second)
	_, ok = m.Get("key")
	assert.False(t, ok)
	assert.Equal(t, 0, m.Len())
}

func TestMap_setRefreshesExpiration(t *testing.T) {
	m, now := withClock[string](t, 10)

	require.NoError(t, m.Set("key", "value", 10))

	*now = now.Add(8 * time.Second)
	require.NoError(t, m.Set("key", "updated", 10))

	// Original expiry would have passed, but Set pushed it back.
	*now = now.Add(8 * time.Second)
	value, ok := m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "updated", value)
	assert.Equal(t, 1, m.Len())
}

func TestMap_capacityEvictsClosestToExpiry(t *testing.T) {
	m, now := withClock[string](t, 2)

	require.NoError(t, m.Set("a", "a", 100)) // expires first.
	require.NoError(t, m.Set("b", "b", 200))

	// Inserting a third key evicts "a", the entry closest to expiration.
	require.NoError(t, m.Set("c", "c", 300))

	_, ok := m.Get("a")
	assert.False(t, ok)

	_, ok = m.Get("b")
	assert.True(t, ok)

	_, ok = m.Get("c")
	assert.True(t, ok)

	assert.Equal(t, 2, m.Len())

	// now is unused beyond construction here but kept for symmetry.
	_ = now
}

func TestMap_concurrentAccess(t *testing.T) {
	m, err := New[int](100)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			key := strconv.Itoa(i)
			require.NoError(t, m.Set(key, i, 60))
			m.Get(key)
		})
	}
	wg.Wait()
}
