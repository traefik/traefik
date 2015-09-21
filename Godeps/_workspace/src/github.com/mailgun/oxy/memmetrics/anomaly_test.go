package memmetrics

import (
	"time"

	. "gopkg.in/check.v1"
)

type AnomalySuite struct {
}

var _ = Suite(&AnomalySuite{})

func (s *AnomalySuite) TestMedian(c *C) {
	c.Assert(median([]float64{0.1, 0.2}), Equals, (float64(0.1)+float64(0.2))/2.0)
	c.Assert(median([]float64{0.3, 0.2, 0.5}), Equals, 0.3)
}

func (s *AnomalySuite) TestSplitRatios(c *C) {
	vals := []struct {
		values []float64
		good   []float64
		bad    []float64
	}{
		{
			values: []float64{0, 0},
			good:   []float64{0},
			bad:    []float64{},
		},

		{
			values: []float64{0, 1},
			good:   []float64{0},
			bad:    []float64{1},
		},
		{
			values: []float64{0.1, 0.1},
			good:   []float64{0.1},
			bad:    []float64{},
		},

		{
			values: []float64{0.15, 0.1},
			good:   []float64{0.15, 0.1},
			bad:    []float64{},
		},
		{
			values: []float64{0.01, 0.01},
			good:   []float64{0.01},
			bad:    []float64{},
		},
		{
			values: []float64{0.012, 0.01, 1},
			good:   []float64{0.012, 0.01},
			bad:    []float64{1},
		},
		{
			values: []float64{0, 0, 1, 1},
			good:   []float64{0},
			bad:    []float64{1},
		},
		{
			values: []float64{0, 0.1, 0.1, 0},
			good:   []float64{0},
			bad:    []float64{0.1},
		},
		{
			values: []float64{0, 0.01, 0.1, 0},
			good:   []float64{0},
			bad:    []float64{0.01, 0.1},
		},
		{
			values: []float64{0, 0.01, 0.02, 1},
			good:   []float64{0, 0.01, 0.02},
			bad:    []float64{1},
		},
		{
			values: []float64{0, 0, 0, 0, 0, 0.01, 0.02, 1},
			good:   []float64{0},
			bad:    []float64{0.01, 0.02, 1},
		},
	}
	for _, v := range vals {
		good, bad := SplitRatios(v.values)
		vgood, vbad := make(map[float64]bool, len(v.good)), make(map[float64]bool, len(v.bad))
		for _, v := range v.good {
			vgood[v] = true
		}
		for _, v := range v.bad {
			vbad[v] = true
		}

		c.Assert(good, DeepEquals, vgood)
		c.Assert(bad, DeepEquals, vbad)
	}
}

func (s *AnomalySuite) TestSplitLatencies(c *C) {
	vals := []struct {
		values []int
		good   []int
		bad    []int
	}{
		{
			values: []int{0, 0},
			good:   []int{0},
			bad:    []int{},
		},
		{
			values: []int{1, 2},
			good:   []int{1, 2},
			bad:    []int{},
		},
		{
			values: []int{1, 2, 4},
			good:   []int{1, 2, 4},
			bad:    []int{},
		},
		{
			values: []int{8, 8, 18},
			good:   []int{8},
			bad:    []int{18},
		},
		{
			values: []int{32, 28, 11, 26, 19, 51, 25, 39, 28, 26, 8, 97},
			good:   []int{32, 28, 11, 26, 19, 51, 25, 39, 28, 26, 8},
			bad:    []int{97},
		},
		{
			values: []int{1, 2, 4, 40},
			good:   []int{1, 2, 4},
			bad:    []int{40},
		},
		{
			values: []int{40, 60, 1000},
			good:   []int{40, 60},
			bad:    []int{1000},
		},
	}
	for _, v := range vals {
		vvalues := make([]time.Duration, len(v.values))
		for i, d := range v.values {
			vvalues[i] = time.Millisecond * time.Duration(d)
		}
		good, bad := SplitLatencies(vvalues, time.Millisecond)

		vgood, vbad := make(map[time.Duration]bool, len(v.good)), make(map[time.Duration]bool, len(v.bad))
		for _, v := range v.good {
			vgood[time.Duration(v)*time.Millisecond] = true
		}
		for _, v := range v.bad {
			vbad[time.Duration(v)*time.Millisecond] = true
		}

		c.Assert(good, DeepEquals, vgood)
		c.Assert(bad, DeepEquals, vbad)
	}
}
