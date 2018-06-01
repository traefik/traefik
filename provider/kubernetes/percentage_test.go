package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPercentageValueFromFloat64(t *testing.T) {
	testCases := []struct {
		desc            string
		value           float64
		expectedString  string
		expectedFloat64 float64
	}{
		{
			value:           0.01,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			value:           0.5,
			expectedString:  "50.000%",
			expectedFloat64: 0.5,
		},
		{
			value:           0.99,
			expectedString:  "99.000%",
			expectedFloat64: 0.99,
		},
		{
			value:           0.99999,
			expectedString:  "99.999%",
			expectedFloat64: 0.99999,
		},
		{
			value:           -0.99999,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			value:           -0.9999999,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			value:           0,
			expectedString:  "0.000%",
			expectedFloat64: 0,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pvFromFloat64 := newPercentageValueFromFloat64(test.value)

			assert.Equal(t, test.expectedString, pvFromFloat64.String(), "percentage string value mismatched")
			assert.Equal(t, test.expectedFloat64, pvFromFloat64.toFloat64(), "percentage float64 value mismatched")
		})
	}
}

func TestNewPercentageValueFromString(t *testing.T) {
	testCases := []struct {
		desc            string
		value           string
		expectError     bool
		expectedString  string
		expectedFloat64 float64
	}{
		{
			value:           "1%",
			expectError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			value:           "0.5",
			expectError:     false,
			expectedString:  "0.500%",
			expectedFloat64: 0.005,
		},
		{
			value:           "99%",
			expectError:     false,
			expectedString:  "99.000%",
			expectedFloat64: 0.99,
		},
		{
			value:           "99.9%",
			expectError:     false,
			expectedString:  "99.900%",
			expectedFloat64: 0.999,
		},
		{
			value:           "-99.9%",
			expectError:     false,
			expectedString:  "-99.900%",
			expectedFloat64: -0.999,
		},
		{
			value:           "-99.99999%",
			expectError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			value:           "0%",
			expectError:     false,
			expectedString:  "0.000%",
			expectedFloat64: 0,
		},
		{
			value:       "%",
			expectError: true,
		},
		{
			value:       "foo",
			expectError: true,
		},
		{
			value:       "",
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pvFromString, err := newPercentageValueFromString(test.value)

			if test.expectError {
				require.Error(t, err, "expecting error but not happening")
			} else {
				require.NoError(t, err, "fail to parse percentage value")

				assert.Equal(t, test.expectedString, pvFromString.String(), "percentage string value mismatched")
				assert.Equal(t, test.expectedFloat64, pvFromString.toFloat64(), "percentage float64 value mismatched")
			}
		})
	}
}

func TestNewPercentageValue(t *testing.T) {
	testCases := []struct {
		desc        string
		stringValue string
		floatValue  float64
	}{
		{
			desc:        "percentage",
			stringValue: "1%",
			floatValue:  0.01,
		},
		{
			desc:        "decimal",
			stringValue: "0.5",
			floatValue:  0.005,
		},
		{
			desc:        "negative percentage",
			stringValue: "-99.999%",
			floatValue:  -0.99999,
		},
		{
			desc:        "negative decimal",
			stringValue: "-0.99999",
			floatValue:  -0.0099999,
		},
		{
			desc:        "zero",
			stringValue: "0%",
			floatValue:  0,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pvFromString, err := newPercentageValueFromString(test.stringValue)
			require.NoError(t, err, "fail to parse percentage value")

			pvFromFloat64 := newPercentageValueFromFloat64(test.floatValue)

			assert.Equal(t, pvFromString, pvFromFloat64)
		})
	}
}
