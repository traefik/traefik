package kubernetes

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	kubeerror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestTranslateNotFoundError(t *testing.T) {
	testCases := []struct {
		err            error
		expectedExists bool
		expectedError  error
	}{
		{
			err:            kubeerror.NewNotFound(schema.GroupResource{}, "foo"),
			expectedExists: false,
			expectedError:  nil,
		},
		{
			err:            nil,
			expectedExists: true,
			expectedError:  nil,
		},
		{
			err:            fmt.Errorf("bar error"),
			expectedExists: false,
			expectedError:  fmt.Errorf("bar error"),
		},
	}
	for _, testCase := range testCases {
		exists, err := translateNotFoundError(testCase.err)
		assert.Equal(t, testCase.expectedExists, exists)
		assert.Equal(t, testCase.expectedError, err)
	}
}
