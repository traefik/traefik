package kubernetes

import (
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
)

type KubernetesWatcher struct {
	EventsChan    chan interface{}
	NamespaceChan chan interface{}
	ErrChan       chan error
	stopChan      chan struct{}
	refreshChan   chan struct{}
	k8sClient     Client
	namespaces    Namespaces
}

func NewKubernetesWatcher(namespaces Namespaces, client Client) *KubernetesWatcher {
	return &KubernetesWatcher{
		EventsChan:    make(chan interface{}),
		NamespaceChan: make(chan interface{}),
		ErrChan:       make(chan error),
		stopChan:      make(chan struct{}),
		refreshChan:   make(chan struct{}),
		namespaces:    namespaces,
		k8sClient:     client,
	}
}

func (k *KubernetesWatcher) Watch() {
	err := k.k8sClient.WatchNamespaces(k.namespaces, k.stopChan, k.NamespaceChan)
	if err != nil {
		log.Errorf("Error watching kubernetes namespace events: %s", err)
		k.ErrChan <- err
		return
	}

	safe.Go(func() {
		for {
			stopWatch := make(chan struct{}, 1)
			err = k.k8sClient.WatchAll(k.namespaces, stopWatch, k.EventsChan)
			if err != nil {
				log.Errorf("Error watching kubernetes namespace events: %s", err)
				k.ErrChan <- err
				close(stopWatch)
				return
			}
			select {
			case <-k.stopChan:
				close(stopWatch)
				return
			case <-k.refreshChan:
				close(stopWatch)
			}
		}
	})

}

func (k *KubernetesWatcher) Refresh() {
	k.refreshChan <- struct{}{}
}

func (k *KubernetesWatcher) Stop() {
	close(k.stopChan)
	close(k.EventsChan)
	close(k.NamespaceChan)
	close(k.refreshChan)
	close(k.ErrChan)
}
