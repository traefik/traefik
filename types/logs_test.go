package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCodesSet(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
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
	tests := []struct {
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

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatusCodesString(t *testing.T) {
	tests := []struct {
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
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStatusCodesSetValue(t *testing.T) {
	tests := []struct {
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
	for _, test := range tests {
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
	tests := []struct {
		desc     string
		value    string
		expected *FieldsNames
	}{
		{
			desc:  "One value should return FieldsNames of size 1",
			value: "field-1=foo",
			expected: &FieldsNames{
				"field-1": "foo",
			},
		},
		{
			desc:  "Two values separated by space should return FieldsNames of size 2",
			value: "field-1=foo field-2=bar",
			expected: &FieldsNames{
				"field-1": "foo",
				"field-2": "bar",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fieldsNames := &FieldsNames{}
			err := fieldsNames.Set(test.value)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, fieldsNames)
		})
	}
}

func TestFieldsNamesGet(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsNames
		expected FieldsNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsNames{"field-1": "foo"},
			expected: FieldsNames{"field-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldsNames{"field-1": "foo", "field-2": "bar"},
			expected: FieldsNames{"field-1": "foo", "field-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldsNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
			expected: FieldsNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsNamesString(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsNames
		expected string
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsNames{"field-1": "foo"},
			expected: "map[field-1:foo]",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsNamesSetValue(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsNames
		expected *FieldsNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsNames{"field-1": "foo"},
			expected: &FieldsNames{"field-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldsNames{"field-1": "foo", "field-2": "bar"},
			expected: &FieldsNames{"field-1": "foo", "field-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldsNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
			expected: &FieldsNames{"field-1": "foo", "field-2": "bar", "field-3": "powpow"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fieldsNames := &FieldsNames{}
			fieldsNames.SetValue(test.values)
			assert.Equal(t, test.expected, fieldsNames)
		})
	}
}

func TestFieldsHeadersNamesSet(t *testing.T) {
	tests := []struct {
		desc     string
		value    string
		expected *FieldsHeadersNames
	}{
		{
			desc:  "One value should return FieldsNames of size 1",
			value: "X-HEADER-1=foo",
			expected: &FieldsHeadersNames{
				"X-HEADER-1": "foo",
			},
		},
		{
			desc:  "Two values separated by space should return FieldsNames of size 2",
			value: "X-HEADER-1=foo X-HEADER-2=bar",
			expected: &FieldsHeadersNames{
				"X-HEADER-1": "foo",
				"X-HEADER-2": "bar",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headersNames := &FieldsHeadersNames{}
			err := headersNames.Set(test.value)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, headersNames)
		})
	}
}

func TestFieldsHeadersNamesGet(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsHeadersNames
		expected FieldsHeadersNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo"},
			expected: FieldsHeadersNames{"X-HEADER-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
			expected: FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
			expected: FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsHeadersNamesString(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsHeadersNames
		expected string
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo"},
			expected: "map[X-HEADER-1:foo]",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := test.values.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestFieldsHeadersNamesSetValue(t *testing.T) {
	tests := []struct {
		desc     string
		values   FieldsHeadersNames
		expected *FieldsHeadersNames
	}{
		{
			desc:     "Should return 1 value",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo"},
			expected: &FieldsHeadersNames{"X-HEADER-1": "foo"},
		},
		{
			desc:     "Should return 2 values",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
			expected: &FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar"},
		},
		{
			desc:     "Should return 3 values",
			values:   FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
			expected: &FieldsHeadersNames{"X-HEADER-1": "foo", "X-HEADER-2": "bar", "X-HEADER-3": "powpow"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headersNames := &FieldsHeadersNames{}
			headersNames.SetValue(test.values)
			assert.Equal(t, test.expected, headersNames)
		})
	}
}
