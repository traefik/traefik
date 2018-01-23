package leadership

import (
	"errors"

	"github.com/abronan/valkeyrie/store"
)

// Follower can follow an election in real-time and push notifications whenever
// there is a change in leadership.
type Follower struct {
	client store.Store
	key    string

	leader   string
	leaderCh chan string
	stopCh   chan struct{}
	errCh    chan error
}

// NewFollower creates a new follower.
func NewFollower(client store.Store, key string) *Follower {
	return &Follower{
		client: client,
		key:    key,
		stopCh: make(chan struct{}),
	}
}

// Leader returns the current leader.
func (f *Follower) Leader() string {
	return f.leader
}

// FollowElection starts monitoring the election.
func (f *Follower) FollowElection() (<-chan string, <-chan error) {
	f.leaderCh = make(chan string)
	f.errCh = make(chan error)

	go f.follow()

	return f.leaderCh, f.errCh
}

// Stop stops monitoring an election.
func (f *Follower) Stop() {
	close(f.stopCh)
}

func (f *Follower) follow() {
	defer close(f.leaderCh)
	defer close(f.errCh)

	ch, err := f.client.Watch(f.key, f.stopCh, nil)
	if err != nil {
		f.errCh <- err
	}

	f.leader = ""
	for kv := range ch {
		if kv == nil {
			continue
		}
		curr := string(kv.Value)
		if curr == f.leader {
			continue
		}
		f.leader = curr
		f.leaderCh <- f.leader
	}

	// Channel closed, we return an error
	f.errCh <- errors.New("Leader Election: watch leader channel closed, the store may be unavailable...")
}
