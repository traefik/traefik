package accesslog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type rwWithCloseNotify struct {
	*httptest.ResponseRecorder
}

func (r *rwWithCloseNotify) CloseNotify() <-chan bool {
	panic("implement me")
}

func TestCloseNotifier(t *testing.T) {
	testCases := []struct {
		rw                      http.ResponseWriter
		desc                    string
		implementsCloseNotifier bool
	}{
		{
			rw:                      httptest.NewRecorder(),
			desc:                    "does not implement CloseNotifier",
			implementsCloseNotifier: false,
		},
		{
			rw:                      &rwWithCloseNotify{httptest.NewRecorder()},
			desc:                    "implements CloseNotifier",
			implementsCloseNotifier: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, ok := test.rw.(http.CloseNotifier)
			assert.Equal(t, test.implementsCloseNotifier, ok)

			rw := newCaptureResponseWriter(test.rw)
			_, impl := rw.(http.CloseNotifier)
			assert.Equal(t, test.implementsCloseNotifier, impl)
		})
	}
}
