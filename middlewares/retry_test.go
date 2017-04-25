package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		rbody         *bytes.Buffer
		rt            *Retry
		responseCodes []int
	}{
		{
			rt: &Retry{
				attempts: 4,
			},
			rbody:         &bytes.Buffer{},
			responseCodes: []int{},
		},
		{
			rt: &Retry{
				attempts: 3,
			},
			rbody:         &bytes.Buffer{},
			responseCodes: []int{http.StatusGatewayTimeout, http.StatusGatewayTimeout},
		},
		{
			rt: &Retry{
				attempts: 1,
			},
			rbody:         bytes.NewBuffer([]byte([]byte("this is a test"))),
			responseCodes: []int{http.StatusGatewayTimeout, http.StatusGatewayTimeout},
		},
	}

	for i, test := range tests {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			callCount := 0
			test.rt.next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				reqBodyBytes := &bytes.Buffer{}
				io.Copy(reqBodyBytes, r.Body)
				if bytes.Compare(reqBodyBytes.Bytes(), test.rbody.Bytes()) != 0 {
					t.Fatalf("expected request body %q, got %q", reqBodyBytes.Bytes(), test.rbody.Bytes())
				}
				callCount++
				if callCount-1 > len(test.responseCodes) {
					t.Fatalf("Called too many times")
					return
				}
				if callCount-1 == len(test.responseCodes) {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(test.responseCodes[callCount-1])
			})

			req, _ := http.NewRequest("GET", "/path", test.rbody)
			rr := httptest.NewRecorder()

			test.rt.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK && callCount != test.rt.attempts {
				t.Fatalf("not called enough times, expected %#v, got %#v", test.rt.attempts, callCount)
			}
		})
	}
}
