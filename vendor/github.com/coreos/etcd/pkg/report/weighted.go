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

// the file is borrowed from github.com/rakyll/boom/boomer/print.go

package report

import (
	"time"
)

type weightedReport struct {
	baseReport Report

	report      *report
	results     chan Result
	weightTotal float64
}

// NewWeightedReport returns a report that includes
// both weighted and unweighted statistics.
func NewWeightedReport(r Report, precision string) Report {
	return &weightedReport{
		baseReport: r,
		report:     newReport(precision),
		results:    make(chan Result, 16),
	}
}

func (wr *weightedReport) Results() chan<- Result { return wr.results }

func (wr *weightedReport) Run() <-chan string {
	donec := make(chan string, 2)
	go func() {
		defer close(donec)
		basec, rc := make(chan string, 1), make(chan Stats, 1)
		go func() { basec <- (<-wr.baseReport.Run()) }()
		go func() { rc <- (<-wr.report.Stats()) }()
		go wr.processResults()
		wr.report.stats = wr.reweighStat(<-rc)
		donec <- wr.report.String()
		donec <- (<-basec)
	}()
	return donec
}

func (wr *weightedReport) Stats() <-chan Stats {
	donec := make(chan Stats, 2)
	go func() {
		defer close(donec)
		basec, rc := make(chan Stats, 1), make(chan Stats, 1)
		go func() { basec <- (<-wr.baseReport.Stats()) }()
		go func() { rc <- (<-wr.report.Stats()) }()
		go wr.processResults()
		donec <- wr.reweighStat(<-rc)
		donec <- (<-basec)
	}()
	return donec
}

func (wr *weightedReport) processResults() {
	defer close(wr.report.results)
	defer close(wr.baseReport.Results())
	for res := range wr.results {
		wr.processResult(res)
		wr.baseReport.Results() <- res
	}
}

func (wr *weightedReport) processResult(res Result) {
	if res.Err != nil {
		wr.report.results <- res
		return
	}
	if res.Weight == 0 {
		res.Weight = 1.0
	}
	wr.weightTotal += res.Weight
	res.End = res.Start.Add(time.Duration(float64(res.End.Sub(res.Start)) / res.Weight))
	res.Weight = 1.0
	wr.report.results <- res
}

func (wr *weightedReport) reweighStat(s Stats) Stats {
	weightCoef := wr.weightTotal / float64(len(s.Lats))
	// weight > 1 => processing more than one request
	s.RPS *= weightCoef
	s.AvgTotal *= weightCoef * weightCoef
	return s
}
