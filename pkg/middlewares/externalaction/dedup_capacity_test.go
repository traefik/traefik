package externalaction

import (
	"crypto/sha256"
	"sync"
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
	// An old cleanup timestamp must not force a scan of live entries.
	store.lastCleanup = time.Now().Add(-time.Hour)
	lastCleanup := store.lastCleanup
	_, _, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.ErrorContains(t, err, "capacity is full")
	assert.Equal(t, lastCleanup, store.lastCleanup)
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

func TestDeduplicationConcurrentExpiryReclaim(t *testing.T) {
	store := &memoryDeduplicator{entries: make(map[[32]byte]memoryDeduplicationEntry), maxEntries: 1}
	oldClaim, claimed, err := store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, store.Complete(t.Context(), "existing", oldClaim, -time.Second))

	type result struct {
		claimID string
		claimed bool
		err     error
	}
	results := make(chan result, 64)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for range cap(results) {
		workers.Go(func() {
			<-start
			claimID, claimed, err := store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
			results <- result{claimID: claimID, claimed: claimed, err: err}
		})
	}
	close(start)
	workers.Wait()
	close(results)
	var claims []string
	for result := range results {
		require.NoError(t, result.err)
		if result.claimed {
			claims = append(claims, result.claimID)
		}
	}
	require.Len(t, claims, 1)
	require.Error(t, store.Complete(t.Context(), "existing", oldClaim, time.Hour))
	require.NoError(t, store.Release(t.Context(), "existing", oldClaim))
	require.NoError(t, store.Complete(t.Context(), "existing", claims[0], time.Hour))
	_, claimed, err = store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
	require.NoError(t, err)
	assert.False(t, claimed)
}

func TestDeduplicationCleanupBeforeCapacity(t *testing.T) {
	store := &memoryDeduplicator{entries: make(map[[32]byte]memoryDeduplicationEntry), maxEntries: 3}
	claim, claimed, err := store.Claim(t.Context(), "rule", "existing", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, store.Complete(t.Context(), "existing", claim, time.Hour))
	store.lastCleanup = time.Now().Add(-time.Hour)
	lastCleanup := store.lastCleanup
	_, claimed, err = store.Claim(t.Context(), "rule", "new", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	assert.Equal(t, lastCleanup, store.lastCleanup)

	// Expiry must still be reclaimed promptly when the map is not full.
	hash := sha256.Sum256([]byte("existing"))
	entry := store.entries[hash]
	entry.expiresAt = time.Now().Add(-time.Second)
	store.entries[hash] = entry
	store.trackExpiry(entry.expiresAt)
	_, claimed, err = store.Claim(t.Context(), "rule", "another", time.Hour, time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
	assert.NotContains(t, store.entries, hash)
	assert.Len(t, store.entries, 2)
	assert.True(t, store.lastCleanup.After(lastCleanup))
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
