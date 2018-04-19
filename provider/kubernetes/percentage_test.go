package kubernetes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPercentageValueParse(t *testing.T) {
	newInt := func(i int) *int { return &i }
	testCases := []struct {
		s               string
		precision       *int
		shouldError     bool
		expectedString  string
		expectedFloat64 float64
	}{
		{"1%", nil, false, "1.000%", 0.01},
		{"1%", newInt(4), false, "1.0000%", 0.01},
		{"1%", newInt(0), false, "1.000%", 0.01},
		{"1%", newInt(-1), false, "1.000%", 0.01},
		{"0.5", nil, false, "50.000%", 0.5},
		{"99%", nil, false, "99.000%", 0.99},
		{"99.999%", nil, false, "99.999%", 0.99999},
		{"99.9999%", nil, false, "99.999%", 0.99999},
		{"-99.999%", nil, false, "-99.999%", -0.99999},
		{"-99.9999%", nil, false, "-99.999%", -0.99999},
		{"0%", nil, false, "0.000%", 0},
		{"%", nil, true, "", 0},
		{"foo", nil, true, "", 0},
		{"", nil, true, "", 0},
	}
	for _, testCase := range testCases {
		var pv *PercentageValue
		var err error
		if testCase.precision == nil {
			pv, err = PercentageValueFromString(testCase.s)
		} else {
			pv, err = PercentageValueFromString(testCase.s, *testCase.precision)
		}
		if testCase.shouldError {
			assert.Error(t, err, "expecting error but not happening")
			continue
		}
		assert.NoError(t, err, "fail to parse percentage value")
		assert.Equal(t, testCase.expectedString, pv.String(), "percentage string value mismatched")
		assert.Equal(t, testCase.expectedFloat64, pv.Float64(), "percentage float64 value mismatched")
	}
}
