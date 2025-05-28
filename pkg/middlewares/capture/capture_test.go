package capture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapture(t *testing.T) {
	wrapMiddleware := func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			capt, err := FromContext(req.Context())
			require.NoError(t, err)

			_, err = fmt.Fprintf(rw, "%d,%d,%d,", capt.RequestSize(), capt.ResponseSize(), capt.StatusCode())
			require.NoError(t, err)

			next.ServeHTTP(rw, req)

			_, err = fmt.Fprintf(rw, ",%d,%d,%d", capt.RequestSize(), capt.ResponseSize(), capt.StatusCode())
			require.NoError(t, err)
		}), nil
	}

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte("foo"))
		require.NoError(t, err)

		all, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, "bar", string(all))
	})

	chain := alice.New()
	chain = chain.Append(Wrap)
	chain = chain.Append(wrapMiddleware)
	handlers, err := chain.Then(handler)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte("bar")))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	handlers.ServeHTTP(recorder, request)
	// 3 = len("bar")
	// 9 = len("0,0,0,toto")
	assert.Equal(t, "0,0,0,foo,3,9,200", recorder.Body.String())
}

// BenchmarkCapture with response writer and request reader
// $ go test -bench=. ./pkg/middlewares/capture/
// goos: linux
// goarch: amd64
// pkg: github.com/traefik/traefik/v3/pkg/middlewares/capture
// cpu: Intel(R) Core(TM) i7-10750H CPU @ 2.60GHz
// BenchmarkCapture/2k-12				280507	 4015 ns/op	 510.03 MB/s	  5072 B/op	14 allocs/op
// BenchmarkCapture/20k-12				135726	 8301 ns/op	2467.26 MB/s	 41936 B/op	14 allocs/op
// BenchmarkCapture/100k-12				 45494	26059 ns/op	3929.54 MB/s	213968 B/op	14 allocs/op
// BenchmarkCapture/2k_captured-12		263713	 4356 ns/op	 470.20 MB/s	  5552 B/op	18 allocs/op
// BenchmarkCapture/20k_captured-12		132243	 8790 ns/op	2329.98 MB/s	 42416 B/op	18 allocs/op
// BenchmarkCapture/100k_captured-12	 45650	26587 ns/op	3851.57 MB/s	214448 B/op	18 allocs/op
// BenchmarkCapture/2k_body-12			274135	 7471 ns/op	 274.12 MB/s	  5624 B/op	20 allocs/op
// BenchmarkCapture/20k_body-12			130206	21149 ns/op	 968.36 MB/s	 42488 B/op	20 allocs/op
// BenchmarkCapture/100k_body-12		 41600	51716 ns/op	1980.06 MB/s	214520 B/op	20 allocs/op
// PASS
func BenchmarkCapture(b *testing.B) {
	testCases := []struct {
		name    string
		size    int
		capture bool
		body    bool
	}{
		{
			name: "2k",
			size: 2048,
		},
		{
			name: "20k",
			size: 20480,
		},
		{
			name: "100k",
			size: 102400,
		},
		{
			name:    "2k captured",
			size:    2048,
			capture: true,
		},
		{
			name:    "20k captured",
			size:    20480,
			capture: true,
		},
		{
			name:    "100k captured",
			size:    102400,
			capture: true,
		},
		{
			name: "2k body",
			size: 2048,
			body: true,
		},
		{
			name: "20k body",
			size: 20480,
			body: true,
		},
		{
			name: "100k body",
			size: 102400,
			body: true,
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			baseBody := generateBytes(test.size)

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				n, err := rw.Write(baseBody)
				require.Equal(b, test.size, n)
				require.NoError(b, err)
			})

			var body io.Reader
			if test.body {
				body = bytes.NewReader(baseBody)
			}

			req, err := http.NewRequest(http.MethodGet, "http://foo/", body)
			require.NoError(b, err)

			chain := alice.New()
			if test.capture || test.body {
				chain = chain.Append(Wrap)
			}
			handlers, err := chain.Then(next)
			require.NoError(b, err)

			b.ReportAllocs()
			b.SetBytes(int64(test.size))
			b.ResetTimer()
			for range b.N {
				runBenchmark(b, test.size, req, handlers)
			}
		})
	}
}

func runBenchmark(b *testing.B, size int, req *http.Request, handler http.Handler) {
	b.Helper()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if code := recorder.Code; code != 200 {
		b.Fatalf("Expected 200 but got %d", code)
	}

	assert.Len(b, recorder.Body.String(), size)
}

func generateBytes(length int) []byte {
	var value []byte
	for i := range length {
		value = append(value, 0x61+byte(i%26))
	}
	return value
}

func TestRequestReader(t *testing.T) {
	buff := bytes.NewBufferString("foo")
	rr := readCounter{source: io.NopCloser(buff)}
	assert.Equal(t, int64(0), rr.size)

	n, err := rr.Read([]byte("bar"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	err = rr.Close()
	require.NoError(t, err)
	assert.Equal(t, int64(3), rr.size)
}
