package cmap

import (
	"encoding/json"
	"hash/fnv"
	"sort"
	"strconv"
	"testing"
)

type Animal struct {
	name string
}

func TestMapCreation(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("map is null.")
	}

	if m.Count() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestInsert(t *testing.T) {
	m := New()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.Set("elephant", elephant)
	m.Set("monkey", monkey)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestInsertAbsent(t *testing.T) {
	m := New()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.SetIfAbsent("elephant", elephant)
	if ok := m.SetIfAbsent("elephant", monkey); ok {
		t.Error("map set a new value even the entry is already present")
	}
}

func TestGet(t *testing.T) {
	m := New()

	// Get a missing element.
	val, ok := m.Get("Money")

	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	if val != nil {
		t.Error("Missing values should return as null.")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	// Retrieve inserted element.

	tmp, ok := m.Get("elephant")
	elephant = tmp.(Animal) // Type assertion.

	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	if &elephant == nil {
		t.Error("expecting an element, not null.")
	}

	if elephant.name != "elephant" {
		t.Error("item was modified.")
	}
}

func TestHas(t *testing.T) {
	m := New()

	// Get a missing element.
	if m.Has("Money") == true {
		t.Error("element shouldn't exists")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	if m.Has("elephant") == false {
		t.Error("element exists, expecting Has to return True.")
	}
}

func TestRemove(t *testing.T) {
	m := New()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	m.Remove("monkey")

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was removed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if temp != nil {
		t.Error("Expecting item to be nil after its removal.")
	}

	// Remove a none existing element.
	m.Remove("noone")
}

func TestPop(t *testing.T) {
	m := New()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	v, exists := m.Pop("monkey")

	if !exists {
		t.Error("Pop didn't find a monkey.")
	}

	m1, ok := v.(Animal)

	if !ok || m1 != monkey {
		t.Error("Pop found something else, but monkey.")
	}

	v2, exists2 := m.Pop("monkey")
	m1, ok = v2.(Animal)

	if exists2 || ok || m1 == monkey {
		t.Error("Pop keeps finding monkey")
	}

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was Pop'ed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if temp != nil {
		t.Error("Expecting item to be nil after its removal.")
	}
}

func TestCount(t *testing.T) {
	m := New()
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	if m.Count() != 100 {
		t.Error("Expecting 100 element within map.")
	}
}

func TestIsEmpty(t *testing.T) {
	m := New()

	if m.IsEmpty() == false {
		t.Error("new map should be empty")
	}

	m.Set("elephant", Animal{"elephant"})

	if m.IsEmpty() != false {
		t.Error("map shouldn't be empty.")
	}
}

func TestIterator(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	counter := 0
	// Iterate over elements.
	for item := range m.Iter() {
		val := item.Val

		if val == nil {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestBufferedIterator(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	counter := 0
	// Iterate over elements.
	for item := range m.IterBuffered() {
		val := item.Val

		if val == nil {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestIterCb(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	counter := 0
	// Iterate over elements.
	m.IterCb(func(key string, v interface{}) {
		_, ok := v.(Animal)
		if !ok {
			t.Error("Expecting an animal object")
		}

		counter++
	})
	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestItems(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	items := m.Items()

	if len(items) != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestConcurrent(t *testing.T) {
	m := New()
	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	// Using go routines insert 1000 ints into our map.
	go func() {
		for i := 0; i < iterations/2; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val.(int)
		} // Call go routine with current index.
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val.(int)
		} // Call go routine with current index.
	}()

	// Wait for all go routines to finish.
	counter := 0
	for elem := range ch {
		a[counter] = elem
		counter++
		if counter == iterations {
			break
		}
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(a[0:iterations])

	// Make sure map contains 1000 elements.
	if m.Count() != iterations {
		t.Error("Expecting 1000 elements.")
	}

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		if i != a[i] {
			t.Error("missing value", i)
		}
	}
}

func TestJsonMarshal(t *testing.T) {
	SHARD_COUNT = 2
	defer func() {
		SHARD_COUNT = 32
	}()
	expected := "{\"a\":1,\"b\":2}"
	m := New()
	m.Set("a", 1)
	m.Set("b", 2)
	j, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(j) != expected {
		t.Error("json", string(j), "differ from expected", expected)
		return
	}
}

func TestKeys(t *testing.T) {
	m := New()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	keys := m.Keys()
	if len(keys) != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestMInsert(t *testing.T) {
	animals := map[string]interface{}{
		"elephant": Animal{"elephant"},
		"monkey":   Animal{"monkey"},
	}
	m := New()
	m.MSet(animals)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestFnv32(t *testing.T) {
	key := []byte("ABC")

	hasher := fnv.New32()
	hasher.Write(key)
	if fnv32(string(key)) != hasher.Sum32() {
		t.Errorf("Bundled fnv32 produced %d, expected result from hash/fnv32 is %d", fnv32(string(key)), hasher.Sum32())
	}
}

func TestUpsert(t *testing.T) {
	dolphin := Animal{"dolphin"}
	whale := Animal{"whale"}
	tiger := Animal{"tiger"}
	lion := Animal{"lion"}

	cb := func(exists bool, valueInMap interface{}, newValue interface{}) interface{} {
		nv := newValue.(Animal)
		if !exists {
			return []Animal{nv}
		}
		res := valueInMap.([]Animal)
		return append(res, nv)
	}

	m := New()
	m.Set("marine", []Animal{dolphin})
	m.Upsert("marine", whale, cb)
	m.Upsert("predator", tiger, cb)
	m.Upsert("predator", lion, cb)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}

	compare := func(a, b []Animal) bool {
		if a == nil || b == nil {
			return false
		}

		if len(a) != len(b) {
			return false
		}

		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	marineAnimals, ok := m.Get("marine")
	if !ok || !compare(marineAnimals.([]Animal), []Animal{dolphin, whale}) {
		t.Error("Set, then Upsert failed")
	}

	predators, ok := m.Get("predator")
	if !ok || !compare(predators.([]Animal), []Animal{tiger, lion}) {
		t.Error("Upsert, then Upsert failed")
	}
}
