// Copyright 2017 The etcd Authors
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

package report

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPercentiles(t *testing.T) {
	nums := make([]float64, 100)
	nums[99] = 1 // 99-percentile (1 out of 100)
	data := percentiles(nums)
	if data[len(pctls)-2] != 1 {
		t.Fatalf("99-percentile expected 1, got %f", data[len(pctls)-2])
	}

	nums = make([]float64, 1000)
	nums[999] = 1 // 99.9-percentile (1 out of 1000)
	data = percentiles(nums)
	if data[len(pctls)-1] != 1 {
		t.Fatalf("99.9-percentile expected 1, got %f", data[len(pctls)-1])
	}
}

func TestReport(t *testing.T) {
	r := NewReportSample("%f")
	go func() {
		start := time.Now()
		for i := 0; i < 5; i++ {
			end := start.Add(time.Second)
			r.Results() <- Result{Start: start, End: end}
			start = end
		}
		r.Results() <- Result{Start: start, End: start.Add(time.Second), Err: fmt.Errorf("oops")}
		close(r.Results())
	}()

	stats := <-r.Stats()
	stats.TimeSeries = nil // ignore timeseries since it uses wall clock
	wStats := Stats{
		AvgTotal:  5.0,
		Fastest:   1.0,
		Slowest:   1.0,
		Average:   1.0,
		Stddev:    0.0,
		Total:     stats.Total,
		RPS:       5.0 / stats.Total.Seconds(),
		ErrorDist: map[string]int{"oops": 1},
		Lats:      []float64{1.0, 1.0, 1.0, 1.0, 1.0},
	}
	if !reflect.DeepEqual(stats, wStats) {
		t.Fatalf("got %+v, want %+v", stats, wStats)
	}

	wstrs := []string{
		"Stddev:\t0",
		"Average:\t1.0",
		"Slowest:\t1.0",
		"Fastest:\t1.0",
	}
	ss := <-r.Run()
	for i, ws := range wstrs {
		if !strings.Contains(ss, ws) {
			t.Errorf("#%d: stats string missing %s", i, ws)
		}
	}
}

func TestWeightedReport(t *testing.T) {
	r := NewWeightedReport(NewReport("%f"), "%f")
	go func() {
		start := time.Now()
		for i := 0; i < 5; i++ {
			end := start.Add(time.Second)
			r.Results() <- Result{Start: start, End: end, Weight: 2.0}
			start = end
		}
		r.Results() <- Result{Start: start, End: start.Add(time.Second), Err: fmt.Errorf("oops")}
		close(r.Results())
	}()

	stats := <-r.Stats()
	stats.TimeSeries = nil // ignore timeseries since it uses wall clock
	wStats := Stats{
		AvgTotal:  10.0,
		Fastest:   0.5,
		Slowest:   0.5,
		Average:   0.5,
		Stddev:    0.0,
		Total:     stats.Total,
		RPS:       10.0 / stats.Total.Seconds(),
		ErrorDist: map[string]int{"oops": 1},
		Lats:      []float64{0.5, 0.5, 0.5, 0.5, 0.5},
	}
	if !reflect.DeepEqual(stats, wStats) {
		t.Fatalf("got %+v, want %+v", stats, wStats)
	}
}
