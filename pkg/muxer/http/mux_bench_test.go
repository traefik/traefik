package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func benchMuxer(tb testing.TB, routes int, hostRule bool) *Muxer {
	tb.Helper()

	parser, err := NewSyntaxParser()
	if err != nil {
		tb.Fatal(err)
	}

	muxer := NewMuxer(parser, []string{"file"})

	for i := range routes {
		rule := fmt.Sprintf("PathPrefix(`/api/v1/svc%05d`)", i)
		if hostRule {
			rule = fmt.Sprintf("Host(`t%05d.example.com`) && PathPrefix(`/api/v1/svc%05d`)", i, i)
		}

		if err := muxer.AddRoute(rule, "v3", 100, "file", http.NotFoundHandler()); err != nil {
			tb.Fatal(err)
		}
	}

	return muxer
}

func BenchmarkMuxerServeHTTPPathMatchers(b *testing.B) {
	for _, routes := range []int{100, 1000} {
		b.Run(fmt.Sprintf("routes=%d", routes), func(b *testing.B) {
			muxer := benchMuxer(b, routes, false)
			req := httptest.NewRequest(http.MethodGet, "http://example.com/nothing/here", nil)
			rw := httptest.NewRecorder()

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				muxer.ServeHTTP(rw, req)
			}
		})
	}
}

func BenchmarkWithRoutingPath(b *testing.B) {
	b.Run("unencoded", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/service/resource/subresource", nil)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := withRoutingPath(req); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("encoded", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/service%20name/resource", nil)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			if _, err := withRoutingPath(req); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAddRoute(b *testing.B) {
	for _, routes := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("routes=%d", routes), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				benchMuxer(b, routes, true)
			}
		})
	}
}
