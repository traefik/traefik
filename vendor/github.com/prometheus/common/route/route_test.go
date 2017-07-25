package route

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestRedirect(t *testing.T) {
	router := New(nil).WithPrefix("/test/prefix")
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://localhost:9090/foo", nil)
	if err != nil {
		t.Fatalf("Error building test request: %s", err)
	}

	router.Redirect(w, r, "/some/endpoint", http.StatusFound)
	if w.Code != http.StatusFound {
		t.Fatalf("Unexpected redirect status code: got %d, want %d", w.Code, http.StatusFound)
	}

	want := "/test/prefix/some/endpoint"
	got := w.Header()["Location"][0]
	if want != got {
		t.Fatalf("Unexpected redirect location: got %s, want %s", got, want)
	}
}

func TestContextFunc(t *testing.T) {
	router := New(func(r *http.Request) (context.Context, error) {
		return context.WithValue(context.Background(), "testkey", "testvalue"), nil
	})

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		want := "testvalue"
		got := Context(r).Value("testkey")
		if want != got {
			t.Fatalf("Unexpected context value: want %q, got %q", want, got)
		}
	})

	r, err := http.NewRequest("GET", "http://localhost:9090/test", nil)
	if err != nil {
		t.Fatalf("Error building test request: %s", err)
	}
	router.ServeHTTP(nil, r)
}

func TestContextFnError(t *testing.T) {
	router := New(func(r *http.Request) (context.Context, error) {
		return context.Background(), fmt.Errorf("test error")
	})

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {})

	r, err := http.NewRequest("GET", "http://localhost:9090/test", nil)
	if err != nil {
		t.Fatalf("Error building test request: %s", err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Unexpected response status: got %q, want %q", w.Code, http.StatusBadRequest)
	}

	want := "Error creating request context: test error\n"
	got := w.Body.String()
	if want != got {
		t.Fatalf("Unexpected response body: got %q, want %q", got, want)
	}
}
