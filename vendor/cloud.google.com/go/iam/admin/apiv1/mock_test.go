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

package admin

import (
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
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

type mockIamServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	adminpb.IAMServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockIamServer) ListServiceAccounts(_ context.Context, req *adminpb.ListServiceAccountsRequest) (*adminpb.ListServiceAccountsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ListServiceAccountsResponse), nil
}

func (s *mockIamServer) GetServiceAccount(_ context.Context, req *adminpb.GetServiceAccountRequest) (*adminpb.ServiceAccount, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ServiceAccount), nil
}

func (s *mockIamServer) CreateServiceAccount(_ context.Context, req *adminpb.CreateServiceAccountRequest) (*adminpb.ServiceAccount, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ServiceAccount), nil
}

func (s *mockIamServer) UpdateServiceAccount(_ context.Context, req *adminpb.ServiceAccount) (*adminpb.ServiceAccount, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ServiceAccount), nil
}

func (s *mockIamServer) DeleteServiceAccount(_ context.Context, req *adminpb.DeleteServiceAccountRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockIamServer) ListServiceAccountKeys(_ context.Context, req *adminpb.ListServiceAccountKeysRequest) (*adminpb.ListServiceAccountKeysResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ListServiceAccountKeysResponse), nil
}

func (s *mockIamServer) GetServiceAccountKey(_ context.Context, req *adminpb.GetServiceAccountKeyRequest) (*adminpb.ServiceAccountKey, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ServiceAccountKey), nil
}

func (s *mockIamServer) CreateServiceAccountKey(_ context.Context, req *adminpb.CreateServiceAccountKeyRequest) (*adminpb.ServiceAccountKey, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.ServiceAccountKey), nil
}

func (s *mockIamServer) DeleteServiceAccountKey(_ context.Context, req *adminpb.DeleteServiceAccountKeyRequest) (*google_protobuf.Empty, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*google_protobuf.Empty), nil
}

func (s *mockIamServer) SignBlob(_ context.Context, req *adminpb.SignBlobRequest) (*adminpb.SignBlobResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.SignBlobResponse), nil
}

func (s *mockIamServer) GetIamPolicy(_ context.Context, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockIamServer) SetIamPolicy(_ context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.Policy), nil
}

func (s *mockIamServer) TestIamPermissions(_ context.Context, req *iampb.TestIamPermissionsRequest) (*iampb.TestIamPermissionsResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*iampb.TestIamPermissionsResponse), nil
}

func (s *mockIamServer) QueryGrantableRoles(_ context.Context, req *adminpb.QueryGrantableRolesRequest) (*adminpb.QueryGrantableRolesResponse, error) {
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	return s.resps[0].(*adminpb.QueryGrantableRolesResponse), nil
}

// clientOpt is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOpt option.ClientOption

var (
	mockIam mockIamServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	adminpb.RegisterIAMServer(serv, &mockIam)

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

func TestIamListServiceAccounts(t *testing.T) {
	var nextPageToken string = ""
	var accountsElement *adminpb.ServiceAccount = &adminpb.ServiceAccount{}
	var accounts = []*adminpb.ServiceAccount{accountsElement}
	var expectedResponse = &adminpb.ListServiceAccountsResponse{
		NextPageToken: nextPageToken,
		Accounts:      accounts,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamProjectPath("[PROJECT]")
	var request = &adminpb.ListServiceAccountsRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListServiceAccounts(context.Background(), request).Next()

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	want := (interface{})(expectedResponse.Accounts[0])
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

func TestIamListServiceAccountsError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamProjectPath("[PROJECT]")
	var request = &adminpb.ListServiceAccountsRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListServiceAccounts(context.Background(), request).Next()

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamGetServiceAccount(t *testing.T) {
	var name2 string = "name2-1052831874"
	var projectId string = "projectId-1969970175"
	var uniqueId string = "uniqueId-538310583"
	var email string = "email96619420"
	var displayName string = "displayName1615086568"
	var etag []byte = []byte("21")
	var oauth2ClientId string = "oauth2ClientId-1833466037"
	var expectedResponse = &adminpb.ServiceAccount{
		Name:           name2,
		ProjectId:      projectId,
		UniqueId:       uniqueId,
		Email:          email,
		DisplayName:    displayName,
		Etag:           etag,
		Oauth2ClientId: oauth2ClientId,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.GetServiceAccountRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetServiceAccount(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamGetServiceAccountError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.GetServiceAccountRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetServiceAccount(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamCreateServiceAccount(t *testing.T) {
	var name2 string = "name2-1052831874"
	var projectId string = "projectId-1969970175"
	var uniqueId string = "uniqueId-538310583"
	var email string = "email96619420"
	var displayName string = "displayName1615086568"
	var etag []byte = []byte("21")
	var oauth2ClientId string = "oauth2ClientId-1833466037"
	var expectedResponse = &adminpb.ServiceAccount{
		Name:           name2,
		ProjectId:      projectId,
		UniqueId:       uniqueId,
		Email:          email,
		DisplayName:    displayName,
		Etag:           etag,
		Oauth2ClientId: oauth2ClientId,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamProjectPath("[PROJECT]")
	var accountId string = "accountId-803333011"
	var request = &adminpb.CreateServiceAccountRequest{
		Name:      formattedName,
		AccountId: accountId,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateServiceAccount(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamCreateServiceAccountError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamProjectPath("[PROJECT]")
	var accountId string = "accountId-803333011"
	var request = &adminpb.CreateServiceAccountRequest{
		Name:      formattedName,
		AccountId: accountId,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateServiceAccount(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamUpdateServiceAccount(t *testing.T) {
	var name string = "name3373707"
	var projectId string = "projectId-1969970175"
	var uniqueId string = "uniqueId-538310583"
	var email string = "email96619420"
	var displayName string = "displayName1615086568"
	var etag2 []byte = []byte("-120")
	var oauth2ClientId string = "oauth2ClientId-1833466037"
	var expectedResponse = &adminpb.ServiceAccount{
		Name:           name,
		ProjectId:      projectId,
		UniqueId:       uniqueId,
		Email:          email,
		DisplayName:    displayName,
		Etag:           etag2,
		Oauth2ClientId: oauth2ClientId,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var etag []byte = []byte("21")
	var request = &adminpb.ServiceAccount{
		Etag: etag,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.UpdateServiceAccount(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamUpdateServiceAccountError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var etag []byte = []byte("21")
	var request = &adminpb.ServiceAccount{
		Etag: etag,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.UpdateServiceAccount(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamDeleteServiceAccount(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.DeleteServiceAccountRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteServiceAccount(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestIamDeleteServiceAccountError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.DeleteServiceAccountRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteServiceAccount(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestIamListServiceAccountKeys(t *testing.T) {
	var expectedResponse *adminpb.ListServiceAccountKeysResponse = &adminpb.ListServiceAccountKeysResponse{}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.ListServiceAccountKeysRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListServiceAccountKeys(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamListServiceAccountKeysError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.ListServiceAccountKeysRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ListServiceAccountKeys(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamGetServiceAccountKey(t *testing.T) {
	var name2 string = "name2-1052831874"
	var privateKeyData []byte = []byte("-58")
	var publicKeyData []byte = []byte("-96")
	var expectedResponse = &adminpb.ServiceAccountKey{
		Name:           name2,
		PrivateKeyData: privateKeyData,
		PublicKeyData:  publicKeyData,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamKeyPath("[PROJECT]", "[SERVICE_ACCOUNT]", "[KEY]")
	var request = &adminpb.GetServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetServiceAccountKey(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamGetServiceAccountKeyError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamKeyPath("[PROJECT]", "[SERVICE_ACCOUNT]", "[KEY]")
	var request = &adminpb.GetServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.GetServiceAccountKey(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamCreateServiceAccountKey(t *testing.T) {
	var name2 string = "name2-1052831874"
	var privateKeyData []byte = []byte("-58")
	var publicKeyData []byte = []byte("-96")
	var expectedResponse = &adminpb.ServiceAccountKey{
		Name:           name2,
		PrivateKeyData: privateKeyData,
		PublicKeyData:  publicKeyData,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.CreateServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateServiceAccountKey(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamCreateServiceAccountKeyError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &adminpb.CreateServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.CreateServiceAccountKey(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamDeleteServiceAccountKey(t *testing.T) {
	var expectedResponse *google_protobuf.Empty = &google_protobuf.Empty{}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamKeyPath("[PROJECT]", "[SERVICE_ACCOUNT]", "[KEY]")
	var request = &adminpb.DeleteServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteServiceAccountKey(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

}

func TestIamDeleteServiceAccountKeyError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamKeyPath("[PROJECT]", "[SERVICE_ACCOUNT]", "[KEY]")
	var request = &adminpb.DeleteServiceAccountKeyRequest{
		Name: formattedName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	err = c.DeleteServiceAccountKey(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
}
func TestIamSignBlob(t *testing.T) {
	var keyId string = "keyId-1134673157"
	var signature []byte = []byte("-72")
	var expectedResponse = &adminpb.SignBlobResponse{
		KeyId:     keyId,
		Signature: signature,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var bytesToSign []byte = []byte("45")
	var request = &adminpb.SignBlobRequest{
		Name:        formattedName,
		BytesToSign: bytesToSign,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.SignBlob(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamSignBlobError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedName string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var bytesToSign []byte = []byte("45")
	var request = &adminpb.SignBlobRequest{
		Name:        formattedName,
		BytesToSign: bytesToSign,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.SignBlob(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamGetIamPolicy(t *testing.T) {
	var version int32 = 351608024
	var etag []byte = []byte("21")
	var expectedResponse = &iampb.Policy{
		Version: version,
		Etag:    etag,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &iampb.GetIamPolicyRequest{
		Resource: formattedResource,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.getIamPolicy(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamGetIamPolicyError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var request = &iampb.GetIamPolicyRequest{
		Resource: formattedResource,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.getIamPolicy(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamSetIamPolicy(t *testing.T) {
	var version int32 = 351608024
	var etag []byte = []byte("21")
	var expectedResponse = &iampb.Policy{
		Version: version,
		Etag:    etag,
	}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var policy *iampb.Policy = &iampb.Policy{}
	var request = &iampb.SetIamPolicyRequest{
		Resource: formattedResource,
		Policy:   policy,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.setIamPolicy(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamSetIamPolicyError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var policy *iampb.Policy = &iampb.Policy{}
	var request = &iampb.SetIamPolicyRequest{
		Resource: formattedResource,
		Policy:   policy,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.setIamPolicy(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamTestIamPermissions(t *testing.T) {
	var expectedResponse *iampb.TestIamPermissionsResponse = &iampb.TestIamPermissionsResponse{}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var permissions []string = nil
	var request = &iampb.TestIamPermissionsRequest{
		Resource:    formattedResource,
		Permissions: permissions,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.TestIamPermissions(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamTestIamPermissionsError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var formattedResource string = IamServiceAccountPath("[PROJECT]", "[SERVICE_ACCOUNT]")
	var permissions []string = nil
	var request = &iampb.TestIamPermissionsRequest{
		Resource:    formattedResource,
		Permissions: permissions,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.TestIamPermissions(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
func TestIamQueryGrantableRoles(t *testing.T) {
	var expectedResponse *adminpb.QueryGrantableRolesResponse = &adminpb.QueryGrantableRolesResponse{}

	mockIam.err = nil
	mockIam.reqs = nil

	mockIam.resps = append(mockIam.resps[:0], expectedResponse)

	var fullResourceName string = "fullResourceName1300993644"
	var request = &adminpb.QueryGrantableRolesRequest{
		FullResourceName: fullResourceName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.QueryGrantableRoles(context.Background(), request)

	if err != nil {
		t.Fatal(err)
	}

	if want, got := request, mockIam.reqs[0]; !proto.Equal(want, got) {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if want, got := expectedResponse, resp; !proto.Equal(want, got) {
		t.Errorf("wrong response %q, want %q)", got, want)
	}
}

func TestIamQueryGrantableRolesError(t *testing.T) {
	errCode := codes.Internal
	mockIam.err = grpc.Errorf(errCode, "test error")

	var fullResourceName string = "fullResourceName1300993644"
	var request = &adminpb.QueryGrantableRolesRequest{
		FullResourceName: fullResourceName,
	}

	c, err := NewIamClient(context.Background(), clientOpt)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.QueryGrantableRoles(context.Background(), request)

	if c := grpc.Code(err); c != errCode {
		t.Errorf("got error code %q, want %q", c, errCode)
	}
	_ = resp
}
