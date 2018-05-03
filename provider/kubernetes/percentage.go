package kubernetes

import (
	"strconv"
	"strings"
)

var (
	defaultPercentageValuePrecision = 3
)

// percentageValue is int64 form of percentage value with 10^-3 precision.
type percentageValue int64

// toFloat64 returns its decimal float64 value.
func (v percentageValue) toFloat64() float64 {
	return float64(int64(v)) / (1000 * 100)
}

func (v percentageValue) sub(value percentageValue) percentageValue {
	return percentageValue(v - value)
}

func (v percentageValue) add(value percentageValue) percentageValue {
	return percentageValue(v + value)
}

func (v percentageValue) computeWeight(count int) int {
	if count == 0 {
		return 0
	}
	return int(float64(v) / float64(count))
}

// String returns its string form of percentage value.
func (v percentageValue) String() string {
	return strconv.FormatFloat(v.toFloat64()*100, 'f', defaultPercentageValuePrecision, 64) + "%"
}

// newPercentageValueFromString tries to read percentage value from string, it can be either "1.1" or "1.1%", "6%".
// It will lose the extra precision if there are more digits after decimal point.
func newPercentageValueFromString(rawValue string) (percentageValue, error) {
	value, err := strconv.ParseFloat(strings.TrimSuffix(rawValue, "%"), 64)
	if err != nil {
		return 0, err
	}

	percentageValue := newPercentageValueFromFloat64(value)
	if strings.HasSuffix(rawValue, "%") {
		percentageValue /= 100
	}
	return percentageValue, nil
}

// newPercentageValueFromFloat64 reads percentage value from float64
func newPercentageValueFromFloat64(f float64) percentageValue {
	return percentageValue(int64(f * 100 * 1000))
}
