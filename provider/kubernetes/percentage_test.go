package kubernetes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPercentageValueParse(t *testing.T) {
	newInt := func(i int) *int { return &i }
	testCases := []struct {
		parseString     string
		precision       *int
		shouldError     bool
		expectedString  string
		expectedFloat64 float64
	}{
		{
			parseString:     "1%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "1%",
			precision:       newInt(4),
			shouldError:     false,
			expectedString:  "1.0000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "1%",
			precision:       newInt(0),
			shouldError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "1%",
			precision:       newInt(-1),
			shouldError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "0.5",
			precision:       nil,
			shouldError:     false,
			expectedString:  "50.000%",
			expectedFloat64: 0.5,
		},
		{
			parseString:     "99%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "99.000%",
			expectedFloat64: 0.99,
		},
		{
			parseString:     "99.999%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "99.999%",
			expectedFloat64: 0.99999,
		},
		{
			parseString:     "-99.999%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "-99.9990%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "0%",
			precision:       nil,
			shouldError:     false,
			expectedString:  "0.000%",
			expectedFloat64: 0,
		},
		{
			parseString:     "%",
			precision:       nil,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "foo",
			precision:       nil,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "",
			precision:       nil,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
	}
	for _, testCase := range testCases {
		var pv *PercentageValue
		var err error
		if testCase.precision == nil {
			pv, err = PercentageValueFromString(testCase.parseString)
		} else {
			pv, err = PercentageValueFromString(testCase.parseString, *testCase.precision)
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
