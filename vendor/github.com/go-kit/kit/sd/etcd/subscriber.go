package etcd

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/cache"
)

// Subscriber yield endpoints stored in a certain etcd keyspace. Any kind of
// change in that keyspace is watched and will update the Subscriber endpoints.
type Subscriber struct {
	client Client
	prefix string
	cache  *cache.Cache
	logger log.Logger
	quitc  chan struct{}
}

var _ sd.Subscriber = &Subscriber{}

// NewSubscriber returns an etcd subscriber. It will start watching the given
// prefix for changes, and update the endpoints.
func NewSubscriber(c Client, prefix string, factory sd.Factory, logger log.Logger) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		prefix: prefix,
		cache:  cache.New(factory, logger),
		logger: logger,
		quitc:  make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		logger.Log("prefix", s.prefix, "instances", len(instances))
	} else {
		logger.Log("prefix", s.prefix, "err", err)
	}
	s.cache.Update(instances)

	go s.loop()
	return s, nil
}

func (s *Subscriber) loop() {
	ch := make(chan struct{})
	go s.client.WatchPrefix(s.prefix, ch)
	for {
		select {
		case <-ch:
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				s.logger.Log("msg", "failed to retrieve entries", "err", err)
				continue
			}
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
