package goque

import (
	"bytes"
	"encoding/gob"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// Stack is a standard LIFO (last in, first out) stack.
type Stack struct {
	sync.RWMutex
	DataDir string
	db      *leveldb.DB
	head    uint64
	tail    uint64
	isOpen  bool
}

// OpenStack opens a stack if one exists at the given directory. If one
// does not already exist, a new stack is created.
func OpenStack(dataDir string) (*Stack, error) {
	var err error

	// Create a new Stack.
	s := &Stack{
		DataDir: dataDir,
		db:      &leveldb.DB{},
		head:    0,
		tail:    0,
		isOpen:  false,
	}

	// Open database for the stack.
	s.db, err = leveldb.OpenFile(dataDir, nil)
	if err != nil {
		return s, err
	}

	// Check if this Goque type can open the requested data directory.
	ok, err := checkGoqueType(dataDir, goqueStack)
	if err != nil {
		return s, err
	}
	if !ok {
		return s, ErrIncompatibleType
	}

	// Set isOpen and return.
	s.isOpen = true
	return s, s.init()
}

// Push adds an item to the stack.
func (s *Stack) Push(value []byte) (*Item, error) {
	s.Lock()
	defer s.Unlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	// Create new Item.
	item := &Item{
		ID:    s.head + 1,
		Key:   idToKey(s.head + 1),
		Value: value,
	}

	// Add it to the stack.
	if err := s.db.Put(item.Key, item.Value, nil); err != nil {
		return nil, err
	}

	// Increment head position.
	s.head++

	return item, nil
}

// PushString is a helper function for Push that accepts a
// value as a string rather than a byte slice.
func (s *Stack) PushString(value string) (*Item, error) {
	return s.Push([]byte(value))
}

// PushObject is a helper function for Push that accepts any
// value type, which is then encoded into a byte slice using
// encoding/gob.
func (s *Stack) PushObject(value interface{}) (*Item, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return s.Push(buffer.Bytes())
}

// Pop removes the next item in the stack and returns it.
func (s *Stack) Pop() (*Item, error) {
	s.Lock()
	defer s.Unlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	// Try to get the next item in the stack.
	item, err := s.getItemByID(s.head)
	if err != nil {
		return nil, err
	}

	// Remove this item from the stack.
	if err := s.db.Delete(item.Key, nil); err != nil {
		return nil, err
	}

	// Decrement head position.
	s.head--

	return item, nil
}

// Peek returns the next item in the stack without removing it.
func (s *Stack) Peek() (*Item, error) {
	s.RLock()
	defer s.RUnlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	return s.getItemByID(s.head)
}

// PeekByOffset returns the item located at the given offset,
// starting from the head of the stack, without removing it.
func (s *Stack) PeekByOffset(offset uint64) (*Item, error) {
	s.RLock()
	defer s.RUnlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	return s.getItemByID(s.head - offset)
}

// PeekByID returns the item with the given ID without removing it.
func (s *Stack) PeekByID(id uint64) (*Item, error) {
	s.RLock()
	defer s.RUnlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	return s.getItemByID(id)
}

// Update updates an item in the stack without changing its position.
func (s *Stack) Update(id uint64, newValue []byte) (*Item, error) {
	s.Lock()
	defer s.Unlock()

	// Check if stack is closed.
	if !s.isOpen {
		return nil, ErrDBClosed
	}

	// Check if item exists in stack.
	if id > s.head || id <= s.tail {
		return nil, ErrOutOfBounds
	}

	// Create new Item.
	item := &Item{
		ID:    id,
		Key:   idToKey(id),
		Value: newValue,
	}

	// Update this item in the stack.
	if err := s.db.Put(item.Key, item.Value, nil); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateString is a helper function for Update that accepts a value
// as a string rather than a byte slice.
func (s *Stack) UpdateString(id uint64, newValue string) (*Item, error) {
	return s.Update(id, []byte(newValue))
}

// UpdateObject is a helper function for Update that accepts any
// value type, which is then encoded into a byte slice using
// encoding/gob.
func (s *Stack) UpdateObject(id uint64, newValue interface{}) (*Item, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(newValue); err != nil {
		return nil, err
	}
	return s.Update(id, buffer.Bytes())
}

// Length returns the total number of items in the stack.
func (s *Stack) Length() uint64 {
	return s.head - s.tail
}

// Close closes the LevelDB database of the stack.
func (s *Stack) Close() {
	s.Lock()
	defer s.Unlock()

	// Check if stack is already closed.
	if !s.isOpen {
		return
	}

	// Reset stack head and tail.
	s.head = 0
	s.tail = 0

	s.db.Close()
	s.isOpen = false
}

// Drop closes and deletes the LevelDB database of the stack.
func (s *Stack) Drop() {
	s.Close()
	os.RemoveAll(s.DataDir)
}

// getItemByID returns an item, if found, for the given ID.
func (s *Stack) getItemByID(id uint64) (*Item, error) {
	// Check if empty or out of bounds.
	if s.Length() == 0 {
		return nil, ErrEmpty
	} else if id <= s.tail || id > s.head {
		return nil, ErrOutOfBounds
	}

	// Get item from database.
	var err error
	item := &Item{ID: id, Key: idToKey(id)}
	if item.Value, err = s.db.Get(item.Key, nil); err != nil {
		return nil, err
	}

	return item, nil
}

// init initializes the stack data.
func (s *Stack) init() error {
	// Create a new LevelDB Iterator.
	iter := s.db.NewIterator(nil, nil)
	defer iter.Release()

	// Set stack head to the last item.
	if iter.Last() {
		s.head = keyToID(iter.Key())
	}

	// Set stack tail to the first item.
	if iter.First() {
		s.tail = keyToID(iter.Key()) - 1
	}

	return iter.Error()
}
