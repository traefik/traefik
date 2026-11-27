package k8s

import (
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

const (
	defaultDedupTTL     = time.Minute
	maxEntryStaleFactor = 10
)

type dedupEntry struct {
	windowStart time.Time
	lastSeen    time.Time
	count       int
}

// ErrorLogOptions describes a deduplicated Kubernetes provider error log event.
type ErrorLogOptions struct {
	Provider    string
	Namespace   string
	Resource    string
	ServiceName string
	ServicePort string
	Err         error
	Msg         string
}

// ErrorDeduper suppresses repeated identical error logs within a TTL window.
// The zero value defaults to a 1-minute TTL and system clock, so callers do
// not need explicit initialization.
type ErrorDeduper struct {
	mu        sync.Mutex
	entries   map[string]dedupEntry
	ttl       time.Duration
	now       func() time.Time
	lastSweep time.Time
}

// NewErrorDeduper creates an ErrorDeduper with the default 1-minute TTL.
func NewErrorDeduper() *ErrorDeduper {
	return &ErrorDeduper{
		entries: make(map[string]dedupEntry),
		ttl:     defaultDedupTTL,
		now:     time.Now,
	}
}

// Reset clears any tracked deduplication windows.
func (d *ErrorDeduper) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	clear(d.entries)
	d.lastSweep = time.Time{}
}

// ShouldLog returns true if the error for key should be logged now, along with
// the number of occurrences suppressed during the expired window.
func (d *ErrorDeduper) ShouldLog(key string) (bool, int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.initLocked()
	now := d.now()

	entry, exists := d.entries[key]
	if !exists {
		d.maybeSweepLocked(now)
		d.entries[key] = dedupEntry{
			windowStart: now,
			lastSeen:    now,
		}
		return true, 0
	}

	if now.Sub(entry.windowStart) < d.ttl {
		entry.count++
		entry.lastSeen = now
		d.entries[key] = entry
		return false, 0
	}

	suppressed := entry.count
	d.entries[key] = dedupEntry{
		windowStart: now,
		lastSeen:    now,
	}
	d.maybeSweepLocked(now)
	return true, suppressed
}

// LogError emits an error event on logger when the deduplication window allows it,
// attaching a "suppressed" count field if earlier occurrences were elided.
func (d *ErrorDeduper) LogError(logger zerolog.Logger, opts ErrorLogOptions) {
	errText := ""
	if opts.Err != nil {
		errText = opts.Err.Error()
	}

	key := strings.Join([]string{
		opts.Provider,
		opts.Namespace,
		opts.Resource,
		opts.ServiceName,
		opts.ServicePort,
		opts.Msg,
		errText,
	}, "\x00")

	shouldLog, suppressed := d.ShouldLog(key)
	if !shouldLog {
		return
	}

	event := logger.Error()
	if opts.ServiceName != "" {
		event = event.Str(logs.ServiceName, opts.ServiceName)
	}
	if opts.ServicePort != "" {
		event = event.Str("servicePort", opts.ServicePort)
	}
	if opts.Err != nil {
		event = event.Err(opts.Err)
	}
	if suppressed > 0 {
		event = event.Int("suppressed", suppressed)
	}

	if opts.Msg != "" {
		event.Msg(opts.Msg)
		return
	}
	event.Send()
}

func (d *ErrorDeduper) initLocked() {
	if d.entries == nil {
		d.entries = make(map[string]dedupEntry)
	}
	if d.ttl <= 0 {
		d.ttl = defaultDedupTTL
	}
	if d.now == nil {
		d.now = time.Now
	}
}

func (d *ErrorDeduper) maybeSweepLocked(now time.Time) {
	if d.lastSweep.IsZero() {
		d.lastSweep = now
		return
	}
	if now.Sub(d.lastSweep) < d.ttl {
		return
	}

	for key, entry := range d.entries {
		if entry.count == 0 && now.Sub(entry.windowStart) >= d.ttl {
			delete(d.entries, key)
			continue
		}
		if now.Sub(entry.lastSeen) >= maxEntryStaleFactor*d.ttl {
			delete(d.entries, key)
		}
	}
	d.lastSweep = now
}
