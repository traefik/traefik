package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	kubeerror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestTranslateNotFoundError(t *testing.T) {
	testCases := []struct {
		desc           string
		err            error
		expectedExists bool
		expectedError  error
	}{
		{
			desc:           "kubernetes not found error",
			err:            kubeerror.NewNotFound(schema.GroupResource{}, "foo"),
			expectedExists: false,
			expectedError:  nil,
		},
		{
			desc:           "nil error",
			err:            nil,
			expectedExists: true,
			expectedError:  nil,
		},
		{
			desc:           "not a kubernetes not found error",
			err:            fmt.Errorf("bar error"),
			expectedExists: false,
			expectedError:  fmt.Errorf("bar error"),
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			exists, err := translateNotFoundError(test.err)
			assert.Equal(t, test.expectedExists, exists)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
