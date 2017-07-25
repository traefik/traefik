package circonus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/circonus-labs/circonus-gometrics"
	"github.com/circonus-labs/circonus-gometrics/checkmgr"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestCounter(t *testing.T) {
	// The only way to extract values from Circonus is to pose as a Circonus
	// server and receive real HTTP writes.
	const name = "abc"
	var val int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res map[string]struct {
			Value int64 `json:"_value"` // reverse-engineered :\
		}
		json.NewDecoder(r.Body).Decode(&res)
		val = res[name].Value
	}))
	defer s.Close()

	// Set up a Circonus object, submitting to our HTTP server.
	m := newCirconusMetrics(s.URL)
	counter := New(m).NewCounter(name).With("label values", "not supported")
	value := func() float64 { m.Flush(); return float64(val) }

	// Engage.
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestGauge(t *testing.T) {
	const name = "def"
	var val float64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res map[string]struct {
			Value string `json:"_value"`
		}
		json.NewDecoder(r.Body).Decode(&res)
		val, _ = strconv.ParseFloat(res[name].Value, 64)
	}))
	defer s.Close()

	m := newCirconusMetrics(s.URL)
	gauge := New(m).NewGauge(name).With("label values", "not supported")
	value := func() float64 { m.Flush(); return val }

	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	const name = "ghi"

	// Circonus just emits bucketed counts. We'll dump them into a generic
	// histogram (losing some precision) and take statistics from there. Note
	// this does assume that the generic histogram computes statistics properly,
	// but we have another test for that :)
	re := regexp.MustCompile(`^H\[([0-9\.e\+]+)\]=([0-9]+)$`) // H[1.2e+03]=456

	var p50, p90, p95, p99 float64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res map[string]struct {
			Values []string `json:"_value"` // reverse-engineered :\
		}
		json.NewDecoder(r.Body).Decode(&res)

		h := generic.NewHistogram("dummy", len(res[name].Values)) // match tbe bucket counts
		for _, v := range res[name].Values {
			match := re.FindStringSubmatch(v)
			f, _ := strconv.ParseFloat(match[1], 64)
			n, _ := strconv.ParseInt(match[2], 10, 64)
			for i := int64(0); i < n; i++ {
				h.Observe(f)
			}
		}

		p50 = h.Quantile(0.50)
		p90 = h.Quantile(0.90)
		p95 = h.Quantile(0.95)
		p99 = h.Quantile(0.99)
	}))
	defer s.Close()

	m := newCirconusMetrics(s.URL)
	histogram := New(m).NewHistogram(name).With("label values", "not supported")
	quantiles := func() (float64, float64, float64, float64) { m.Flush(); return p50, p90, p95, p99 }

	// Circonus metrics, because they do their own bucketing, are less precise
	// than other systems. So, we bump the tolerance to 5 percent.
	if err := teststat.TestHistogram(histogram, quantiles, 0.05); err != nil {
		t.Fatal(err)
	}
}

func newCirconusMetrics(url string) *circonusgometrics.CirconusMetrics {
	m, err := circonusgometrics.NewCirconusMetrics(&circonusgometrics.Config{
		CheckManager: checkmgr.Config{
			Check: checkmgr.CheckConfig{
				SubmissionURL: url,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return m
}
