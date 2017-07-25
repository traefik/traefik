// Copyright 2017, Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTO-GENERATED CODE. DO NOT EDIT.

package trace

import (
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	cloudtracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v1"
)

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var _ = io.EOF
var _ = ptypes.MarshalAny
var _ status.Status

type mockTraceServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	cloudtracepb.TraceServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockTraceServer) ListTraces(_ context.Context, req *cloudtracepb.ListTracesRequest) (*cloudtracepb.ListTracesResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*cloudtracepb.ListTracesResponse), nil
}

func (s *mockTraceServer) GetTrace(_ context.Context, req *cloudtracepb.GetTraceRequest) (*cloudtracepb.Trace, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*cloudtracepb.Trace), nil
}

func (s *mockTraceServer) PatchTraces(_ context.Context, req *cloudtracepb.PatchTracesRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockTrace mockTraceServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	cloudtracepb.RegisterTraceServiceServer(serv, &mockTrace)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	go serv.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	clientOpt = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}

func TestTraceServicePatchTraces(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockTrace.err = nil
	mockTrace.reqs = nil

	mockTrace.resps = append(mockTrace.resps[:0], expectedResponse)

	var projectId string = "projectId-1969970175"
	var traces *cloudtracepb.Traces = &cloudtracepb.Traces{}
	var request = &cloudtracepb.PatchTracesRequest{
		ProjectId: projectId,
		Traces:    traces,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.PatchTraces(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockTrace.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestTraceServicePatchTracesError(t *testing.T) {
	errCode := codes.Internal
	mockTrace.err = grpc.Errorf(errCode, "test error")

	var projectId string = "projectId-1969970175"
	var traces *cloudtracepb.Traces = &cloudtracepb.Traces{}
	var request = &cloudtracepb.PatchTracesRequest{
		ProjectId: projectId,
		Traces:    traces,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.PatchTraces(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestTraceServiceGetTrace(t *testing.T) {
	var projectId2 string = "projectId2939242356"
	var traceId2 string = "traceId2987826376"
	var expectedResponse = &cloudtracepb.Trace{
		ProjectId: projectId2,
		TraceId:   traceId2,
	}

	mockTrace.err = nil
	mockTrace.reqs = nil

	mockTrace.resps = append(mockTrace.resps[:0], expectedResponse)

	var projectId string = "projectId-1969970175"
	var traceId string = "traceId1270300245"
	var request = &cloudtracepb.GetTraceRequest{
		ProjectId: projectId,
		TraceId:   traceId,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetTrace(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockTrace.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestTraceServiceGetTraceError(t *testing.T) {
	errCode := codes.Internal
	mockTrace.err = grpc.Errorf(errCode, "test error")

	var projectId string = "projectId-1969970175"
	var traceId string = "traceId1270300245"
	var request = &cloudtracepb.GetTraceRequest{
		ProjectId: projectId,
		TraceId:   traceId,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetTrace(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestTraceServiceListTraces(t *testing.T) {
	var nextPageToken string = ""
	var tracesElement *cloudtracepb.Trace = &cloudtracepb.Trace{}
	var traces = []*cloudtracepb.Trace{tracesElement}
	var expectedResponse = &cloudtracepb.ListTracesResponse{
		NextPageToken: nextPageToken,
		Traces:        traces,
	}

	mockTrace.err = nil
	mockTrace.reqs = nil

	mockTrace.resps = append(mockTrace.resps[:0], expectedResponse)

	var projectId string = "projectId-1969970175"
	var request = &cloudtracepb.ListTracesRequest{
		ProjectId: projectId,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTraces(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockTrace.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Traces[0])
	got := (interface{})(resp)
	var ok bool

	switch want := (want).(type) {
	case proto.Message:
		ok = proto.Equal(want, got.(proto.Message))
	default:
		ok = want == got
	}
	if !ok {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestTraceServiceListTracesError(t *testing.T) {
	errCode := codes.Internal
	mockTrace.err = grpc.Errorf(errCode, "test error")

	var projectId string = "projectId-1969970175"
	var request = &cloudtracepb.ListTracesRequest{
		ProjectId: projectId,
	}

	c, err := NewClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTraces(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
