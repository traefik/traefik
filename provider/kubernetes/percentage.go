package kubernetes

import (
	"strconv"
	"strings"
)

const (
	defaultPercentageValuePrecision = 3
)

// PercentageValue is int64 form of percentage value with 10^-3 precision.
type PercentageValue struct {
	value     int64
	precision int
}

// PercentageValueFromString tries to read percentage value from string, it can be
// either "1.1" or "1.1%", "6%". It will lose the extra precision if there are more
// digits after decimal point.
func PercentageValueFromString(s string, precision ...int) (*PercentageValue, error) {
	if len(precision) > 1 {
		precision = []int{precision[0]}
	}
	hasPercentageTag := strings.HasSuffix(s, "%")
	f, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	if err != nil {
		return nil, err
	}
	percentageValue := PercentageValueFromFloat64(f)
	if len(precision) > 0 && precision[0] > 0 {
		percentageValue.precision = precision[0]
	}
	if hasPercentageTag {
		percentageValue.value /= 100
		return percentageValue, nil
	}
	return percentageValue, nil
}

// PercentageValueFromFloat64 reads percentage value from float64
func PercentageValueFromFloat64(f float64) *PercentageValue {
	return &PercentageValue{
		value:     int64(f * 100 * 1000),
		precision: defaultPercentageValuePrecision,
	}
}

// RawValue returns its internal raw int64 form of percentage value.
func (v *PercentageValue) RawValue() int64 {
	return v.value
}

// Float64 returns its decimal float64 value.
func (v *PercentageValue) Float64() float64 {
	return float64(v.value) / (1000 * 100)
}

// String returns its string form of percentage value.
func (v *PercentageValue) String() string {
	return strconv.FormatFloat(v.Float64()*100, 'f', v.precision, 64) + "%"
}
