package kubernetes

import (
	"strconv"
	"strings"
)

var (
	defaultPercentageValuePrecision = 3
)

// percentageValue is int64 form of percentage value with 10^-3 precision.
type percentageValue struct {
	value int64
}

// PercentageValueFromString tries to read percentage value from string, it can be
// either "1.1" or "1.1%", "6%". It will lose the extra precision if there are more
// digits after decimal point.
func percentageValueFromString(s string) (*percentageValue, error) {
	hasPercentageTag := strings.HasSuffix(s, "%")
	f, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	if err != nil {
		return nil, err
	}
	percentageValue := percentageValueFromFloat64(f)
	if hasPercentageTag {
		percentageValue.value /= 100
	}
	return percentageValue, nil
}

// PercentageValueFromFloat64 reads percentage value from float64
func percentageValueFromFloat64(f float64) *percentageValue {
	return &percentageValue{
		value: int64(f * 100 * 1000),
	}
}

// rawValue returns its internal raw int64 form of percentage value.
func (v *percentageValue) rawValue() float64 {
	return float64(v.value)
}

// Float64 returns its decimal float64 value.
func (v *percentageValue) toFloat64() float64 {
	return float64(v.value) / (1000 * 100)
}

// String returns its string form of percentage value.
func (v *percentageValue) toString() string {
	return strconv.FormatFloat(v.toFloat64()*100, 'f', defaultPercentageValuePrecision, 64) + "%"
}

func (v *percentageValue) sub(value *percentageValue) *percentageValue {
	return &percentageValue{
		value: v.value - value.value,
	}
}

func (v *percentageValue) add(value *percentageValue) *percentageValue {
	return &percentageValue{
		value: v.value + value.value,
	}
}

func (p *percentageValue) computeWeight(count int) int {
	if count == 0 {
		return 0
	}
	return int(p.rawValue() / float64(count))
}
