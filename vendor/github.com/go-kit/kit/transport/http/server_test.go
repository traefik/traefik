package http_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestServerBadDecode(t *testing.T) {
	handler := httptransport.NewServer(
		context.Background(),
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, *http.Request) (interface{}, error) { return struct{}{}, errors.New("dang") },
		func(context.Context, http.ResponseWriter, interface{}) error { return nil },
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Get(server.URL)
	if want, have := http.StatusBadRequest, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerBadEndpoint(t *testing.T) {
	handler := httptransport.NewServer(
		context.Background(),
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errors.New("dang") },
		func(context.Context, *http.Request) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, http.ResponseWriter, interface{}) error { return nil },
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Get(server.URL)
	if want, have := http.StatusServiceUnavailable, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerBadEncode(t *testing.T) {
	handler := httptransport.NewServer(
		context.Background(),
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, *http.Request) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, http.ResponseWriter, interface{}) error { return errors.New("dang") },
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Get(server.URL)
	if want, have := http.StatusInternalServerError, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerErrorEncoder(t *testing.T) {
	errTeapot := errors.New("teapot")
	code := func(err error) int {
		if e, ok := err.(httptransport.Error); ok && e.Err == errTeapot {
			return http.StatusTeapot
		}
		return http.StatusInternalServerError
	}
	handler := httptransport.NewServer(
		context.Background(),
		func(context.Context, interface{}) (interface{}, error) { return struct{}{}, errTeapot },
		func(context.Context, *http.Request) (interface{}, error) { return struct{}{}, nil },
		func(context.Context, http.ResponseWriter, interface{}) error { return nil },
		httptransport.ServerErrorEncoder(func(_ context.Context, err error, w http.ResponseWriter) { w.WriteHeader(code(err)) }),
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	resp, _ := http.Get(server.URL)
	if want, have := http.StatusTeapot, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestServerHappyPath(t *testing.T) {
	_, step, response := testServer(t)
	step()
	resp := <-response
	defer resp.Body.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("want %d, have %d (%s)", want, have, buf)
	}
}

func testServer(t *testing.T) (cancel, step func(), resp <-chan *http.Response) {
	var (
		ctx, cancelfn = context.WithCancel(context.Background())
		stepch        = make(chan bool)
		endpoint      = func(context.Context, interface{}) (interface{}, error) { <-stepch; return struct{}{}, nil }
		response      = make(chan *http.Response)
		handler       = httptransport.NewServer(
			ctx,
			endpoint,
			func(context.Context, *http.Request) (interface{}, error) { return struct{}{}, nil },
			func(context.Context, http.ResponseWriter, interface{}) error { return nil },
			httptransport.ServerBefore(func(ctx context.Context, r *http.Request) context.Context { return ctx }),
			httptransport.ServerAfter(func(ctx context.Context, w http.ResponseWriter) context.Context { return ctx }),
		)
	)
	go func() {
		server := httptest.NewServer(handler)
		defer server.Close()
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Error(err)
			return
		}
		response <- resp
	}()
	return cancelfn, func() { stepch <- true }, response
}
