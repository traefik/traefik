package k8s

import (
	"sync"
	"time"
)

const defaultDedupTTL = 1 * time.Minute

type dedupEntry struct {
	lastSeen time.Time
	count    int
}

// ErrorDeduper suppresses repeated identical error logs within a TTL window.
type ErrorDeduper struct {
	mu      sync.Mutex
	entries map[string]dedupEntry
	ttl     time.Duration
}

// NewErrorDeduper creates an ErrorDeduper with the default 1-minute TTL.
func NewErrorDeduper() *ErrorDeduper {
	return &ErrorDeduper{
		entries: make(map[string]dedupEntry),
		ttl:     defaultDedupTTL,
	}
}

// ShouldLog returns true if the error should be logged, and the number of
// suppressed occurrences since the last time it was logged.
func (d *ErrorDeduper) ShouldLog(key string) (bool, int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	entry, exists := d.entries[key]

	if exists && now.Sub(entry.lastSeen) < d.ttl {
		// Within TTL: suppress and increment count.
		d.entries[key] = dedupEntry{
			lastSeen: entry.lastSeen,
			count:    entry.count + 1,
		}
		return false, 0
	}

	// Either first time or TTL expired: allow log.
	suppressed := 0
	if exists {
		suppressed = entry.count
	}
	d.entries[key] = dedupEntry{
		lastSeen: now,
		count:    0,
	}
	return true, suppressed
}
