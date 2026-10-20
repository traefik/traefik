package externalaction

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullDeduplicationDoesNotRepeatCleanup(t *testing.T) {
	store := &memoryDeduplicator{entries: make(map[[32]byte]memoryDeduplicationEntry), maxEntries: 1}
	claim, claimed, err := store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, store.Complete(t.Context(), "existing", claim, time.Hour))
	// Force one full-map scan to establish its next possible expiry.
	store.lastCleanup = time.Now().Add(-time.Minute)
	_, _, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.ErrorContains(t, err, "capacity is full")
	lastCleanup := store.lastCleanup
	for range 1000 {
		_, claimed, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
		require.ErrorContains(t, err, "capacity is full")
		require.False(t, claimed)
		_, claimed, err = store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
		require.NoError(t, err)
		require.False(t, claimed)
	}
	assert.Equal(t, lastCleanup, store.lastCleanup)
}

func TestFullDeduplicationReclaimsDueExpiry(t *testing.T) {
	store := &memoryDeduplicator{entries: make(map[[32]byte]memoryDeduplicationEntry), maxEntries: 1}
	claim, claimed, err := store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	// Completing with a shorter TTL must invalidate the previous expiry bound.
	require.NoError(t, store.Complete(t.Context(), "existing", claim, -time.Second))
	require.True(t, store.nextExpiry.Before(time.Now()))
	_, claimed, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	assert.NotContains(t, store.entries, sha256.Sum256([]byte("existing")))
	assert.Len(t, store.entries, 1)
}

func TestDeduplicationRecomputesExpiryAfterExtension(t *testing.T) {
	store := &memoryDeduplicator{entries: make(map[[32]byte]memoryDeduplicationEntry), maxEntries: 1}
	claim, claimed, err := store.Claim(t.Context(), "rule", "existing", -time.Second, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, store.Complete(t.Context(), "existing", claim, time.Hour))
	_, _, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.ErrorContains(t, err, "capacity is full")
	assert.True(t, store.nextExpiry.After(time.Now()))
	lastCleanup := store.lastCleanup
	_, _, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.Error(t, err)
	assert.Equal(t, lastCleanup, store.lastCleanup)
	require.NoError(t, store.Release(t.Context(), "existing", claim))
	// Release cannot remove a completed delivery.
	assert.Len(t, store.entries, 1)
}
