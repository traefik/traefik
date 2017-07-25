// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// TODO(jba): test that OnError is getting called appropriately.

package logadmin

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/internal/testutil"
	"cloud.google.com/go/logging"
	ltesting "cloud.google.com/go/logging/internal/testing"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	durpb "github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	audit "google.golang.org/genproto/googleapis/cloud/audit"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
	logpb "google.golang.org/genproto/googleapis/logging/v2"
	"google.golang.org/grpc"
)

var (
	client        *Client
	testProjectID string
)

var (
	// If true, this test is using the production service, not a fake.
	integrationTest bool

	newClient func(ctx context.Context, projectID string) *Client
)

func TestMain(m *testing.M) {
	flag.Parse() // needed for testing.Short()
	ctx := context.Background()
	testProjectID = testutil.ProjID()
	if testProjectID == "" || testing.Short() {
		integrationTest = false
		if testProjectID != "" {
			log.Print("Integration tests skipped in short mode (using fake instead)")
		}
		testProjectID = "PROJECT_ID"
		addr, err := ltesting.NewServer()
		if err != nil {
			log.Fatalf("creating fake server: %v", err)
		}
		newClient = func(ctx context.Context, projectID string) *Client {
			conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("dialing %q: %v", addr, err)
			}
			c, err := NewClient(ctx, projectID, option.WithGRPCConn(conn))
			if err != nil {
				log.Fatalf("creating client for fake at %q: %v", addr, err)
			}
			return c
		}
	} else {
		integrationTest = true
		ts := testutil.TokenSource(ctx, logging.AdminScope)
		if ts == nil {
			log.Fatal("The project key must be set. See CONTRIBUTING.md for details")
		}
		log.Printf("running integration tests with project %s", testProjectID)
		newClient = func(ctx context.Context, projectID string) *Client {
			c, err := NewClient(ctx, projectID, option.WithTokenSource(ts),
				option.WithGRPCDialOption(grpc.WithBlock()))
			if err != nil {
				log.Fatalf("creating prod client: %v", err)
			}
			return c
		}
	}
	client = newClient(ctx, testProjectID)
	initMetrics(ctx)
	cleanup := initSinks(ctx)
	exit := m.Run()
	cleanup()
	client.Close()
	os.Exit(exit)
}

// EntryIterator and DeleteLog are tested in the logging package.

func TestClientClose(t *testing.T) {
	c := newClient(context.Background(), testProjectID)
	if err := c.Close(); err != nil {
		t.Errorf("want got %v, want nil", err)
	}
}

func TestFromLogEntry(t *testing.T) {
	now := time.Now()
	res := &mrpb.MonitoredResource{Type: "global"}
	ts, err := ptypes.TimestampProto(now)
	if err != nil {
		t.Fatal(err)
	}
	logEntry := logpb.LogEntry{
		LogName:   "projects/PROJECT_ID/logs/LOG_ID",
		Resource:  res,
		Payload:   &logpb.LogEntry_TextPayload{"hello"},
		Timestamp: ts,
		Severity:  logtypepb.LogSeverity_INFO,
		InsertId:  "123",
		HttpRequest: &logtypepb.HttpRequest{
			RequestMethod:                  "GET",
			RequestUrl:                     "http:://example.com/path?q=1",
			RequestSize:                    100,
			Status:                         200,
			ResponseSize:                   25,
			Latency:                        &durpb.Duration{Seconds: 100},
			UserAgent:                      "user-agent",
			RemoteIp:                       "127.0.0.1",
			Referer:                        "referer",
			CacheHit:                       true,
			CacheValidatedWithOriginServer: true,
		},
		Labels: map[string]string{
			"a": "1",
			"b": "two",
			"c": "true",
		},
	}
	u, err := url.Parse("http:://example.com/path?q=1")
	if err != nil {
		t.Fatal(err)
	}
	want := &logging.Entry{
		LogName:   "projects/PROJECT_ID/logs/LOG_ID",
		Resource:  res,
		Timestamp: now.In(time.UTC),
		Severity:  logging.Info,
		Payload:   "hello",
		Labels: map[string]string{
			"a": "1",
			"b": "two",
			"c": "true",
		},
		InsertID: "123",
		HTTPRequest: &logging.HTTPRequest{
			Request: &http.Request{
				Method: "GET",
				URL:    u,
				Header: map[string][]string{
					"User-Agent": []string{"user-agent"},
					"Referer":    []string{"referer"},
				},
			},
			RequestSize:                    100,
			Status:                         200,
			ResponseSize:                   25,
			Latency:                        100 * time.Second,
			RemoteIP:                       "127.0.0.1",
			CacheHit:                       true,
			CacheValidatedWithOriginServer: true,
		},
	}
	got, err := fromLogEntry(&logEntry)
	if err != nil {
		t.Fatal(err)
	}
	// Test sub-values separately because %+v and %#v do not follow pointers.
	// TODO(jba): use a differ or pretty-printer.
	if !reflect.DeepEqual(got.HTTPRequest.Request, want.HTTPRequest.Request) {
		t.Fatalf("HTTPRequest.Request:\ngot  %+v\nwant %+v", got.HTTPRequest.Request, want.HTTPRequest.Request)
	}
	if !reflect.DeepEqual(got.HTTPRequest, want.HTTPRequest) {
		t.Fatalf("HTTPRequest:\ngot  %+v\nwant %+v", got.HTTPRequest, want.HTTPRequest)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FullEntry:\ngot  %+v\nwant %+v", got, want)
	}

	// Proto payload.
	alog := &audit.AuditLog{
		ServiceName:  "svc",
		MethodName:   "method",
		ResourceName: "shelves/S/books/B",
	}
	any, err := ptypes.MarshalAny(alog)
	if err != nil {
		t.Fatal(err)
	}
	logEntry = logpb.LogEntry{
		LogName:   "projects/PROJECT_ID/logs/LOG_ID",
		Resource:  res,
		Timestamp: ts,
		Payload:   &logpb.LogEntry_ProtoPayload{any},
	}
	got, err = fromLogEntry(&logEntry)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Payload, alog) {
		t.Errorf("got %+v, want %+v", got.Payload, alog)
	}

	// JSON payload.
	jstruct := &structpb.Struct{map[string]*structpb.Value{
		"f": &structpb.Value{&structpb.Value_NumberValue{3.1}},
	}}
	logEntry = logpb.LogEntry{
		LogName:   "projects/PROJECT_ID/logs/LOG_ID",
		Resource:  res,
		Timestamp: ts,
		Payload:   &logpb.LogEntry_JsonPayload{jstruct},
	}
	got, err = fromLogEntry(&logEntry)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Payload, jstruct) {
		t.Errorf("got %+v, want %+v", got.Payload, jstruct)
	}
}

func TestListLogEntriesRequest(t *testing.T) {
	for _, test := range []struct {
		opts       []EntriesOption
		projectIDs []string
		filter     string
		orderBy    string
	}{
		// Default is client's project ID, empty filter and orderBy.
		{nil,
			[]string{"PROJECT_ID"}, "", ""},
		{[]EntriesOption{NewestFirst(), Filter("f")},
			[]string{"PROJECT_ID"}, "f", "timestamp desc"},
		{[]EntriesOption{ProjectIDs([]string{"foo"})},
			[]string{"foo"}, "", ""},
		{[]EntriesOption{NewestFirst(), Filter("f"), ProjectIDs([]string{"foo"})},
			[]string{"foo"}, "f", "timestamp desc"},
		{[]EntriesOption{NewestFirst(), Filter("f"), ProjectIDs([]string{"foo"})},
			[]string{"foo"}, "f", "timestamp desc"},
		// If there are repeats, last one wins.
		{[]EntriesOption{NewestFirst(), Filter("no"), ProjectIDs([]string{"foo"}), Filter("f")},
			[]string{"foo"}, "f", "timestamp desc"},
	} {
		got := listLogEntriesRequest("PROJECT_ID", test.opts)
		want := &logpb.ListLogEntriesRequest{
			ProjectIds: test.projectIDs,
			Filter:     test.filter,
			OrderBy:    test.orderBy,
		}
		if !proto.Equal(got, want) {
			t.Errorf("%v:\ngot  %v\nwant %v", test.opts, got, want)
		}
	}
}
