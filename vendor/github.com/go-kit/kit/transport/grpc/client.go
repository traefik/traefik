package grpc

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/go-kit/kit/endpoint"
)

// Client wraps a gRPC connection and provides a method that implements
// endpoint.Endpoint.
type Client struct {
	client      *grpc.ClientConn
	serviceName string
	method      string
	enc         EncodeRequestFunc
	dec         DecodeResponseFunc
	grpcReply   reflect.Type
	before      []RequestFunc
}

// NewClient constructs a usable Client for a single remote endpoint.
// Pass an zero-value protobuf message of the RPC response type as
// the grpcReply argument.
func NewClient(
	cc *grpc.ClientConn,
	serviceName string,
	method string,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	grpcReply interface{},
	options ...ClientOption,
) *Client {
	if strings.IndexByte(serviceName, '.') == -1 {
		serviceName = "pb." + serviceName
	}
	c := &Client{
		client: cc,
		method: fmt.Sprintf("/%s/%s", serviceName, method),
		enc:    enc,
		dec:    dec,
		// We are using reflect.Indirect here to allow both reply structs and
		// pointers to these reply structs. New consumers of the client should
		// use structs directly, while existing consumers will not break if they
		// remain to use pointers to structs.
		grpcReply: reflect.TypeOf(
			reflect.Indirect(
				reflect.ValueOf(grpcReply),
			).Interface(),
		),
		before: []RequestFunc{},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// ClientOption sets an optional parameter for clients.
type ClientOption func(*Client)

// ClientBefore sets the RequestFuncs that are applied to the outgoing gRPC
// request before it's invoked.
func ClientBefore(before ...RequestFunc) ClientOption {
	return func(c *Client) { c.before = before }
}

// Endpoint returns a usable endpoint that will invoke the gRPC specified by the
// client.
func (c Client) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		req, err := c.enc(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("Encode: %v", err)
		}

		md := &metadata.MD{}
		for _, f := range c.before {
			ctx = f(ctx, md)
		}
		ctx = metadata.NewContext(ctx, *md)

		grpcReply := reflect.New(c.grpcReply).Interface()
		if err = grpc.Invoke(ctx, c.method, req, grpcReply, c.client); err != nil {
			return nil, fmt.Errorf("Invoke: %v", err)
		}

		response, err := c.dec(ctx, grpcReply)
		if err != nil {
			return nil, fmt.Errorf("Decode: %v", err)
		}
		return response, nil
	}
}
