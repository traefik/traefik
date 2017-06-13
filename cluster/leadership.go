package cluster

import (
	"context"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/docker/leadership"
)

// Leadership allows leadership election using a KV store
type Leadership struct {
	*safe.Pool
	*types.Cluster
	candidate *leadership.Candidate
	leader    *safe.Safe
	listeners []LeaderListener
}

// NewLeadership creates a leadership
func NewLeadership(ctx context.Context, cluster *types.Cluster) *Leadership {
	return &Leadership{
		Pool:      safe.NewPool(ctx),
		Cluster:   cluster,
		candidate: leadership.NewCandidate(cluster.Store, cluster.Store.Prefix+"/leader", cluster.Node, 20*time.Second),
		listeners: []LeaderListener{},
		leader:    safe.New(false),
	}
}

// LeaderListener is called when leadership has changed
type LeaderListener func(elected bool) error

// Participate tries to be a leader
func (l *Leadership) Participate(pool *safe.Pool) {
	pool.GoCtx(func(ctx context.Context) {
		log.Debugf("Node %s running for election", l.Cluster.Node)
		defer log.Debugf("Node %s no more running for election", l.Cluster.Node)
		backOff := backoff.NewExponentialBackOff()
		operation := func() error {
			return l.run(ctx, l.candidate)
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Leadership election error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backOff, notify)
		if err != nil {
			log.Errorf("Cannot elect leadership %+v", err)
		}
	})
}

// AddListener adds a leadership listener
func (l *Leadership) AddListener(listener LeaderListener) {
	l.listeners = append(l.listeners, listener)
}

// Resign resigns from being a leader
func (l *Leadership) Resign() {
	l.candidate.Resign()
	log.Infof("Node %s resigned", l.Cluster.Node)
}

func (l *Leadership) run(ctx context.Context, candidate *leadership.Candidate) error {
	electedCh, errCh := candidate.RunForElection()
	for {
		select {
		case elected := <-electedCh:
			l.onElection(elected)
		case err := <-errCh:
			return err
		case <-ctx.Done():
			l.candidate.Resign()
			return nil
		}
	}
}

func (l *Leadership) onElection(elected bool) {
	if elected {
		log.Infof("Node %s elected leader ♚", l.Cluster.Node)
		l.leader.Set(true)
		l.Start()
	} else {
		log.Infof("Node %s elected worker ♝", l.Cluster.Node)
		l.leader.Set(false)
		l.Stop()
	}
	for _, listener := range l.listeners {
		err := listener(elected)
		if err != nil {
			log.Errorf("Error calling Leadership listener: %s", err)
		}
	}
}

// IsLeader returns true if current node is leader
func (l *Leadership) IsLeader() bool {
	return l.leader.Get().(bool)
}
