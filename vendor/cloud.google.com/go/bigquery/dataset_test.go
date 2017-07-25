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
	"errors"
	"strconv"
	"testing"

	"golang.org/x/net/context"
	itest "google.golang.org/api/iterator/testing"
)

// readServiceStub services read requests by returning data from an in-memory list of values.
type listTablesServiceStub struct {
	expectedProject, expectedDataset string
	tables                           []*Table
	service
}

func (s *listTablesServiceStub) listTables(ctx context.Context, projectID, datasetID string, pageSize int, pageToken string) ([]*Table, string, error) {
	if projectID != s.expectedProject {
		return nil, "", errors.New("wrong project id")
	}
	if datasetID != s.expectedDataset {
		return nil, "", errors.New("wrong dataset id")
	}
	const maxPageSize = 2
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	start := 0
	if pageToken != "" {
		var err error
		start, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", err
		}
	}
	end := start + pageSize
	if end > len(s.tables) {
		end = len(s.tables)
	}
	nextPageToken := ""
	if end < len(s.tables) {
		nextPageToken = strconv.Itoa(end)
	}
	return s.tables[start:end], nextPageToken, nil
}

func TestTables(t *testing.T) {
	t1 := &Table{ProjectID: "p1", DatasetID: "d1", TableID: "t1"}
	t2 := &Table{ProjectID: "p1", DatasetID: "d1", TableID: "t2"}
	t3 := &Table{ProjectID: "p1", DatasetID: "d1", TableID: "t3"}
	allTables := []*Table{t1, t2, t3}
	c := &Client{
		service: &listTablesServiceStub{
			expectedProject: "x",
			expectedDataset: "y",
			tables:          allTables,
		},
		projectID: "x",
	}
	msg, ok := itest.TestIterator(allTables,
		func() interface{} { return c.Dataset("y").Tables(context.Background()) },
		func(it interface{}) (interface{}, error) { return it.(*TableIterator).Next() })
	if !ok {
		t.Error(msg)
	}
}

type listDatasetsFake struct {
	service

	projectID string
	datasets  []*Dataset
	hidden    map[*Dataset]bool
}

func (df *listDatasetsFake) listDatasets(_ context.Context, projectID string, pageSize int, pageToken string, listHidden bool, filter string) ([]*Dataset, string, error) {
	const maxPageSize = 2
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if filter != "" {
		return nil, "", errors.New("filter not supported")
	}
	if projectID != df.projectID {
		return nil, "", errors.New("bad project ID")
	}
	start := 0
	if pageToken != "" {
		var err error
		start, err = strconv.Atoi(pageToken)
		if err != nil {
			return nil, "", err
		}
	}
	var (
		i             int
		result        []*Dataset
		nextPageToken string
	)
	for i = start; len(result) < pageSize && i < len(df.datasets); i++ {
		if df.hidden[df.datasets[i]] && !listHidden {
			continue
		}
		result = append(result, df.datasets[i])
	}
	if i < len(df.datasets) {
		nextPageToken = strconv.Itoa(i)
	}
	return result, nextPageToken, nil
}

func TestDatasets(t *testing.T) {
	service := &listDatasetsFake{projectID: "p"}
	client := &Client{service: service}
	datasets := []*Dataset{
		{"p", "a", client},
		{"p", "b", client},
		{"p", "hidden", client},
		{"p", "c", client},
	}
	service.datasets = datasets
	service.hidden = map[*Dataset]bool{datasets[2]: true}
	c := &Client{
		projectID: "p",
		service:   service,
	}
	msg, ok := itest.TestIterator(datasets,
		func() interface{} { it := c.Datasets(context.Background()); it.ListHidden = true; return it },
		func(it interface{}) (interface{}, error) { return it.(*DatasetIterator).Next() })
	if !ok {
		t.Fatalf("ListHidden=true: %s", msg)
	}

	msg, ok = itest.TestIterator([]*Dataset{datasets[0], datasets[1], datasets[3]},
		func() interface{} { it := c.Datasets(context.Background()); it.ListHidden = false; return it },
		func(it interface{}) (interface{}, error) { return it.(*DatasetIterator).Next() })
	if !ok {
		t.Fatalf("ListHidden=false: %s", msg)
	}
}
