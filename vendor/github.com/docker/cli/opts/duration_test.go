package opts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationOptString(t *testing.T) {
	dur := time.Duration(300 * 10e8)
	duration := DurationOpt{value: &dur}
	assert.Equal(t, "5m0s", duration.String())
}

func TestDurationOptSetAndValue(t *testing.T) {
	var duration DurationOpt
	assert.NoError(t, duration.Set("300s"))
	assert.Equal(t, time.Duration(300*10e8), *duration.Value())
	assert.NoError(t, duration.Set("-300s"))
	assert.Equal(t, time.Duration(-300*10e8), *duration.Value())
}

func TestPositiveDurationOptSetAndValue(t *testing.T) {
	var duration PositiveDurationOpt
	assert.NoError(t, duration.Set("300s"))
	assert.Equal(t, time.Duration(300*10e8), *duration.Value())
	assert.EqualError(t, duration.Set("-300s"), "duration cannot be negative")
}
