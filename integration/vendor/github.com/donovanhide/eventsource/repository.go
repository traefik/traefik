package eventsource

import (
	"sort"
	"sync"
)

// Example repository that uses a slice as storage for past events.
type SliceRepository struct {
	events map[string][]Event
	lock   sync.RWMutex
}

func NewSliceRepository() *SliceRepository {
	return &SliceRepository{
		events: make(map[string][]Event),
	}
}

func (repo SliceRepository) indexOfEvent(channel, id string) int {
	return sort.Search(len(repo.events[channel]), func(i int) bool {
		return repo.events[channel][i].Id() >= id
	})
}

func (repo SliceRepository) Replay(channel, id string) (out chan Event) {
	out = make(chan Event)
	go func() {
		defer close(out)
		repo.lock.RLock()
		defer repo.lock.RUnlock()
		events := repo.events[channel][repo.indexOfEvent(channel, id):]
		for i := range events {
			out <- events[i]
		}
	}()
	return
}

func (repo *SliceRepository) Add(channel string, event Event) {
	repo.lock.Lock()
	defer repo.lock.Unlock()
	i := repo.indexOfEvent(channel, event.Id())
	if i < len(repo.events[channel]) && repo.events[channel][i].Id() == event.Id() {
		repo.events[channel][i] = event
	} else {
		repo.events[channel] = append(repo.events[channel][:i], append([]Event{event}, repo.events[channel][i:]...)...)
	}
	return
}
