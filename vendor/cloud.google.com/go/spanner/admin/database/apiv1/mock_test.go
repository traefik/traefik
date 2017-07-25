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

package database

import (
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	longrunningpb "google.golang.org/genproto/googleapis/longrunning"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
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

type mockDatabaseAdminServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	databasepb.DatabaseAdminServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockDatabaseAdminServer) ListDatabases(_ context.Context, req *databasepb.ListDatabasesRequest) (*databasepb.ListDatabasesResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*databasepb.ListDatabasesResponse), nil
}

func (s *mockDatabaseAdminServer) CreateDatabase(_ context.Context, req *databasepb.CreateDatabaseRequest) (*longrunningpb.Operation, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*longrunningpb.Operation), nil
}

func (s *mockDatabaseAdminServer) GetDatabase(_ context.Context, req *databasepb.GetDatabaseRequest) (*databasepb.Database, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*databasepb.Database), nil
}

func (s *mockDatabaseAdminServer) UpdateDatabaseDdl(_ context.Context, req *databasepb.UpdateDatabaseDdlRequest) (*longrunningpb.Operation, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*longrunningpb.Operation), nil
}

func (s *mockDatabaseAdminServer) DropDatabase(_ context.Context, req *databasepb.DropDatabaseRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockDatabaseAdminServer) GetDatabaseDdl(_ context.Context, req *databasepb.GetDatabaseDdlRequest) (*databasepb.GetDatabaseDdlResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*databasepb.GetDatabaseDdlResponse), nil
}

func (s *mockDatabaseAdminServer) SetIamPolicy(_ context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockDatabaseAdminServer) GetIamPolicy(_ context.Context, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockDatabaseAdminServer) TestIamPermissions(_ context.Context, req *iampb.TestIamPermissionsRequest) (*iampb.TestIamPermissionsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.TestIamPermissionsResponse), nil
}

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockDatabaseAdmin mockDatabaseAdminServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	databasepb.RegisterDatabaseAdminServer(serv, &mockDatabaseAdmin)

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

func TestDatabaseAdminListDatabases(t *testing.T) {
	var nextPageToken string = ""
	var databasesElement *databasepb.Database = &databasepb.Database{}
	var databases = []*databasepb.Database{databasesElement}
	var expectedResponse = &databasepb.ListDatabasesResponse{
		NextPageToken: nextPageToken,
		Databases:     databases,
	}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedParent string = DatabaseAdminInstancePath("[PROJECT]", "[INSTANCE]")
	var request = &databasepb.ListDatabasesRequest{
		Parent: formattedParent,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListDatabases(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Databases[0])
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

func TestDatabaseAdminListDatabasesError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedParent string = DatabaseAdminInstancePath("[PROJECT]", "[INSTANCE]")
	var request = &databasepb.ListDatabasesRequest{
		Parent: formattedParent,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListDatabases(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminCreateDatabase(t *testing.T) {
	var name string = "name3373707"
	var expectedResponse = &databasepb.Database{
		Name: name,
	}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	any, err := ptypes.MarshalAny(expectedResponse)
	if err != nil {
		t.Fatal(err)
	}
	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], &longrunningpb.Operation{
		Name:   "longrunning-test",
		Done:   true,
		Result: &longrunningpb.Operation_Response{Response: any},
	})

	var formattedParent string = DatabaseAdminInstancePath("[PROJECT]", "[INSTANCE]")
	var createStatement string = "createStatement552974828"
	var request = &databasepb.CreateDatabaseRequest{
		Parent:          formattedParent,
		CreateStatement: createStatement,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	respLRO, err := c.CreateDatabase(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := respLRO.Wait(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminCreateDatabaseError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], &longrunningpb.Operation{
		Name: "longrunning-test",
		Done: true,
		Result: &longrunningpb.Operation_Error{
			Error: &status.Status{
				Code:    int32(errCode),
				Message: "test error",
			},
		},
	})

	var formattedParent string = DatabaseAdminInstancePath("[PROJECT]", "[INSTANCE]")
	var createStatement string = "createStatement552974828"
	var request = &databasepb.CreateDatabaseRequest{
		Parent:          formattedParent,
		CreateStatement: createStatement,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	respLRO, err := c.CreateDatabase(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := respLRO.Wait(context.Background())

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminGetDatabase(t *testing.T) {
	var name2 string = "name2-1052831874"
	var expectedResponse = &databasepb.Database{
		Name: name2,
	}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedName string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.GetDatabaseRequest{
		Name: formattedName,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetDatabase(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminGetDatabaseError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedName string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.GetDatabaseRequest{
		Name: formattedName,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetDatabase(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminUpdateDatabaseDdl(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	any, err := ptypes.MarshalAny(expectedResponse)
	if err != nil {
		t.Fatal(err)
	}
	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], &longrunningpb.Operation{
		Name:   "longrunning-test",
		Done:   true,
		Result: &longrunningpb.Operation_Response{Response: any},
	})

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var statements []string = nil
	var request = &databasepb.UpdateDatabaseDdlRequest{
		Database:   formattedDatabase,
		Statements: statements,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	respLRO, err := c.UpdateDatabaseDdl(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	err = respLRO.Wait(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestDatabaseAdminUpdateDatabaseDdlError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], &longrunningpb.Operation{
		Name: "longrunning-test",
		Done: true,
		Result: &longrunningpb.Operation_Error{
			Error: &status.Status{
				Code:    int32(errCode),
				Message: "test error",
			},
		},
	})

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var statements []string = nil
	var request = &databasepb.UpdateDatabaseDdlRequest{
		Database:   formattedDatabase,
		Statements: statements,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	respLRO, err := c.UpdateDatabaseDdl(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	err = respLRO.Wait(context.Background())

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestDatabaseAdminDropDatabase(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.DropDatabaseRequest{
		Database: formattedDatabase,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DropDatabase(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestDatabaseAdminDropDatabaseError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.DropDatabaseRequest{
		Database: formattedDatabase,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DropDatabase(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestDatabaseAdminGetDatabaseDdl(t *testing.T) {
	var expectedResponse *databasepb.GetDatabaseDdlResponse = &databasepb.GetDatabaseDdlResponse{}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.GetDatabaseDdlRequest{
		Database: formattedDatabase,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetDatabaseDdl(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminGetDatabaseDdlError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedDatabase string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &databasepb.GetDatabaseDdlRequest{
		Database: formattedDatabase,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetDatabaseDdl(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminSetIamPolicy(t *testing.T) {
	var version int32 = 351608024
	var etag []byte = []byte("21")
	var expectedResponse = &iampb.Policy{
		Version: version,
		Etag:    etag,
	}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var policy *iampb.Policy = &iampb.Policy{}
	var request = &iampb.SetIamPolicyRequest{
		Resource: formattedResource,
		Policy:   policy,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.SetIamPolicy(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminSetIamPolicyError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var policy *iampb.Policy = &iampb.Policy{}
	var request = &iampb.SetIamPolicyRequest{
		Resource: formattedResource,
		Policy:   policy,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.SetIamPolicy(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminGetIamPolicy(t *testing.T) {
	var version int32 = 351608024
	var etag []byte = []byte("21")
	var expectedResponse = &iampb.Policy{
		Version: version,
		Etag:    etag,
	}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &iampb.GetIamPolicyRequest{
		Resource: formattedResource,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetIamPolicy(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminGetIamPolicyError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var request = &iampb.GetIamPolicyRequest{
		Resource: formattedResource,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetIamPolicy(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestDatabaseAdminTestIamPermissions(t *testing.T) {
	var expectedResponse *iampb.TestIamPermissionsResponse = &iampb.TestIamPermissionsResponse{}

	mockDatabaseAdmin.err = nil
	mockDatabaseAdmin.reqs = nil

	mockDatabaseAdmin.resps = append(mockDatabaseAdmin.resps[:0], expectedResponse)

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var permissions []string = nil
	var request = &iampb.TestIamPermissionsRequest{
		Resource:    formattedResource,
		Permissions: permissions,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.TestIamPermissions(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockDatabaseAdmin.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestDatabaseAdminTestIamPermissionsError(t *testing.T) {
	errCode := codes.Internal
	mockDatabaseAdmin.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = DatabaseAdminDatabasePath("[PROJECT]", "[INSTANCE]", "[DATABASE]")
	var permissions []string = nil
	var request = &iampb.TestIamPermissionsRequest{
		Resource:    formattedResource,
		Permissions: permissions,
	}

	c, err := NewDatabaseAdminClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.TestIamPermissions(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
