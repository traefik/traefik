package k8s

import (
	"testing"
	"time"
)

func TestErrorDeduper_ShouldLog(t *testing.T) {
	d := NewErrorDeduper()

	// First occurrence should always be logged.
	shouldLog, suppressed := d.ShouldLog("key1")
	if !shouldLog {
		t.Error("expected first occurrence to be logged")
	}
	if suppressed != 0 {
		t.Errorf("expected suppressed=0, got %d", suppressed)
	}

	// Duplicate within TTL should be suppressed.
	shouldLog, suppressed = d.ShouldLog("key1")
	if shouldLog {
		t.Error("expected duplicate within TTL to be suppressed")
	}
	if suppressed != 0 {
		t.Errorf("expected suppressed=0 for suppressed entry, got %d", suppressed)
	}

	// Multiple suppressions should accumulate.
	for i := 0; i < 3; i++ {
		d.ShouldLog("key1")
	}

	// Expire the TTL manually.
	d.mu.Lock()
	entry := d.entries["key1"]
	entry.lastSeen = time.Now().Add(-2 * time.Minute)
	d.entries["key1"] = entry
	d.mu.Unlock()

	// After TTL expires, should log again with suppressed count.
	shouldLog, suppressed = d.ShouldLog("key1")
	if !shouldLog {
		t.Error("expected log after TTL expiry")
	}
	if suppressed != 4 {
		t.Errorf("expected suppressed=4, got %d", suppressed)
	}

	// Different key should log independently.
	shouldLog, suppressed = d.ShouldLog("key2")
	if !shouldLog {
		t.Error("expected different key to be logged")
	}
	if suppressed != 0 {
		t.Errorf("expected suppressed=0 for new key, got %d", suppressed)
	}
}
