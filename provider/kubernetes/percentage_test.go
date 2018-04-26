package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentageValueParse(t *testing.T) {
	testCases := []struct {
		desc            string
		parseString     string
		parseFloat64    float64
		expectError     bool
		expectedString  string
		expectedFloat64 float64
	}{
		{
			parseString:     "1%",
			parseFloat64:    0.01,
			expectError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "0.5",
			parseFloat64:    0.5,
			expectError:     false,
			expectedString:  "50.000%",
			expectedFloat64: 0.5,
		},
		{
			parseString:     "99%",
			parseFloat64:    0.99,
			expectError:     false,
			expectedString:  "99.000%",
			expectedFloat64: 0.99,
		},
		{
			parseString:     "99.999%",
			parseFloat64:    0.99999,
			expectError:     false,
			expectedString:  "99.999%",
			expectedFloat64: 0.99999,
		},
		{
			parseString:     "-99.999%",
			parseFloat64:    -0.99999,
			expectError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "-99.9990%",
			parseFloat64:    -0.99999,
			expectError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "0%",
			parseFloat64:    0,
			expectError:     false,
			expectedString:  "0.000%",
			expectedFloat64: 0,
		},
		{
			parseString:     "%",
			parseFloat64:    0,
			expectError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "foo",
			parseFloat64:    0,
			expectError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "",
			parseFloat64:    0,
			expectError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pvFromString, err := percentageValueFromString(test.parseString)
			pvFromFloat64 := percentageValueFromFloat64(test.parseFloat64)

			if test.expectError {
				require.Error(t, err, "expecting error but not happening")
			} else {
				assert.NoError(t, err, "fail to parse percentage value")

				assert.Equal(t, pvFromString, pvFromFloat64)
				assert.Equal(t, test.expectedString, pvFromFloat64.toString(), "percentage string value mismatched")
				assert.Equal(t, test.expectedFloat64, pvFromFloat64.toFloat64(), "percentage float64 value mismatched")
			}
		})
	}
}
