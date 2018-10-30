package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCodesSet(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected StatusCodes
	}{
		{
			desc:     "One value should return StatusCodes of size 1",
			value:    "200",
			expected: StatusCodes{"200"},
		},
		{
			desc:     "Two values separated by comma should return StatusCodes of size 2",
			value:    "200,400",
			expected: StatusCodes{"200", "400"},
		},
		{
			desc:     "Two values separated by semicolon should return StatusCodes of size 2",
			value:    "200;400",
			expected: StatusCodes{"200", "400"},
		},
		{
			desc:     "Three values separated by comma and semicolon should return StatusCodes of size 3",
			value:    "200,400;500",
			expected: StatusCodes{"200", "400", "500"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var statusCodes StatusCodes
			err := statusCodes.Set(test.value)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, statusCodes)
		})
	}
}

func TestStatusCodesGet(t *testing.T) {
	testCases := []struct {
		desc     string
		values   StatusCodes
		expected StatusCodes
	}{
		{
			desc:     "Should return 1 value",
			values:   StatusCodes{"200"},
			expected: StatusCodes{"200"},
		},
		{
			desc:     "Should return 2 values",
			values:   StatusCodes{"200", "400"},
			expected: StatusCodes{"200", "400"},
		},
		{
			desc:     "Should return 3 values",
			values:   StatusCodes{"200", "400", "500"},
			expected: StatusCodes{"200", "400", "500"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatusCodesString(t *testing.T) {
	testCases := []struct {
		desc     string
		values   StatusCodes
		expected string
	}{
		{
			desc:     "Should return 1 value",
			values:   StatusCodes{"200"},
			expected: "[200]",
		},
		{
			desc:     "Should return 2 values",
			values:   StatusCodes{"200", "400"},
			expected: "[200 400]",
		},
		{
			desc:     "Should return 3 values",
			values:   StatusCodes{"200", "400", "500"},
			expected: "[200 400 500]",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatusCodesSetValue(t *testing.T) {
	testCases := []struct {
		desc     string
		values   StatusCodes
		expected StatusCodes
	}{
		{
			desc:     "Should return 1 value",
			values:   StatusCodes{"200"},
			expected: StatusCodes{"200"},
		},
		{
			desc:     "Should return 2 values",
			values:   StatusCodes{"200", "400"},
			expected: StatusCodes{"200", "400"},
		},
		{
			desc:     "Should return 3 values",
			values:   StatusCodes{"200", "400", "500"},
			expected: StatusCodes{"200", "400", "500"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var slice StatusCodes
			slice.SetValue(test.values)
			assert.Equal(t, test.expected, slice)
		})
	}
}

func TestFieldsNamesSet(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected *FieldNames
	}{
		{
			desc:  "One value should return FieldNames of size 1",
			value: "field-1=foo",
			expected: &FieldNames{
				"field-1": "foo",
			},
		},
		{
			desc:  "Two values separated by space should return FieldNames of size 2",
			value: "field-1=foo field-2=bar",
			expected: &FieldNames{
				"field-1": "foo",
				"field-2": "bar",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fieldsNames := &FieldNames{}
			err := fieldsNames.Set(test.value)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, fieldsNames)
		})
	}
}

func TestFieldsNamesGet(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldNames
		expected FieldNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldNames{"field-1": "foo"},
			expected: FieldNames{"field-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldNames{"field-1": "foo", "field-2": "bar"},
			expected: FieldNames{"field-1": "foo", "field-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
			expected: FieldNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsNamesString(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldNames
		expected string
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldNames{"field-1": "foo"},
			expected: "map[field-1:foo]",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsNamesSetValue(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldNames
		expected *FieldNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldNames{"field-1": "foo"},
			expected: &FieldNames{"field-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldNames{"field-1": "foo", "field-2": "bar"},
			expected: &FieldNames{"field-1": "foo", "field-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
			expected: &FieldNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fieldsNames := &FieldNames{}
			fieldsNames.SetValue(test.values)
			assert.Equal(t, test.expected, fieldsNames)
		})
	}
}

func TestFieldsHeadersNamesSet(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected *FieldHeaderNames
	}{
		{
			desc:  "One value should return FieldNames of size 1",
			value: "X-HEADER-1=foo",
			expected: &FieldHeaderNames{
				"X-HEADER-1": "foo",
			},
		},
		{
			desc:  "Two values separated by space should return FieldNames of size 2",
			value: "X-HEADER-1=foo X-HEADER-2=bar",
			expected: &FieldHeaderNames{
				"X-HEADER-1": "foo",
				"X-HEADER-2": "bar",
			},
		},
		{
			desc:  "Two values separated by space with escaped double quotes should return FieldNames of size 2",
			value: "\"X-HEADER-1=foo X-HEADER-2=bar\"",
			expected: &FieldHeaderNames{
				"X-HEADER-1": "foo",
				"X-HEADER-2": "bar",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headersNames := &FieldHeaderNames{}
			err := headersNames.Set(test.value)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, headersNames)
		})
	}
}

func TestFieldsHeadersNamesGet(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldHeaderNames
		expected FieldHeaderNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldHeaderNames{"X-HEADER-1": "foo"},
			expected: FieldHeaderNames{"X-HEADER-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
			expected: FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
			expected: FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsHeadersNamesString(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldHeaderNames
		expected string
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldHeaderNames{"X-HEADER-1": "foo"},
			expected: "map[X-HEADER-1:foo]",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsHeadersNamesSetValue(t *testing.T) {
	testCases := []struct {
		desc     string
		values   FieldHeaderNames
		expected *FieldHeaderNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldHeaderNames{"X-HEADER-1": "foo"},
			expected: &FieldHeaderNames{"X-HEADER-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
			expected: &FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
			expected: &FieldHeaderNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headersNames := &FieldHeaderNames{}
			headersNames.SetValue(test.values)
			assert.Equal(t, test.expected, headersNames)
		})
	}
}
