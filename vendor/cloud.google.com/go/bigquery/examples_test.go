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

package bigquery_test

import (
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func ExampleNewClient() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	_ = client // TODO: Use client.
}

func ExampleClient_Dataset() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.Dataset("my_dataset")
	fmt.Println(ds)
}

func ExampleClient_DatasetInProject() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.DatasetInProject("their-project-id", "their-dataset")
	fmt.Println(ds)
}

func ExampleClient_Datasets() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Datasets(ctx)
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleClient_DatasetsInProject() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.DatasetsInProject(ctx, "their-project-id")
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func getJobID() string { return "" }

func ExampleClient_JobFromID() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	jobID := getJobID() // Get a job ID using Job.ID, the console or elsewhere.
	job, err := client.JobFromID(ctx, jobID)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(job)
}

func ExampleNewGCSReference() {
	gcsRef := bigquery.NewGCSReference("gs://my-bucket/my-object")
	fmt.Println(gcsRef)
}

func ExampleClient_Query() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	q := client.Query("select name, num from t1")
	q.DefaultProjectID = "project-id"
	// TODO: set other options on the Query.
	// TODO: Call Query.Run or Query.Read.
}

func ExampleClient_Query_parameters() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	q := client.Query("select num from t1 where name = @user")
	q.Parameters = []bigquery.QueryParameter{
		{Name: "user", Value: "Elizabeth"},
	}
	// TODO: set other options on the Query.
	// TODO: Call Query.Run or Query.Read.
}

func ExampleQuery_Read() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	q := client.Query("select name, num from t1")
	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleRowIterator_Next() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	q := client.Query("select name, num from t1")
	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(row)
	}
}

func ExampleRowIterator_Next_struct() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	type score struct {
		Name string
		Num  int
	}

	q := client.Query("select name, num from t1")
	it, err := q.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	for {
		var s score
		err := it.Next(&s)
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(s)
	}
}

func ExampleJob_Read() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	q := client.Query("select name, num from t1")
	// Call Query.Run to get a Job, then call Read on the job.
	// Note: Query.Read is a shorthand for this.
	job, err := q.Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	it, err := job.Read(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleJob_Wait() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.Dataset("my_dataset")
	job, err := ds.Table("t1").CopierFrom(ds.Table("t2")).Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	status, err := job.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if status.Err() != nil {
		// TODO: Handle error.
	}
}

func ExampleDataset_Create() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Dataset("my_dataset").Create(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleDataset_Delete() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Dataset("my_dataset").Delete(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleDataset_Metadata() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	md, err := client.Dataset("my_dataset").Metadata(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(md)
}

func ExampleDataset_Table() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	// Table creates a reference to the table. It does not create the actual
	// table in BigQuery; to do so, use Table.Create.
	t := client.Dataset("my_dataset").Table("my_table")
	fmt.Println(t)
}

func ExampleDataset_Tables() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Dataset("my_dataset").Tables(ctx)
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleDatasetIterator_Next() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Datasets(ctx)
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(ds)
	}
}

func ExampleInferSchema() {
	type Item struct {
		Name  string
		Size  float64
		Count int
	}
	schema, err := bigquery.InferSchema(Item{})
	if err != nil {
		fmt.Println(err)
		// TODO: Handle error.
	}
	for _, fs := range schema {
		fmt.Println(fs.Name, fs.Type)
	}
	// Output:
	// Name STRING
	// Size FLOAT
	// Count INTEGER
}

func ExampleInferSchema_tags() {
	type Item struct {
		Name   string
		Size   float64
		Count  int    `bigquery:"number"`
		Secret []byte `bigquery:"-"`
	}
	schema, err := bigquery.InferSchema(Item{})
	if err != nil {
		fmt.Println(err)
		// TODO: Handle error.
	}
	for _, fs := range schema {
		fmt.Println(fs.Name, fs.Type)
	}
	// Output:
	// Name STRING
	// Size FLOAT
	// number INTEGER
}

func ExampleTable_Create() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	t := client.Dataset("my_dataset").Table("new-table")
	if err := t.Create(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Create_schema() {
	ctx := context.Background()
	// Infer table schema from a Go type.
	schema, err := bigquery.InferSchema(Item{})
	if err != nil {
		// TODO: Handle error.
	}
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	t := client.Dataset("my_dataset").Table("new-table")
	if err := t.Create(ctx, schema); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Delete() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Dataset("my_dataset").Table("my_table").Delete(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Metadata() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	md, err := client.Dataset("my_dataset").Table("my_table").Metadata(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(md)
}

func ExampleTable_Uploader() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my_dataset").Table("my_table").Uploader()
	_ = u // TODO: Use u.
}

func ExampleTable_Uploader_options() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my_dataset").Table("my_table").Uploader()
	u.SkipInvalidRows = true
	u.IgnoreUnknownValues = true
	_ = u // TODO: Use u.
}

func ExampleTable_CopierFrom() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.Dataset("my_dataset")
	c := ds.Table("combined").CopierFrom(ds.Table("t1"), ds.Table("t2"))
	c.WriteDisposition = bigquery.WriteTruncate
	// TODO: set other options on the Copier.
	job, err := c.Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	status, err := job.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if status.Err() != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_ExtractorTo() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	gcsRef := bigquery.NewGCSReference("gs://my-bucket/my-object")
	gcsRef.FieldDelimiter = ":"
	// TODO: set other options on the GCSReference.
	ds := client.Dataset("my_dataset")
	extractor := ds.Table("my_table").ExtractorTo(gcsRef)
	extractor.DisableHeader = true
	// TODO: set other options on the Extractor.
	job, err := extractor.Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	status, err := job.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if status.Err() != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_LoaderFrom() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	gcsRef := bigquery.NewGCSReference("gs://my-bucket/my-object")
	gcsRef.AllowJaggedRows = true
	// TODO: set other options on the GCSReference.
	ds := client.Dataset("my_dataset")
	loader := ds.Table("my_table").LoaderFrom(gcsRef)
	loader.CreateDisposition = bigquery.CreateNever
	// TODO: set other options on the Loader.
	job, err := loader.Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	status, err := job.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if status.Err() != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_LoaderFrom_reader() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	f, err := os.Open("data.csv")
	if err != nil {
		// TODO: Handle error.
	}
	rs := bigquery.NewReaderSource(f)
	rs.AllowJaggedRows = true
	// TODO: set other options on the GCSReference.
	ds := client.Dataset("my_dataset")
	loader := ds.Table("my_table").LoaderFrom(rs)
	loader.CreateDisposition = bigquery.CreateNever
	// TODO: set other options on the Loader.
	job, err := loader.Run(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	status, err := job.Wait(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	if status.Err() != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Read() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Dataset("my_dataset").Table("my_table").Read(ctx)
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleTable_Update() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	t := client.Dataset("my_dataset").Table("my_table")
	tm, err := t.Update(ctx, bigquery.TableMetadataToUpdate{
		Description: "my favorite table",
	})
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(tm)
}

func ExampleTableIterator_Next() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Dataset("my_dataset").Tables(ctx)
	for {
		t, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(t)
	}
}

type Item struct {
	Name  string
	Size  float64
	Count int
}

// Save implements the ValueSaver interface.
func (i *Item) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"Name":  i.Name,
		"Size":  i.Size,
		"Count": i.Count,
	}, "", nil
}

func ExampleUploader_Put() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my_dataset").Table("my_table").Uploader()
	// Item implements the ValueSaver interface.
	items := []*Item{
		{Name: "n1", Size: 32.6, Count: 7},
		{Name: "n2", Size: 4, Count: 2},
		{Name: "n3", Size: 101.5, Count: 1},
	}
	if err := u.Put(ctx, items); err != nil {
		// TODO: Handle error.
	}
}

var schema bigquery.Schema

func ExampleUploader_Put_structSaver() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my_dataset").Table("my_table").Uploader()

	type score struct {
		Name string
		Num  int
	}

	// Assume schema holds the table's schema.
	savers := []*bigquery.StructSaver{
		{Struct: score{Name: "n1", Num: 12}, Schema: schema, InsertID: "id1"},
		{Struct: score{Name: "n2", Num: 31}, Schema: schema, InsertID: "id2"},
		{Struct: score{Name: "n3", Num: 7}, Schema: schema, InsertID: "id3"},
	}
	if err := u.Put(ctx, savers); err != nil {
		// TODO: Handle error.
	}
}

func ExampleUploader_Put_struct() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my_dataset").Table("my_table").Uploader()

	type score struct {
		Name string
		Num  int
	}
	scores := []score{
		{Name: "n1", Num: 12},
		{Name: "n2", Num: 31},
		{Name: "n3", Num: 7},
	}
	// Schema is inferred from the score type.
	if err := u.Put(ctx, scores); err != nil {
		// TODO: Handle error.
	}
}
