package ttlmap

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Map is a thread-safe map with a bounded capacity and per-entry expiration.
type Map[V any] struct {
	capacity int

	mu       sync.Mutex
	items    map[string]*item[V]
	expiries expiryHeap[V]

	clock func() time.Time
}

// New creates a Map holding at most capacity entries.
func New[V any](capacity int) (*Map[V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be greater than 0")
	}

	return &Map[V]{
		capacity: capacity,
		items:    make(map[string]*item[V]),
		clock:    time.Now,
	}, nil
}

// Get returns the value stored for the key. Expired entries are removed lazily and
// reported as absent.
func (m *Map[V]) Get(key string) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var zero V
	it, ok := m.items[key]
	if !ok {
		return zero, false
	}

	if m.expired(it) {
		m.remove(it)
		return zero, false
	}

	return it.value, true
}

// Set stores value for a key and resets its expiration to ttlSeconds from now.
// When the map is full, inserting a new key evicts the entry closest to expiration.
func (m *Map[V]) Set(key string, value V, ttlSeconds int) error {
	if ttlSeconds <= 0 {
		return fmt.Errorf("ttlSeconds must be greater than 0, got %d", ttlSeconds)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	expiry := m.clock().Add(time.Duration(ttlSeconds) * time.Second).Unix()

	if it, ok := m.items[key]; ok {
		it.value = value
		it.expiry = expiry
		heap.Fix(&m.expiries, it.index)
		return nil
	}

	if len(m.items) >= m.capacity {
		m.freeSpace()
	}

	it := &item[V]{key: key, value: value, expiry: expiry}
	m.items[key] = it
	heap.Push(&m.expiries, it)

	return nil
}

// Len returns the number of entries currently held, including any that have
// expired but not yet been removed.
func (m *Map[V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.items)
}

func (m *Map[V]) expired(it *item[V]) bool {
	return it.expiry <= m.clock().Unix()
}

// freeSpace evicts the entry closest to expiration to make room for a new one.
func (m *Map[V]) freeSpace() {
	if len(m.expiries) == 0 {
		return
	}

	it := heap.Pop(&m.expiries).(*item[V])
	delete(m.items, it.key)
}

func (m *Map[V]) remove(it *item[V]) {
	heap.Remove(&m.expiries, it.index)
	delete(m.items, it.key)
}

// item is an entry of the Map.
type item[V any] struct {
	key    string
	value  V
	expiry int64 // expiration as a Unix timestamp in seconds.
	index  int   // position in the expiries heap, maintained by the heap.
}

// expiryHeap is a min-heap of items ordered by expiration, so that the entry
// closest to expiration sits at the root. It implements heap.Interface.
type expiryHeap[V any] []*item[V]

func (h expiryHeap[V]) Len() int { return len(h) }

func (h expiryHeap[V]) Less(i, j int) bool { return h[i].expiry < h[j].expiry }

func (h expiryHeap[V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *expiryHeap[V]) Push(x any) {
	it := x.(*item[V])
	it.index = len(*h)
	*h = append(*h, it)
}

func (h *expiryHeap[V]) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	old[n-1] = nil // avoid memory leak.
	it.index = -1
	*h = old[:n-1]
	return it
}
