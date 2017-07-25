package command

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/buger/goterm"
	"github.com/vulcand/vulcand/engine"
)

func frontendsOverview(frontends []engine.Frontend) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tRoute\tR/sec\t50ile[ms]\t95ile[ms]\t99ile[ms]\tStatus codes %%\tNet. errors %%\n")

	if len(frontends) == 0 {
		return t.String()
	}
	for _, l := range frontends {
		frontendOverview(t, l)
	}
	return t.String()
}

func serversOverview(servers []engine.Server) string {
	t := goterm.NewTable(0, 10, 5, ' ', 0)
	fmt.Fprint(t, "Id\tURL\tReqs/sec\t50ile[ms]\t95ile[ms]\t99ile[ms]\tStatus codes %%\tNet. errors %%\tMessages\n")

	for _, e := range servers {
		serverOverview(t, e)
	}
	return t.String()
}

func frontendOverview(w io.Writer, l engine.Frontend) {
	s := l.Stats

	fmt.Fprintf(w, "%s\t%s\t%0.1f\t%0.2f\t%0.2f\t%0.2f\t%s\t%s\n",
		l.Id,
		l.Route,
		s.RequestsPerSecond(),
		latencyAtQuantile(50.0, s),
		latencyAtQuantile(95.0, s),
		latencyAtQuantile(99.0, s),
		statusCodesToString(s),
		errRatioToString(s.NetErrorRatio()),
	)
}

func serverOverview(w io.Writer, srv engine.Server) {
	s := srv.Stats

	anomalies := ""
	if s.Verdict.IsBad {
		anomalies = fmt.Sprintf("%v", s.Verdict.Anomalies)
	}

	fmt.Fprintf(w, "%s\t%s\t%0.1f\t%0.2f\t%0.2f\t%0.2f\t%s\t%s\t%s\n",
		srv.Id,
		srv.URL,
		s.RequestsPerSecond(),
		latencyAtQuantile(50.0, s),
		latencyAtQuantile(95.0, s),
		latencyAtQuantile(99.0, s),
		statusCodesToString(s),
		errRatioToString(s.NetErrorRatio()),
		anomalies)
}

func latencyAtQuantile(q float64, s *engine.RoundTripStats) float64 {
	v, err := s.LatencyBrackets.GetQuantile(q)
	if err != nil {
		log.Errorf("Failed to get latency %f from %v, err: %v", q, s, err)
		return -1
	}
	return float64(v.Value) / float64(time.Millisecond)
}

func errRatioToString(r float64) string {
	failRatioS := fmt.Sprintf("%0.2f", r*100)
	if r != 0 {
		return goterm.Color(failRatioS, goterm.RED)
	} else {
		return goterm.Color(failRatioS, goterm.GREEN)
	}
}

func statusCodesToString(s *engine.RoundTripStats) string {
	if s.Counters.Total == 0 {
		return ""
	}

	sort.Sort(&codeSorter{codes: s.Counters.StatusCodes})

	codes := make([]string, 0, len(s.Counters.StatusCodes))
	for _, c := range s.Counters.StatusCodes {
		percent := 100 * (float64(c.Count) / float64(s.Counters.Total))
		out := fmt.Sprintf("%d: %0.2f", c.Code, percent)
		codes = append(codes, out)
	}

	return strings.Join(codes, ", ")
}

func getColor(code int) int {
	if code < 300 {
		return goterm.GREEN
	} else if code < 500 {
		return goterm.YELLOW
	}
	return goterm.RED
}

type codeSorter struct {
	codes []engine.StatusCode
}

func (c *codeSorter) Len() int {
	return len(c.codes)
}

func (c *codeSorter) Swap(i, j int) {
	c.codes[i], c.codes[j] = c.codes[j], c.codes[i]
}

func (c *codeSorter) Less(i, j int) bool {
	return c.codes[i].Code < c.codes[j].Code
}
