package goque

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

// Item represents an entry in either a stack or queue.
type Item struct {
	ID    uint64
	Key   []byte
	Value []byte
}

// ToString returns the item value as a string.
func (i *Item) ToString() string {
	return string(i.Value)
}

// ToObject decodes the item value into the given value type using
// encoding/gob.
//
// The value passed to this method should be a pointer to a variable
// of the type you wish to decode into. The variable pointed to will
// hold the decoded object.
func (i *Item) ToObject(value interface{}) error {
	buffer := bytes.NewBuffer(i.Value)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(value)
}

// PriorityItem represents an entry in a priority queue.
type PriorityItem struct {
	ID       uint64
	Priority uint8
	Key      []byte
	Value    []byte
}

// ToString returns the priority item value as a string.
func (pi *PriorityItem) ToString() string {
	return string(pi.Value)
}

// ToObject decodes the item value into the given value type using
// encoding/gob.
//
// The value passed to this method should be a pointer to a variable
// of the type you wish to decode into. The variable pointed to will
// hold the decoded object.
func (pi *PriorityItem) ToObject(value interface{}) error {
	buffer := bytes.NewBuffer(pi.Value)
	dec := gob.NewDecoder(buffer)
	return dec.Decode(value)
}

// idToKey converts and returns the given ID to a key.
func idToKey(id uint64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)
	return key
}

// keyToID converts and returns the given key to an ID.
func keyToID(key []byte) uint64 {
	return binary.BigEndian.Uint64(key)
}
