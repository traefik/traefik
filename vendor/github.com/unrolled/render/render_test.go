package render

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

/* Benchmarks */
func BenchmarkNormalJSON(b *testing.B) {
	render := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 200, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(res, req)
	}
}

func BenchmarkStreamingJSON(b *testing.B) {
	render := New(Options{
		StreamingJSON: true,
	})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 200, Greeting{"hello", "world"})
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foo", nil)

	for i := 0; i < b.N; i++ {
		h.ServeHTTP(res, req)
	}
}

/* Test Helper */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected ||%#v|| (type %v) - Got ||%#v|| (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func expectNil(t *testing.T, a interface{}) {
	if a != nil {
		t.Errorf("Expected ||nil|| - Got ||%#v|| (type %v)", a, reflect.TypeOf(a))
	}
}

func expectNotNil(t *testing.T, a interface{}) {
	if a == nil {
		t.Errorf("Expected ||not nil|| - Got ||nil|| (type %v)", reflect.TypeOf(a))
	}
}
