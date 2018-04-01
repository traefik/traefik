package kubernetes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPercentageValueParse(t *testing.T) {
	testCases := []struct {
		s               string
		shouldError     bool
		expected        PercentageValue
		expectedString  string
		expectedFloat64 float64
	}{
		{"1%", false, 1000, "1.000%", 0.01},
		{"99%", false, 99000, "99.000%", 0.99},
		{"99.999%", false, 99999, "99.999%", 0.99999},
		{"99.9999%", false, 99999, "99.999%", 0.99999},
		{"0%", false, 0, "0.000%", 0},
		{"%", true, 0, "", 0},
		{"foo", true, 0, "", 0},
		{"1.11", true, 0, "", 0},
		{"", true, 0, "", 0},
	}
	for _, testCase := range testCases {
		pv, err := PercentageValueFromString(testCase.s)
		if testCase.shouldError {
			assert.Error(t, err, "expecting error but not happening")
			continue
		}
		assert.NoError(t, err, "fail to parse percentage value")
		assert.Equal(t, testCase.expected, pv, "percentage value mismatched")
		assert.Equal(t, testCase.expectedString, pv.String(), "percentage string value mismatched")
		assert.Equal(t, testCase.expectedFloat64, pv.Float64(), "percentage float64 value mismatched")
	}
}
