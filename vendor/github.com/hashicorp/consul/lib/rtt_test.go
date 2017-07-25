package lib

import (
	"math"
	"testing"
	"time"

	"github.com/hashicorp/serf/coordinate"
)

func TestRTT(t *testing.T) {
	cases := []struct {
		a    *coordinate.Coordinate
		b    *coordinate.Coordinate
		dist float64
	}{
		{
			GenerateCoordinate(0),
			GenerateCoordinate(10 * time.Millisecond),
			0.010,
		},
		{
			GenerateCoordinate(10 * time.Millisecond),
			GenerateCoordinate(10 * time.Millisecond),
			0.0,
		},
		{
			GenerateCoordinate(8 * time.Millisecond),
			GenerateCoordinate(10 * time.Millisecond),
			0.002,
		},
		{
			GenerateCoordinate(10 * time.Millisecond),
			GenerateCoordinate(8 * time.Millisecond),
			0.002,
		},
		{
			nil,
			GenerateCoordinate(8 * time.Millisecond),
			math.Inf(1.0),
		},
		{
			GenerateCoordinate(8 * time.Millisecond),
			nil,
			math.Inf(1.0),
		},
	}
	for i, c := range cases {
		dist := ComputeDistance(c.a, c.b)
		if c.dist != dist {
			t.Fatalf("bad (%d): %9.6f != %9.6f", i, c.dist, dist)
		}
	}
}
