package proxy

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/oxy/memmetrics"
	"github.com/vulcand/vulcand/engine"
)

func (mx *mux) emitMetrics() error {
	c := mx.options.MetricsClient

	// Emit connection stats
	counts := mx.connTracker.Counts()
	for state, values := range counts {
		for addr, count := range values {
			c.Gauge(c.Metric("conns", addr, state.String()), count, 1)
		}
	}

	// Emit frontend metrics stats
	frontends, err := mx.topFrontends(nil)
	if err != nil {
		log.Errorf("failed to get top frontends: %v", err)
		return err
	}
	for _, f := range frontends {
		m := c.Metric("frontend", strings.Replace(f.Id, ".", "_", -1))
		s := f.Stats
		for _, scode := range s.Counters.StatusCodes {
			// response codes counters
			c.Gauge(m.Metric("code", strconv.Itoa(scode.Code)), scode.Count, 1)
		}
		// network errors
		c.Gauge(m.Metric("neterr"), s.Counters.NetErrors, 1)
		// requests
		c.Gauge(m.Metric("reqs"), s.Counters.Total, 1)

		// round trip times in microsecond resolution
		for _, b := range s.LatencyBrackets {
			c.Gauge(m.Metric("rtt", strconv.Itoa(int(b.Quantile*10.0))), int64(b.Value/time.Microsecond), 1)
		}
	}

	return nil
}

func (mx *mux) frontendStats(key engine.FrontendKey) (*engine.RoundTripStats, error) {
	f, ok := mx.frontends[key]
	if !ok {
		return nil, fmt.Errorf("%v not found", key)
	}
	return f.watcher.rtStats()
}

func (mx *mux) backendStats(key engine.BackendKey) (*engine.RoundTripStats, error) {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		return nil, err
	}
	for _, f := range mx.frontends {
		if f.backend.backend.Id != key.Id {
			continue
		}
		if err := f.watcher.collectMetrics(m); err != nil {
			return nil, err
		}
	}
	return engine.NewRoundTripStats(m)
}

func (mx *mux) serverStats(key engine.ServerKey) (*engine.RoundTripStats, error) {
	b, ok := mx.backends[key.BackendKey]
	if !ok {
		return nil, fmt.Errorf("%v not found", key.BackendKey)
	}
	srv, ok := b.findServer(key)
	if !ok {
		return nil, fmt.Errorf("%v not found", key)
	}

	u, err := url.Parse(srv.URL)
	if err != nil {
		return nil, err
	}

	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		return nil, err
	}
	for _, f := range mx.frontends {
		if f.backend.backend.Id != key.BackendKey.Id {
			continue
		}
		if err := f.watcher.collectServerMetrics(m, u); err != nil {
			return nil, err
		}
	}
	return engine.NewRoundTripStats(m)
}

func (mx *mux) topFrontends(key *engine.BackendKey) ([]engine.Frontend, error) {
	frontends := []engine.Frontend{}
	for _, m := range mx.frontends {
		if key != nil && key.Id != m.backend.backend.Id {
			continue
		}
		f := m.frontend
		stats, err := m.watcher.rtStats()
		if err != nil {
			return nil, err
		}
		f.Stats = stats
		frontends = append(frontends, f)
	}
	sort.Stable(&frontendSorter{frontends: frontends})
	return frontends, nil
}

func (mx *mux) topServers(key *engine.BackendKey) ([]engine.Server, error) {
	metrics := map[string]*sval{}
	for _, f := range mx.frontends {
		if key != nil && key.Id != f.backend.backend.Id {
			continue
		}
		for _, s := range f.backend.servers {
			val, ok := metrics[s.URL]
			if !ok {
				sval, err := newSval(s)
				if err != nil {
					return nil, err
				}
				metrics[s.URL] = sval
				val = sval
			}
			if err := f.watcher.collectServerMetrics(val.m, val.u); err != nil {
				return nil, err
			}
		}
	}
	servers := make([]engine.Server, 0, len(metrics))
	for _, v := range metrics {
		stats, err := engine.NewRoundTripStats(v.m)
		if err != nil {
			return nil, err
		}
		v.srv.Stats = stats
		servers = append(servers, *v.srv)
	}
	sort.Stable(&serverSorter{es: servers})
	return servers, nil
}

type sval struct {
	u   *url.URL
	srv *engine.Server
	m   *memmetrics.RTMetrics
}

func newSval(s engine.Server) (*sval, error) {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(s.URL)
	if err != nil {
		return nil, err
	}
	return &sval{srv: &s, m: m, u: u}, nil
}

type frontendSorter struct {
	frontends []engine.Frontend
}

func (s *frontendSorter) Len() int {
	return len(s.frontends)
}

func (s *frontendSorter) Swap(i, j int) {
	s.frontends[i], s.frontends[j] = s.frontends[j], s.frontends[i]
}

func (s *frontendSorter) Less(i, j int) bool {
	return cmpStats(s.frontends[i].Stats, s.frontends[j].Stats)
}

type serverSorter struct {
	es []engine.Server
}

func (s *serverSorter) Len() int {
	return len(s.es)
}

func (s *serverSorter) Swap(i, j int) {
	s.es[i], s.es[j] = s.es[j], s.es[i]
}

func (s *serverSorter) Less(i, j int) bool {
	return cmpStats(s.es[i].Stats, s.es[j].Stats)
}

func cmpStats(s1, s2 *engine.RoundTripStats) bool {
	// Items that have network errors go first
	if s1.NetErrorRatio() != 0 || s2.NetErrorRatio() != 0 {
		return s1.NetErrorRatio() > s2.NetErrorRatio()
	}

	// Items that have application level errors go next
	if s1.AppErrorRatio() != 0 || s2.AppErrorRatio() != 0 {
		return s1.AppErrorRatio() > s2.AppErrorRatio()
	}

	// More highly loaded items go next
	return s1.Counters.Total > s2.Counters.Total
}
