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

package monitoring

import (
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
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

type mockGroupServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	monitoringpb.GroupServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockGroupServer) ListGroups(_ context.Context, req *monitoringpb.ListGroupsRequest) (*monitoringpb.ListGroupsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.ListGroupsResponse), nil
}

func (s *mockGroupServer) GetGroup(_ context.Context, req *monitoringpb.GetGroupRequest) (*monitoringpb.Group, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.Group), nil
}

func (s *mockGroupServer) CreateGroup(_ context.Context, req *monitoringpb.CreateGroupRequest) (*monitoringpb.Group, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.Group), nil
}

func (s *mockGroupServer) UpdateGroup(_ context.Context, req *monitoringpb.UpdateGroupRequest) (*monitoringpb.Group, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.Group), nil
}

func (s *mockGroupServer) DeleteGroup(_ context.Context, req *monitoringpb.DeleteGroupRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockGroupServer) ListGroupMembers(_ context.Context, req *monitoringpb.ListGroupMembersRequest) (*monitoringpb.ListGroupMembersResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.ListGroupMembersResponse), nil
}

type mockMetricServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	monitoringpb.MetricServiceServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockMetricServer) ListMonitoredResourceDescriptors(_ context.Context, req *monitoringpb.ListMonitoredResourceDescriptorsRequest) (*monitoringpb.ListMonitoredResourceDescriptorsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.ListMonitoredResourceDescriptorsResponse), nil
}

func (s *mockMetricServer) GetMonitoredResourceDescriptor(_ context.Context, req *monitoringpb.GetMonitoredResourceDescriptorRequest) (*monitoredrespb.MonitoredResourceDescriptor, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoredrespb.MonitoredResourceDescriptor), nil
}

func (s *mockMetricServer) ListMetricDescriptors(_ context.Context, req *monitoringpb.ListMetricDescriptorsRequest) (*monitoringpb.ListMetricDescriptorsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.ListMetricDescriptorsResponse), nil
}

func (s *mockMetricServer) GetMetricDescriptor(_ context.Context, req *monitoringpb.GetMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*metricpb.MetricDescriptor), nil
}

func (s *mockMetricServer) CreateMetricDescriptor(_ context.Context, req *monitoringpb.CreateMetricDescriptorRequest) (*metricpb.MetricDescriptor, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*metricpb.MetricDescriptor), nil
}

func (s *mockMetricServer) DeleteMetricDescriptor(_ context.Context, req *monitoringpb.DeleteMetricDescriptorRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockMetricServer) ListTimeSeries(_ context.Context, req *monitoringpb.ListTimeSeriesRequest) (*monitoringpb.ListTimeSeriesResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*monitoringpb.ListTimeSeriesResponse), nil
}

func (s *mockMetricServer) CreateTimeSeries(_ context.Context, req *monitoringpb.CreateTimeSeriesRequest) (*google_protobuf.Empty, error) {
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
	mockGroup  mockGroupServer
	mockMetric mockMetricServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	monitoringpb.RegisterGroupServiceServer(serv, &mockGroup)
	monitoringpb.RegisterMetricServiceServer(serv, &mockMetric)

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

func TestGroupServiceListGroups(t *testing.T) {
	var nextPageToken string = ""
	var groupElement *monitoringpb.Group = &monitoringpb.Group{}
	var group = []*monitoringpb.Group{groupElement}
	var expectedResponse = &monitoringpb.ListGroupsResponse{
		NextPageToken: nextPageToken,
		Group:         group,
	}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var formattedName string = GroupProjectPath("[PROJECT]")
	var request = &monitoringpb.ListGroupsRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListGroups(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Group[0])
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

func TestGroupServiceListGroupsError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var formattedName string = GroupProjectPath("[PROJECT]")
	var request = &monitoringpb.ListGroupsRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListGroups(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestGroupServiceGetGroup(t *testing.T) {
	var name2 string = "name2-1052831874"
	var displayName string = "displayName1615086568"
	var parentName string = "parentName1015022848"
	var filter string = "filter-1274492040"
	var isCluster bool = false
	var expectedResponse = &monitoringpb.Group{
		Name:        name2,
		DisplayName: displayName,
		ParentName:  parentName,
		Filter:      filter,
		IsCluster:   isCluster,
	}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.GetGroupRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetGroup(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestGroupServiceGetGroupError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.GetGroupRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetGroup(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestGroupServiceCreateGroup(t *testing.T) {
	var name2 string = "name2-1052831874"
	var displayName string = "displayName1615086568"
	var parentName string = "parentName1015022848"
	var filter string = "filter-1274492040"
	var isCluster bool = false
	var expectedResponse = &monitoringpb.Group{
		Name:        name2,
		DisplayName: displayName,
		ParentName:  parentName,
		Filter:      filter,
		IsCluster:   isCluster,
	}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var formattedName string = GroupProjectPath("[PROJECT]")
	var group *monitoringpb.Group = &monitoringpb.Group{}
	var request = &monitoringpb.CreateGroupRequest{
		Name:  formattedName,
		Group: group,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateGroup(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestGroupServiceCreateGroupError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var formattedName string = GroupProjectPath("[PROJECT]")
	var group *monitoringpb.Group = &monitoringpb.Group{}
	var request = &monitoringpb.CreateGroupRequest{
		Name:  formattedName,
		Group: group,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateGroup(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestGroupServiceUpdateGroup(t *testing.T) {
	var name string = "name3373707"
	var displayName string = "displayName1615086568"
	var parentName string = "parentName1015022848"
	var filter string = "filter-1274492040"
	var isCluster bool = false
	var expectedResponse = &monitoringpb.Group{
		Name:        name,
		DisplayName: displayName,
		ParentName:  parentName,
		Filter:      filter,
		IsCluster:   isCluster,
	}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var group *monitoringpb.Group = &monitoringpb.Group{}
	var request = &monitoringpb.UpdateGroupRequest{
		Group: group,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.UpdateGroup(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestGroupServiceUpdateGroupError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var group *monitoringpb.Group = &monitoringpb.Group{}
	var request = &monitoringpb.UpdateGroupRequest{
		Group: group,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.UpdateGroup(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestGroupServiceDeleteGroup(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.DeleteGroupRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteGroup(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestGroupServiceDeleteGroupError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.DeleteGroupRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteGroup(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestGroupServiceListGroupMembers(t *testing.T) {
	var nextPageToken string = ""
	var totalSize int32 = -705419236
	var membersElement *monitoredrespb.MonitoredResource = &monitoredrespb.MonitoredResource{}
	var members = []*monitoredrespb.MonitoredResource{membersElement}
	var expectedResponse = &monitoringpb.ListGroupMembersResponse{
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
		Members:       members,
	}

	mockGroup.err = nil
	mockGroup.reqs = nil

	mockGroup.resps = append(mockGroup.resps[:0], expectedResponse)

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.ListGroupMembersRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListGroupMembers(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockGroup.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Members[0])
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

func TestGroupServiceListGroupMembersError(t *testing.T) {
	errCode := codes.Internal
	mockGroup.err = grpc.Errorf(errCode, "test error")

	var formattedName string = GroupGroupPath("[PROJECT]", "[GROUP]")
	var request = &monitoringpb.ListGroupMembersRequest{
		Name: formattedName,
	}

	c, err := NewGroupClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListGroupMembers(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceListMonitoredResourceDescriptors(t *testing.T) {
	var nextPageToken string = ""
	var resourceDescriptorsElement *monitoredrespb.MonitoredResourceDescriptor = &monitoredrespb.MonitoredResourceDescriptor{}
	var resourceDescriptors = []*monitoredrespb.MonitoredResourceDescriptor{resourceDescriptorsElement}
	var expectedResponse = &monitoringpb.ListMonitoredResourceDescriptorsResponse{
		NextPageToken:       nextPageToken,
		ResourceDescriptors: resourceDescriptors,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricProjectPath("[PROJECT]")
	var request = &monitoringpb.ListMonitoredResourceDescriptorsRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListMonitoredResourceDescriptors(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.ResourceDescriptors[0])
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

func TestMetricServiceListMonitoredResourceDescriptorsError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricProjectPath("[PROJECT]")
	var request = &monitoringpb.ListMonitoredResourceDescriptorsRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListMonitoredResourceDescriptors(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceGetMonitoredResourceDescriptor(t *testing.T) {
	var name2 string = "name2-1052831874"
	var type_ string = "type3575610"
	var displayName string = "displayName1615086568"
	var description string = "description-1724546052"
	var expectedResponse = &monitoredrespb.MonitoredResourceDescriptor{
		Name:        name2,
		Type:        type_,
		DisplayName: displayName,
		Description: description,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricMonitoredResourceDescriptorPath("[PROJECT]", "[MONITORED_RESOURCE_DESCRIPTOR]")
	var request = &monitoringpb.GetMonitoredResourceDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMonitoredResourceDescriptor(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestMetricServiceGetMonitoredResourceDescriptorError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricMonitoredResourceDescriptorPath("[PROJECT]", "[MONITORED_RESOURCE_DESCRIPTOR]")
	var request = &monitoringpb.GetMonitoredResourceDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMonitoredResourceDescriptor(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceListMetricDescriptors(t *testing.T) {
	var nextPageToken string = ""
	var metricDescriptorsElement *metricpb.MetricDescriptor = &metricpb.MetricDescriptor{}
	var metricDescriptors = []*metricpb.MetricDescriptor{metricDescriptorsElement}
	var expectedResponse = &monitoringpb.ListMetricDescriptorsResponse{
		NextPageToken:     nextPageToken,
		MetricDescriptors: metricDescriptors,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricProjectPath("[PROJECT]")
	var request = &monitoringpb.ListMetricDescriptorsRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListMetricDescriptors(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.MetricDescriptors[0])
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

func TestMetricServiceListMetricDescriptorsError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricProjectPath("[PROJECT]")
	var request = &monitoringpb.ListMetricDescriptorsRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListMetricDescriptors(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceGetMetricDescriptor(t *testing.T) {
	var name2 string = "name2-1052831874"
	var type_ string = "type3575610"
	var unit string = "unit3594628"
	var description string = "description-1724546052"
	var displayName string = "displayName1615086568"
	var expectedResponse = &metricpb.MetricDescriptor{
		Name:        name2,
		Type:        type_,
		Unit:        unit,
		Description: description,
		DisplayName: displayName,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricMetricDescriptorPath("[PROJECT]", "[METRIC_DESCRIPTOR]")
	var request = &monitoringpb.GetMetricDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMetricDescriptor(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestMetricServiceGetMetricDescriptorError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricMetricDescriptorPath("[PROJECT]", "[METRIC_DESCRIPTOR]")
	var request = &monitoringpb.GetMetricDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetMetricDescriptor(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceCreateMetricDescriptor(t *testing.T) {
	var name2 string = "name2-1052831874"
	var type_ string = "type3575610"
	var unit string = "unit3594628"
	var description string = "description-1724546052"
	var displayName string = "displayName1615086568"
	var expectedResponse = &metricpb.MetricDescriptor{
		Name:        name2,
		Type:        type_,
		Unit:        unit,
		Description: description,
		DisplayName: displayName,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricProjectPath("[PROJECT]")
	var metricDescriptor *metricpb.MetricDescriptor = &metricpb.MetricDescriptor{}
	var request = &monitoringpb.CreateMetricDescriptorRequest{
		Name:             formattedName,
		MetricDescriptor: metricDescriptor,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateMetricDescriptor(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestMetricServiceCreateMetricDescriptorError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricProjectPath("[PROJECT]")
	var metricDescriptor *metricpb.MetricDescriptor = &metricpb.MetricDescriptor{}
	var request = &monitoringpb.CreateMetricDescriptorRequest{
		Name:             formattedName,
		MetricDescriptor: metricDescriptor,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateMetricDescriptor(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceDeleteMetricDescriptor(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricMetricDescriptorPath("[PROJECT]", "[METRIC_DESCRIPTOR]")
	var request = &monitoringpb.DeleteMetricDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteMetricDescriptor(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestMetricServiceDeleteMetricDescriptorError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricMetricDescriptorPath("[PROJECT]", "[METRIC_DESCRIPTOR]")
	var request = &monitoringpb.DeleteMetricDescriptorRequest{
		Name: formattedName,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteMetricDescriptor(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestMetricServiceListTimeSeries(t *testing.T) {
	var nextPageToken string = ""
	var timeSeriesElement *monitoringpb.TimeSeries = &monitoringpb.TimeSeries{}
	var timeSeries = []*monitoringpb.TimeSeries{timeSeriesElement}
	var expectedResponse = &monitoringpb.ListTimeSeriesResponse{
		NextPageToken: nextPageToken,
		TimeSeries:    timeSeries,
	}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricProjectPath("[PROJECT]")
	var filter string = "filter-1274492040"
	var interval *monitoringpb.TimeInterval = &monitoringpb.TimeInterval{}
	var view monitoringpb.ListTimeSeriesRequest_TimeSeriesView = monitoringpb.ListTimeSeriesRequest_FULL
	var request = &monitoringpb.ListTimeSeriesRequest{
		Name:     formattedName,
		Filter:   filter,
		Interval: interval,
		View:     view,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTimeSeries(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.TimeSeries[0])
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

func TestMetricServiceListTimeSeriesError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricProjectPath("[PROJECT]")
	var filter string = "filter-1274492040"
	var interval *monitoringpb.TimeInterval = &monitoringpb.TimeInterval{}
	var view monitoringpb.ListTimeSeriesRequest_TimeSeriesView = monitoringpb.ListTimeSeriesRequest_FULL
	var request = &monitoringpb.ListTimeSeriesRequest{
		Name:     formattedName,
		Filter:   filter,
		Interval: interval,
		View:     view,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListTimeSeries(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestMetricServiceCreateTimeSeries(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockMetric.err = nil
	mockMetric.reqs = nil

	mockMetric.resps = append(mockMetric.resps[:0], expectedResponse)

	var formattedName string = MetricProjectPath("[PROJECT]")
	var timeSeries []*monitoringpb.TimeSeries = nil
	var request = &monitoringpb.CreateTimeSeriesRequest{
		Name:       formattedName,
		TimeSeries: timeSeries,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.CreateTimeSeries(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockMetric.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestMetricServiceCreateTimeSeriesError(t *testing.T) {
	errCode := codes.Internal
	mockMetric.err = grpc.Errorf(errCode, "test error")

	var formattedName string = MetricProjectPath("[PROJECT]")
	var timeSeries []*monitoringpb.TimeSeries = nil
	var request = &monitoringpb.CreateTimeSeriesRequest{
		Name:       formattedName,
		TimeSeries: timeSeries,
	}

	c, err := NewMetricClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.CreateTimeSeries(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
