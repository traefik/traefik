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

package spanner

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/internal/testutil"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	adminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
)

var (
	// testProjectID specifies the project used for testing.
	// It can be changed by setting environment variable GCLOUD_TESTS_GOLANG_PROJECT_ID.
	testProjectID = testutil.ProjID()
	// testInstanceID specifies the Cloud Spanner instance used for testing.
	testInstanceID = "go-integration-test"

	// client is a spanner.Client.
	client *Client
	// admin is a spanner.DatabaseAdminClient.
	admin *database.DatabaseAdminClient
	// db is the path of the testing database.
	db string
	// dbName is the short name of the testing database.
	dbName string
)

// prepare initializes Cloud Spanner testing DB and clients.
func prepare(ctx context.Context, t *testing.T) error {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	if testProjectID == "" {
		t.Skip("Integration tests skipped: GCLOUD_TESTS_GOLANG_PROJECT_ID is missing")
	}
	ts := testutil.TokenSource(ctx, AdminScope, Scope)
	if ts == nil {
		t.Skip("Integration test skipped: cannot get service account credential from environment variable %v", "GCLOUD_TESTS_GOLANG_KEY")
	}
	var err error
	// Create Admin client and Data client.
	// TODO: Remove the EndPoint option once this is the default.
	admin, err = database.NewDatabaseAdminClient(ctx, option.WithTokenSource(ts), option.WithEndpoint("spanner.googleapis.com:443"))
	if err != nil {
		t.Errorf("cannot create admin client: %v", err)
		return err
	}
	// Construct test DB name.
	dbName = fmt.Sprintf("gotest_%v", time.Now().UnixNano())
	db = fmt.Sprintf("projects/%v/instances/%v/databases/%v", testProjectID, testInstanceID, dbName)
	// Create database and tables.
	op, err := admin.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%v/instances/%v", testProjectID, testInstanceID),
		CreateStatement: "CREATE DATABASE " + dbName,
		ExtraStatements: []string{
			`CREATE TABLE Singers (
				SingerId	INT64 NOT NULL,
				FirstName	STRING(1024),
				LastName	STRING(1024),
				SingerInfo	BYTES(MAX)
			) PRIMARY KEY (SingerId)`,
			`CREATE INDEX SingerByName ON Singers(FirstName, LastName)`,
			`CREATE TABLE Accounts (
				AccountId	INT64 NOT NULL,
				Nickname	STRING(100),
				Balance		INT64 NOT NULL,
			) PRIMARY KEY (AccountId)`,
			`CREATE INDEX AccountByNickname ON Accounts(Nickname) STORING (Balance)`,
			`CREATE TABLE Types (
				RowID		INT64 NOT NULL,
				String		STRING(MAX),
				StringArray	ARRAY<STRING(MAX)>,
				Bytes		BYTES(MAX),
				BytesArray	ARRAY<BYTES(MAX)>,
				Int64a		INT64,
				Int64Array	ARRAY<INT64>,
				Bool		BOOL,
				BoolArray	ARRAY<BOOL>,
				Float64		FLOAT64,
				Float64Array	ARRAY<FLOAT64>,
				Date		DATE,
				DateArray	ARRAY<DATE>,
				Timestamp	TIMESTAMP,
				TimestampArray	ARRAY<TIMESTAMP>,
			) PRIMARY KEY (RowID)`,
		},
	})
	if err != nil {
		t.Errorf("cannot create testing DB %v: %v", db, err)
		return err
	}
	if _, err := op.Wait(ctx); err != nil {
		t.Errorf("cannot create testing DB %v: %v", db, err)
		return err
	}
	client, err = NewClientWithConfig(ctx, db, ClientConfig{
		SessionPoolConfig: SessionPoolConfig{
			WriteSessions: 0.2,
		},
	}, option.WithTokenSource(ts))
	if err != nil {
		t.Errorf("cannot create data client on DB %v: %v", db, err)
		return err
	}
	return nil
}

// tearDown tears down the testing environment created by prepare().
func tearDown(ctx context.Context, t *testing.T) {
	if admin != nil {
		if err := admin.DropDatabase(ctx, &adminpb.DropDatabaseRequest{db}); err != nil {
			t.Logf("failed to drop testing database: %v, might need a manual removal", db)
		}
		admin.Close()
	}
	if client != nil {
		client.Close()
	}
	admin = nil
	client = nil
	db = ""
}

// Test SingleUse transaction.
func TestSingleUse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Set up testing environment.
	if err := prepare(ctx, t); err != nil {
		// If prepare() fails, tear down whatever that's already up.
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	// After all tests, tear down testing environment.
	defer tearDown(ctx, t)

	writes := []struct {
		row []interface{}
		ts  time.Time
	}{
		{row: []interface{}{1, "Marc", "Foo"}},
		{row: []interface{}{2, "Tars", "Bar"}},
		{row: []interface{}{3, "Alpha", "Beta"}},
		{row: []interface{}{4, "Last", "End"}},
	}
	// Try to write four rows through the Apply API.
	for i, w := range writes {
		var err error
		m := InsertOrUpdate("Singers",
			[]string{"SingerId", "FirstName", "LastName"},
			w.row)
		if writes[i].ts, err = client.Apply(ctx, []*Mutation{m}, ApplyAtLeastOnce()); err != nil {
			t.Fatal(err)
		}
	}

	// For testing timestamp bound staleness.
	<-time.After(time.Second)

	// Test reading rows with different timestamp bounds.
	for i, test := range []struct {
		want    [][]interface{}
		tb      TimestampBound
		checkTs func(time.Time) error
	}{
		{
			// strong
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}, {int64(4), "Last", "End"}},
			StrongRead(),
			func(ts time.Time) error {
				// writes[3] is the last write, all subsequent strong read should have a timestamp larger than that.
				if ts.Before(writes[3].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no later than %v", ts, writes[3].ts)
				}
				return nil
			},
		},
		{
			// min_read_timestamp
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}, {int64(4), "Last", "End"}},
			MinReadTimestamp(writes[3].ts),
			func(ts time.Time) error {
				if ts.Before(writes[3].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no later than %v", ts, writes[3].ts)
				}
				return nil
			},
		},
		{
			// max_staleness
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}, {int64(4), "Last", "End"}},
			MaxStaleness(time.Second),
			func(ts time.Time) error {
				if ts.Before(writes[3].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no later than %v", ts, writes[3].ts)
				}
				return nil
			},
		},
		{
			// read_timestamp
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}},
			ReadTimestamp(writes[2].ts),
			func(ts time.Time) error {
				if ts != writes[2].ts {
					return fmt.Errorf("read got timestamp %v, expect %v", ts, writes[2].ts)
				}
				return nil
			},
		},
		{
			// exact_staleness
			nil,
			// Specify a staleness which should be already before this test because
			// context timeout is set to be 10s.
			ExactStaleness(11 * time.Second),
			func(ts time.Time) error {
				if ts.After(writes[0].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no earlier than %v", ts, writes[0].ts)
				}
				return nil
			},
		},
	} {
		// SingleUse.Query
		su := client.Single().WithTimestampBound(test.tb)
		got, err := readAll(su.Query(
			ctx,
			Statement{
				"SELECT SingerId, FirstName, LastName FROM Singers WHERE SingerId IN (@id1, @id3, @id4)",
				map[string]interface{}{"id1": int64(1), "id3": int64(3), "id4": int64(4)},
			}))
		if err != nil {
			t.Errorf("%d: SingleUse.Query returns error %v, want nil", i, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected result from SingleUse.Query: %v, want %v", i, got, test.want)
		}
		rts, err := su.Timestamp()
		if err != nil {
			t.Errorf("%d: SingleUse.Query doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: SingleUse.Query doesn't return expected timestamp: %v", i, err)
		}
		// SingleUse.Read
		su = client.Single().WithTimestampBound(test.tb)
		got, err = readAll(su.Read(ctx, "Singers", Keys(Key{1}, Key{3}, Key{4}), []string{"SingerId", "FirstName", "LastName"}))
		if err != nil {
			t.Errorf("%d: SingleUse.Read returns error %v, want nil", i, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected result from SingleUse.Read: %v, want %v", i, got, test.want)
		}
		rts, err = su.Timestamp()
		if err != nil {
			t.Errorf("%d: SingleUse.Read doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: SingleUse.Read doesn't return expected timestamp: %v", i, err)
		}
		// SingleUse.ReadRow
		got = nil
		for _, k := range []Key{Key{1}, Key{3}, Key{4}} {
			su = client.Single().WithTimestampBound(test.tb)
			r, err := su.ReadRow(ctx, "Singers", k, []string{"SingerId", "FirstName", "LastName"})
			if err != nil {
				continue
			}
			v, err := rowToValues(r)
			if err != nil {
				continue
			}
			got = append(got, v)
			rts, err = su.Timestamp()
			if err != nil {
				t.Errorf("%d: SingleUse.ReadRow(%v) doesn't return a timestamp, error: %v", i, k, err)
			}
			if err := test.checkTs(rts); err != nil {
				t.Errorf("%d: SingleUse.ReadRow(%v) doesn't return expected timestamp: %v", i, k, err)
			}
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected results from SingleUse.ReadRow: %v, want %v", i, got, test.want)
		}
		// SingleUse.ReadUsingIndex
		su = client.Single().WithTimestampBound(test.tb)
		got, err = readAll(su.ReadUsingIndex(ctx, "Singers", "SingerByName", Keys(Key{"Marc", "Foo"}, Key{"Alpha", "Beta"}, Key{"Last", "End"}), []string{"SingerId", "FirstName", "LastName"}))
		if err != nil {
			t.Errorf("%d: SingleUse.ReadUsingIndex returns error %v, want nil", i, err)
		}
		// The results from ReadUsingIndex is sorted by the index rather than primary key.
		if len(got) != len(test.want) {
			t.Errorf("%d: got unexpected result from SingleUse.ReadUsingIndex: %v, want %v", i, got, test.want)
		}
		for j, g := range got {
			if j > 0 {
				prev := got[j-1][1].(string) + got[j-1][2].(string)
				curr := got[j][1].(string) + got[j][2].(string)
				if strings.Compare(prev, curr) > 0 {
					t.Errorf("%d: SingleUse.ReadUsingIndex fails to order rows by index keys, %v should be after %v", i, got[j-1], got[j])
				}
			}
			found := false
			for _, w := range test.want {
				if reflect.DeepEqual(g, w) {
					found = true
				}
			}
			if !found {
				t.Errorf("%d: got unexpected result from SingleUse.ReadUsingIndex: %v, want %v", i, got, test.want)
				break
			}
		}
		rts, err = su.Timestamp()
		if err != nil {
			t.Errorf("%d: SingleUse.ReadUsingIndex doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: SingleUse.ReadUsingIndex doesn't return expected timestamp: %v", i, err)
		}
	}
}

// Test ReadOnlyTransaction. The testsuite is mostly like SingleUse, except it
// also tests for a single timestamp across multiple reads.
func TestReadOnlyTransaction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// Set up testing environment.
	if err := prepare(ctx, t); err != nil {
		// If prepare() fails, tear down whatever that's already up.
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	// After all tests, tear down testing environment.
	defer tearDown(ctx, t)

	writes := []struct {
		row []interface{}
		ts  time.Time
	}{
		{row: []interface{}{1, "Marc", "Foo"}},
		{row: []interface{}{2, "Tars", "Bar"}},
		{row: []interface{}{3, "Alpha", "Beta"}},
		{row: []interface{}{4, "Last", "End"}},
	}
	// Try to write four rows through the Apply API.
	for i, w := range writes {
		var err error
		m := InsertOrUpdate("Singers",
			[]string{"SingerId", "FirstName", "LastName"},
			w.row)
		if writes[i].ts, err = client.Apply(ctx, []*Mutation{m}, ApplyAtLeastOnce()); err != nil {
			t.Fatal(err)
		}
	}

	// For testing timestamp bound staleness.
	<-time.After(time.Second)

	// Test reading rows with different timestamp bounds.
	for i, test := range []struct {
		want    [][]interface{}
		tb      TimestampBound
		checkTs func(time.Time) error
	}{
		// Note: min_read_timestamp and max_staleness are not supported by ReadOnlyTransaction. See
		// API document for more details.
		{
			// strong
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}, {int64(4), "Last", "End"}},
			StrongRead(),
			func(ts time.Time) error {
				if ts.Before(writes[3].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no later than %v", ts, writes[3].ts)
				}
				return nil
			},
		},
		{
			// read_timestamp
			[][]interface{}{{int64(1), "Marc", "Foo"}, {int64(3), "Alpha", "Beta"}},
			ReadTimestamp(writes[2].ts),
			func(ts time.Time) error {
				if ts != writes[2].ts {
					return fmt.Errorf("read got timestamp %v, expect %v", ts, writes[2].ts)
				}
				return nil
			},
		},
		{
			// exact_staleness
			nil,
			// Specify a staleness which should be already before this test because
			// context timeout is set to be 10s.
			ExactStaleness(11 * time.Second),
			func(ts time.Time) error {
				if ts.After(writes[0].ts) {
					return fmt.Errorf("read got timestamp %v, want it to be no earlier than %v", ts, writes[0].ts)
				}
				return nil
			},
		},
	} {
		// ReadOnlyTransaction.Query
		ro := client.ReadOnlyTransaction().WithTimestampBound(test.tb)
		got, err := readAll(ro.Query(
			ctx,
			Statement{
				"SELECT SingerId, FirstName, LastName FROM Singers WHERE SingerId IN (@id1, @id3, @id4)",
				map[string]interface{}{"id1": int64(1), "id3": int64(3), "id4": int64(4)},
			}))
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Query returns error %v, want nil", i, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected result from ReadOnlyTransaction.Query: %v, want %v", i, got, test.want)
		}
		rts, err := ro.Timestamp()
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Query doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Query doesn't return expected timestamp: %v", i, err)
		}
		roTs := rts
		// ReadOnlyTransaction.Read
		got, err = readAll(ro.Read(ctx, "Singers", Keys(Key{1}, Key{3}, Key{4}), []string{"SingerId", "FirstName", "LastName"}))
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Read returns error %v, want nil", i, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected result from ReadOnlyTransaction.Read: %v, want %v", i, got, test.want)
		}
		rts, err = ro.Timestamp()
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Read doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: ReadOnlyTransaction.Read doesn't return expected timestamp: %v", i, err)
		}
		if roTs != rts {
			t.Errorf("%d: got two read timestamps: %v, %v, want ReadOnlyTransaction to return always the same read timestamp", i, roTs, rts)
		}
		// ReadOnlyTransaction.ReadRow
		got = nil
		for _, k := range []Key{Key{1}, Key{3}, Key{4}} {
			r, err := ro.ReadRow(ctx, "Singers", k, []string{"SingerId", "FirstName", "LastName"})
			if err != nil {
				continue
			}
			v, err := rowToValues(r)
			if err != nil {
				continue
			}
			got = append(got, v)
			rts, err = ro.Timestamp()
			if err != nil {
				t.Errorf("%d: ReadOnlyTransaction.ReadRow(%v) doesn't return a timestamp, error: %v", i, k, err)
			}
			if err := test.checkTs(rts); err != nil {
				t.Errorf("%d: ReadOnlyTransaction.ReadRow(%v) doesn't return expected timestamp: %v", i, k, err)
			}
			if roTs != rts {
				t.Errorf("%d: got two read timestamps: %v, %v, want ReadOnlyTransaction to return always the same read timestamp", i, roTs, rts)
			}
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%d: got unexpected results from ReadOnlyTransaction.ReadRow: %v, want %v", i, got, test.want)
		}
		// SingleUse.ReadUsingIndex
		got, err = readAll(ro.ReadUsingIndex(ctx, "Singers", "SingerByName", Keys(Key{"Marc", "Foo"}, Key{"Alpha", "Beta"}, Key{"Last", "End"}), []string{"SingerId", "FirstName", "LastName"}))
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.ReadUsingIndex returns error %v, want nil", i, err)
		}
		// The results from ReadUsingIndex is sorted by the index rather than primary key.
		if len(got) != len(test.want) {
			t.Errorf("%d: got unexpected result from ReadOnlyTransaction.ReadUsingIndex: %v, want %v", i, got, test.want)
		}
		for j, g := range got {
			if j > 0 {
				prev := got[j-1][1].(string) + got[j-1][2].(string)
				curr := got[j][1].(string) + got[j][2].(string)
				if strings.Compare(prev, curr) > 0 {
					t.Errorf("%d: ReadOnlyTransaction.ReadUsingIndex fails to order rows by index keys, %v should be after %v", i, got[j-1], got[j])
				}
			}
			found := false
			for _, w := range test.want {
				if reflect.DeepEqual(g, w) {
					found = true
				}
			}
			if !found {
				t.Errorf("%d: got unexpected result from ReadOnlyTransaction.ReadUsingIndex: %v, want %v", i, got, test.want)
				break
			}
		}
		rts, err = ro.Timestamp()
		if err != nil {
			t.Errorf("%d: ReadOnlyTransaction.ReadUsingIndex doesn't return a timestamp, error: %v", i, err)
		}
		if err := test.checkTs(rts); err != nil {
			t.Errorf("%d: ReadOnlyTransaction.ReadUsingIndex doesn't return expected timestamp: %v", i, err)
		}
		if roTs != rts {
			t.Errorf("%d: got two read timestamps: %v, %v, want ReadOnlyTransaction to return always the same read timestamp", i, roTs, rts)
		}
		ro.Close()
	}
}

// Test ReadWriteTransaction.
func TestReadWriteTransaction(t *testing.T) {
	// Give a longer deadline because of transaction backoffs.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := prepare(ctx, t); err != nil {
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	defer tearDown(ctx, t)

	// Set up two accounts
	accounts := []*Mutation{
		Insert("Accounts", []string{"AccountId", "Nickname", "Balance"}, []interface{}{int64(1), "Foo", int64(50)}),
		Insert("Accounts", []string{"AccountId", "Nickname", "Balance"}, []interface{}{int64(2), "Bar", int64(1)}),
	}
	if _, err := client.Apply(ctx, accounts, ApplyAtLeastOnce()); err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}

	readBalance := func(iter *RowIterator) (int64, error) {
		defer iter.Stop()
		var bal int64
		for {
			row, err := iter.Next()
			if err == iterator.Done {
				return bal, nil
			}
			if err != nil {
				return 0, err
			}
			if err := row.Column(0, &bal); err != nil {
				return 0, err
			}
		}
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(iter int) {
			defer wg.Done()
			_, err := client.ReadWriteTransaction(ctx, func(tx *ReadWriteTransaction) error {
				// Query Foo's balance and Bar's balance.
				bf, e := readBalance(tx.Query(ctx,
					Statement{"SELECT Balance FROM Accounts WHERE AccountId = @id", map[string]interface{}{"id": int64(1)}}))
				if e != nil {
					return e
				}
				bb, e := readBalance(tx.Read(ctx, "Accounts", Keys(Key{int64(2)}), []string{"Balance"}))
				if e != nil {
					return e
				}
				if bf <= 0 {
					return nil
				}
				bf--
				bb++
				tx.BufferWrite([]*Mutation{
					Update("Accounts", []string{"AccountId", "Balance"}, []interface{}{int64(1), bf}),
					Update("Accounts", []string{"AccountId", "Balance"}, []interface{}{int64(2), bb}),
				})
				return nil
			})
			if err != nil {
				t.Fatalf("%d: failed to execute transaction: %v", iter, err)
			}
		}(i)
	}
	// Because of context timeout, all goroutines will eventually return.
	wg.Wait()
	_, err := client.ReadWriteTransaction(ctx, func(tx *ReadWriteTransaction) error {
		var bf, bb int64
		r, e := tx.ReadRow(ctx, "Accounts", Key{int64(1)}, []string{"Balance"})
		if e != nil {
			return e
		}
		if ce := r.Column(0, &bf); ce != nil {
			return ce
		}
		bb, e = readBalance(tx.ReadUsingIndex(ctx, "Accounts", "AccountByNickname", Keys(Key{"Bar"}), []string{"Balance"}))
		if e != nil {
			return e
		}
		if bf != 30 || bb != 21 {
			t.Errorf("Foo's balance is now %v and Bar's balance is now %v, want %v and %v", bf, bb, 30, 21)
		}
		return nil
	})
	if err != nil {
		t.Errorf("failed to check balances: %v", err)
	}
}

// Test client recovery on database recreation.
func TestDbRemovalRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := prepare(ctx, t); err != nil {
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	defer tearDown(ctx, t)

	// Drop the testing database.
	if err := admin.DropDatabase(ctx, &adminpb.DropDatabaseRequest{db}); err != nil {
		t.Fatalf("failed to drop testing database %v: %v", db, err)
	}

	// Now, send the query.
	iter := client.Single().Query(ctx, Statement{SQL: "SELECT SingerId FROM Singers"})
	defer iter.Stop()
	_, err := iter.Next()
	if err == nil {
		t.Errorf("client sends query to removed database successfully, want it to fail")
	}

	// Recreate database and table.
	op, err := admin.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%v/instances/%v", testProjectID, testInstanceID),
		CreateStatement: "CREATE DATABASE " + dbName,
		ExtraStatements: []string{
			`CREATE TABLE Singers (
				SingerId        INT64 NOT NULL,
				FirstName       STRING(1024),
				LastName        STRING(1024),
				SingerInfo      BYTES(MAX)
			) PRIMARY KEY (SingerId)`,
		},
	})
	if _, err := op.Wait(ctx); err != nil {
		t.Errorf("cannot recreate testing DB %v: %v", db, err)
	}

	// Now, send the query again.
	iter = client.Single().Query(ctx, Statement{SQL: "SELECT SingerId FROM Singers"})
	defer iter.Stop()
	_, err = iter.Next()
	if err != nil && err != iterator.Done {
		t.Fatalf("failed to send query to database %v: %v", db, err)
	}
}

// Test encoding/decoding non-struct Cloud Spanner types.
func TestBasicTypes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := prepare(ctx, t); err != nil {
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	defer tearDown(ctx, t)
	t1, _ := time.Parse(time.RFC3339Nano, "2016-11-15T15:04:05.999999999Z")
	// Boundaries
	t2, _ := time.Parse(time.RFC3339Nano, "0001-01-01T00:00:00.000000000Z")
	t3, _ := time.Parse(time.RFC3339Nano, "9999-12-31T23:59:59.999999999Z")
	d1, _ := civil.ParseDate("2016-11-15")
	// Boundaries
	d2, _ := civil.ParseDate("0001-01-01")
	d3, _ := civil.ParseDate("9999-12-31")

	tests := []struct {
		col  string
		val  interface{}
		want interface{}
	}{
		{col: "String", val: ""},
		{col: "String", val: "", want: NullString{"", true}},
		{col: "String", val: "foo"},
		{col: "String", val: "foo", want: NullString{"foo", true}},
		{col: "String", val: NullString{"bar", true}, want: "bar"},
		{col: "String", val: NullString{"bar", false}, want: NullString{"", false}},
		{col: "StringArray", val: []string(nil), want: []NullString(nil)},
		{col: "StringArray", val: []string{}, want: []NullString{}},
		{col: "StringArray", val: []string{"foo", "bar"}, want: []NullString{{"foo", true}, {"bar", true}}},
		{col: "StringArray", val: []NullString(nil)},
		{col: "StringArray", val: []NullString{}},
		{col: "StringArray", val: []NullString{{"foo", true}, {}}},
		{col: "Bytes", val: []byte{}},
		{col: "Bytes", val: []byte{1, 2, 3}},
		{col: "Bytes", val: []byte(nil)},
		{col: "BytesArray", val: [][]byte(nil)},
		{col: "BytesArray", val: [][]byte{}},
		{col: "BytesArray", val: [][]byte{[]byte{1}, []byte{2, 3}}},
		{col: "Int64a", val: 0, want: int64(0)},
		{col: "Int64a", val: -1, want: int64(-1)},
		{col: "Int64a", val: 2, want: int64(2)},
		{col: "Int64a", val: int64(3)},
		{col: "Int64a", val: 4, want: NullInt64{4, true}},
		{col: "Int64a", val: NullInt64{5, true}, want: int64(5)},
		{col: "Int64a", val: NullInt64{6, true}, want: int64(6)},
		{col: "Int64a", val: NullInt64{7, false}, want: NullInt64{0, false}},
		{col: "Int64Array", val: []int(nil), want: []NullInt64(nil)},
		{col: "Int64Array", val: []int{}, want: []NullInt64{}},
		{col: "Int64Array", val: []int{1, 2}, want: []NullInt64{{1, true}, {2, true}}},
		{col: "Int64Array", val: []int64(nil), want: []NullInt64(nil)},
		{col: "Int64Array", val: []int64{}, want: []NullInt64{}},
		{col: "Int64Array", val: []int64{1, 2}, want: []NullInt64{{1, true}, {2, true}}},
		{col: "Int64Array", val: []NullInt64(nil)},
		{col: "Int64Array", val: []NullInt64{}},
		{col: "Int64Array", val: []NullInt64{{1, true}, {}}},
		{col: "Bool", val: false},
		{col: "Bool", val: true},
		{col: "Bool", val: false, want: NullBool{false, true}},
		{col: "Bool", val: true, want: NullBool{true, true}},
		{col: "Bool", val: NullBool{true, true}},
		{col: "Bool", val: NullBool{false, false}},
		{col: "BoolArray", val: []bool(nil), want: []NullBool(nil)},
		{col: "BoolArray", val: []bool{}, want: []NullBool{}},
		{col: "BoolArray", val: []bool{true, false}, want: []NullBool{{true, true}, {false, true}}},
		{col: "BoolArray", val: []NullBool(nil)},
		{col: "BoolArray", val: []NullBool{}},
		{col: "BoolArray", val: []NullBool{{false, true}, {true, true}, {}}},
		{col: "Float64", val: 0.0},
		{col: "Float64", val: 3.14},
		{col: "Float64", val: math.NaN()},
		{col: "Float64", val: math.Inf(1)},
		{col: "Float64", val: math.Inf(-1)},
		{col: "Float64", val: 2.78, want: NullFloat64{2.78, true}},
		{col: "Float64", val: NullFloat64{2.71, true}, want: 2.71},
		{col: "Float64", val: NullFloat64{1.41, true}, want: NullFloat64{1.41, true}},
		{col: "Float64", val: NullFloat64{0, false}},
		{col: "Float64Array", val: []float64(nil), want: []NullFloat64(nil)},
		{col: "Float64Array", val: []float64{}, want: []NullFloat64{}},
		{col: "Float64Array", val: []float64{2.72, 3.14, math.Inf(1)}, want: []NullFloat64{{2.72, true}, {3.14, true}, {math.Inf(1), true}}},
		{col: "Float64Array", val: []NullFloat64(nil)},
		{col: "Float64Array", val: []NullFloat64{}},
		{col: "Float64Array", val: []NullFloat64{{2.72, true}, {math.Inf(1), true}, {}}},
		{col: "Date", val: d1},
		{col: "Date", val: d1, want: NullDate{d1, true}},
		{col: "Date", val: NullDate{d1, true}},
		{col: "Date", val: NullDate{d1, true}, want: d1},
		{col: "Date", val: NullDate{civil.Date{}, false}},
		{col: "DateArray", val: []civil.Date(nil), want: []NullDate(nil)},
		{col: "DateArray", val: []civil.Date{}, want: []NullDate{}},
		{col: "DateArray", val: []civil.Date{d1, d2, d3}, want: []NullDate{{d1, true}, {d2, true}, {d3, true}}},
		{col: "Timestamp", val: t1},
		{col: "Timestamp", val: t1, want: NullTime{t1, true}},
		{col: "Timestamp", val: NullTime{t1, true}},
		{col: "Timestamp", val: NullTime{t1, true}, want: t1},
		{col: "Timestamp", val: NullTime{}},
		{col: "TimestampArray", val: []time.Time(nil), want: []NullTime(nil)},
		{col: "TimestampArray", val: []time.Time{}, want: []NullTime{}},
		{col: "TimestampArray", val: []time.Time{t1, t2, t3}, want: []NullTime{{t1, true}, {t2, true}, {t3, true}}},
	}

	// Write rows into table first.
	var muts []*Mutation
	for i, test := range tests {
		muts = append(muts, InsertOrUpdate("Types", []string{"RowID", test.col}, []interface{}{i, test.val}))
	}
	if _, err := client.Apply(ctx, muts, ApplyAtLeastOnce()); err != nil {
		t.Fatal(err)
	}

	for i, test := range tests {
		row, err := client.Single().ReadRow(ctx, "Types", []interface{}{i}, []string{test.col})
		if err != nil {
			t.Fatalf("Unable to fetch row %v: %v", i, err)
		}
		// Create new instance of type of test.want.
		want := test.want
		if want == nil {
			want = test.val
		}
		gotp := reflect.New(reflect.TypeOf(want))
		if err := row.Column(0, gotp.Interface()); err != nil {
			t.Errorf("%d: col:%v val:%#v, %v", i, test.col, test.val, err)
			continue
		}
		got := reflect.Indirect(gotp).Interface()

		// One of the test cases is checking NaN handling.  Given
		// NaN!=NaN, we can't use reflect to test for it.
		isNaN := func(t interface{}) bool {
			f, ok := t.(float64)
			if !ok {
				return false
			}
			return math.IsNaN(f)
		}
		if isNaN(got) && isNaN(want) {
			continue
		}

		// Check non-NaN cases.
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: col:%v val:%#v, got %#v, want %#v", i, test.col, test.val, got, want)
			continue
		}
	}
}

// Test decoding Cloud Spanner STRUCT type.
func TestStructTypes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := prepare(ctx, t); err != nil {
		tearDown(ctx, t)
		t.Fatalf("cannot set up testing environment: %v", err)
	}
	defer tearDown(ctx, t)

	tests := []struct {
		q    Statement
		want func(r *Row) error
	}{
		{
			q: Statement{SQL: `SELECT ARRAY(SELECT STRUCT(1, 2))`},
			want: func(r *Row) error {
				// Test STRUCT ARRAY decoding to []NullRow.
				var rows []NullRow
				if err := r.Column(0, &rows); err != nil {
					return err
				}
				if len(rows) != 1 {
					return fmt.Errorf("len(rows) = %d; want 1", len(rows))
				}
				if !rows[0].Valid {
					return fmt.Errorf("rows[0] is NULL")
				}
				var i, j int64
				if err := rows[0].Row.Columns(&i, &j); err != nil {
					return err
				}
				if i != 1 || j != 2 {
					return fmt.Errorf("got (%d,%d), want (1,2)", i, j)
				}
				return nil
			},
		},
		{
			q: Statement{SQL: `SELECT ARRAY(SELECT STRUCT(1 as foo, 2 as bar)) as col1`},
			want: func(r *Row) error {
				// Test Row.ToStruct.
				s := struct {
					Col1 []*struct {
						Foo int64 `spanner:"foo"`
						Bar int64 `spanner:"bar"`
					} `spanner:"col1"`
				}{}
				if err := r.ToStruct(&s); err != nil {
					return err
				}
				want := struct {
					Col1 []*struct {
						Foo int64 `spanner:"foo"`
						Bar int64 `spanner:"bar"`
					} `spanner:"col1"`
				}{
					Col1: []*struct {
						Foo int64 `spanner:"foo"`
						Bar int64 `spanner:"bar"`
					}{
						{
							Foo: 1,
							Bar: 2,
						},
					},
				}
				if !reflect.DeepEqual(want, s) {
					return fmt.Errorf("unexpected decoding result: %v, want %v", s, want)
				}
				return nil
			},
		},
	}
	for i, test := range tests {
		iter := client.Single().Query(ctx, test.q)
		defer iter.Stop()
		row, err := iter.Next()
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		if err := test.want(row); err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
	}
}

func rowToValues(r *Row) ([]interface{}, error) {
	var x int64
	var y, z string
	if err := r.Column(0, &x); err != nil {
		return nil, err
	}
	if err := r.Column(1, &y); err != nil {
		return nil, err
	}
	if err := r.Column(2, &z); err != nil {
		return nil, err
	}
	return []interface{}{x, y, z}, nil
}

func readAll(iter *RowIterator) ([][]interface{}, error) {
	defer iter.Stop()
	var vals [][]interface{}
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			return vals, nil
		}
		if err != nil {
			return nil, err
		}
		v, err := rowToValues(row)
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
}
