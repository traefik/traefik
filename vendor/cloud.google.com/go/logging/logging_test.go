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

package logging_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	gax "github.com/googleapis/gax-go"

	cinternal "cloud.google.com/go/internal"
	"cloud.google.com/go/internal/testutil"
	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/internal"
	ltesting "cloud.google.com/go/logging/internal/testing"
	"cloud.google.com/go/logging/logadmin"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/grpc"
)

const testLogIDPrefix = "GO-LOGGING-CLIENT/TEST-LOG"

var (
	client        *logging.Client
	aclient       *logadmin.Client
	testProjectID string
	testLogID     string
	testFilter    string
	errorc        chan error
	ctx           context.Context

	// Adjust the fields of a FullEntry received from the production service
	// before comparing it with the expected result. We can't correctly
	// compare certain fields, like times or server-generated IDs.
	clean func(*logging.Entry)

	// Create a new client with the given project ID.
	newClients func(ctx context.Context, projectID string) (*logging.Client, *logadmin.Client)
)

func testNow() time.Time {
	return time.Unix(1000, 0)
}

// If true, this test is using the production service, not a fake.
var integrationTest bool

func TestMain(m *testing.M) {
	flag.Parse() // needed for testing.Short()
	ctx = context.Background()
	testProjectID = testutil.ProjID()
	errorc = make(chan error, 100)
	if testProjectID == "" || testing.Short() {
		integrationTest = false
		if testProjectID != "" {
			log.Print("Integration tests skipped in short mode (using fake instead)")
		}
		testProjectID = "PROJECT_ID"
		clean = func(e *logging.Entry) {
			// Remove the insert ID for consistency with the integration test.
			e.InsertID = ""
		}

		addr, err := ltesting.NewServer()
		if err != nil {
			log.Fatalf("creating fake server: %v", err)
		}
		logging.SetNow(testNow)

		newClients = func(ctx context.Context, projectID string) (*logging.Client, *logadmin.Client) {
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("dialing %q: %v", addr, err)
			}
			c, err := logging.NewClient(ctx, projectID, option.WithGRPCConn(conn))
			if err != nil {
				log.Fatalf("creating client for fake at %q: %v", addr, err)
			}
			ac, err := logadmin.NewClient(ctx, projectID, option.WithGRPCConn(conn))
			if err != nil {
				log.Fatalf("creating client for fake at %q: %v", addr, err)
			}
			return c, ac
		}

	} else {
		integrationTest = true
		clean = func(e *logging.Entry) {
			// We cannot compare timestamps, so set them to the test time.
			// Also, remove the insert ID added by the service.
			e.Timestamp = testNow().UTC()
			e.InsertID = ""
		}
		ts := testutil.TokenSource(ctx, logging.AdminScope)
		if ts == nil {
			log.Fatal("The project key must be set. See CONTRIBUTING.md for details")
		}
		log.Printf("running integration tests with project %s", testProjectID)
		newClients = func(ctx context.Context, projectID string) (*logging.Client, *logadmin.Client) {
			c, err := logging.NewClient(ctx, projectID, option.WithTokenSource(ts))
			if err != nil {
				log.Fatalf("creating prod client: %v", err)
			}
			ac, err := logadmin.NewClient(ctx, projectID, option.WithTokenSource(ts))
			if err != nil {
				log.Fatalf("creating prod client: %v", err)
			}
			return c, ac
		}

	}
	client, aclient = newClients(ctx, testProjectID)
	client.OnError = func(e error) { errorc <- e }

	exit := m.Run()
	client.Close()
	os.Exit(exit)
}

func initLogs(ctx context.Context) {
	testLogID = ltesting.UniqueID(testLogIDPrefix)
	testFilter = fmt.Sprintf(`logName = "projects/%s/logs/%s"`, testProjectID,
		strings.Replace(testLogID, "/", "%2F", -1))
	// TODO(jba): Clean up from previous aborted tests by deleting old logs; requires ListLogs RPC.
}

// Testing of Logger.Log is done in logadmin_test.go, TestEntries.

func TestLogSync(t *testing.T) {
	initLogs(ctx) // Generate new testLogID
	ctx := context.Background()
	lg := client.Logger(testLogID)
	err := lg.LogSync(ctx, logging.Entry{Payload: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	err = lg.LogSync(ctx, logging.Entry{Payload: "goodbye"})
	if err != nil {
		t.Fatal(err)
	}
	// Allow overriding the MonitoredResource.
	err = lg.LogSync(ctx, logging.Entry{Payload: "mr", Resource: &mrpb.MonitoredResource{Type: "global"}})
	if err != nil {
		t.Fatal(err)
	}

	want := []*logging.Entry{
		entryForTesting("hello"),
		entryForTesting("goodbye"),
		entryForTesting("mr"),
	}
	var got []*logging.Entry
	ok := waitFor(func() bool {
		got, err = allTestLogEntries(ctx)
		if err != nil {
			t.Log("fetching log entries: ", err)
			return false
		}
		return len(got) == len(want)
	})
	if !ok {
		t.Fatalf("timed out; got: %d, want: %d\n", len(got), len(want))
	}
	if msg, ok := compareEntries(got, want); !ok {
		t.Error(msg)
	}
}

func TestLogAndEntries(t *testing.T) {
	initLogs(ctx) // Generate new testLogID
	ctx := context.Background()
	payloads := []string{"p1", "p2", "p3", "p4", "p5"}
	lg := client.Logger(testLogID)
	for _, p := range payloads {
		// Use the insert ID to guarantee iteration order.
		lg.Log(logging.Entry{Payload: p, InsertID: p})
	}
	lg.Flush()
	var want []*logging.Entry
	for _, p := range payloads {
		want = append(want, entryForTesting(p))
	}
	var got []*logging.Entry
	ok := waitFor(func() bool {
		var err error
		got, err = allTestLogEntries(ctx)
		if err != nil {
			t.Log("fetching log entries: ", err)
			return false
		}
		return len(got) == len(want)
	})
	if !ok {
		t.Fatalf("timed out; got: %d, want: %d\n", len(got), len(want))
	}
	if msg, ok := compareEntries(got, want); !ok {
		t.Error(msg)
	}
}

// compareEntries compares most fields list of Entries against expected. compareEntries does not compare:
//   - HTTPRequest
//   - Operation
//   - Resource
func compareEntries(got, want []*logging.Entry) (string, bool) {
	if len(got) != len(want) {
		return fmt.Sprintf("got %d entries, want %d", len(got), len(want)), false
	}
	for i := range got {
		if !compareEntry(got[i], want[i]) {
			return fmt.Sprintf("#%d:\ngot  %+v\nwant %+v", i, got[i], want[i]), false
		}
	}
	return "", true
}

func compareEntry(got, want *logging.Entry) bool {
	if got.Timestamp.Unix() != want.Timestamp.Unix() {
		return false
	}

	if got.Severity != want.Severity {
		return false
	}

	if !reflect.DeepEqual(got.Payload, want.Payload) {
		return false
	}

	if !reflect.DeepEqual(got.Labels, want.Labels) {
		return false
	}

	if got.InsertID != want.InsertID {
		return false
	}

	if got.LogName != want.LogName {
		return false
	}

	return true
}

func entryForTesting(payload interface{}) *logging.Entry {
	return &logging.Entry{
		Timestamp: testNow().UTC(),
		Payload:   payload,
		LogName:   "projects/" + testProjectID + "/logs/" + testLogID,
		Resource:  &mrpb.MonitoredResource{Type: "global", Labels: map[string]string{"project_id": testProjectID}},
	}
}

func countLogEntries(ctx context.Context, filter string) int {
	it := aclient.Entries(ctx, logadmin.Filter(filter))
	n := 0
	for {
		_, err := it.Next()
		if err == iterator.Done {
			return n
		}
		if err != nil {
			log.Fatalf("counting log entries: %v", err)
		}
		n++
	}
}

func allTestLogEntries(ctx context.Context) ([]*logging.Entry, error) {
	var es []*logging.Entry
	it := aclient.Entries(ctx, logadmin.Filter(testFilter))
	for {
		e, err := cleanNext(it)
		switch err {
		case nil:
			es = append(es, e)
		case iterator.Done:
			return es, nil
		default:
			return nil, err
		}
	}
}

func cleanNext(it *logadmin.EntryIterator) (*logging.Entry, error) {
	e, err := it.Next()
	if err != nil {
		return nil, err
	}
	clean(e)
	return e, nil
}

func TestStandardLogger(t *testing.T) {
	initLogs(ctx) // Generate new testLogID
	ctx := context.Background()
	lg := client.Logger(testLogID)
	slg := lg.StandardLogger(logging.Info)

	if slg != lg.StandardLogger(logging.Info) {
		t.Error("There should be only one standard logger at each severity.")
	}
	if slg == lg.StandardLogger(logging.Debug) {
		t.Error("There should be a different standard logger for each severity.")
	}

	slg.Print("info")
	lg.Flush()
	var got []*logging.Entry
	ok := waitFor(func() bool {
		var err error
		got, err = allTestLogEntries(ctx)
		if err != nil {
			t.Log("fetching log entries: ", err)
			return false
		}
		return len(got) == 1
	})
	if !ok {
		t.Fatalf("timed out; got: %d, want: %d\n", len(got), 1)
	}
	if len(got) != 1 {
		t.Fatalf("expected non-nil request with one entry; got:\n%+v", got)
	}
	if got, want := got[0].Payload.(string), "info\n"; got != want {
		t.Errorf("payload: got %q, want %q", got, want)
	}
	if got, want := logging.Severity(got[0].Severity), logging.Info; got != want {
		t.Errorf("severity: got %s, want %s", got, want)
	}
}

func TestSeverity(t *testing.T) {
	if got, want := logging.Info.String(), "Info"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := logging.Severity(-99).String(), "-99"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseSeverity(t *testing.T) {
	for _, test := range []struct {
		in   string
		want logging.Severity
	}{
		{"", logging.Default},
		{"whatever", logging.Default},
		{"Default", logging.Default},
		{"ERROR", logging.Error},
		{"Error", logging.Error},
		{"error", logging.Error},
	} {
		got := logging.ParseSeverity(test.in)
		if got != test.want {
			t.Errorf("%q: got %s, want %s\n", test.in, got, test.want)
		}
	}
}

func TestErrors(t *testing.T) {
	initLogs(ctx) // Generate new testLogID
	// Drain errors already seen.
loop:
	for {
		select {
		case <-errorc:
		default:
			break loop
		}
	}
	// Try to log something that can't be JSON-marshalled.
	lg := client.Logger(testLogID)
	lg.Log(logging.Entry{Payload: func() {}})
	// Expect an error.
	select {
	case <-errorc: // pass
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected an error but timed out")
	}
}

type badTokenSource struct{}

func (badTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{}, nil
}

func TestPing(t *testing.T) {
	// Ping twice, in case the service's InsertID logic messes with the error code.
	ctx := context.Background()
	// The global client should be valid.
	if err := client.Ping(ctx); err != nil {
		t.Errorf("project %s: got %v, expected nil", testProjectID, err)
	}
	if err := client.Ping(ctx); err != nil {
		t.Errorf("project %s, #2: got %v, expected nil", testProjectID, err)
	}
	// nonexistent project
	c, _ := newClients(ctx, testProjectID+"-BAD")
	if err := c.Ping(ctx); err == nil {
		t.Errorf("nonexistent project: want error pinging logging api, got nil")
	}
	if err := c.Ping(ctx); err == nil {
		t.Errorf("nonexistent project, #2: want error pinging logging api, got nil")
	}

	// Bad creds. We cannot test this with the fake, since it doesn't do auth.
	if integrationTest {
		c, err := logging.NewClient(ctx, testProjectID, option.WithTokenSource(badTokenSource{}))
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Ping(ctx); err == nil {
			t.Errorf("bad creds: want error pinging logging api, got nil")
		}
		if err := c.Ping(ctx); err == nil {
			t.Errorf("bad creds, #2: want error pinging logging api, got nil")
		}
		if err := c.Close(); err != nil {
			t.Fatalf("error closing client: %v", err)
		}
	}
}

func TestDeleteLog(t *testing.T) {
	initLogs(ctx) // Generate new testLogID
	// Write some log entries.
	ctx := context.Background()
	payloads := []string{"p1", "p2"}
	lg := client.Logger(testLogID)
	for _, p := range payloads {
		// Use the insert ID to guarantee iteration order.
		lg.Log(logging.Entry{Payload: p, InsertID: p})
	}
	lg.Flush()

	var got []*logging.Entry
	ok := waitFor(func() bool {
		var err error
		got, err = allTestLogEntries(ctx)
		if err != nil {
			t.Log("fetching log entries: ", err)
			return false
		}
		return len(got) == 2
	})
	if !ok {
		t.Fatalf("timed out; got: %d, want: %d\n", len(got), 2)
	}

	// Sleep.
	// Write timestamp uses client-provided timestamp, delete uses server
	// timestamp. We sleep to reduce the possibility that the logs are never
	// "deleted" because of clock skew.
	// This is the recommended approach by Stackdriver team.
	time.Sleep(3 * time.Second)

	// Delete the log
	err := aclient.DeleteLog(ctx, testLogID)
	if err != nil {
		log.Fatalf("error deleting log: %v", err)
	}

	// DeleteLog can take some time to happen, so we wait for the log to
	// disappear. There is no direct way to determine if a log exists, so we
	// just wait until there are no log entries associated with the ID.
	filter := fmt.Sprintf(`logName = "%s"`, internal.LogPath("projects/"+testProjectID, testLogID))
	ok = waitFor(func() bool { return countLogEntries(ctx, filter) == 0 })
	if !ok {
		t.Fatalf("timed out waiting for log entries to be deleted")
	}
}

// waitFor calls f repeatedly with exponential backoff, blocking until it returns true.
// It returns false after a while (if it times out).
func waitFor(f func() bool) bool {
	// TODO(shadams): Find a better way to deflake these tests.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err := cinternal.Retry(ctx,
		gax.Backoff{Initial: time.Second, Multiplier: 2},
		func() (bool, error) { return f(), nil })
	return err == nil
}
