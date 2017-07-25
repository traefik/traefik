package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

type TestResponse struct {
	Body   io.ReadCloser
	String string
}

func TestHTTPClient(t *testing.T) {
	var (
		testbody = "testbody"
		encode   = func(context.Context, *http.Request, interface{}) error { return nil }
		decode   = func(_ context.Context, r *http.Response) (interface{}, error) {
			buffer := make([]byte, len(testbody))
			r.Body.Read(buffer)
			return TestResponse{r.Body, string(buffer)}, nil
		}
		headers        = make(chan string, 1)
		headerKey      = "X-Foo"
		headerVal      = "abcde"
		afterHeaderKey = "X-The-Dude"
		afterHeaderVal = "Abides"
		afterVal       = ""
		afterFunc      = func(ctx context.Context, r *http.Response) context.Context {
			afterVal = r.Header.Get(afterHeaderKey)
			return ctx
		}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers <- r.Header.Get(headerKey)
		w.Header().Set(afterHeaderKey, afterHeaderVal)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testbody))
	}))

	client := httptransport.NewClient(
		"GET",
		mustParse(server.URL),
		encode,
		decode,
		httptransport.ClientBefore(httptransport.SetRequestHeader(headerKey, headerVal)),
		httptransport.ClientAfter(afterFunc),
	)

	res, err := client.Endpoint()(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	var have string
	select {
	case have = <-headers:
	case <-time.After(time.Millisecond):
		t.Fatalf("timeout waiting for %s", headerKey)
	}
	// Check that Request Header was successfully received
	if want := headerVal; want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// Check that Response header set from server was received in SetClientAfter
	if want, have := afterVal, afterHeaderVal; want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// Check that the response was successfully decoded
	response, ok := res.(TestResponse)
	if !ok {
		t.Fatal("response should be TestResponse")
	}
	if want, have := testbody, response.String; want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	// Check that response body was closed
	b := make([]byte, 1)
	_, err = response.Body.Read(b)
	if err == nil {
		t.Fatal("wanted error, got none")
	}
	if doNotWant, have := io.EOF, err; doNotWant == have {
		t.Errorf("do not want %q, have %q", doNotWant, have)
	}
}

func TestHTTPClientBufferedStream(t *testing.T) {
	var (
		testbody = "testbody"
		encode   = func(context.Context, *http.Request, interface{}) error { return nil }
		decode   = func(_ context.Context, r *http.Response) (interface{}, error) {
			return TestResponse{r.Body, ""}, nil
		}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testbody))
	}))

	client := httptransport.NewClient(
		"GET",
		mustParse(server.URL),
		encode,
		decode,
		httptransport.BufferedStream(true),
	)

	res, err := client.Endpoint()(context.Background(), struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	// Check that the response was successfully decoded
	response, ok := res.(TestResponse)
	if !ok {
		t.Fatal("response should be TestResponse")
	}

	// Check that response body was NOT closed
	b := make([]byte, len(testbody))
	_, err = response.Body.Read(b)
	if want, have := io.EOF, err; have != want {
		t.Fatalf("want %q, have %q", want, have)
	}
	if want, have := testbody, string(b); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
