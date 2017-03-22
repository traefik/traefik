package goque

import (
	"bytes"
	"encoding/gob"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// prefixSep is the prefix separator for each item key.
var prefixSep []byte = []byte(":")

// order defines the priority ordering of the queue.
type order int

// Defines which priority order to dequeue in.
const (
	ASC  order = iota // Set priority level 0 as most important.
	DESC              // Set priority level 255 as most important.
)

// priorityLevel holds the head and tail position of a priority
// level within the queue.
type priorityLevel struct {
	head uint64
	tail uint64
}

// length returns the total number of items in this priority level.
func (pl *priorityLevel) length() uint64 {
	return pl.tail - pl.head
}

// PriorityQueue is a standard FIFO (first in, first out) queue with
// priority levels.
type PriorityQueue struct {
	sync.RWMutex
	DataDir  string
	db       *leveldb.DB
	order    order
	levels   [256]*priorityLevel
	curLevel uint8
	isOpen   bool
}

// OpenPriorityQueue opens a priority queue if one exists at the given
// directory. If one does not already exist, a new priority queue is
// created.
func OpenPriorityQueue(dataDir string, order order) (*PriorityQueue, error) {
	var err error

	// Create a new PriorityQueue.
	pq := &PriorityQueue{
		DataDir: dataDir,
		db:      &leveldb.DB{},
		order:   order,
		isOpen:  false,
	}

	// Open database for the priority queue.
	pq.db, err = leveldb.OpenFile(dataDir, nil)
	if err != nil {
		return pq, err
	}

	// Check if this Goque type can open the requested data directory.
	ok, err := checkGoqueType(dataDir, goquePriorityQueue)
	if err != nil {
		return pq, err
	}
	if !ok {
		return pq, ErrIncompatibleType
	}

	// Set isOpen and return.
	pq.isOpen = true
	return pq, pq.init()
}

// Enqueue adds an item to the priority queue.
func (pq *PriorityQueue) Enqueue(priority uint8, value []byte) (*PriorityItem, error) {
	pq.Lock()
	defer pq.Unlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	// Get the priorityLevel.
	level := pq.levels[priority]

	// Create new PriorityItem.
	item := &PriorityItem{
		ID:       level.tail + 1,
		Priority: priority,
		Key:      pq.generateKey(priority, level.tail+1),
		Value:    value,
	}

	// Add it to the priority queue.
	if err := pq.db.Put(item.Key, item.Value, nil); err != nil {
		return nil, err
	}

	// Increment tail position.
	level.tail++

	// If this priority level is more important than the curLevel.
	if pq.cmpAsc(priority) || pq.cmpDesc(priority) {
		pq.curLevel = priority
	}

	return item, nil
}

// EnqueueString is a helper function for Enqueue that accepts a
// value as a string rather than a byte slice.
func (pq *PriorityQueue) EnqueueString(priority uint8, value string) (*PriorityItem, error) {
	return pq.Enqueue(priority, []byte(value))
}

// EnqueueObject is a helper function for Enqueue that accepts any
// value type, which is then encoded into a byte slice using
// encoding/gob.
func (pq *PriorityQueue) EnqueueObject(priority uint8, value interface{}) (*PriorityItem, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return pq.Enqueue(priority, buffer.Bytes())
}

// Dequeue removes the next item in the priority queue and returns it.
func (pq *PriorityQueue) Dequeue() (*PriorityItem, error) {
	pq.Lock()
	defer pq.Unlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	// Try to get the next item.
	item, err := pq.getNextItem()
	if err != nil {
		return nil, err
	}

	// Remove this item from the priority queue.
	if err = pq.db.Delete(item.Key, nil); err != nil {
		return nil, err
	}

	// Increment head position.
	pq.levels[pq.curLevel].head++

	return item, nil
}

// DequeueByPriority removes the next item in the given priority level
// and returns it.
func (pq *PriorityQueue) DequeueByPriority(priority uint8) (*PriorityItem, error) {
	pq.Lock()
	defer pq.Unlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	// Try to get the next item in the given priority level.
	item, err := pq.getItemByPriorityID(priority, pq.levels[priority].head+1)
	if err != nil {
		return nil, err
	}

	// Remove this item from the priority queue.
	if err = pq.db.Delete(item.Key, nil); err != nil {
		return nil, err
	}

	// Increment head position.
	pq.levels[priority].head++

	return item, nil
}

// Peek returns the next item in the priority queue without removing it.
func (pq *PriorityQueue) Peek() (*PriorityItem, error) {
	pq.RLock()
	defer pq.RUnlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	return pq.getNextItem()
}

// PeekByOffset returns the item located at the given offset,
// starting from the head of the queue, without removing it.
func (pq *PriorityQueue) PeekByOffset(offset uint64) (*PriorityItem, error) {
	pq.RLock()
	defer pq.RUnlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	// Check if queue is empty.
	if pq.Length() == 0 {
		return nil, ErrEmpty
	}

	// If the offset is within the current priority level.
	if pq.levels[pq.curLevel].length() >= offset+1 {
		return pq.getItemByPriorityID(pq.curLevel, pq.levels[pq.curLevel].head+offset+1)
	}

	return pq.findOffset(offset)
}

// PeekByPriorityID returns the item with the given ID and priority without
// removing it.
func (pq *PriorityQueue) PeekByPriorityID(priority uint8, id uint64) (*PriorityItem, error) {
	pq.RLock()
	defer pq.RUnlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	return pq.getItemByPriorityID(priority, id)
}

// Update updates an item in the priority queue without changing its
// position.
func (pq *PriorityQueue) Update(priority uint8, id uint64, newValue []byte) (*PriorityItem, error) {
	pq.Lock()
	defer pq.Unlock()

	// Check if queue is closed.
	if !pq.isOpen {
		return nil, ErrDBClosed
	}

	// Check if item exists in queue.
	if id <= pq.levels[priority].head || id > pq.levels[priority].tail {
		return nil, ErrOutOfBounds
	}

	// Create new PriorityItem.
	item := &PriorityItem{
		ID:       id,
		Priority: priority,
		Key:      pq.generateKey(priority, id),
		Value:    newValue,
	}

	// Update this item in the queue.
	if err := pq.db.Put(item.Key, item.Value, nil); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateString is a helper function for Update that accepts a value
// as a string rather than a byte slice.
func (pq *PriorityQueue) UpdateString(priority uint8, id uint64, newValue string) (*PriorityItem, error) {
	return pq.Update(priority, id, []byte(newValue))
}

// UpdateObject is a helper function for Update that accepts any
// value type, which is then encoded into a byte slice using
// encoding/gob.
func (pq *PriorityQueue) UpdateObject(priority uint8, id uint64, newValue interface{}) (*PriorityItem, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(newValue); err != nil {
		return nil, err
	}
	return pq.Update(priority, id, buffer.Bytes())
}

// Length returns the total number of items in the priority queue.
func (pq *PriorityQueue) Length() uint64 {
	pq.RLock()
	defer pq.RUnlock()

	var length uint64
	for _, v := range pq.levels {
		length += v.length()
	}

	return length
}

// Close closes the LevelDB database of the priority queue.
func (pq *PriorityQueue) Close() {
	pq.Lock()
	defer pq.Unlock()

	// Check if queue is already closed.
	if !pq.isOpen {
		return
	}

	// Reset head and tail of each priority level.
	for i := 0; i <= 255; i++ {
		pq.levels[uint8(i)].head = 0
		pq.levels[uint8(i)].tail = 0
	}

	pq.db.Close()
	pq.isOpen = false
}

// Drop closes and deletes the LevelDB database of the priority queue.
func (pq *PriorityQueue) Drop() {
	pq.Close()
	os.RemoveAll(pq.DataDir)
}

// cmpAsc returns wehther the given priority level is higher than the
// current priority level based on ascending order.
func (pq *PriorityQueue) cmpAsc(priority uint8) bool {
	return pq.order == ASC && priority < pq.curLevel
}

// cmpAsc returns wehther the given priority level is higher than the
// current priority level based on descending order.
func (pq *PriorityQueue) cmpDesc(priority uint8) bool {
	return pq.order == DESC && priority > pq.curLevel
}

// resetCurrentLevel resets the current priority level of the queue
// so the highest level can be found.
func (pq *PriorityQueue) resetCurrentLevel() {
	if pq.order == ASC {
		pq.curLevel = 255
	} else if pq.order == DESC {
		pq.curLevel = 0
	}
}

// findOffset finds the given offset from the current queue position
// based on priority order.
func (pq *PriorityQueue) findOffset(offset uint64) (*PriorityItem, error) {
	var length uint64
	var curLevel uint8 = pq.curLevel
	var newLevel int

	// Handle newLevel initialization for descending order.
	if pq.order == DESC {
		newLevel = 255
	}

	// For condition expression.
	condExpr := func(level int) bool {
		if pq.order == ASC {
			return level <= 255
		}
		return level >= 0
	}

	// For loop expression.
	loopExpr := func(level *int) {
		if pq.order == ASC {
			*level++
		} else if pq.order == DESC {
			*level--
		}
	}

	// Level comparison.
	cmpLevels := func(newLevel, curLevel uint8) bool {
		if pq.order == ASC {
			return newLevel >= curLevel
		}
		return newLevel <= curLevel
	}

	// Loop through the priority levels.
	for ; condExpr(newLevel); loopExpr(&newLevel) {
		// If this level is lower than the current level based on ordering and contains items.
		if cmpLevels(uint8(newLevel), curLevel) && pq.levels[uint8(newLevel)].length() > 0 {
			curLevel = uint8(newLevel)
			newLength := pq.levels[curLevel].length()

			// If the offset is within the current priority level.
			if length+newLength >= offset+1 {
				return pq.getItemByPriorityID(curLevel, offset-length+1)
			}

			length += newLength
		}
	}

	return nil, ErrOutOfBounds
}

// getNextItem returns the next item in the priority queue, updating
// the current priority level of the queue if necessary.
func (pq *PriorityQueue) getNextItem() (*PriorityItem, error) {
	// If the current priority level is empty.
	if pq.levels[pq.curLevel].length() == 0 {
		// Set starting value for curLevel.
		pq.resetCurrentLevel()

		// Try to get the next priority level.
		for i := 0; i <= 255; i++ {
			if (pq.cmpAsc(uint8(i)) || pq.cmpDesc(uint8(i))) && pq.levels[uint8(i)].length() > 0 {
				pq.curLevel = uint8(i)
			}
		}

		// If still empty, return queue empty error.
		if pq.levels[pq.curLevel].length() == 0 {
			return nil, ErrEmpty
		}
	}

	// Try to get the next item in the current priority level.
	return pq.getItemByPriorityID(pq.curLevel, pq.levels[pq.curLevel].head+1)
}

// getItemByID returns an item, if found, for the given ID.
func (pq *PriorityQueue) getItemByPriorityID(priority uint8, id uint64) (*PriorityItem, error) {
	// Check if empty or out of bounds.
	if pq.levels[priority].length() == 0 {
		return nil, ErrEmpty
	} else if id <= pq.levels[priority].head || id > pq.levels[priority].tail {
		return nil, ErrOutOfBounds
	}

	// Get item from database.
	var err error
	item := &PriorityItem{ID: id, Priority: priority, Key: pq.generateKey(priority, id)}
	if item.Value, err = pq.db.Get(item.Key, nil); err != nil {
		return nil, err
	}

	return item, nil
}

// generatePrefix creates the key prefix for the given priority level.
func (pq *PriorityQueue) generatePrefix(level uint8) []byte {
	// priority + prefixSep = 1 + 1 = 2
	prefix := make([]byte, 2)
	prefix[0] = byte(level)
	prefix[1] = prefixSep[0]
	return prefix
}

// generateKey create a key to be used with LevelDB.
func (pq *PriorityQueue) generateKey(priority uint8, id uint64) []byte {
	// prefix + key = 2 + 8 = 10
	key := make([]byte, 10)
	copy(key[0:2], pq.generatePrefix(priority))
	copy(key[2:], idToKey(id))
	return key
}

// init initializes the priority queue data.
func (pq *PriorityQueue) init() error {
	// Set starting value for curLevel.
	pq.resetCurrentLevel()

	// Loop through each priority level.
	for i := 0; i <= 255; i++ {
		// Create a new LevelDB Iterator for this priority level.
		prefix := pq.generatePrefix(uint8(i))
		iter := pq.db.NewIterator(util.BytesPrefix(prefix), nil)

		// Create a new priorityLevel.
		pl := &priorityLevel{
			head: 0,
			tail: 0,
		}

		// Set priority level head to the first item.
		if iter.First() {
			pl.head = keyToID(iter.Key()[2:]) - 1

			// Since this priority level has item(s), handle updating curLevel.
			if pq.cmpAsc(uint8(i)) || pq.cmpDesc(uint8(i)) {
				pq.curLevel = uint8(i)
			}
		}

		// Set priority level tail to the last item.
		if iter.Last() {
			pl.tail = keyToID(iter.Key()[2:])
		}

		if iter.Error() != nil {
			return iter.Error()
		}

		pq.levels[i] = pl
		iter.Release()
	}

	return nil
}
