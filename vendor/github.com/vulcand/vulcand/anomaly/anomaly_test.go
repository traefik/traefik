package anomaly

import (
	"fmt"
	"testing"
	"time"

	. "github.com/vulcand/vulcand/engine"
	. "gopkg.in/check.v1"
)

func TestAnomaly(t *testing.T) { TestingT(t) }

type AnomalySuite struct {
}

var _ = Suite(&AnomalySuite{})

func (s *AnomalySuite) TestMarkEmptyDoesNotCrash(c *C) {
	var servers []Server
	MarkServerAnomalies(servers)

	var stats []RoundTripStats
	MarkAnomalies(stats)
}

func (s *AnomalySuite) TestMarkAnomalies(c *C) {
	tc := []struct {
		Servers  []Server
		Verdicts []Verdict
	}{
		{
			Servers: []Server{
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:    time.Second,
							NetErrors: 0,
							Total:     100,
						},
					},
				},
			},
			Verdicts: []Verdict{{IsBad: false}},
		},
		{
			Servers: []Server{
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:    time.Second,
							NetErrors: 10,
							Total:     100,
						},
					},
				},
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:    time.Second,
							NetErrors: 0,
							Total:     100,
						},
					},
				},
			},
			Verdicts: []Verdict{{IsBad: true, Anomalies: []Anomaly{{Code: CodeNetErrorRate, Message: MessageNetErrRate}}}, {}},
		},
		{
			Servers: []Server{
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:      time.Second,
							Total:       100,
							StatusCodes: []StatusCode{{Code: 500, Count: 10}, {Code: 200, Count: 90}},
						},
					},
				},
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:    time.Second,
							NetErrors: 0,
							Total:     100,
						},
					},
				},
			},
			Verdicts: []Verdict{{IsBad: true, Anomalies: []Anomaly{{Code: CodeAppErrorRate, Message: MessageAppErrRate}}}, {}},
		},
		{
			Servers: []Server{
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:      time.Second,
							Total:       100,
							StatusCodes: []StatusCode{{Code: 500, Count: 10}, {Code: 200, Count: 90}},
						},
					},
				},
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
						Counters: Counters{
							Period:    time.Second,
							NetErrors: 0,
							Total:     100,
						},
					},
				},
			},
			Verdicts: []Verdict{{IsBad: true, Anomalies: []Anomaly{{Code: CodeAppErrorRate, Message: MessageAppErrRate}}}, {}},
		},
		{
			Servers: []Server{
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Second,
							},
						},
					},
				},
				Server{
					Stats: &RoundTripStats{
						LatencyBrackets: []Bracket{
							{
								Quantile: 50,
								Value:    time.Millisecond,
							},
						},
					},
				},
			},
			Verdicts: []Verdict{{IsBad: true, Anomalies: []Anomaly{{Code: CodeLatency, Message: fmt.Sprintf(MessageLatency, 50.0)}}}, {}},
		},
	}

	for i, t := range tc {
		comment := Commentf("Test case #%d", i)
		err := MarkServerAnomalies(t.Servers)
		c.Assert(err, IsNil)
		for j, e := range t.Servers {
			c.Assert(e.Stats.Verdict, DeepEquals, t.Verdicts[j], comment)
		}
	}
}
