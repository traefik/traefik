/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/ptypes/empty"
	proto3 "github.com/golang/protobuf/ptypes/struct"
	pbt "github.com/golang/protobuf/ptypes/timestamp"

	sppb "google.golang.org/genproto/googleapis/spanner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Action is a mocked RPC activity that MockCloudSpannerClient will take.
type Action struct {
	method string
	err    error
}

// NewAction creates Action objects.
func NewAction(m string, e error) Action {
	return Action{m, e}
}

// MockCloudSpannerClient is a mock implementation of sppb.SpannerClient.
type MockCloudSpannerClient struct {
	mu sync.Mutex
	t  *testing.T
	// Live sessions on the client.
	sessions map[string]bool
	// Expected set of actions that will be executed by the client.
	actions []Action
	// Session ping history
	pings []string
	// Injected error, will be returned by all APIs
	injErr map[string]error
	// nice client will not fail on any request
	nice bool
}

// NewMockCloudSpannerClient creates new MockCloudSpannerClient instance.
func NewMockCloudSpannerClient(t *testing.T, acts ...Action) *MockCloudSpannerClient {
	mc := &MockCloudSpannerClient{t: t, sessions: map[string]bool{}, injErr: map[string]error{}}
	mc.SetActions(acts...)
	return mc
}

// MakeNice makes this a nice mock which will not fail on any request.
func (m *MockCloudSpannerClient) MakeNice() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nice = true
}

// MakeStrict makes this a strict mock which will fail on any unexpected request.
func (m *MockCloudSpannerClient) MakeStrict() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nice = false
}

// InjectError injects a global error that will be returned by all APIs regardless of
// the actions array.
func (m *MockCloudSpannerClient) InjectError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.injErr[method] = err
}

// SetActions sets the new set of expected actions to MockCloudSpannerClient.
func (m *MockCloudSpannerClient) SetActions(acts ...Action) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.actions = []Action{}
	for _, act := range acts {
		m.actions = append(m.actions, act)
	}
}

// DumpPings dumps the ping history.
func (m *MockCloudSpannerClient) DumpPings() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.pings...)
}

// DumpSessions dumps the internal session table.
func (m *MockCloudSpannerClient) DumpSessions() map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := map[string]bool{}
	for s, v := range m.sessions {
		st[s] = v
	}
	return st
}

// CreateSession is a placeholder for SpannerClient.CreateSession.
func (m *MockCloudSpannerClient) CreateSession(c context.Context, r *sppb.CreateSessionRequest, opts ...grpc.CallOption) (*sppb.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.injErr["CreateSession"]; err != nil {
		return nil, err
	}
	s := &sppb.Session{}
	if r.Database != "mockdb" {
		// Reject other databases
		return s, grpc.Errorf(codes.NotFound, fmt.Sprintf("database not found: %v", r.Database))
	}
	// Generate & record session name.
	s.Name = fmt.Sprintf("mockdb-%v", time.Now().UnixNano())
	m.sessions[s.Name] = true
	return s, nil
}

// GetSession is a placeholder for SpannerClient.GetSession.
func (m *MockCloudSpannerClient) GetSession(c context.Context, r *sppb.GetSessionRequest, opts ...grpc.CallOption) (*sppb.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.injErr["GetSession"]; err != nil {
		return nil, err
	}
	m.pings = append(m.pings, r.Name)
	if _, ok := m.sessions[r.Name]; !ok {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("Session not found: %v", r.Name))
	}
	return &sppb.Session{Name: r.Name}, nil
}

// DeleteSession is a placeholder for SpannerClient.DeleteSession.
func (m *MockCloudSpannerClient) DeleteSession(c context.Context, r *sppb.DeleteSessionRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.injErr["DeleteSession"]; err != nil {
		return nil, err
	}
	if _, ok := m.sessions[r.Name]; !ok {
		// Session not found.
		return &empty.Empty{}, grpc.Errorf(codes.NotFound, fmt.Sprintf("Session not found: %v", r.Name))
	}
	// Delete session from in-memory table.
	delete(m.sessions, r.Name)
	return &empty.Empty{}, nil
}

// ExecuteSql is a placeholder for SpannerClient.ExecuteSql.
func (m *MockCloudSpannerClient) ExecuteSql(c context.Context, r *sppb.ExecuteSqlRequest, opts ...grpc.CallOption) (*sppb.ResultSet, error) {
	return nil, errors.New("Unimplemented")
}

// ExecuteStreamingSql is a mock implementation of SpannerClient.ExecuteStreamingSql.
func (m *MockCloudSpannerClient) ExecuteStreamingSql(c context.Context, r *sppb.ExecuteSqlRequest, opts ...grpc.CallOption) (sppb.Spanner_ExecuteStreamingSqlClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.injErr["ExecuteStreamingSql"]; err != nil {
		return nil, err
	}
	if len(m.actions) == 0 {
		m.t.Fatalf("unexpected ExecuteStreamingSql executed")
	}
	act := m.actions[0]
	m.actions = m.actions[1:]
	if act.method != "ExecuteStreamingSql" {
		m.t.Fatalf("unexpected ExecuteStreamingSql call, want action: %v", act)
	}
	wantReq := &sppb.ExecuteSqlRequest{
		Session: "mocksession",
		Transaction: &sppb.TransactionSelector{
			Selector: &sppb.TransactionSelector_SingleUse{
				SingleUse: &sppb.TransactionOptions{
					Mode: &sppb.TransactionOptions_ReadOnly_{
						ReadOnly: &sppb.TransactionOptions_ReadOnly{
							TimestampBound: &sppb.TransactionOptions_ReadOnly_Strong{
								Strong: true,
							},
							ReturnReadTimestamp: false,
						},
					},
				},
			},
		},
		Sql: "mockquery",
		Params: &proto3.Struct{
			Fields: map[string]*proto3.Value{"var1": &proto3.Value{Kind: &proto3.Value_StringValue{StringValue: "abc"}}},
		},
		ParamTypes: map[string]*sppb.Type{"var1": &sppb.Type{Code: sppb.TypeCode_STRING}},
	}
	if !reflect.DeepEqual(r, wantReq) {
		return nil, fmt.Errorf("got query request: %v, want: %v", r, wantReq)
	}
	if act.err != nil {
		return nil, act.err
	}
	return nil, errors.New("query never succeeds on mock client")
}

// Read is a placeholder for SpannerClient.Read.
func (m *MockCloudSpannerClient) Read(c context.Context, r *sppb.ReadRequest, opts ...grpc.CallOption) (*sppb.ResultSet, error) {
	m.t.Fatalf("Read is unimplemented")
	return nil, errors.New("Unimplemented")
}

// StreamingRead is a placeholder for SpannerClient.StreamingRead.
func (m *MockCloudSpannerClient) StreamingRead(c context.Context, r *sppb.ReadRequest, opts ...grpc.CallOption) (sppb.Spanner_StreamingReadClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.injErr["StreamingRead"]; err != nil {
		return nil, err
	}
	if len(m.actions) == 0 {
		m.t.Fatalf("unexpected StreamingRead executed")
	}
	act := m.actions[0]
	m.actions = m.actions[1:]
	if act.method != "StreamingRead" && act.method != "StreamingIndexRead" {
		m.t.Fatalf("unexpected read call, want action: %v", act)
	}
	wantReq := &sppb.ReadRequest{
		Session: "mocksession",
		Transaction: &sppb.TransactionSelector{
			Selector: &sppb.TransactionSelector_SingleUse{
				SingleUse: &sppb.TransactionOptions{
					Mode: &sppb.TransactionOptions_ReadOnly_{
						ReadOnly: &sppb.TransactionOptions_ReadOnly{
							TimestampBound: &sppb.TransactionOptions_ReadOnly_Strong{
								Strong: true,
							},
							ReturnReadTimestamp: false,
						},
					},
				},
			},
		},
		Table:   "t_mock",
		Columns: []string{"col1", "col2"},
		KeySet: &sppb.KeySet{
			[]*proto3.ListValue{
				&proto3.ListValue{
					Values: []*proto3.Value{
						&proto3.Value{Kind: &proto3.Value_StringValue{StringValue: "foo"}},
					},
				},
			},
			[]*sppb.KeyRange{},
			false,
		},
	}
	if act.method == "StreamingIndexRead" {
		wantReq.Index = "idx1"
	}
	if !reflect.DeepEqual(r, wantReq) {
		return nil, fmt.Errorf("got query request: %v, want: %v", r, wantReq)
	}
	if act.err != nil {
		return nil, act.err
	}
	return nil, errors.New("read never succeeds on mock client")
}

// BeginTransaction is a placeholder for SpannerClient.BeginTransaction.
func (m *MockCloudSpannerClient) BeginTransaction(c context.Context, r *sppb.BeginTransactionRequest, opts ...grpc.CallOption) (*sppb.Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.nice {
		if err := m.injErr["BeginTransaction"]; err != nil {
			return nil, err
		}
		if len(m.actions) == 0 {
			m.t.Fatalf("unexpected Begin executed")
		}
		act := m.actions[0]
		m.actions = m.actions[1:]
		if act.method != "Begin" {
			m.t.Fatalf("unexpected Begin call, want action: %v", act)
		}
		if act.err != nil {
			return nil, act.err
		}
	}
	resp := &sppb.Transaction{Id: []byte("transaction-1")}
	if _, ok := r.Options.Mode.(*sppb.TransactionOptions_ReadOnly_); ok {
		resp.ReadTimestamp = &pbt.Timestamp{Seconds: 3, Nanos: 4}
	}
	return resp, nil
}

// Commit is a placeholder for SpannerClient.Commit.
func (m *MockCloudSpannerClient) Commit(c context.Context, r *sppb.CommitRequest, opts ...grpc.CallOption) (*sppb.CommitResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.nice {
		if err := m.injErr["Commit"]; err != nil {
			return nil, err
		}
		if len(m.actions) == 0 {
			m.t.Fatalf("unexpected Commit executed")
		}
		act := m.actions[0]
		m.actions = m.actions[1:]
		if act.method != "Commit" {
			m.t.Fatalf("unexpected Commit call, want action: %v", act)
		}
		if act.err != nil {
			return nil, act.err
		}
	}
	return &sppb.CommitResponse{CommitTimestamp: &pbt.Timestamp{Seconds: 1, Nanos: 2}}, nil
}

// Rollback is a placeholder for SpannerClient.Rollback.
func (m *MockCloudSpannerClient) Rollback(c context.Context, r *sppb.RollbackRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.nice {
		if err := m.injErr["Rollback"]; err != nil {
			return nil, err
		}
		if len(m.actions) == 0 {
			m.t.Fatalf("unexpected Rollback executed")
		}
		act := m.actions[0]
		m.actions = m.actions[1:]
		if act.method != "Rollback" {
			m.t.Fatalf("unexpected Rollback call, want action: %v", act)
		}
		if act.err != nil {
			return nil, act.err
		}
	}
	return nil, nil
}
