package cluster

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/leadership"
	"time"
)

// Leadership allows leadership election using a KV store
type Leadership struct {
	types.Cluster
	candidate *leadership.Candidate
}

// Participate tries to be a leader
func (l *Leadership) Participate(pool *safe.Pool, isElected func(bool)) {
	pool.Go(func(stop chan bool) {
		l.candidate = leadership.NewCandidate(l.Store, l.Store.Prefix+"/leader", l.Node, 30*time.Second)
		backOff := backoff.NewExponentialBackOff()
		operation := func() error {
			return l.run(l.candidate, stop, isElected)
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Leadership election error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, backOff, notify)
		if err != nil {
			log.Errorf("Cannot elect leadership %+v", err)
		}
	})
}

// Resign resigns from being a leader
func (l *Leadership) Resign() {
	if l.candidate != nil {
		l.candidate.Resign()
	}
}

func (l *Leadership) run(candidate *leadership.Candidate, stop chan bool, isElected func(bool)) error {
	electedCh, errCh := candidate.RunForElection()
	for {
		select {
		case elected := <-electedCh:
			isElected(elected)
		case err := <-errCh:
			return err
		case <-stop:
			return nil
		}
	}
}
