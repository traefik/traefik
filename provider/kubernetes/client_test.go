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
		err            error
		desc           string
		expectedExists bool
		expectedError  error
	}{
		{
			err:            kubeerror.NewNotFound(schema.GroupResource{}, "foo"),
			desc:           "kubernetes not found error",
			expectedExists: false,
			expectedError:  nil,
		},
		{
			err:            nil,
			desc:           "nil error",
			expectedExists: true,
			expectedError:  nil,
		},
		{
			err:            fmt.Errorf("bar error"),
			desc:           "not a kubernetes not found error",
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
