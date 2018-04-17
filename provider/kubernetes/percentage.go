package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
)

// PercentageValue is int64 form of percentage value with 10^-3 precision.
type PercentageValue int64

// PercentageValueFromString tries to read percentage value from string like "1.1%", "6%".
// It will lose the extra precision if there are more than 3 digit after decimal point.
func PercentageValueFromString(s string) (PercentageValue, error) {
	idx := strings.Index(s, "%")
	if idx < 0 {
		return 0, fmt.Errorf("missing %% for parsing percentage value: %q", s)
	}
	f, err := strconv.ParseFloat(s[:idx], 64)
	if err != nil {
		return 0, err
	}
	return PercentageValue(f * 1000), nil
}

// PercentageValueFromFloat64 reads percentage value from float64
func PercentageValueFromFloat64(f float64) PercentageValue {
	return PercentageValue(f * 100 * 1000)
}

// RawValue returns its internal raw int64 form of percentage value.
func (v PercentageValue) RawValue() int64 {
	return int64(v)
}

// Float64 returns its decimal float64 value.
func (v PercentageValue) Float64() float64 {
	return float64(v) / (1000 * 100)
}

// String returns its string form of percentage value.
func (v PercentageValue) String() string {
	return fmt.Sprintf("%.3f%%", v.Float64()*100)
}
