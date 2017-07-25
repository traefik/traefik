package zk

import (
	"github.com/samuel/go-zookeeper/zk"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/cache"
)

// Subscriber yield endpoints stored in a certain ZooKeeper path. Any kind of
// change in that path is watched and will update the Subscriber endpoints.
type Subscriber struct {
	client Client
	path   string
	cache  *cache.Cache
	logger log.Logger
	quitc  chan struct{}
}

var _ sd.Subscriber = &Subscriber{}

// NewSubscriber returns a ZooKeeper subscriber. ZooKeeper will start watching
// the given path for changes and update the Subscriber endpoints.
func NewSubscriber(c Client, path string, factory sd.Factory, logger log.Logger) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		path:   path,
		cache:  cache.New(factory, logger),
		logger: logger,
		quitc:  make(chan struct{}),
	}

	err := s.client.CreateParentNodes(s.path)
	if err != nil {
		return nil, err
	}

	instances, eventc, err := s.client.GetEntries(s.path)
	if err != nil {
		logger.Log("path", s.path, "msg", "failed to retrieve entries", "err", err)
		return nil, err
	}
	logger.Log("path", s.path, "instances", len(instances))
	s.cache.Update(instances)

	go s.loop(eventc)

	return s, nil
}

func (s *Subscriber) loop(eventc <-chan zk.Event) {
	var (
		instances []string
		err       error
	)
	for {
		select {
		case <-eventc:
			// We received a path update notification. Call GetEntries to
			// retrieve child node data, and set a new watch, as ZK watches are
			// one-time triggers.
			instances, eventc, err = s.client.GetEntries(s.path)
			if err != nil {
				s.logger.Log("path", s.path, "msg", "failed to retrieve entries", "err", err)
				continue
			}
			s.logger.Log("path", s.path, "instances", len(instances))
			s.cache.Update(instances)

		case <-s.quitc:
			return
		}
	}
}

// Endpoints implements the Subscriber interface.
func (s *Subscriber) Endpoints() ([]endpoint.Endpoint, error) {
	return s.cache.Endpoints(), nil
}

// Stop terminates the Subscriber.
func (s *Subscriber) Stop() {
	close(s.quitc)
}
