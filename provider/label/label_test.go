package label

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/types"
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

func TestGetStringValueP(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       *map[string]string
		labelName    string
		defaultValue string
		expected     string
	}{
		{
			desc:         "nil labels map",
			labels:       nil,
			labelName:    "foo",
			defaultValue: "default",
			expected:     "default",
		},
		{
			desc: "existing label",
			labels: &map[string]string{
				"foo": "bar",
			},
			labelName:    "foo",
			defaultValue: "default",
			expected:     "bar",
		},
		{
			desc: "non existing label",
			labels: &map[string]string{
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

			got := GetStringValueP(test.labels, test.labelName, test.defaultValue)
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

func TestGetIntValueP(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       *map[string]string
		labelName    string
		defaultValue int
		expected     int
	}{
		{
			desc:         "nil map",
			labels:       nil,
			labelName:    "foo",
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "invalid int value",
			labelName: "foo",
			labels: &map[string]string{
				"foo": "bar",
			},
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "negative int value",
			labelName: "foo",
			labels: &map[string]string{
				"foo": "-1",
			},
			defaultValue: 666,
			expected:     -1,
		},
		{
			desc:      "positive int value",
			labelName: "foo",
			labels: &map[string]string{
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

			got := GetIntValueP(test.labels, test.labelName, test.defaultValue)
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

func TestGetInt64ValueP(t *testing.T) {
	testCases := []struct {
		desc         string
		labels       *map[string]string
		labelName    string
		defaultValue int64
		expected     int64
	}{
		{
			desc:         "nil map",
			labels:       nil,
			labelName:    "foo",
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "invalid int value",
			labelName: "foo",
			labels: &map[string]string{
				"foo": "bar",
			},
			defaultValue: 666,
			expected:     666,
		},
		{
			desc:      "negative int value",
			labelName: "foo",
			labels: &map[string]string{
				"foo": "-1",
			},
			defaultValue: 666,
			expected:     -1,
		},
		{
			desc:      "positive int value",
			labelName: "foo",
			labels: &map[string]string{
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

			got := GetInt64ValueP(test.labels, test.labelName, test.defaultValue)
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

func TestGetSliceStringValueP(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    *map[string]string
		labelName string
		expected  []string
	}{
		{
			desc:      "nil map",
			labels:    nil,
			labelName: "foo",
			expected:  nil,
		},
		{
			desc: "one value, not split",
			labels: &map[string]string{
				"foo": "bar",
			},
			labelName: "foo",
			expected:  []string{"bar"},
		},
		{
			desc: "several values",
			labels: &map[string]string{
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

			got := GetSliceStringValueP(test.labels, test.labelName)
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

func TestHasP(t *testing.T) {
	testCases := []struct {
		desc      string
		labels    *map[string]string
		labelName string
		expected  bool
	}{
		{
			desc:      "nil labels map",
			labelName: "foo",
		},
		{
			desc: "nonexistent label",
			labels: &map[string]string{
				"foo": "bar",
			},
			labelName: "fii",
			expected:  false,
		},
		{
			desc: "existent label",
			labels: &map[string]string{
				"foo": "bar",
			},
			labelName: "foo",
			expected:  true,
		},
		{
			desc: "existent label with empty value",
			labels: &map[string]string{
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

			got := HasP(test.labels, test.labelName)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestExtractServiceProperties(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected RoadProperties
	}{
		{
			desc:     "empty labels map",
			expected: RoadProperties{},
		},
		{
			desc: "valid label names",
			labels: map[string]string{
				"traefik.foo.port":         "bar",
				"traefik.foo.frontend.bar": "1bar",
				"traefik.foo.backend":      "3bar",
			},
			expected: RoadProperties{
				"foo": RoadPropertyValues{
					"port":         "bar",
					"frontend.bar": "1bar",
					"backend":      "3bar",
				},
			},
		},
		{
			desc: "invalid label names",
			labels: map[string]string{
				"foo.frontend.bar":             "1bar",
				"traefik.foo.frontend.":        "2bar",
				"traefik.foo.port.bar":         "barbar",
				"traefik.foo.frontend":         "0bar",
				"traefik.frontend.foo.backend": "0bar",
			},
			expected: RoadProperties{},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := ExtractServiceProperties(test.labels)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestExtractServicePropertiesP(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   *map[string]string
		expected RoadProperties
	}{
		{
			desc:     "nil labels map",
			expected: RoadProperties{},
		},
		{
			desc: "valid label names",
			labels: &map[string]string{
				"traefik.foo.port":         "bar",
				"traefik.foo.frontend.bar": "1bar",
				"traefik.foo.backend":      "3bar",
			},
			expected: RoadProperties{
				"foo": RoadPropertyValues{
					"port":         "bar",
					"frontend.bar": "1bar",
					"backend":      "3bar",
				},
			},
		},
		{
			desc: "invalid label names",
			labels: &map[string]string{
				"foo.frontend.bar":             "1bar",
				"traefik.foo.frontend.":        "2bar",
				"traefik.foo.port.bar":         "barbar",
				"traefik.foo.frontend":         "0bar",
				"traefik.frontend.foo.backend": "0bar",
			},
			expected: RoadProperties{},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := ExtractServicePropertiesP(test.labels)
			assert.EqualValues(t, test.expected, got)
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

func TestIsEnabledP(t *testing.T) {
	testCases := []struct {
		desc             string
		labels           *map[string]string
		exposedByDefault bool
		expected         bool
	}{
		{
			desc:             "nil labels map & exposedByDefault true",
			exposedByDefault: true,
			expected:         true,
		},
		{
			desc:             "nil labels map & exposedByDefault false",
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "exposedByDefault false and label enable true",
			labels: &map[string]string{
				TraefikEnable: "true",
			},
			exposedByDefault: false,
			expected:         true,
		},
		{
			desc: "exposedByDefault false and label enable false",
			labels: &map[string]string{
				TraefikEnable: "false",
			},
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "exposedByDefault true and label enable false",
			labels: &map[string]string{
				TraefikEnable: "false",
			},
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "exposedByDefault true and label enable true",
			labels: &map[string]string{
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

			got := IsEnabledP(test.labels, test.exposedByDefault)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetServiceLabel(t *testing.T) {
	testCases := []struct {
		desc        string
		labelName   string
		serviceName string
		expected    string
	}{
		{
			desc:      "without service name",
			labelName: TraefikPort,
			expected:  TraefikPort,
		},
		{
			desc:        "with service name",
			labelName:   TraefikPort,
			serviceName: "bar",
			expected:    "traefik.bar.port",
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetServiceLabel(test.labelName, test.serviceName)
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

func TestParseErrorPages(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected map[string]*types.ErrorPage
	}{
		{
			desc: "2 errors pages",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageStatus:  "404",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageBackend: "foo_backend",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageQuery:   "foo_query",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageStatus:  "500,600",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageBackend: "bar_backend",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageQuery:   "bar_query",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status:  []string{"404"},
					Query:   "foo_query",
					Backend: "foo_backend",
				},
				"bar": {
					Status:  []string{"500", "600"},
					Query:   "bar_query",
					Backend: "bar_backend",
				},
			},
		},
		{
			desc: "only status field",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageStatus: "404",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status: []string{"404"},
				},
			},
		},
		{
			desc: "invalid field",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + "courgette": "404",
			},
			expected: map[string]*types.ErrorPage{"foo": {}},
		},
		{
			desc:     "no error pages labels",
			labels:   map[string]string{},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pages := ParseErrorPages(test.labels, Prefix+BaseFrontendErrorPage, RegexpFrontendErrorPage)

			assert.EqualValues(t, test.expected, pages)
		})
	}
}

func TestParseRateSets(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected map[string]*types.Rate
	}{
		{
			desc: "2 rate limits",
			labels: map[string]string{
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitPeriod:  "6",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitAverage: "12",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitBurst:   "18",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitPeriod:  "3",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitAverage: "6",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitBurst:   "9",
			},
			expected: map[string]*types.Rate{
				"foo": {
					Period:  flaeg.Duration(6 * time.Second),
					Average: 12,
					Burst:   18,
				},
				"bar": {
					Period:  flaeg.Duration(3 * time.Second),
					Average: 6,
					Burst:   9,
				},
			},
		},

		{
			desc:     "no rate limits labels",
			labels:   map[string]string{},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rateSets := ParseRateSets(test.labels, Prefix+BaseFrontendRateLimit, RegexpFrontendRateLimit)

			assert.EqualValues(t, test.expected, rateSets)
		})
	}
}

func TestExtractTraefikLabels(t *testing.T) {
	testCases := []struct {
		desc         string
		prefix       string
		originLabels map[string]string
		expected     RoadProperties
	}{
		{
			desc:         "nil labels map",
			prefix:       "traefik",
			originLabels: nil,
			expected:     RoadProperties{"": {}},
		},
		{
			desc:   "container labels",
			prefix: "traefik",
			originLabels: map[string]string{
				"frontend.priority": "foo", // missing prefix: skip
				"traefik.port":      "bar",
			},
			expected: RoadProperties{
				"": {
					"traefik.port": "bar",
				},
			},
		},
		{
			desc:   "road labels: only road no default",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.goo.frontend.priority": "A",
				"traefik.goo.port":              "D",
				"traefik.port":                  "C",
			},
			expected: RoadProperties{
				"goo": {
					"traefik.frontend.priority": "A",
					"traefik.port":              "D",
				},
			},
		},
		{
			desc:   "road labels: use default",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.guu.frontend.priority": "B",
				"traefik.port":                  "C",
			},
			expected: RoadProperties{
				"guu": {
					"traefik.frontend.priority": "B",
					"traefik.port":              "C",
				},
			},
		},
		{
			desc:   "road labels: several roads",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.goo.frontend.priority": "A",
				"traefik.goo.port":              "D",
				"traefik.guu.frontend.priority": "B",
				"traefik.port":                  "C",
			},
			expected: RoadProperties{
				"goo": {
					"traefik.frontend.priority": "A",
					"traefik.port":              "D",
				},
				"guu": {
					"traefik.frontend.priority": "B",
					"traefik.port":              "C",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := ExtractTraefikLabels(test.originLabels)
			assert.Equal(t, test.expected, actual)
		})
	}
}
