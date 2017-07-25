// Copyright 2015 Google Inc. All Rights Reserved.
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

package bigquery

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/internal/pretty"
	"cloud.google.com/go/internal/testutil"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	client  *Client
	dataset *Dataset
	schema  = Schema{
		{Name: "name", Type: StringFieldType},
		{Name: "num", Type: IntegerFieldType},
	}
	fiveMinutesFromNow time.Time
)

func TestMain(m *testing.M) {
	initIntegrationTest()
	os.Exit(m.Run())
}

func getClient(t *testing.T) *Client {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	return client
}

// If integration tests will be run, create a unique bucket for them.
func initIntegrationTest() {
	flag.Parse() // needed for testing.Short()
	if testing.Short() {
		return
	}
	ctx := context.Background()
	ts := testutil.TokenSource(ctx, Scope)
	if ts == nil {
		log.Println("Integration tests skipped. See CONTRIBUTING.md for details")
		return
	}
	projID := testutil.ProjID()
	var err error
	client, err = NewClient(ctx, projID, option.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	dataset = client.Dataset("bigquery_integration_test")
	if err := dataset.Create(ctx); err != nil && !hasStatusCode(err, http.StatusConflict) { // AlreadyExists is 409
		log.Fatalf("creating dataset: %v", err)
	}
}

func TestIntegration_Create(t *testing.T) {
	// Check that creating a record field with an empty schema is an error.
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	table := dataset.Table("t_bad")
	schema := Schema{
		{Name: "rec", Type: RecordFieldType, Schema: Schema{}},
	}
	err := table.Create(context.Background(), schema, TableExpiration(time.Now().Add(5*time.Minute)))
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !hasStatusCode(err, http.StatusBadRequest) {
		t.Fatalf("want a 400 error, got %v", err)
	}
}

func TestIntegration_CreateView(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Test that standard SQL views work.
	view := dataset.Table("t_view_standardsql")
	query := ViewQuery(fmt.Sprintf("SELECT APPROX_COUNT_DISTINCT(name) FROM `%s.%s.%s`", dataset.ProjectID, dataset.DatasetID, table.TableID))
	err := view.Create(context.Background(), UseStandardSQL(), query)
	if err != nil {
		t.Fatalf("table.create: Did not expect an error, got: %v", err)
	}
	view.Delete(ctx)
}

func TestIntegration_TableMetadata(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)
	// Check table metadata.
	md, err := table.Metadata(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(jba): check md more thorougly.
	if got, want := md.ID, fmt.Sprintf("%s:%s.%s", dataset.ProjectID, dataset.DatasetID, table.TableID); got != want {
		t.Errorf("metadata.ID: got %q, want %q", got, want)
	}
	if got, want := md.Type, RegularTable; got != want {
		t.Errorf("metadata.Type: got %v, want %v", got, want)
	}
	if got, want := md.ExpirationTime, fiveMinutesFromNow; !got.Equal(want) {
		t.Errorf("metadata.Type: got %v, want %v", got, want)
	}

	// Check that timePartitioning is nil by default
	if md.TimePartitioning != nil {
		t.Errorf("metadata.TimePartitioning: got %v, want %v", md.TimePartitioning, nil)
	}

	// Create tables that have time partitioning
	partitionCases := []struct {
		timePartitioning   TimePartitioning
		expectedExpiration time.Duration
	}{
		{TimePartitioning{}, time.Duration(0)},
		{TimePartitioning{time.Second}, time.Second},
	}
	for i, c := range partitionCases {
		table := dataset.Table(fmt.Sprintf("t_metadata_partition_%v", i))
		err = table.Create(context.Background(), schema, c.timePartitioning, TableExpiration(time.Now().Add(5*time.Minute)))
		if err != nil {
			t.Fatal(err)
		}
		defer table.Delete(ctx)
		md, err = table.Metadata(ctx)
		if err != nil {
			t.Fatal(err)
		}

		got := md.TimePartitioning
		want := &TimePartitioning{c.expectedExpiration}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("metadata.TimePartitioning: got %v, want %v", got, want)
		}
	}
}

func TestIntegration_DatasetMetadata(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	md, err := dataset.Metadata(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := md.ID, fmt.Sprintf("%s:%s", dataset.ProjectID, dataset.DatasetID); got != want {
		t.Errorf("ID: got %q, want %q", got, want)
	}
	jan2016 := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	if md.CreationTime.Before(jan2016) {
		t.Errorf("CreationTime: got %s, want > 2016-1-1", md.CreationTime)
	}
	if md.LastModifiedTime.Before(jan2016) {
		t.Errorf("LastModifiedTime: got %s, want > 2016-1-1", md.LastModifiedTime)
	}

	// Verify that we get a NotFound for a nonexistent dataset.
	_, err = client.Dataset("does_not_exist").Metadata(ctx)
	if err == nil || !hasStatusCode(err, http.StatusNotFound) {
		t.Errorf("got %v, want NotFound error", err)
	}
}

func TestIntegration_DatasetDelete(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	ds := client.Dataset("delete_test")
	if err := ds.Create(ctx); err != nil && !hasStatusCode(err, http.StatusConflict) { // AlreadyExists is 409
		t.Fatalf("creating dataset %s: %v", ds, err)
	}
	if err := ds.Delete(ctx); err != nil {
		t.Fatalf("deleting dataset %s: %v", ds, err)
	}
}

func TestIntegration_Tables(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Iterate over tables in the dataset.
	it := dataset.Tables(ctx)
	var tables []*Table
	for {
		tbl, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		tables = append(tables, tbl)
	}
	// Other tests may be running with this dataset, so there might be more
	// than just our table in the list. So don't try for an exact match; just
	// make sure that our table is there somewhere.
	found := false
	for _, tbl := range tables {
		if reflect.DeepEqual(tbl, table) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Tables: got %v\nshould see %v in the list", pretty.Value(tables), pretty.Value(table))
	}
}

func TestIntegration_UploadAndRead(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Populate the table.
	upl := table.Uploader()
	var (
		wantRows  [][]Value
		saverRows []*ValuesSaver
	)
	for i, name := range []string{"a", "b", "c"} {
		row := []Value{name, int64(i)}
		wantRows = append(wantRows, row)
		saverRows = append(saverRows, &ValuesSaver{
			Schema:   schema,
			InsertID: name,
			Row:      row,
		})
	}
	if err := upl.Put(ctx, saverRows); err != nil {
		t.Fatal(putError(err))
	}

	// Wait until the data has been uploaded. This can take a few seconds, according
	// to https://cloud.google.com/bigquery/streaming-data-into-bigquery.
	if err := waitForRow(ctx, table); err != nil {
		t.Fatal(err)
	}

	// Read the table.
	checkRead(t, "upload", table.Read(ctx), wantRows)

	// Query the table.
	q := client.Query(fmt.Sprintf("select name, num from %s", table.TableID))
	q.DefaultProjectID = dataset.ProjectID
	q.DefaultDatasetID = dataset.DatasetID

	rit, err := q.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkRead(t, "query", rit, wantRows)

	// Query the long way.
	job1, err := q.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	job2, err := client.JobFromID(ctx, job1.ID())
	if err != nil {
		t.Fatal(err)
	}
	rit, err = job2.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	checkRead(t, "job.Read", rit, wantRows)

	// Test reading directly into a []Value.
	valueLists, err := readAll(table.Read(ctx))
	if err != nil {
		t.Fatal(err)
	}
	it := table.Read(ctx)
	for i, vl := range valueLists {
		var got []Value
		if err := it.Next(&got); err != nil {
			t.Fatal(err)
		}
		want := []Value(vl)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v, want %v", i, got, want)
		}
	}

	// Test reading into a map.
	it = table.Read(ctx)
	for _, vl := range valueLists {
		var vm map[string]Value
		if err := it.Next(&vm); err != nil {
			t.Fatal(err)
		}
		if got, want := len(vm), len(vl); got != want {
			t.Fatalf("valueMap len: got %d, want %d", got, want)
		}
		for i, v := range vl {
			if got, want := vm[schema[i].Name], v; got != want {
				t.Errorf("%d, name=%s: got %v, want %v",
					i, schema[i].Name, got, want)
			}
		}
	}
}

type TestStruct struct {
	Name string
	Nums []int
	Sub  Sub
	Subs []*Sub
}

type Sub struct {
	B       bool
	SubSub  SubSub
	SubSubs []*SubSub
}

type SubSub struct{ Count int }

func TestIntegration_UploadAndReadStructs(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	schema, err := InferSchema(TestStruct{})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Populate the table.
	upl := table.Uploader()
	want := []*TestStruct{
		{Name: "a", Nums: []int{1, 2}, Sub: Sub{B: true}, Subs: []*Sub{{B: false}, {B: true}}},
		{Name: "b", Nums: []int{1}, Subs: []*Sub{{B: false}, {B: false}, {B: true}}},
		{Name: "c", Sub: Sub{B: true}},
		{
			Name: "d",
			Sub:  Sub{SubSub: SubSub{12}, SubSubs: []*SubSub{{1}, {2}, {3}}},
			Subs: []*Sub{{B: false, SubSub: SubSub{4}}, {B: true, SubSubs: []*SubSub{{5}, {6}}}},
		},
	}
	var savers []*StructSaver
	for _, s := range want {
		savers = append(savers, &StructSaver{Schema: schema, Struct: s})
	}
	if err := upl.Put(ctx, savers); err != nil {
		t.Fatal(putError(err))
	}

	// Wait until the data has been uploaded. This can take a few seconds, according
	// to https://cloud.google.com/bigquery/streaming-data-into-bigquery.
	if err := waitForRow(ctx, table); err != nil {
		t.Fatal(err)
	}

	// Test iteration with structs.
	it := table.Read(ctx)
	var got []*TestStruct
	for {
		var g TestStruct
		err := it.Next(&g)
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, &g)
	}
	sort.Sort(byName(got))

	// BigQuery does not elide nils. It reports an error for nil fields.
	for i, g := range got {
		if i >= len(want) {
			t.Errorf("%d: got %v, past end of want", i, pretty.Value(g))
		} else if w := want[i]; !reflect.DeepEqual(g, w) {
			t.Errorf("%d: got %v, want %v", i, pretty.Value(g), pretty.Value(w))
		}
	}
}

type byName []*TestStruct

func (b byName) Len() int           { return len(b) }
func (b byName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byName) Less(i, j int) bool { return b[i].Name < b[j].Name }

func TestIntegration_Update(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Test Update of non-schema fields.
	tm, err := table.Metadata(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantDescription := tm.Description + "more"
	wantName := tm.Name + "more"
	got, err := table.Update(ctx, TableMetadataToUpdate{
		Description: wantDescription,
		Name:        wantName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Description != wantDescription {
		t.Errorf("Description: got %q, want %q", got.Description, wantDescription)
	}
	if got.Name != wantName {
		t.Errorf("Name: got %q, want %q", got.Name, wantName)
	}
	if !reflect.DeepEqual(got.Schema, schema) {
		t.Errorf("Schema: got %v, want %v", pretty.Value(got.Schema), pretty.Value(schema))
	}

	// Test schema update.
	// Columns can be added. schema2 is the same as schema, except for the
	// added column in the middle.
	nested := Schema{
		{Name: "nested", Type: BooleanFieldType},
		{Name: "other", Type: StringFieldType},
	}
	schema2 := Schema{
		schema[0],
		{Name: "rec", Type: RecordFieldType, Schema: nested},
		schema[1],
	}

	got, err = table.Update(ctx, TableMetadataToUpdate{Schema: schema2})
	if err != nil {
		t.Fatal(err)
	}

	// Wherever you add the column, it appears at the end.
	schema3 := Schema{schema2[0], schema2[2], schema2[1]}
	if !reflect.DeepEqual(got.Schema, schema3) {
		t.Errorf("add field:\ngot  %v\nwant %v",
			pretty.Value(got.Schema), pretty.Value(schema3))
	}

	// Updating with the empty schema succeeds, but is a no-op.
	got, err = table.Update(ctx, TableMetadataToUpdate{Schema: Schema{}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Schema, schema3) {
		t.Errorf("empty schema:\ngot  %v\nwant %v",
			pretty.Value(got.Schema), pretty.Value(schema3))
	}

	// Error cases.
	for _, test := range []struct {
		desc   string
		fields []*FieldSchema
	}{
		{"change from optional to required", []*FieldSchema{
			schema3[0],
			{Name: "num", Type: IntegerFieldType, Required: true},
			schema3[2],
		}},
		{"add a required field", []*FieldSchema{
			schema3[0], schema3[1], schema3[2],
			{Name: "req", Type: StringFieldType, Required: true},
		}},
		{"remove a field", []*FieldSchema{schema3[0], schema3[1]}},
		{"remove a nested field", []*FieldSchema{
			schema3[0], schema3[1],
			{Name: "rec", Type: RecordFieldType, Schema: Schema{nested[0]}}}},
		{"remove all nested fields", []*FieldSchema{
			schema3[0], schema3[1],
			{Name: "rec", Type: RecordFieldType, Schema: Schema{}}}},
	} {
		for {
			_, err = table.Update(ctx, TableMetadataToUpdate{Schema: Schema(test.fields)})
			if !hasStatusCode(err, 403) {
				break
			}
			// We've hit the rate limit for updates. Wait a bit and retry.
			t.Logf("%s: retrying after getting %v", test.desc, err)
			time.Sleep(4 * time.Second)
		}
		if err == nil {
			t.Errorf("%s: want error, got nil", test.desc)
		} else if !hasStatusCode(err, 400) {
			t.Errorf("%s: want 400, got %v", test.desc, err)
		}
	}
}

func TestIntegration_Load(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Load the table from a reader.
	r := strings.NewReader("a,0\nb,1\nc,2\n")
	wantRows := [][]Value{
		[]Value{"a", int64(0)},
		[]Value{"b", int64(1)},
		[]Value{"c", int64(2)},
	}
	rs := NewReaderSource(r)
	loader := table.LoaderFrom(rs)
	loader.WriteDisposition = WriteTruncate
	job, err := loader.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := wait(ctx, job); err != nil {
		t.Fatal(err)
	}
	checkRead(t, "reader load", table.Read(ctx), wantRows)
}

func TestIntegration_DML(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	table := newTable(t, schema)
	defer table.Delete(ctx)

	// Use DML to insert.
	wantRows := [][]Value{
		[]Value{"a", int64(0)},
		[]Value{"b", int64(1)},
		[]Value{"c", int64(2)},
	}
	query := fmt.Sprintf("INSERT bigquery_integration_test.%s (name, num) "+
		"VALUES ('a', 0), ('b', 1), ('c', 2)",
		table.TableID)
	q := client.Query(query)
	q.UseStandardSQL = true // necessary for DML
	job, err := q.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := wait(ctx, job); err != nil {
		t.Fatal(err)
	}
	checkRead(t, "INSERT", table.Read(ctx), wantRows)
}

func TestIntegration_TimeTypes(t *testing.T) {
	if client == nil {
		t.Skip("Integration tests skipped")
	}
	ctx := context.Background()
	dtSchema := Schema{
		{Name: "d", Type: DateFieldType},
		{Name: "t", Type: TimeFieldType},
		{Name: "dt", Type: DateTimeFieldType},
		{Name: "ts", Type: TimestampFieldType},
	}
	table := newTable(t, dtSchema)
	defer table.Delete(ctx)

	d := civil.Date{2016, 3, 20}
	tm := civil.Time{12, 30, 0, 0}
	ts := time.Date(2016, 3, 20, 15, 04, 05, 0, time.UTC)
	wantRows := [][]Value{
		[]Value{d, tm, civil.DateTime{d, tm}, ts},
	}
	upl := table.Uploader()
	if err := upl.Put(ctx, []*ValuesSaver{
		{Schema: dtSchema, Row: wantRows[0]},
	}); err != nil {
		t.Fatal(putError(err))
	}
	if err := waitForRow(ctx, table); err != nil {
		t.Fatal(err)
	}

	// SQL wants DATETIMEs with a space between date and time, but the service
	// returns them in RFC3339 form, with a "T" between.
	query := fmt.Sprintf("INSERT bigquery_integration_test.%s (d, t, dt, ts) "+
		"VALUES ('%s', '%s', '%s %s', '%s')",
		table.TableID, d, tm, d, tm, ts.Format("2006-01-02 15:04:05"))
	q := client.Query(query)
	q.UseStandardSQL = true // necessary for DML
	job, err := q.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := wait(ctx, job); err != nil {
		t.Fatal(err)
	}
	wantRows = append(wantRows, wantRows[0])
	checkRead(t, "TimeTypes", table.Read(ctx), wantRows)
}

// Creates a new, temporary table with a unique name and the given schema.
func newTable(t *testing.T, s Schema) *Table {
	fiveMinutesFromNow = time.Now().Add(5 * time.Minute).Round(time.Second)
	name := fmt.Sprintf("t%d", time.Now().UnixNano())
	table := dataset.Table(name)
	err := table.Create(context.Background(), s, TableExpiration(fiveMinutesFromNow))
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func checkRead(t *testing.T, msg string, it *RowIterator, want [][]Value) {
	got, err := readAll(it)
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
	if len(got) != len(want) {
		t.Errorf("%s: got %d rows, want %d", msg, len(got), len(want))
	}
	sort.Sort(byCol0(got))
	for i, r := range got {
		gotRow := []Value(r)
		wantRow := want[i]
		if !reflect.DeepEqual(gotRow, wantRow) {
			t.Errorf("%s #%d: got %v, want %v", msg, i, gotRow, wantRow)
		}
	}
}

func readAll(it *RowIterator) ([][]Value, error) {
	var rows [][]Value
	for {
		var vals []Value
		err := it.Next(&vals)
		if err == iterator.Done {
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, vals)
	}
}

type byCol0 [][]Value

func (b byCol0) Len() int      { return len(b) }
func (b byCol0) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byCol0) Less(i, j int) bool {
	switch a := b[i][0].(type) {
	case string:
		return a < b[j][0].(string)
	case civil.Date:
		return a.Before(b[j][0].(civil.Date))
	default:
		panic("unknown type")
	}
}

func hasStatusCode(err error, code int) bool {
	if e, ok := err.(*googleapi.Error); ok && e.Code == code {
		return true
	}
	return false
}

// wait polls the job until it is complete or an error is returned.
func wait(ctx context.Context, job *Job) error {
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("getting job status: %v", err)
	}
	if status.Err() != nil {
		return fmt.Errorf("job status error: %#v", status.Err())
	}
	return nil
}

// waitForRow polls the table until it contains a row.
// TODO(jba): use internal.Retry.
func waitForRow(ctx context.Context, table *Table) error {
	for {
		it := table.Read(ctx)
		var v []Value
		err := it.Next(&v)
		if err == nil {
			return nil
		}
		if err != iterator.Done {
			return err
		}
		time.Sleep(1 * time.Second)
	}
}

func putError(err error) string {
	pme, ok := err.(PutMultiError)
	if !ok {
		return err.Error()
	}
	var msgs []string
	for _, err := range pme {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}
