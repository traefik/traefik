package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
)

// MillisValue is used to store decimal values of -3 scale precision.
type MilliValue int64

type PercentageValue int64

func PercentageValueFromString(s string) (PercentageValue, error) {
	idx := strings.Index(s, "%")
	if idx < 0 {
		return 0, fmt.Errorf("missing % for parsing percentage value: %q", s)
	}
	f, err := strconv.ParseFloat(s[:idx], 64)
	if err != nil {
		return 0, err
	}
	return PercentageValue(f * 1000), nil
}

func (v PercentageValue) RawValue() int64 {
	return int64(v)
}

func (v PercentageValue) Float64() float64 {
	return float64(v) / (1000 * 100)
}

func (v PercentageValue) String() string {
	return fmt.Sprintf("%.3f%%", v.Float64()*100)
}
