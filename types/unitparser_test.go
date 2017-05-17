package types

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestAsSIHappy(t *testing.T) {
	cases := []struct {
		in, unit string
		exp      int64
	}{
		{"1", "", 1},
		{"1 km", "m", 1000},
		{"1KB", "B", 1000},
		{"2 KiB", "B", 2048},
		{"1MJ", "J", 1000000},
		{"2 MiB", "B", 2 * 1024 * 1024},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("in: %s, unit %s", tc.in, tc.unit), func(t *testing.T) {
			i, u, e := AsSI(tc.in)
			assert.NoError(t, e)
			assert.Equal(t, tc.unit, u)
			assert.Equal(t, tc.exp, i)
		})
	}
}

func TestAsSIUnhappy(t *testing.T) {
	cases := []struct {
		in, unit string
		exp      int64
		err      bool
	}{
		{"", "", 0, true},
		{"1x", "", 1, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("in: %s, unit %s", tc.in, tc.unit), func(t *testing.T) {
			i, u, e := AsSI(tc.in)
			if tc.err {
				assert.Error(t, e)
			} else {
				assert.NoError(t, e)
			}
			assert.Equal(t, tc.unit, u)
			assert.Equal(t, tc.exp, i)
		})
	}
}
