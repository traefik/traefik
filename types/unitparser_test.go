package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsSI_happy(t *testing.T) {
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

	for _, c := range cases {
		i, u, e := AsSI(c.in)
		assert.NoError(t, e)
		assert.Equal(t, c.unit, u)
		assert.Equal(t, c.exp, i)
	}
}
