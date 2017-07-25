package http

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a URL and provides a method that implements endpoint.Endpoint.
type Client struct {
	client         *http.Client
	method         string
	tgt            *url.URL
	enc            EncodeRequestFunc
	dec            DecodeResponseFunc
	before         []RequestFunc
	after          []ClientResponseFunc
	bufferedStream bool
}

// NewClient constructs a usable Client for a single remote method.
func NewClient(
	method string,
	tgt *url.URL,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...ClientOption,
) *Client {
	c := &Client{
		client:         http.DefaultClient,
		method:         method,
		tgt:            tgt,
		enc:            enc,
		dec:            dec,
		before:         []RequestFunc{},
		after:          []ClientResponseFunc{},
		bufferedStream: false,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// SetClient sets the underlying HTTP client used for requests.
// By default, http.DefaultClient is used.
func SetClient(client *http.Client) ClientOption {
	return func(c *Client) { c.client = client }
}

// ClientBefore sets the RequestFuncs that are applied to the outgoing HTTP
// request before it's invoked.
func ClientBefore(before ...RequestFunc) ClientOption {
	return func(c *Client) { c.before = before }
}

// ClientAfter sets the ClientResponseFuncs applied to the incoming HTTP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func ClientAfter(after ...ClientResponseFunc) ClientOption {
	return func(c *Client) { c.after = after }
}

// BufferedStream sets whether the Response.Body is left open, allowing it
// to be read from later. Useful for transporting a file as a buffered stream.
func BufferedStream(buffered bool) ClientOption {
	return func(c *Client) { c.bufferedStream = buffered }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		req, err := http.NewRequest(c.method, c.tgt.String(), nil)
		if err != nil {
			return nil, Error{Domain: DomainNewRequest, Err: err}
		}

		if err = c.enc(ctx, req, request); err != nil {
			return nil, Error{Domain: DomainEncode, Err: err}
		}

		for _, f := range c.before {
			ctx = f(ctx, req)
		}

		resp, err := ctxhttp.Do(ctx, c.client, req)
		if err != nil {
			return nil, Error{Domain: DomainDo, Err: err}
		}
		if !c.bufferedStream {
			defer resp.Body.Close()
		}

		for _, f := range c.after {
			ctx = f(ctx, resp)
		}

		response, err := c.dec(ctx, resp)
		if err != nil {
			return nil, Error{Domain: DomainDecode, Err: err}
		}

		return response, nil
	}
}
