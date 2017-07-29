// +build !race

package utils

import (
	"fmt"
	"sync"
	"testing"
)

type safeMap struct {
	mu sync.RWMutex
	m  map[int]bool
}

func (s *safeMap) Add(index int, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[index] = ok
}

func (s *safeMap) Read() map[int]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m
}

func TestInParallel(t *testing.T) {
	size := 5
	booleanMap := safeMap{
		m: make(map[int]bool, size+1),
	}
	tasks := InParallel{}
	for i := 0; i < size; i++ {
		task := func(index int) func() error {
			return func() error {
				booleanMap.Add(index, true)
				return nil
			}
		}(i)
		tasks.Add(task)
	}
	err := tasks.Wait()
	if err != nil {
		t.Fatal(err)
	}
	// Make sure every value is true
	for _, value := range booleanMap.Read() {
		if !value {
			t.Fatalf("booleanMap expected to contain only true values, got at least one false")
		}
	}
}

func TestInParallelError(t *testing.T) {
	size := 5
	booleanMap := safeMap{
		m: make(map[int]bool, size+1),
	}
	tasks := InParallel{}
	for i := 0; i < size; i++ {
		task := func(index int) func() error {
			return func() error {
				booleanMap.Add(index, false)
				t.Log("index", index)
				if index%2 == 0 {
					t.Log("return an error for", index)
					return fmt.Errorf("Error with %v", index)
				}
				booleanMap.Add(index, true)
				return nil
			}
		}(i)
		tasks.Add(task)
	}
	err := tasks.Wait()
	if err == nil {
		t.Fatalf("Expected an error on Wait, got nothing.")
	}
	for key, value := range booleanMap.Read() {
		if key%2 != 0 && !value {
			t.Fatalf("booleanMap expected to contain true values on odd number, got %v", booleanMap)
		}
	}
}
