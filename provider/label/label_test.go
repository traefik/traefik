package label

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAndTrimString(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []string
	}{
		{
			desc:     "empty string",
			input:    "",
			expected: nil,
		}, {
			desc:     "one piece",
			input:    "foo",
			expected: []string{"foo"},
		}, {
			desc:     "two pieces",
			input:    "foo,bar",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "three pieces",
			input:    "foo,bar,zoo",
			expected: []string{"foo", "bar", "zoo"},
		}, {
			desc:     "two pieces with whitespace",
			input:    " foo   ,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "consecutive commas",
			input:    " foo   ,,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "consecutive commas with whitespace",
			input:    " foo   , ,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "leading and trailing commas",
			input:    ",, foo   , ,  bar,, , ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "no valid pieces",
			input:    ",  , , ,, ,",
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := SplitAndTrimString(test.input, ",")
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetStringValue(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       map[string]string
		labelName    string
		defaultValue string
		expected     string
	}{
		{
			desc:         "empty labels map",
			labelName:    "foo",
			defaultValue: "default",
			expected:     "default",
		},
		{
			desc: "existing label",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName:    "foo",
			defaultValue: "default",
			expected:     "bar",
		},
		{
			desc: "non existing label",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName:    "fii",
			defaultValue: "default",
			expected:     "default",
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetStringValue(test.labels, test.labelName, test.defaultValue)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetBoolValue(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       map[string]string
		labelName    string
		defaultValue bool
		expected     bool
	}{
		{
			desc:      "empty map",
			labelName: "foo",
		},
		{
			desc: "invalid boolean value",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName:    "foo",
			defaultValue: true,
			expected:     true,
		},
		{
			desc: "valid boolean value: true",
			labels: map[string]string{
				"foo": "true",
			},
			labelName:    "foo",
			defaultValue: false,
			expected:     true,
		},
		{
			desc: "valid boolean value: false",
			labels: map[string]string{
				"foo": "false",
			},
			labelName:    "foo",
			defaultValue: true,
			expected:     false,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetBoolValue(test.labels, test.labelName, test.defaultValue)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetIntValue(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       map[string]string
		labelName    string
		defaultValue int
		expected     int
	}{
		{
			desc:      "empty map",
			labelName: "foo",
		},
		{
			desc:      "invalid int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "bar",
			},
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "negative int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "-1",
			},
			defaultValue: 666,
			expected:     -1,
		},
		{
			desc:      "positive int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "1",
			},
			defaultValue: 666,
			expected:     1,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetIntValue(test.labels, test.labelName, test.defaultValue)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetInt64Value(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       map[string]string
		labelName    string
		defaultValue int64
		expected     int64
	}{
		{
			desc:      "empty map",
			labelName: "foo",
		},
		{
			desc:      "invalid int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "bar",
			},
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "negative int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "-1",
			},
			defaultValue: 666,
			expected:     -1,
		},
		{
			desc:      "positive int value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "1",
			},
			defaultValue: 666,
			expected:     1,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetInt64Value(test.labels, test.labelName, test.defaultValue)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetSliceStringValue(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    map[string]string
		labelName string
		expected  []string
	}{
		{
			desc:      "empty map",
			labelName: "foo",
		},
		{
			desc: "empty value",
			labels: map[string]string{
				"foo": "",
			},
			labelName: "foo",
			expected:  nil,
		},
		{
			desc: "one value, not split",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName: "foo",
			expected:  []string{"bar"},
		},
		{
			desc: "several values",
			labels: map[string]string{
				"foo": "bar,bir ,bur",
			},
			labelName: "foo",
			expected:  []string{"bar", "bir", "bur"},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetSliceStringValue(test.labels, test.labelName)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestGetMapValue(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    map[string]string
		labelName string
		expected  map[string]string
	}{
		{
			desc:      "empty map",
			labelName: "foo",
		},
		{
			desc:      "existent label with empty entry",
			labelName: "foo",
			labels: map[string]string{
				"foo": "",
			},
			expected: nil,
		},
		{
			desc:      "existent label with invalid entry",
			labelName: "foo",
			labels: map[string]string{
				"foo": "bar",
			},
			expected: nil,
		},
		{
			desc:      "existent label with empty value",
			labelName: "foo",
			labels: map[string]string{
				"foo": "bar:",
			},
			expected: map[string]string{
				"Bar": "",
			},
		},
		{
			desc:      "one entry",
			labelName: "foo",
			labels: map[string]string{
				"foo": " Access-Control-Allow-Methods:POST,GET,OPTIONS ",
			},
			expected: map[string]string{
				"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
			},
		},
		{
			desc:      "several entry",
			labelName: "foo",
			labels: map[string]string{
				"foo": "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
			},
			expected: map[string]string{
				"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
				"Content-Type":                 "application/json; charset=utf-8",
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetMapValue(test.labels, test.labelName)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestGetStringMultipleStrict(t *testing.T) {
	testCases := []struct {
		desc        string
		labels      map[string]string
		labelNames  []string
		expected    map[string]string
		expectedErr bool
	}{
		{
			desc:     "empty labels names and empty labels map",
			labels:   map[string]string{},
			expected: map[string]string{},
		},
		{
			desc: "empty labels names",
			labels: map[string]string{
				"foo": "bar",
				"fii": "bir",
			},
			expected: map[string]string{},
		},
		{
			desc: "one label missing",
			labels: map[string]string{
				"foo": "bar",
				"fii": "bir",
				"fyy": "byr",
			},
			labelNames:  []string{"foo", "fii", "fuu"},
			expected:    nil,
			expectedErr: true,
		},
		{
			desc: "all labels are present",
			labels: map[string]string{
				"foo": "bar",
				"fii": "bir",
				"fyy": "byr",
			},
			labelNames: []string{"foo", "fii"},
			expected: map[string]string{
				"foo": "bar",
				"fii": "bir",
			},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, err := GetStringMultipleStrict(test.labels, test.labelNames...)
			if (err != nil) != test.expectedErr {
				t.Errorf("error = %v, wantErr %v", err, test.expectedErr)
				return
			}
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestHas(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    map[string]string
		labelName string
		expected  bool
	}{
		{
			desc:      "nil labels map",
			labelName: "foo",
		},
		{
			desc: "nonexistent label",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName: "fii",
			expected:  false,
		},
		{
			desc: "existent label",
			labels: map[string]string{
				"foo": "bar",
			},
			labelName: "foo",
			expected:  true,
		},
		{
			desc: "existent label with empty value",
			labels: map[string]string{
				"foo": "",
			},
			labelName: "foo",
			expected:  false,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := Has(test.labels, test.labelName)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestIsEnabled(t *testing.T) {
	testCases := []struct {
		desc             string
		labels           map[string]string
		exposedByDefault bool
		expected         bool
	}{
		{
			desc:             "empty labels map & exposedByDefault true",
			exposedByDefault: true,
			expected:         true,
		},
		{
			desc:             "empty labels map & exposedByDefault false",
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "exposedByDefault false and label enable true",
			labels: map[string]string{
				TraefikEnable: "true",
			},
			exposedByDefault: false,
			expected:         true,
		},
		{
			desc: "exposedByDefault false and label enable false",
			labels: map[string]string{
				TraefikEnable: "false",
			},
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "exposedByDefault true and label enable false",
			labels: map[string]string{
				TraefikEnable: "false",
			},
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "exposedByDefault true and label enable true",
			labels: map[string]string{
				TraefikEnable: "true",
			},
			exposedByDefault: true,
			expected:         true,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := IsEnabled(test.labels, test.exposedByDefault)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		prefix   string
		expected bool
	}{
		{
			desc:     "nil labels map",
			prefix:   "foo",
			expected: false,
		},
		{
			desc: "nonexistent prefix",
			labels: map[string]string{
				"foo.carotte": "bar",
			},
			prefix:   "fii",
			expected: false,
		},
		{
			desc: "existent prefix",
			labels: map[string]string{
				"foo.carotte": "bar",
			},
			prefix:   "foo",
			expected: true,
		},
		{
			desc: "existent prefix with empty value",
			labels: map[string]string{
				"foo.carotte": "",
			},
			prefix:   "foo",
			expected: false,
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := HasPrefix(test.labels, test.prefix)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetFuncString(t *testing.T) {
	testCases := []struct {
		labels       map[string]string
		labelName    string
		defaultValue string
		expected     string
	}{
		{
			labels:       nil,
			labelName:    TraefikProtocol,
			defaultValue: DefaultProtocol,
			expected:     "http",
		},
		{
			labels: map[string]string{
				TraefikProtocol: "https",
			},
			labelName:    TraefikProtocol,
			defaultValue: DefaultProtocol,
			expected:     "https",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			actual := GetFuncString(test.labelName, test.defaultValue)(test.labels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetSliceString(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    map[string]string
		labelName string
		expected  []string
	}{
		{
			desc:     "no whitelist-label",
			labels:   nil,
			expected: nil,
		},
		{
			desc: "whitelist-label with empty string",
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange: "",
			},
			labelName: TraefikFrontendWhiteListSourceRange,
			expected:  nil,
		},
		{
			desc: "whitelist-label with IPv4 mask",
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange: "1.2.3.4/16",
			},
			labelName: TraefikFrontendWhiteListSourceRange,
			expected: []string{
				"1.2.3.4/16",
			},
		},
		{
			desc: "whitelist-label with IPv6 mask",
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange: "fe80::/16",
			},
			labelName: TraefikFrontendWhiteListSourceRange,
			expected: []string{
				"fe80::/16",
			},
		},
		{
			desc: "whitelist-label with multiple masks",
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange: "1.1.1.1/24, 1234:abcd::42/32",
			},
			labelName: TraefikFrontendWhiteListSourceRange,
			expected: []string{
				"1.1.1.1/24",
				"1234:abcd::42/32",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetFuncSliceString(test.labelName)(test.labels)
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
