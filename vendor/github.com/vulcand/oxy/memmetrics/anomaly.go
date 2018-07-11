package memmetrics

import (
	"math"
	"sort"
	"time"
)

// SplitLatencies provides simple anomaly detection for requests latencies.
// it splits values into good or bad category based on the threshold and the median value.
// If all values are not far from the median, it will return all values in 'good' set.
// Precision is the smallest value to consider, e.g. if set to millisecond, microseconds will be ignored.
func SplitLatencies(values []time.Duration, precision time.Duration) (good map[time.Duration]bool, bad map[time.Duration]bool) {
	// Find the max latency M and then map each latency L to the ratio L/M and then call SplitFloat64
	v2r := map[float64]time.Duration{}
	ratios := make([]float64, len(values))
	m := maxTime(values)
	for i, v := range values {
		ratio := float64(v/precision+1) / float64(m/precision+1) // +1 is to avoid division by 0
		v2r[ratio] = v
		ratios[i] = ratio
	}
	good, bad = make(map[time.Duration]bool), make(map[time.Duration]bool)
	// Note that multiplier makes this function way less sensitive than ratios detector, this is to avoid noise.
	vgood, vbad := SplitFloat64(2, 0, ratios)
	for r := range vgood {
		good[v2r[r]] = true
	}
	for r := range vbad {
		bad[v2r[r]] = true
	}
	return good, bad
}

// SplitRatios provides simple anomaly detection for ratio values, that are all in the range [0, 1]
// it splits values into good or bad category based on the threshold and the median value.
// If all values are not far from the median, it will return all values in 'good' set.
func SplitRatios(values []float64) (good map[float64]bool, bad map[float64]bool) {
	return SplitFloat64(1.5, 0, values)
}

// SplitFloat64 provides simple anomaly detection for skewed data sets with no particular distribution.
// In essence it applies the formula if(v > median(values) + threshold * medianAbsoluteDeviation) -> anomaly
// There's a corner case where there are just 2 values, so by definition there's no value that exceeds the threshold.
// This case is solved by introducing additional value that we know is good, e.g. 0. That helps to improve the detection results
// on such data sets.
func SplitFloat64(threshold, sentinel float64, values []float64) (good map[float64]bool, bad map[float64]bool) {
	good, bad = make(map[float64]bool), make(map[float64]bool)
	var newValues []float64
	if len(values)%2 == 0 {
		newValues = make([]float64, len(values)+1)
		copy(newValues, values)
		// Add a sentinel endpoint so we can distinguish outliers better
		newValues[len(newValues)-1] = sentinel
	} else {
		newValues = values
	}

	m := median(newValues)
	mAbs := medianAbsoluteDeviation(newValues)
	for _, v := range values {
		if v > (m+mAbs)*threshold {
			bad[v] = true
		} else {
			good[v] = true
		}
	}
	return good, bad
}

func median(values []float64) float64 {
	vals := make([]float64, len(values))
	copy(vals, values)
	sort.Float64s(vals)
	l := len(vals)
	if l%2 != 0 {
		return vals[l/2]
	}
	return (vals[l/2-1] + vals[l/2]) / 2.0
}

func medianAbsoluteDeviation(values []float64) float64 {
	m := median(values)
	distances := make([]float64, len(values))
	for i, v := range values {
		distances[i] = math.Abs(v - m)
	}
	return median(distances)
}

func maxTime(vals []time.Duration) time.Duration {
	val := vals[0]
	for _, v := range vals {
		if v > val {
			val = v
		}
	}
	return val
}
