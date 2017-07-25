package leadership

import (
	"testing"

	"github.com/docker/libkv/store"
	libkvmock "github.com/docker/libkv/store/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFollower(t *testing.T) {
	kv, err := libkvmock.New([]string{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, kv)

	mockStore := kv.(*libkvmock.Mock)

	kvCh := make(chan *store.KVPair)
	var mockKVCh <-chan *store.KVPair = kvCh
	mockStore.On("Watch", "test_key", mock.Anything).Return(mockKVCh, nil)

	follower := NewFollower(kv, "test_key")
	leaderCh, errCh := follower.FollowElection()

	// Simulate leader updates
	go func() {
		kvCh <- &store.KVPair{Key: "test_key", Value: []byte("leader1")}
		kvCh <- &store.KVPair{Key: "test_key", Value: []byte("leader1")}
		kvCh <- &store.KVPair{Key: "test_key", Value: []byte("leader2")}
		kvCh <- &store.KVPair{Key: "test_key", Value: []byte("leader1")}
	}()

	// We shouldn't see duplicate events.
	assert.Equal(t, <-leaderCh, "leader1")
	assert.Equal(t, <-leaderCh, "leader2")
	assert.Equal(t, <-leaderCh, "leader1")
	assert.Equal(t, follower.Leader(), "leader1")

	// Once stopped, iteration over the leader channel should stop.
	follower.Stop()
	close(kvCh)

	// Assert that we receive an error from the error chan to deal with the failover
	err, open := <-errCh
	assert.True(t, open)
	assert.NotNil(t, err)

	// Ensure that the chan is closed
	_, open = <-leaderCh
	assert.False(t, open)

	mockStore.AssertExpectations(t)
}
