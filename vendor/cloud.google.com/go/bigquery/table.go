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
	"fmt"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/internal/optional"
	bq "google.golang.org/api/bigquery/v2"
)

// A Table is a reference to a BigQuery table.
type Table struct {
	// ProjectID, DatasetID and TableID may be omitted if the Table is the destination for a query.
	// In this case the result will be stored in an ephemeral table.
	ProjectID string
	DatasetID string
	// TableID must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_).
	// The maximum length is 1,024 characters.
	TableID string

	c *Client
}

// TableMetadata contains information about a BigQuery table.
type TableMetadata struct {
	Description string // The user-friendly description of this table.
	Name        string // The user-friendly name for this table.
	Schema      Schema
	View        string

	ID   string // An opaque ID uniquely identifying the table.
	Type TableType

	// The time when this table expires. If not set, the table will persist
	// indefinitely. Expired tables will be deleted and their storage reclaimed.
	ExpirationTime time.Time

	CreationTime     time.Time
	LastModifiedTime time.Time

	// The size of the table in bytes.
	// This does not include data that is being buffered during a streaming insert.
	NumBytes int64

	// The number of rows of data in this table.
	// This does not include data that is being buffered during a streaming insert.
	NumRows uint64

	// The time-based partitioning settings for this table.
	TimePartitioning *TimePartitioning
}

// TableCreateDisposition specifies the circumstances under which destination table will be created.
// Default is CreateIfNeeded.
type TableCreateDisposition string

const (
	// CreateIfNeeded will create the table if it does not already exist.
	// Tables are created atomically on successful completion of a job.
	CreateIfNeeded TableCreateDisposition = "CREATE_IF_NEEDED"

	// CreateNever ensures the table must already exist and will not be
	// automatically created.
	CreateNever TableCreateDisposition = "CREATE_NEVER"
)

// TableWriteDisposition specifies how existing data in a destination table is treated.
// Default is WriteAppend.
type TableWriteDisposition string

const (
	// WriteAppend will append to any existing data in the destination table.
	// Data is appended atomically on successful completion of a job.
	WriteAppend TableWriteDisposition = "WRITE_APPEND"

	// WriteTruncate overrides the existing data in the destination table.
	// Data is overwritten atomically on successful completion of a job.
	WriteTruncate TableWriteDisposition = "WRITE_TRUNCATE"

	// WriteEmpty fails writes if the destination table already contains data.
	WriteEmpty TableWriteDisposition = "WRITE_EMPTY"
)

// TableType is the type of table.
type TableType string

const (
	RegularTable TableType = "TABLE"
	ViewTable    TableType = "VIEW"
)

func (t *Table) tableRefProto() *bq.TableReference {
	return &bq.TableReference{
		ProjectId: t.ProjectID,
		DatasetId: t.DatasetID,
		TableId:   t.TableID,
	}
}

// FullyQualifiedName returns the ID of the table in projectID:datasetID.tableID format.
func (t *Table) FullyQualifiedName() string {
	return fmt.Sprintf("%s:%s.%s", t.ProjectID, t.DatasetID, t.TableID)
}

// implicitTable reports whether Table is an empty placeholder, which signifies that a new table should be created with an auto-generated Table ID.
func (t *Table) implicitTable() bool {
	return t.ProjectID == "" && t.DatasetID == "" && t.TableID == ""
}

// Create creates a table in the BigQuery service.
func (t *Table) Create(ctx context.Context, options ...CreateTableOption) error {
	conf := &createTableConf{
		projectID: t.ProjectID,
		datasetID: t.DatasetID,
		tableID:   t.TableID,
	}
	for _, o := range options {
		o.customizeCreateTable(conf)
	}
	return t.c.service.createTable(ctx, conf)
}

// Metadata fetches the metadata for the table.
func (t *Table) Metadata(ctx context.Context) (*TableMetadata, error) {
	return t.c.service.getTableMetadata(ctx, t.ProjectID, t.DatasetID, t.TableID)
}

// Delete deletes the table.
func (t *Table) Delete(ctx context.Context) error {
	return t.c.service.deleteTable(ctx, t.ProjectID, t.DatasetID, t.TableID)
}

// A CreateTableOption is an optional argument to CreateTable.
type CreateTableOption interface {
	customizeCreateTable(*createTableConf)
}

type tableExpiration time.Time

// TableExpiration returns a CreateTableOption that will cause the created table to be deleted after the expiration time.
func TableExpiration(exp time.Time) CreateTableOption { return tableExpiration(exp) }

func (opt tableExpiration) customizeCreateTable(conf *createTableConf) {
	conf.expiration = time.Time(opt)
}

type viewQuery string

// ViewQuery returns a CreateTableOption that causes the created table to be a virtual table defined by the supplied query.
// For more information see: https://cloud.google.com/bigquery/querying-data#views
func ViewQuery(query string) CreateTableOption { return viewQuery(query) }

func (opt viewQuery) customizeCreateTable(conf *createTableConf) {
	conf.viewQuery = string(opt)
}

type useStandardSQL struct{}

// UseStandardSQL returns a CreateTableOption to set the table to use standard SQL.
// The default setting is false (using legacy SQL).
func UseStandardSQL() CreateTableOption { return useStandardSQL{} }

func (opt useStandardSQL) customizeCreateTable(conf *createTableConf) {
	conf.useStandardSQL = true
}

// TimePartitioning is a CreateTableOption that can be used to set time-based
// date partitioning on a table.
// For more information see: https://cloud.google.com/bigquery/docs/creating-partitioned-tables
type TimePartitioning struct {
	// (Optional) The amount of time to keep the storage for a partition.
	// If the duration is empty (0), the data in the partitions do not expire.
	Expiration time.Duration
}

func (opt TimePartitioning) customizeCreateTable(conf *createTableConf) {
	conf.timePartitioning = &opt
}

// Update modifies specific Table metadata fields.
func (t *Table) Update(ctx context.Context, tm TableMetadataToUpdate) (*TableMetadata, error) {
	var conf patchTableConf
	if tm.Description != nil {
		s := optional.ToString(tm.Description)
		conf.Description = &s
	}
	if tm.Name != nil {
		s := optional.ToString(tm.Name)
		conf.Name = &s
	}
	conf.Schema = tm.Schema
	return t.c.service.patchTable(ctx, t.ProjectID, t.DatasetID, t.TableID, &conf)
}

// TableMetadataToUpdate is used when updating a table's metadata.
// Only non-nil fields will be updated.
type TableMetadataToUpdate struct {
	// Description is the user-friendly description of this table.
	Description optional.String

	// Name is the user-friendly name for this table.
	Name optional.String

	// Schema is the table's schema.
	// When updating a schema, you can add columns but not remove them.
	Schema Schema
	// TODO(jba): support updating the view
}
