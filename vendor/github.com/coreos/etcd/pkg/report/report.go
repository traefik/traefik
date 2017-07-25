// Copyright 2014 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// the file is borrowed from github.com/rakyll/boom/boomer/print.go

package report

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	barChar = "∎"
)

// Result describes the timings for an operation.
type Result struct {
	Start  time.Time
	End    time.Time
	Err    error
	Weight float64
}

func (res *Result) Duration() time.Duration { return res.End.Sub(res.Start) }

type report struct {
	results   chan Result
	precision string

	stats Stats
	sps   *secondPoints
}

// Stats exposes results raw data.
type Stats struct {
	AvgTotal   float64
	Fastest    float64
	Slowest    float64
	Average    float64
	Stddev     float64
	RPS        float64
	Total      time.Duration
	ErrorDist  map[string]int
	Lats       []float64
	TimeSeries TimeSeries
}

func (s *Stats) copy() Stats {
	ss := *s
	ss.ErrorDist = copyMap(ss.ErrorDist)
	ss.Lats = copyFloats(ss.Lats)
	return ss
}

// Report processes a result stream until it is closed, then produces a
// string with information about the consumed result data.
type Report interface {
	Results() chan<- Result

	// Run returns results in print-friendly format.
	Run() <-chan string

	// Stats returns results in raw data.
	Stats() <-chan Stats
}

func NewReport(precision string) Report { return newReport(precision) }

func newReport(precision string) *report {
	r := &report{
		results:   make(chan Result, 16),
		precision: precision,
	}
	r.stats.ErrorDist = make(map[string]int)
	return r
}

func NewReportSample(precision string) Report {
	r := NewReport(precision).(*report)
	r.sps = newSecondPoints()
	return r
}

func (r *report) Results() chan<- Result { return r.results }

func (r *report) Run() <-chan string {
	donec := make(chan string, 1)
	go func() {
		defer close(donec)
		r.processResults()
		donec <- r.String()
	}()
	return donec
}

func (r *report) Stats() <-chan Stats {
	donec := make(chan Stats, 1)
	go func() {
		defer close(donec)
		r.processResults()
		s := r.stats.copy()
		if r.sps != nil {
			s.TimeSeries = r.sps.getTimeSeries()
		}
		donec <- s
	}()
	return donec
}

func copyMap(m map[string]int) (c map[string]int) {
	c = make(map[string]int, len(m))
	for k, v := range m {
		c[k] = v
	}
	return
}

func copyFloats(s []float64) (c []float64) {
	c = make([]float64, len(s))
	copy(c, s)
	return
}

func (r *report) String() (s string) {
	if len(r.stats.Lats) > 0 {
		s += fmt.Sprintf("\nSummary:\n")
		s += fmt.Sprintf("  Total:\t%s.\n", r.sec2str(r.stats.Total.Seconds()))
		s += fmt.Sprintf("  Slowest:\t%s.\n", r.sec2str(r.stats.Slowest))
		s += fmt.Sprintf("  Fastest:\t%s.\n", r.sec2str(r.stats.Fastest))
		s += fmt.Sprintf("  Average:\t%s.\n", r.sec2str(r.stats.Average))
		s += fmt.Sprintf("  Stddev:\t%s.\n", r.sec2str(r.stats.Stddev))
		s += fmt.Sprintf("  Requests/sec:\t"+r.precision+"\n", r.stats.RPS)
		s += r.histogram()
		s += r.sprintLatencies()
		if r.sps != nil {
			s += fmt.Sprintf("%v\n", r.sps.getTimeSeries())
		}
	}
	if len(r.stats.ErrorDist) > 0 {
		s += r.errors()
	}
	return s
}

func (r *report) sec2str(sec float64) string { return fmt.Sprintf(r.precision+" secs", sec) }

type reportRate struct{ *report }

func NewReportRate(precision string) Report {
	return &reportRate{NewReport(precision).(*report)}
}

func (r *reportRate) String() string {
	return fmt.Sprintf(" Requests/sec:\t"+r.precision+"\n", r.stats.RPS)
}

func (r *report) processResult(res *Result) {
	if res.Err != nil {
		r.stats.ErrorDist[res.Err.Error()]++
		return
	}
	dur := res.Duration()
	r.stats.Lats = append(r.stats.Lats, dur.Seconds())
	r.stats.AvgTotal += dur.Seconds()
	if r.sps != nil {
		r.sps.Add(res.Start, dur)
	}
}

func (r *report) processResults() {
	st := time.Now()
	for res := range r.results {
		r.processResult(&res)
	}
	r.stats.Total = time.Since(st)

	r.stats.RPS = float64(len(r.stats.Lats)) / r.stats.Total.Seconds()
	r.stats.Average = r.stats.AvgTotal / float64(len(r.stats.Lats))
	for i := range r.stats.Lats {
		dev := r.stats.Lats[i] - r.stats.Average
		r.stats.Stddev += dev * dev
	}
	r.stats.Stddev = math.Sqrt(r.stats.Stddev / float64(len(r.stats.Lats)))
	sort.Float64s(r.stats.Lats)
	if len(r.stats.Lats) > 0 {
		r.stats.Fastest = r.stats.Lats[0]
		r.stats.Slowest = r.stats.Lats[len(r.stats.Lats)-1]
	}
}

var pctls = []float64{10, 25, 50, 75, 90, 95, 99, 99.9}

// Percentiles returns percentile distribution of float64 slice.
func Percentiles(nums []float64) (pcs []float64, data []float64) {
	return pctls, percentiles(nums)
}

func percentiles(nums []float64) (data []float64) {
	data = make([]float64, len(pctls))
	j := 0
	n := len(nums)
	for i := 0; i < n && j < len(pctls); i++ {
		current := float64(i) * 100.0 / float64(n)
		if current >= pctls[j] {
			data[j] = nums[i]
			j++
		}
	}
	return
}

func (r *report) sprintLatencies() string {
	data := percentiles(r.stats.Lats)
	s := fmt.Sprintf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			s += fmt.Sprintf("  %v%% in %s.\n", pctls[i], r.sec2str(data[i]))
		}
	}
	return s
}

func (r *report) histogram() string {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (r.stats.Slowest - r.stats.Fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = r.stats.Fastest + bs*float64(i)
	}
	buckets[bc] = r.stats.Slowest
	var bi int
	var max int
	for i := 0; i < len(r.stats.Lats); {
		if r.stats.Lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	s := fmt.Sprintf("\nResponse time histogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = counts[i] * 40 / max
		}
		s += fmt.Sprintf("  "+r.precision+" [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat(barChar, barLen))
	}
	return s
}

func (r *report) errors() string {
	s := fmt.Sprintf("\nError distribution:\n")
	for err, num := range r.stats.ErrorDist {
		s += fmt.Sprintf("  [%d]\t%s\n", num, err)
	}
	return s
}
