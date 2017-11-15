// Package gohistogram contains implementations of weighted and exponential histograms.
package gohistogram

// Copyright (c) 2013 VividCortex, Inc. All rights reserved.
// Please see the LICENSE file for applicable license terms.

import "fmt"

// A WeightedHistogram implements Histogram. A WeightedHistogram has bins that have values
// which are exponentially weighted moving averages. This allows you keep inserting large
// amounts of data into the histogram and approximate quantiles with recency factored in.
type WeightedHistogram struct {
	bins    []bin
	maxbins int
	total   float64
	alpha   float64
}

// NewWeightedHistogram returns a new WeightedHistogram with a maximum of n bins with a decay factor
// of alpha.
//
// There is no "optimal" bin count, but somewhere between 20 and 80 bins should be
// sufficient.
//
// Alpha should be set to 2 / (N+1), where N represents the average age of the moving window.
// For example, a 60-second window with an average age of 30 seconds would yield an
// alpha of 0.064516129.
func NewWeightedHistogram(n int, alpha float64) *WeightedHistogram {
	return &WeightedHistogram{
		bins:    make([]bin, 0),
		maxbins: n,
		total:   0,
		alpha:   alpha,
	}
}

func ewma(existingVal float64, newVal float64, alpha float64) (result float64) {
	result = newVal*(1-alpha) + existingVal*alpha
	return
}

func (h *WeightedHistogram) scaleDown(except int) {
	for i := range h.bins {
		if i != except {
			h.bins[i].count = ewma(h.bins[i].count, 0, h.alpha)
		}
	}
}

func (h *WeightedHistogram) Add(n float64) {
	defer h.trim()
	for i := range h.bins {
		if h.bins[i].value == n {
			h.bins[i].count++

			defer h.scaleDown(i)
			return
		}

		if h.bins[i].value > n {

			newbin := bin{value: n, count: 1}
			head := append(make([]bin, 0), h.bins[0:i]...)

			head = append(head, newbin)
			tail := h.bins[i:]
			h.bins = append(head, tail...)

			defer h.scaleDown(i)
			return
		}
	}

	h.bins = append(h.bins, bin{count: 1, value: n})
}

func (h *WeightedHistogram) Quantile(q float64) float64 {
	count := q * h.total
	for i := range h.bins {
		count -= float64(h.bins[i].count)

		if count <= 0 {
			return h.bins[i].value
		}
	}

	return -1
}

// CDF returns the value of the cumulative distribution function
// at x
func (h *WeightedHistogram) CDF(x float64) float64 {
	count := 0.0
	for i := range h.bins {
		if h.bins[i].value <= x {
			count += float64(h.bins[i].count)
		}
	}

	return count / h.total
}

// Mean returns the sample mean of the distribution
func (h *WeightedHistogram) Mean() float64 {
	if h.total == 0 {
		return 0
	}

	sum := 0.0

	for i := range h.bins {
		sum += h.bins[i].value * h.bins[i].count
	}

	return sum / h.total
}

// Variance returns the variance of the distribution
func (h *WeightedHistogram) Variance() float64 {
	if h.total == 0 {
		return 0
	}

	sum := 0.0
	mean := h.Mean()

	for i := range h.bins {
		sum += (h.bins[i].count * (h.bins[i].value - mean) * (h.bins[i].value - mean))
	}

	return sum / h.total
}

func (h *WeightedHistogram) Count() float64 {
	return h.total
}

func (h *WeightedHistogram) trim() {
	total := 0.0
	for i := range h.bins {
		total += h.bins[i].count
	}
	h.total = total
	for len(h.bins) > h.maxbins {

		// Find closest bins in terms of value
		minDelta := 1e99
		minDeltaIndex := 0
		for i := range h.bins {
			if i == 0 {
				continue
			}

			if delta := h.bins[i].value - h.bins[i-1].value; delta < minDelta {
				minDelta = delta
				minDeltaIndex = i
			}
		}

		// We need to merge bins minDeltaIndex-1 and minDeltaIndex
		totalCount := h.bins[minDeltaIndex-1].count + h.bins[minDeltaIndex].count
		mergedbin := bin{
			value: (h.bins[minDeltaIndex-1].value*
				h.bins[minDeltaIndex-1].count +
				h.bins[minDeltaIndex].value*
					h.bins[minDeltaIndex].count) /
				totalCount, // weighted average
			count: totalCount, // summed heights
		}
		head := append(make([]bin, 0), h.bins[0:minDeltaIndex-1]...)
		tail := append([]bin{mergedbin}, h.bins[minDeltaIndex+1:]...)
		h.bins = append(head, tail...)
	}
}

// String returns a string reprentation of the histogram,
// which is useful for printing to a terminal.
func (h *WeightedHistogram) String() (str string) {
	str += fmt.Sprintln("Total:", h.total)

	for i := range h.bins {
		var bar string
		for j := 0; j < int(float64(h.bins[i].count)/float64(h.total)*200); j++ {
			bar += "."
		}
		str += fmt.Sprintln(h.bins[i].value, "\t", bar)
	}

	return
}
