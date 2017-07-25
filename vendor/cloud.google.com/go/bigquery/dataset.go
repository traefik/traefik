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
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// Dataset is a reference to a BigQuery dataset.
type Dataset struct {
	ProjectID string
	DatasetID string
	c         *Client
}

type DatasetMetadata struct {
	CreationTime           time.Time
	LastModifiedTime       time.Time // When the dataset or any of its tables were modified.
	DefaultTableExpiration time.Duration
	Description            string // The user-friendly description of this table.
	Name                   string // The user-friendly name for this table.
	ID                     string
	Location               string            // The geo location of the dataset.
	Labels                 map[string]string // User-provided labels.
	// TODO(jba): access rules
}

// Dataset creates a handle to a BigQuery dataset in the client's project.
func (c *Client) Dataset(id string) *Dataset {
	return c.DatasetInProject(c.projectID, id)
}

// DatasetInProject creates a handle to a BigQuery dataset in the specified project.
func (c *Client) DatasetInProject(projectID, datasetID string) *Dataset {
	return &Dataset{
		ProjectID: projectID,
		DatasetID: datasetID,
		c:         c,
	}
}

// Create creates a dataset in the BigQuery service. An error will be returned
// if the dataset already exists.
func (d *Dataset) Create(ctx context.Context) error {
	return d.c.service.insertDataset(ctx, d.DatasetID, d.ProjectID)
}

// Delete deletes the dataset.
func (d *Dataset) Delete(ctx context.Context) error {
	return d.c.service.deleteDataset(ctx, d.DatasetID, d.ProjectID)
}

// Metadata fetches the metadata for the dataset.
func (d *Dataset) Metadata(ctx context.Context) (*DatasetMetadata, error) {
	return d.c.service.getDatasetMetadata(ctx, d.ProjectID, d.DatasetID)
}

// Table creates a handle to a BigQuery table in the dataset.
// To determine if a table exists, call Table.Metadata.
// If the table does not already exist, use Table.Create to create it.
func (d *Dataset) Table(tableID string) *Table {
	return &Table{ProjectID: d.ProjectID, DatasetID: d.DatasetID, TableID: tableID, c: d.c}
}

// Tables returns an iterator over the tables in the Dataset.
func (d *Dataset) Tables(ctx context.Context) *TableIterator {
	it := &TableIterator{
		ctx:     ctx,
		dataset: d,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.tables) },
		func() interface{} { b := it.tables; it.tables = nil; return b })
	return it
}

// A TableIterator is an iterator over Tables.
type TableIterator struct {
	ctx      context.Context
	dataset  *Dataset
	tables   []*Table
	pageInfo *iterator.PageInfo
	nextFunc func() error
}

// Next returns the next result. Its second return value is Done if there are
// no more results. Once Next returns Done, all subsequent calls will return
// Done.
func (it *TableIterator) Next() (*Table, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	t := it.tables[0]
	it.tables = it.tables[1:]
	return t, nil
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *TableIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

func (it *TableIterator) fetch(pageSize int, pageToken string) (string, error) {
	tables, tok, err := it.dataset.c.service.listTables(it.ctx, it.dataset.ProjectID, it.dataset.DatasetID, pageSize, pageToken)
	if err != nil {
		return "", err
	}
	for _, t := range tables {
		t.c = it.dataset.c
		it.tables = append(it.tables, t)
	}
	return tok, nil
}

// Datasets returns an iterator over the datasets in the Client's project.
func (c *Client) Datasets(ctx context.Context) *DatasetIterator {
	return c.DatasetsInProject(ctx, c.projectID)
}

// DatasetsInProject returns an iterator over the datasets in the provided project.
func (c *Client) DatasetsInProject(ctx context.Context, projectID string) *DatasetIterator {
	it := &DatasetIterator{
		ctx:       ctx,
		c:         c,
		projectID: projectID,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.items) },
		func() interface{} { b := it.items; it.items = nil; return b })
	return it
}

// DatasetIterator iterates over the datasets in a project.
type DatasetIterator struct {
	// ListHidden causes hidden datasets to be listed when set to true.
	ListHidden bool

	// Filter restricts the datasets returned by label. The filter syntax is described in
	// https://cloud.google.com/bigquery/docs/labeling-datasets#filtering_datasets_using_labels
	Filter string

	ctx       context.Context
	projectID string
	c         *Client
	pageInfo  *iterator.PageInfo
	nextFunc  func() error
	items     []*Dataset
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *DatasetIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

func (it *DatasetIterator) Next() (*Dataset, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	item := it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *DatasetIterator) fetch(pageSize int, pageToken string) (string, error) {
	datasets, nextPageToken, err := it.c.service.listDatasets(it.ctx, it.projectID,
		pageSize, pageToken, it.ListHidden, it.Filter)
	if err != nil {
		return "", err
	}
	for _, d := range datasets {
		d.c = it.c
		it.items = append(it.items, d)
	}
	return nextPageToken, nil
}
