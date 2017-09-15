package ttlmap

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mailgun/minheap"
	"github.com/mailgun/timetools"
)

type TtlMapOption func(m *TtlMap) error

// Clock sets the time provider clock, handy for testing
func Clock(c timetools.TimeProvider) TtlMapOption {
	return func(m *TtlMap) error {
		m.clock = c
		return nil
	}
}

type Callback func(key string, el interface{})

// CallOnExpire will call this callback on expiration of elements
func CallOnExpire(cb Callback) TtlMapOption {
	return func(m *TtlMap) error {
		m.onExpire = cb
		return nil
	}
}

type TtlMap struct {
	capacity    int
	elements    map[string]*mapElement
	expiryTimes *minheap.MinHeap
	clock       timetools.TimeProvider
	mutex       *sync.RWMutex
	// onExpire callback will be called when element is expired
	onExpire Callback
}

type mapElement struct {
	key    string
	value  interface{}
	heapEl *minheap.Element
}

func NewMap(capacity int, opts ...TtlMapOption) (*TtlMap, error) {
	if capacity <= 0 {
		return nil, errors.New("Capacity should be > 0")
	}

	m := &TtlMap{
		capacity:    capacity,
		elements:    make(map[string]*mapElement),
		expiryTimes: minheap.NewMinHeap(),
	}

	for _, o := range opts {
		if err := o(m); err != nil {
			return nil, err
		}
	}

	if m.clock == nil {
		m.clock = &timetools.RealTime{}
	}

	return m, nil
}

func NewMapWithProvider(capacity int, timeProvider timetools.TimeProvider) (*TtlMap, error) {
	if timeProvider == nil {
		return nil, errors.New("Please pass timeProvider")
	}
	return NewMap(capacity, Clock(timeProvider))
}

func NewConcurrent(capacity int, opts ...TtlMapOption) (*TtlMap, error) {
	m, err := NewMap(capacity, opts...)
	if err == nil {
		m.mutex = new(sync.RWMutex)
	}
	return m, err
}

func (m *TtlMap) Set(key string, value interface{}, ttlSeconds int) error {
	expiryTime, err := m.toEpochSeconds(ttlSeconds)
	if err != nil {
		return err
	}
	if m.mutex != nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}
	return m.set(key, value, expiryTime)
}

func (m *TtlMap) Len() int {
	if m.mutex != nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}
	return len(m.elements)
}

func (m *TtlMap) Get(key string) (interface{}, bool) {
	value, mapEl, expired := m.lockNGet(key)
	if mapEl == nil {
		return nil, false
	}
	if expired {
		m.lockNDel(mapEl)
		return nil, false
	}
	return value, true
}

func (m *TtlMap) Increment(key string, value int, ttlSeconds int) (int, error) {
	expiryTime, err := m.toEpochSeconds(ttlSeconds)
	if err != nil {
		return 0, err
	}

	if m.mutex != nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}

	mapEl, expired := m.get(key)
	if mapEl == nil || expired {
		m.set(key, value, expiryTime)
		return value, nil
	}

	currentValue, ok := mapEl.value.(int)
	if !ok {
		return 0, fmt.Errorf("Expected existing value to be integer, got %T", mapEl.value)
	}

	currentValue += value
	m.set(key, currentValue, expiryTime)
	return currentValue, nil
}

func (m *TtlMap) GetInt(key string) (int, bool, error) {
	valueI, exists := m.Get(key)
	if !exists {
		return 0, false, nil
	}
	value, ok := valueI.(int)
	if !ok {
		return 0, false, fmt.Errorf("Expected existing value to be integer, got %T", valueI)
	}
	return value, true, nil
}

func (m *TtlMap) set(key string, value interface{}, expiryTime int) error {
	if mapEl, ok := m.elements[key]; ok {
		mapEl.value = value
		m.expiryTimes.UpdateEl(mapEl.heapEl, expiryTime)
		return nil
	}

	if len(m.elements) >= m.capacity {
		m.freeSpace(1)
	}
	heapEl := &minheap.Element{
		Priority: expiryTime,
	}
	mapEl := &mapElement{
		key:    key,
		value:  value,
		heapEl: heapEl,
	}
	heapEl.Value = mapEl
	m.elements[key] = mapEl
	m.expiryTimes.PushEl(heapEl)
	return nil
}

func (m *TtlMap) lockNGet(key string) (value interface{}, mapEl *mapElement, expired bool) {
	if m.mutex != nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}

	mapEl, expired = m.get(key)
	value = nil
	if mapEl != nil {
		value = mapEl.value
	}
	return value, mapEl, expired
}

func (m *TtlMap) get(key string) (*mapElement, bool) {
	mapEl, ok := m.elements[key]
	if !ok {
		return nil, false
	}
	now := int(m.clock.UtcNow().Unix())
	expired := mapEl.heapEl.Priority <= now
	return mapEl, expired
}

func (m *TtlMap) lockNDel(mapEl *mapElement) {
	if m.mutex != nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		// Map element could have been updated. Now that we have a lock
		// retrieve it again and check if it is still expired.
		var ok bool
		if mapEl, ok = m.elements[mapEl.key]; !ok {
			return
		}
		now := int(m.clock.UtcNow().Unix())
		if mapEl.heapEl.Priority > now {
			return
		}
	}
	m.del(mapEl)
}

func (m *TtlMap) del(mapEl *mapElement) {
	if m.onExpire != nil {
		m.onExpire(mapEl.key, mapEl.value)
	}

	delete(m.elements, mapEl.key)
	m.expiryTimes.RemoveEl(mapEl.heapEl)
}

func (m *TtlMap) freeSpace(count int) {
	removed := m.removeExpired(count)
	if removed >= count {
		return
	}
	m.removeLastUsed(count - removed)
}

func (m *TtlMap) removeExpired(iterations int) int {
	removed := 0
	now := int(m.clock.UtcNow().Unix())
	for i := 0; i < iterations; i += 1 {
		if len(m.elements) == 0 {
			break
		}
		heapEl := m.expiryTimes.PeekEl()
		if heapEl.Priority > now {
			break
		}
		m.expiryTimes.PopEl()
		mapEl := heapEl.Value.(*mapElement)
		delete(m.elements, mapEl.key)
		removed += 1
	}
	return removed
}

func (m *TtlMap) removeLastUsed(iterations int) {
	for i := 0; i < iterations; i += 1 {
		if len(m.elements) == 0 {
			return
		}
		heapEl := m.expiryTimes.PopEl()
		mapEl := heapEl.Value.(*mapElement)
		delete(m.elements, mapEl.key)
	}
}

func (m *TtlMap) toEpochSeconds(ttlSeconds int) (int, error) {
	if ttlSeconds <= 0 {
		return 0, fmt.Errorf("ttlSeconds should be >= 0, got %d", ttlSeconds)
	}
	return int(m.clock.UtcNow().Add(time.Second * time.Duration(ttlSeconds)).Unix()), nil
}
