package http_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

func TestClientEndpointEncodeError(t *testing.T) {
	var (
		sampleErr = errors.New("Oh no, an error")
		enc       = func(context.Context, *http.Request, interface{}) error { return sampleErr }
		dec       = func(context.Context, *http.Response) (interface{}, error) { return nil, nil }
	)

	u := &url.URL{
		Scheme: "https",
		Host:   "localhost",
		Path:   "/does/not/matter",
	}

	c := httptransport.NewClient(
		"GET",
		u,
		enc,
		dec,
	)

	_, err := c.Endpoint()(context.Background(), nil)
	if err == nil {
		t.Fatal("err == nil")
	}

	e, ok := err.(httptransport.Error)
	if !ok {
		t.Fatal("err is not of type github.com/go-kit/kit/transport/http.Error")
	}

	if want, have := sampleErr, e.Err; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func ExampleErrorOutput() {
	sampleErr := errors.New("oh no, an error")
	err := httptransport.Error{Domain: httptransport.DomainDo, Err: sampleErr}
	fmt.Println(err)
	// Output:
	// Do: oh no, an error
}
