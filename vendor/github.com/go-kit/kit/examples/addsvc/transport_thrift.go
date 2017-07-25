package addsvc

// This file provides server-side bindings for the Thrift transport.
//
// This file also provides endpoint constructors that utilize a Thrift client,
// for use in client packages, because package transport/thrift doesn't exist
// yet. See https://github.com/go-kit/kit/issues/184.

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	thriftadd "github.com/go-kit/kit/examples/addsvc/thrift/gen-go/addsvc"
)

// MakeThriftHandler makes a set of endpoints available as a Thrift service.
func MakeThriftHandler(ctx context.Context, e Endpoints) thriftadd.AddService {
	return &thriftServer{
		ctx:    ctx,
		sum:    e.SumEndpoint,
		concat: e.ConcatEndpoint,
	}
}

type thriftServer struct {
	ctx    context.Context
	sum    endpoint.Endpoint
	concat endpoint.Endpoint
}

func (s *thriftServer) Sum(a int64, b int64) (*thriftadd.SumReply, error) {
	request := sumRequest{A: int(a), B: int(b)}
	response, err := s.sum(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(sumResponse)
	return &thriftadd.SumReply{Value: int64(resp.V), Err: err2str(resp.Err)}, nil
}

func (s *thriftServer) Concat(a string, b string) (*thriftadd.ConcatReply, error) {
	request := concatRequest{A: a, B: b}
	response, err := s.concat(s.ctx, request)
	if err != nil {
		return nil, err
	}
	resp := response.(concatResponse)
	return &thriftadd.ConcatReply{Value: resp.V, Err: err2str(resp.Err)}, nil
}

// MakeThriftSumEndpoint returns an endpoint that invokes the passed Thrift client.
// Useful only in clients, and only until a proper transport/thrift.Client exists.
func MakeThriftSumEndpoint(client *thriftadd.AddServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(sumRequest)
		reply, err := client.Sum(int64(req.A), int64(req.B))
		if err == ErrIntOverflow {
			return nil, err // special case; see comment on ErrIntOverflow
		}
		return sumResponse{V: int(reply.Value), Err: err}, nil
	}
}

// MakeThriftConcatEndpoint returns an endpoint that invokes the passed Thrift
// client. Useful only in clients, and only until a proper
// transport/thrift.Client exists.
func MakeThriftConcatEndpoint(client *thriftadd.AddServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(concatRequest)
		reply, err := client.Concat(req.A, req.B)
		return concatResponse{V: reply.Value, Err: err}, nil
	}
}
