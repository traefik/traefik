package k8s

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorDeduperZeroValueAndWindow(t *testing.T) {
	t.Parallel()

	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d := ErrorDeduper{
		now: func() time.Time { return current },
	}

	// First occurrence logs immediately without suppression.
	shouldLog, suppressed := d.ShouldLog("key1")
	assert.True(t, shouldLog)
	assert.Zero(t, suppressed)

	// Repeats within the 1-minute TTL window are suppressed.
	for range 4 {
		current = current.Add(10 * time.Second)
		shouldLog, suppressed = d.ShouldLog("key1")
		assert.False(t, shouldLog)
		assert.Zero(t, suppressed)
	}

	// Independent keys are tracked separately.
	shouldLog, suppressed = d.ShouldLog("key2")
	assert.True(t, shouldLog)
	assert.Zero(t, suppressed)

	// Once key1's window expires, it logs again with the accumulated suppressed count.
	current = current.Add(25 * time.Second)
	shouldLog, suppressed = d.ShouldLog("key1")
	assert.True(t, shouldLog)
	assert.Equal(t, 4, suppressed)

	// Subsequent call within the fresh window is suppressed again.
	shouldLog, suppressed = d.ShouldLog("key1")
	assert.False(t, shouldLog)
	assert.Zero(t, suppressed)
}

func TestErrorDeduperPrunesStaleEntries(t *testing.T) {
	t.Parallel()

	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d := NewErrorDeduper()
	d.now = func() time.Time { return current }

	shouldLog, _ := d.ShouldLog("transient-single")
	require.True(t, shouldLog)

	shouldLog, _ = d.ShouldLog("quiet-after-burst")
	require.True(t, shouldLog)
	shouldLog, _ = d.ShouldLog("quiet-after-burst")
	require.False(t, shouldLog)

	// Advance past 1 TTL: single zero-count entry is pruned on sweep while pending count remains.
	current = current.Add(2 * time.Minute)
	shouldLog, _ = d.ShouldLog("active-key")
	require.True(t, shouldLog)

	d.mu.Lock()
	_, hasTransient := d.entries["transient-single"]
	_, hasBurst := d.entries["quiet-after-burst"]
	d.mu.Unlock()
	assert.False(t, hasTransient)
	assert.True(t, hasBurst)

	// Advance past maxEntryStaleFactor * TTL without re-firing: burst entry is also reclaimed.
	current = current.Add(11 * time.Minute)
	shouldLog, _ = d.ShouldLog("active-key-2")
	require.True(t, shouldLog)

	d.mu.Lock()
	_, hasBurstAfterStale := d.entries["quiet-after-burst"]
	d.mu.Unlock()
	assert.False(t, hasBurstAfterStale)
}

func TestErrorDeduperLogError(t *testing.T) {
	t.Parallel()

	current := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d := ErrorDeduper{
		now: func() time.Time { return current },
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Str("namespace", "default").Str("ingress", "broken").Logger()

	opts := ErrorLogOptions{
		Provider:    "kubernetes",
		Namespace:   "default",
		Resource:    "broken",
		ServiceName: "missing-svc",
		ServicePort: "80",
		Err:         errors.New("service not found"),
		Msg:         "Cannot create service",
	}

	d.LogError(logger, opts)
	d.LogError(logger, opts)
	d.LogError(logger, opts)

	firstOutput := buf.String()
	assert.Contains(t, firstOutput, `"serviceName":"missing-svc"`)
	assert.Contains(t, firstOutput, `"servicePort":"80"`)
	assert.Contains(t, firstOutput, `"message":"Cannot create service"`)
	assert.NotContains(t, firstOutput, `"suppressed"`)

	buf.Reset()
	current = current.Add(time.Minute + time.Second)
	d.LogError(logger, opts)

	secondOutput := buf.String()
	assert.Contains(t, secondOutput, `"suppressed":2`)
	assert.Contains(t, secondOutput, `"message":"Cannot create service"`)
}

func TestErrorDeduperConcurrency(t *testing.T) {
	t.Parallel()

	var d ErrorDeduper
	var wg sync.WaitGroup

	for g := range 8 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range 50 {
				d.ShouldLog("key-" + strconv.Itoa((id+i)%5))
			}
		}(g)
	}

	wg.Wait()
}
