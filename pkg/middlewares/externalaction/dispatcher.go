package externalaction

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type actionTask struct {
	rule *compiledRule
	data *actionData
	size int64
}

type dispatcherKey struct {
	done   <-chan struct{}
	config [32]byte
}

var dispatcherRegistry = struct {
	sync.Mutex

	values map[dispatcherKey]*dispatcher
}{
	values: make(map[dispatcherKey]*dispatcher),
}

type dispatcher struct {
	ctx            context.Context
	queue          chan actionTask
	done           chan struct{}
	wait           time.Duration
	logger         zerolog.Logger
	dropped        atomic.Uint64
	lastDropLog    atomic.Int64
	queuedBytes    atomic.Int64
	maxQueuedBytes int64
	wg             sync.WaitGroup
	stopMu         sync.RWMutex
	stopped        bool
}

func getDispatcher(ctx context.Context, config dynamic.ExternalAction, configHash [32]byte, logger zerolog.Logger) *dispatcher {
	key := dispatcherKey{done: ctx.Done(), config: configHash}

	dispatcherRegistry.Lock()
	defer dispatcherRegistry.Unlock()
	if existing := dispatcherRegistry.values[key]; existing != nil {
		return existing
	}

	capacity := config.Queue.Capacity
	if capacity <= 0 {
		capacity = 1024
	}
	workers := config.Queue.Workers
	if workers <= 0 {
		workers = 8
	}
	maxQueuedBytes := config.Queue.MaxQueuedBytes
	if maxQueuedBytes <= 0 {
		maxQueuedBytes = 64 * 1024 * 1024
	}
	value := &dispatcher{
		ctx:            ctx,
		queue:          make(chan actionTask, capacity),
		done:           make(chan struct{}),
		wait:           time.Duration(config.Queue.EnqueueWait),
		logger:         logger,
		maxQueuedBytes: maxQueuedBytes,
	}
	for range workers {
		value.wg.Add(1)
		go value.run()
	}
	go func() {
		<-ctx.Done()
		value.stopMu.Lock()
		value.stopped = true
		close(value.done)
		value.stopMu.Unlock()
		value.wg.Wait()
		dispatcherRegistry.Lock()
		delete(dispatcherRegistry.values, key)
		dispatcherRegistry.Unlock()
	}()
	dispatcherRegistry.values[key] = value
	return value
}

func (d *dispatcher) enqueue(value actionTask) bool {
	d.stopMu.RLock()
	defer d.stopMu.RUnlock()
	if d.stopped {
		return false
	}

	value.size = estimateTaskSize(value)
	if !d.reserve(value.size) {
		return false
	}
	accepted := false
	defer func() {
		if !accepted {
			d.queuedBytes.Add(-value.size)
		}
	}()

	if d.wait <= 0 {
		select {
		case d.queue <- value:
			accepted = true
			return true
		case <-d.done:
			return false
		default:
			return false
		}
	}
	timer := time.NewTimer(d.wait)
	defer timer.Stop()
	select {
	case d.queue <- value:
		accepted = true
		return true
	case <-d.done:
		return false
	case <-timer.C:
		return false
	}
}

func (d *dispatcher) reserve(size int64) bool {
	if size > d.maxQueuedBytes {
		return false
	}
	for {
		current := d.queuedBytes.Load()
		if current+size > d.maxQueuedBytes {
			return false
		}
		if d.queuedBytes.CompareAndSwap(current, current+size) {
			return true
		}
	}
}

func estimateTaskSize(task actionTask) int64 {
	if task.data == nil {
		return 1024
	}
	size := int64(3*len(task.data.RawBody) + 3*len(task.data.Response.RawBody) + len(task.data.Request.RawQuery) + len(task.data.Request.Host) + len(task.data.Request.Path) + 1024)
	for name, values := range task.data.Request.Headers {
		size += int64(len(name))
		for _, value := range values {
			size += int64(len(value))
		}
	}
	return max(size, 1024)
}

func (d *dispatcher) logDrop(rule string) {
	total := d.dropped.Add(1)
	now := time.Now().UnixNano()
	last := d.lastDropLog.Load()
	if last != 0 && time.Duration(now-last) < 10*time.Second {
		return
	}
	if !d.lastDropLog.CompareAndSwap(last, now) {
		return
	}
	d.logger.Warn().
		Str("rule", rule).
		Uint64("dropped", total).
		Msg("External actions dropped because the bounded queue is full")
}

func (d *dispatcher) run() {
	defer d.wg.Done()
	run := func(value actionTask) {
		timeout := time.Duration(value.rule.config.Timeout)
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		// A replaced Middleware cancels d.ctx before its accepted queue has
		// drained. Preserve request-scoped values but bound every final action.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(d.ctx), timeout)
		err := runRule(ctx, value.rule, value.data)
		cancel()
		d.queuedBytes.Add(-value.size)
		if err != nil {
			d.logger.Error().Err(err).Str("rule", value.rule.config.Name).Msg("Asynchronous external action failed")
		}
	}

	for {
		select {
		case value := <-d.queue:
			run(value)
		case <-d.done:
			for {
				select {
				case value := <-d.queue:
					run(value)
				default:
					return
				}
			}
		}
	}
}
