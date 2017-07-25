// Copyright 2016 The etcd Authors
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

package adt

import (
	"math/rand"
	"testing"
	"time"
)

func TestIntervalTreeIntersects(t *testing.T) {
	ivt := &IntervalTree{}
	ivt.Insert(NewStringInterval("1", "3"), 123)

	if ivt.Intersects(NewStringPoint("0")) {
		t.Errorf("contains 0")
	}
	if !ivt.Intersects(NewStringPoint("1")) {
		t.Errorf("missing 1")
	}
	if !ivt.Intersects(NewStringPoint("11")) {
		t.Errorf("missing 11")
	}
	if !ivt.Intersects(NewStringPoint("2")) {
		t.Errorf("missing 2")
	}
	if ivt.Intersects(NewStringPoint("3")) {
		t.Errorf("contains 3")
	}
}

func TestIntervalTreeStringAffine(t *testing.T) {
	ivt := &IntervalTree{}
	ivt.Insert(NewStringAffineInterval("8", ""), 123)
	if !ivt.Intersects(NewStringAffinePoint("9")) {
		t.Errorf("missing 9")
	}
	if ivt.Intersects(NewStringAffinePoint("7")) {
		t.Errorf("contains 7")
	}
}

func TestIntervalTreeStab(t *testing.T) {
	ivt := &IntervalTree{}
	ivt.Insert(NewStringInterval("0", "1"), 123)
	ivt.Insert(NewStringInterval("0", "2"), 456)
	ivt.Insert(NewStringInterval("5", "6"), 789)
	ivt.Insert(NewStringInterval("6", "8"), 999)
	ivt.Insert(NewStringInterval("0", "3"), 0)

	if ivt.root.max.Compare(StringComparable("8")) != 0 {
		t.Fatalf("wrong root max got %v, expected 8", ivt.root.max)
	}
	if x := len(ivt.Stab(NewStringPoint("0"))); x != 3 {
		t.Errorf("got %d, expected 3", x)
	}
	if x := len(ivt.Stab(NewStringPoint("1"))); x != 2 {
		t.Errorf("got %d, expected 2", x)
	}
	if x := len(ivt.Stab(NewStringPoint("2"))); x != 1 {
		t.Errorf("got %d, expected 1", x)
	}
	if x := len(ivt.Stab(NewStringPoint("3"))); x != 0 {
		t.Errorf("got %d, expected 0", x)
	}
	if x := len(ivt.Stab(NewStringPoint("5"))); x != 1 {
		t.Errorf("got %d, expected 1", x)
	}
	if x := len(ivt.Stab(NewStringPoint("55"))); x != 1 {
		t.Errorf("got %d, expected 1", x)
	}
	if x := len(ivt.Stab(NewStringPoint("6"))); x != 1 {
		t.Errorf("got %d, expected 1", x)
	}
}

type xy struct {
	x int64
	y int64
}

func TestIntervalTreeRandom(t *testing.T) {
	// generate unique intervals
	ivs := make(map[xy]struct{})
	ivt := &IntervalTree{}
	maxv := 128
	rand.Seed(time.Now().UnixNano())

	for i := rand.Intn(maxv) + 1; i != 0; i-- {
		x, y := int64(rand.Intn(maxv)), int64(rand.Intn(maxv))
		if x > y {
			t := x
			x = y
			y = t
		} else if x == y {
			y++
		}
		iv := xy{x, y}
		if _, ok := ivs[iv]; ok {
			// don't double insert
			continue
		}
		ivt.Insert(NewInt64Interval(x, y), 123)
		ivs[iv] = struct{}{}
	}

	for ab := range ivs {
		for xy := range ivs {
			v := xy.x + int64(rand.Intn(int(xy.y-xy.x)))
			if slen := len(ivt.Stab(NewInt64Point(v))); slen == 0 {
				t.Fatalf("expected %v stab non-zero for [%+v)", v, xy)
			}
			if !ivt.Intersects(NewInt64Point(v)) {
				t.Fatalf("did not get %d as expected for [%+v)", v, xy)
			}
		}
		if !ivt.Delete(NewInt64Interval(ab.x, ab.y)) {
			t.Errorf("did not delete %v as expected", ab)
		}
		delete(ivs, ab)
	}

	if ivt.Len() != 0 {
		t.Errorf("got ivt.Len() = %v, expected 0", ivt.Len())
	}
}

// TestIntervalTreeSortedVisit tests that intervals are visited in sorted order.
func TestIntervalTreeSortedVisit(t *testing.T) {
	tests := []struct {
		ivls       []Interval
		visitRange Interval
	}{
		{
			ivls:       []Interval{NewInt64Interval(1, 10), NewInt64Interval(2, 5), NewInt64Interval(3, 6)},
			visitRange: NewInt64Interval(0, 100),
		},
		{
			ivls:       []Interval{NewInt64Interval(1, 10), NewInt64Interval(10, 12), NewInt64Interval(3, 6)},
			visitRange: NewInt64Interval(0, 100),
		},
		{
			ivls:       []Interval{NewInt64Interval(2, 3), NewInt64Interval(3, 4), NewInt64Interval(6, 7), NewInt64Interval(5, 6)},
			visitRange: NewInt64Interval(0, 100),
		},
		{
			ivls: []Interval{
				NewInt64Interval(2, 3),
				NewInt64Interval(2, 4),
				NewInt64Interval(3, 7),
				NewInt64Interval(2, 5),
				NewInt64Interval(3, 8),
				NewInt64Interval(3, 5),
			},
			visitRange: NewInt64Interval(0, 100),
		},
	}
	for i, tt := range tests {
		ivt := &IntervalTree{}
		for _, ivl := range tt.ivls {
			ivt.Insert(ivl, struct{}{})
		}
		last := tt.ivls[0].Begin
		count := 0
		chk := func(iv *IntervalValue) bool {
			if last.Compare(iv.Ivl.Begin) > 0 {
				t.Errorf("#%d: expected less than %d, got interval %+v", i, last, iv.Ivl)
			}
			last = iv.Ivl.Begin
			count++
			return true
		}
		ivt.Visit(tt.visitRange, chk)
		if count != len(tt.ivls) {
			t.Errorf("#%d: did not cover all intervals. expected %d, got %d", i, len(tt.ivls), count)
		}
	}
}

// TestIntervalTreeVisitExit tests that visiting can be stopped.
func TestIntervalTreeVisitExit(t *testing.T) {
	ivls := []Interval{NewInt64Interval(1, 10), NewInt64Interval(2, 5), NewInt64Interval(3, 6), NewInt64Interval(4, 8)}
	ivlRange := NewInt64Interval(0, 100)
	tests := []struct {
		f IntervalVisitor

		wcount int
	}{
		{
			f:      func(n *IntervalValue) bool { return false },
			wcount: 1,
		},
		{
			f:      func(n *IntervalValue) bool { return n.Ivl.Begin.Compare(ivls[0].Begin) <= 0 },
			wcount: 2,
		},
		{
			f:      func(n *IntervalValue) bool { return n.Ivl.Begin.Compare(ivls[2].Begin) < 0 },
			wcount: 3,
		},
		{
			f:      func(n *IntervalValue) bool { return true },
			wcount: 4,
		},
	}

	for i, tt := range tests {
		ivt := &IntervalTree{}
		for _, ivl := range ivls {
			ivt.Insert(ivl, struct{}{})
		}
		count := 0
		ivt.Visit(ivlRange, func(n *IntervalValue) bool {
			count++
			return tt.f(n)
		})
		if count != tt.wcount {
			t.Errorf("#%d: expected count %d, got %d", i, tt.wcount, count)
		}
	}
}

// TestIntervalTreeContains tests that contains returns true iff the ivt maps the entire interval.
func TestIntervalTreeContains(t *testing.T) {
	tests := []struct {
		ivls   []Interval
		chkIvl Interval

		wContains bool
	}{
		{
			ivls:   []Interval{NewInt64Interval(1, 10)},
			chkIvl: NewInt64Interval(0, 100),

			wContains: false,
		},
		{
			ivls:   []Interval{NewInt64Interval(1, 10)},
			chkIvl: NewInt64Interval(1, 10),

			wContains: true,
		},
		{
			ivls:   []Interval{NewInt64Interval(1, 10)},
			chkIvl: NewInt64Interval(2, 8),

			wContains: true,
		},
		{
			ivls:   []Interval{NewInt64Interval(1, 5), NewInt64Interval(6, 10)},
			chkIvl: NewInt64Interval(1, 10),

			wContains: false,
		},
		{
			ivls:   []Interval{NewInt64Interval(1, 5), NewInt64Interval(3, 10)},
			chkIvl: NewInt64Interval(1, 10),

			wContains: true,
		},
		{
			ivls:   []Interval{NewInt64Interval(1, 4), NewInt64Interval(4, 7), NewInt64Interval(3, 10)},
			chkIvl: NewInt64Interval(1, 10),

			wContains: true,
		},
		{
			ivls:   []Interval{},
			chkIvl: NewInt64Interval(1, 10),

			wContains: false,
		},
	}
	for i, tt := range tests {
		ivt := &IntervalTree{}
		for _, ivl := range tt.ivls {
			ivt.Insert(ivl, struct{}{})
		}
		if v := ivt.Contains(tt.chkIvl); v != tt.wContains {
			t.Errorf("#%d: ivt.Contains got %v, expected %v", i, v, tt.wContains)
		}
	}
}
