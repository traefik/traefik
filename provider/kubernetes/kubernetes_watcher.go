package kubernetes

import (
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
)

// Watcher contains the channels required to interface with kubernetes
type Watcher struct {
	EventsChan    chan interface{}
	NamespaceChan chan interface{}
	ErrChan       chan error
	stopChan      chan struct{}
	refreshChan   chan struct{}
	k8sClient     Client
	namespaces    Namespaces
}

// NewWatcher returns an initialized watcher
func NewWatcher(namespaces Namespaces, client Client) *Watcher {
	return &Watcher{
		EventsChan:    make(chan interface{}),
		NamespaceChan: make(chan interface{}),
		ErrChan:       make(chan error),
		stopChan:      make(chan struct{}),
		refreshChan:   make(chan struct{}),
		namespaces:    namespaces,
		k8sClient:     client,
	}
}

// Watch contains the main watch loop
func (w *Watcher) Watch() {
	err := w.k8sClient.WatchNamespaces(w.namespaces, w.stopChan, w.NamespaceChan)
	if err != nil {
		log.Errorf("Error watching kubernetes namespace events: %s", err)
		w.ErrChan <- err
		return
	}

	safe.Go(func() {
		for {
			stopWatch := make(chan struct{}, 1)
			err = w.k8sClient.WatchAll(w.namespaces, stopWatch, w.EventsChan)
			if err != nil {
				log.Errorf("Error watching kubernetes namespace events: %s", err)
				w.ErrChan <- err
				close(stopWatch)
				return
			}
			select {
			case <-w.stopChan:
				close(stopWatch)
				return
			case <-w.refreshChan:
				close(stopWatch)
			}
		}
	})

}

// Refresh sends a message to the refresh chan
func (w *Watcher) Refresh() {
	w.refreshChan <- struct{}{}
}

// Stop closes all contained chans
func (w *Watcher) Stop() {
	close(w.stopChan)
	close(w.EventsChan)
	close(w.NamespaceChan)
	close(w.refreshChan)
	close(w.ErrChan)
}
