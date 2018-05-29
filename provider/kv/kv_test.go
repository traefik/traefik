package kv

import (
	"testing"
	"time"

	"github.com/abronan/valkeyrie/store"
	"github.com/containous/traefik/types"
)

func TestKvWatchTree(t *testing.T) {
	returnedChans := make(chan chan []*store.KVPair)
	provider := Provider{
		kvClient: &Mock{
			WatchTreeMethod: func() <-chan []*store.KVPair {
				c := make(chan []*store.KVPair, 10)
				returnedChans <- c
				return c
			},
		},
	}

	configChan := make(chan types.ConfigMessage)
	go func() {
		provider.watchKv(configChan, "prefix", make(chan bool, 1))
	}()

	select {
	case c1 := <-returnedChans:
		c1 <- []*store.KVPair{}
		<-configChan
		close(c1) // WatchTree chans can close due to error
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to create a new WatchTree chan")
	}

	select {
	case c2 := <-returnedChans:
		c2 <- []*store.KVPair{}
		<-configChan
	case <-time.After(1 * time.Second):
		t.Fatalf("Failed to create a new WatchTree chan")
	}

	select {
	case <-configChan:
		t.Fatalf("configChan should be empty")
	default:
	}
}
