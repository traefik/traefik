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
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/internal"
	gax "github.com/googleapis/gax-go"

	"golang.org/x/net/context"
	bq "google.golang.org/api/bigquery/v2"
	"google.golang.org/api/googleapi"
)

// service provides an internal abstraction to isolate the generated
// BigQuery API; most of this package uses this interface instead.
// The single implementation, *bigqueryService, contains all the knowledge
// of the generated BigQuery API.
type service interface {
	// Jobs
	insertJob(ctx context.Context, projectId string, conf *insertJobConf) (*Job, error)
	getJobType(ctx context.Context, projectId, jobID string) (jobType, error)
	jobCancel(ctx context.Context, projectId, jobID string) error
	jobStatus(ctx context.Context, projectId, jobID string) (*JobStatus, error)

	// Tables
	createTable(ctx context.Context, conf *createTableConf) error
	getTableMetadata(ctx context.Context, projectID, datasetID, tableID string) (*TableMetadata, error)
	deleteTable(ctx context.Context, projectID, datasetID, tableID string) error

	// listTables returns a page of Tables and a next page token. Note: the Tables do not have their c field populated.
	listTables(ctx context.Context, projectID, datasetID string, pageSize int, pageToken string) ([]*Table, string, error)
	patchTable(ctx context.Context, projectID, datasetID, tableID string, conf *patchTableConf) (*TableMetadata, error)

	// Table data
	readTabledata(ctx context.Context, conf *readTableConf, pageToken string) (*readDataResult, error)
	insertRows(ctx context.Context, projectID, datasetID, tableID string, rows []*insertionRow, conf *insertRowsConf) error

	// Datasets
	insertDataset(ctx context.Context, datasetID, projectID string) error
	deleteDataset(ctx context.Context, datasetID, projectID string) error
	getDatasetMetadata(ctx context.Context, projectID, datasetID string) (*DatasetMetadata, error)

	// Misc

	// readQuery reads data resulting from a query job. If the job is
	// incomplete, an errIncompleteJob is returned. readQuery may be called
	// repeatedly to poll for job completion.
	readQuery(ctx context.Context, conf *readQueryConf, pageToken string) (*readDataResult, error)

	// listDatasets returns a page of Datasets and a next page token. Note: the Datasets do not have their c field populated.
	listDatasets(ctx context.Context, projectID string, maxResults int, pageToken string, all bool, filter string) ([]*Dataset, string, error)
}

type bigqueryService struct {
	s *bq.Service
}

func newBigqueryService(client *http.Client, endpoint string) (*bigqueryService, error) {
	s, err := bq.New(client)
	if err != nil {
		return nil, fmt.Errorf("constructing bigquery client: %v", err)
	}
	s.BasePath = endpoint

	return &bigqueryService{s: s}, nil
}

// getPages calls the supplied getPage function repeatedly until there are no pages left to get.
// token is the token of the initial page to start from.  Use an empty string to start from the beginning.
func getPages(token string, getPage func(token string) (nextToken string, err error)) error {
	for {
		var err error
		token, err = getPage(token)
		if err != nil {
			return err
		}
		if token == "" {
			return nil
		}
	}
}

type insertJobConf struct {
	job   *bq.Job
	media io.Reader
}

func (s *bigqueryService) insertJob(ctx context.Context, projectID string, conf *insertJobConf) (*Job, error) {
	call := s.s.Jobs.Insert(projectID, conf.job).Context(ctx)
	if conf.media != nil {
		call.Media(conf.media)
	}
	res, err := call.Do()
	if err != nil {
		return nil, err
	}
	return &Job{service: s, projectID: projectID, jobID: res.JobReference.JobId}, nil
}

type pagingConf struct {
	recordsPerRequest    int64
	setRecordsPerRequest bool

	startIndex uint64
}

type readTableConf struct {
	projectID, datasetID, tableID string
	paging                        pagingConf
	schema                        Schema // lazily initialized when the first page of data is fetched.
}

type readDataResult struct {
	pageToken string
	rows      [][]Value
	totalRows uint64
	schema    Schema
}

type readQueryConf struct {
	projectID, jobID string
	paging           pagingConf
}

func (s *bigqueryService) readTabledata(ctx context.Context, conf *readTableConf, pageToken string) (*readDataResult, error) {
	// Prepare request to fetch one page of table data.
	req := s.s.Tabledata.List(conf.projectID, conf.datasetID, conf.tableID)

	if pageToken != "" {
		req.PageToken(pageToken)
	} else {
		req.StartIndex(conf.paging.startIndex)
	}

	if conf.paging.setRecordsPerRequest {
		req.MaxResults(conf.paging.recordsPerRequest)
	}

	// Fetch the table schema in the background, if necessary.
	var schemaErr error
	var schemaFetch sync.WaitGroup
	if conf.schema == nil {
		schemaFetch.Add(1)
		go func() {
			defer schemaFetch.Done()
			var t *bq.Table
			t, schemaErr = s.s.Tables.Get(conf.projectID, conf.datasetID, conf.tableID).
				Fields("schema").
				Context(ctx).
				Do()
			if schemaErr == nil && t.Schema != nil {
				conf.schema = convertTableSchema(t.Schema)
			}
		}()
	}

	res, err := req.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	schemaFetch.Wait()
	if schemaErr != nil {
		return nil, schemaErr
	}

	result := &readDataResult{
		pageToken: res.PageToken,
		totalRows: uint64(res.TotalRows),
		schema:    conf.schema,
	}
	result.rows, err = convertRows(res.Rows, conf.schema)
	if err != nil {
		return nil, err
	}
	return result, nil
}

var errIncompleteJob = errors.New("internal error: query results not available because job is not complete")

// getQueryResultsTimeout controls the maximum duration of a request to the
// BigQuery GetQueryResults endpoint.  Setting a long timeout here does not
// cause increased overall latency, as results are returned as soon as they are
// available.
const getQueryResultsTimeout = time.Minute

func (s *bigqueryService) readQuery(ctx context.Context, conf *readQueryConf, pageToken string) (*readDataResult, error) {
	req := s.s.Jobs.GetQueryResults(conf.projectID, conf.jobID).
		TimeoutMs(getQueryResultsTimeout.Nanoseconds() / 1e6)

	if pageToken != "" {
		req.PageToken(pageToken)
	} else {
		req.StartIndex(conf.paging.startIndex)
	}

	if conf.paging.setRecordsPerRequest {
		req.MaxResults(conf.paging.recordsPerRequest)
	}

	res, err := req.Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	if !res.JobComplete {
		return nil, errIncompleteJob
	}
	schema := convertTableSchema(res.Schema)
	result := &readDataResult{
		pageToken: res.PageToken,
		totalRows: res.TotalRows,
		schema:    schema,
	}
	result.rows, err = convertRows(res.Rows, schema)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type insertRowsConf struct {
	templateSuffix      string
	ignoreUnknownValues bool
	skipInvalidRows     bool
}

func (s *bigqueryService) insertRows(ctx context.Context, projectID, datasetID, tableID string, rows []*insertionRow, conf *insertRowsConf) error {
	req := &bq.TableDataInsertAllRequest{
		TemplateSuffix:      conf.templateSuffix,
		IgnoreUnknownValues: conf.ignoreUnknownValues,
		SkipInvalidRows:     conf.skipInvalidRows,
	}
	for _, row := range rows {
		m := make(map[string]bq.JsonValue)
		for k, v := range row.Row {
			m[k] = bq.JsonValue(v)
		}
		req.Rows = append(req.Rows, &bq.TableDataInsertAllRequestRows{
			InsertId: row.InsertID,
			Json:     m,
		})
	}
	var res *bq.TableDataInsertAllResponse
	err := runWithRetry(ctx, func() error {
		var err error
		res, err = s.s.Tabledata.InsertAll(projectID, datasetID, tableID, req).Context(ctx).Do()
		return err
	})
	if err != nil {
		return err
	}
	if len(res.InsertErrors) == 0 {
		return nil
	}

	var errs PutMultiError
	for _, e := range res.InsertErrors {
		if int(e.Index) > len(rows) {
			return fmt.Errorf("internal error: unexpected row index: %v", e.Index)
		}
		rie := RowInsertionError{
			InsertID: rows[e.Index].InsertID,
			RowIndex: int(e.Index),
		}
		for _, errp := range e.Errors {
			rie.Errors = append(rie.Errors, errorFromErrorProto(errp))
		}
		errs = append(errs, rie)
	}
	return errs
}

type jobType int

const (
	copyJobType jobType = iota
	extractJobType
	loadJobType
	queryJobType
)

func (s *bigqueryService) getJobType(ctx context.Context, projectID, jobID string) (jobType, error) {
	res, err := s.s.Jobs.Get(projectID, jobID).
		Fields("configuration").
		Context(ctx).
		Do()

	if err != nil {
		return 0, err
	}

	switch {
	case res.Configuration.Copy != nil:
		return copyJobType, nil
	case res.Configuration.Extract != nil:
		return extractJobType, nil
	case res.Configuration.Load != nil:
		return loadJobType, nil
	case res.Configuration.Query != nil:
		return queryJobType, nil
	default:
		return 0, errors.New("unknown job type")
	}
}

func (s *bigqueryService) jobCancel(ctx context.Context, projectID, jobID string) error {
	// Jobs.Cancel returns a job entity, but the only relevant piece of
	// data it may contain (the status of the job) is unreliable.  From the
	// docs: "This call will return immediately, and the client will need
	// to poll for the job status to see if the cancel completed
	// successfully".  So it would be misleading to return a status.
	_, err := s.s.Jobs.Cancel(projectID, jobID).
		Fields(). // We don't need any of the response data.
		Context(ctx).
		Do()
	return err
}

func (s *bigqueryService) jobStatus(ctx context.Context, projectID, jobID string) (*JobStatus, error) {
	res, err := s.s.Jobs.Get(projectID, jobID).
		Fields("status"). // Only fetch what we need.
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}
	return jobStatusFromProto(res.Status)
}

var stateMap = map[string]State{"PENDING": Pending, "RUNNING": Running, "DONE": Done}

func jobStatusFromProto(status *bq.JobStatus) (*JobStatus, error) {
	state, ok := stateMap[status.State]
	if !ok {
		return nil, fmt.Errorf("unexpected job state: %v", status.State)
	}

	newStatus := &JobStatus{
		State: state,
		err:   nil,
	}
	if err := errorFromErrorProto(status.ErrorResult); state == Done && err != nil {
		newStatus.err = err
	}

	for _, ep := range status.Errors {
		newStatus.Errors = append(newStatus.Errors, errorFromErrorProto(ep))
	}
	return newStatus, nil
}

// listTables returns a subset of tables that belong to a dataset, and a token for fetching the next subset.
func (s *bigqueryService) listTables(ctx context.Context, projectID, datasetID string, pageSize int, pageToken string) ([]*Table, string, error) {
	var tables []*Table
	req := s.s.Tables.List(projectID, datasetID).
		PageToken(pageToken).
		Context(ctx)
	if pageSize > 0 {
		req.MaxResults(int64(pageSize))
	}
	res, err := req.Do()
	if err != nil {
		return nil, "", err
	}
	for _, t := range res.Tables {
		tables = append(tables, s.convertListedTable(t))
	}
	return tables, res.NextPageToken, nil
}

type createTableConf struct {
	projectID, datasetID, tableID string
	expiration                    time.Time
	viewQuery                     string
	schema                        *bq.TableSchema
	useStandardSQL                bool
	timePartitioning              *TimePartitioning
}

// createTable creates a table in the BigQuery service.
// expiration is an optional time after which the table will be deleted and its storage reclaimed.
// If viewQuery is non-empty, the created table will be of type VIEW.
// Note: expiration can only be set during table creation.
// Note: after table creation, a view can be modified only if its table was initially created with a view.
func (s *bigqueryService) createTable(ctx context.Context, conf *createTableConf) error {
	table := &bq.Table{
		TableReference: &bq.TableReference{
			ProjectId: conf.projectID,
			DatasetId: conf.datasetID,
			TableId:   conf.tableID,
		},
	}
	if !conf.expiration.IsZero() {
		table.ExpirationTime = conf.expiration.UnixNano() / 1e6
	}
	// TODO(jba): make it impossible to provide both a view query and a schema.
	if conf.viewQuery != "" {
		table.View = &bq.ViewDefinition{
			Query: conf.viewQuery,
		}
		if conf.useStandardSQL {
			table.View.UseLegacySql = false
			table.View.ForceSendFields = append(table.View.ForceSendFields, "UseLegacySql")
		}
	}
	if conf.schema != nil {
		table.Schema = conf.schema
	}
	if conf.timePartitioning != nil {
		table.TimePartitioning = &bq.TimePartitioning{
			Type:         "DAY",
			ExpirationMs: int64(conf.timePartitioning.Expiration.Seconds() * 1000),
		}
	}

	_, err := s.s.Tables.Insert(conf.projectID, conf.datasetID, table).Context(ctx).Do()
	return err
}

func (s *bigqueryService) getTableMetadata(ctx context.Context, projectID, datasetID, tableID string) (*TableMetadata, error) {
	table, err := s.s.Tables.Get(projectID, datasetID, tableID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return bqTableToMetadata(table), nil
}

func (s *bigqueryService) deleteTable(ctx context.Context, projectID, datasetID, tableID string) error {
	return s.s.Tables.Delete(projectID, datasetID, tableID).Context(ctx).Do()
}

func bqTableToMetadata(t *bq.Table) *TableMetadata {
	md := &TableMetadata{
		Description:      t.Description,
		Name:             t.FriendlyName,
		Type:             TableType(t.Type),
		ID:               t.Id,
		NumBytes:         t.NumBytes,
		NumRows:          t.NumRows,
		ExpirationTime:   unixMillisToTime(t.ExpirationTime),
		CreationTime:     unixMillisToTime(t.CreationTime),
		LastModifiedTime: unixMillisToTime(int64(t.LastModifiedTime)),
	}
	if t.Schema != nil {
		md.Schema = convertTableSchema(t.Schema)
	}
	if t.View != nil {
		md.View = t.View.Query
	}
	if t.TimePartitioning != nil {
		md.TimePartitioning = &TimePartitioning{time.Duration(t.TimePartitioning.ExpirationMs) * time.Millisecond}
	}

	return md
}

func bqDatasetToMetadata(d *bq.Dataset) *DatasetMetadata {
	/// TODO(jba): access
	return &DatasetMetadata{
		CreationTime:           unixMillisToTime(d.CreationTime),
		LastModifiedTime:       unixMillisToTime(d.LastModifiedTime),
		DefaultTableExpiration: time.Duration(d.DefaultTableExpirationMs) * time.Millisecond,
		Description:            d.Description,
		Name:                   d.FriendlyName,
		ID:                     d.Id,
		Location:               d.Location,
		Labels:                 d.Labels,
	}
}

// Convert a number of milliseconds since the Unix epoch to a time.Time.
// Treat an input of zero specially: convert it to the zero time,
// rather than the start of the epoch.
func unixMillisToTime(m int64) time.Time {
	if m == 0 {
		return time.Time{}
	}
	return time.Unix(0, m*1e6)
}

func (s *bigqueryService) convertListedTable(t *bq.TableListTables) *Table {
	return &Table{
		ProjectID: t.TableReference.ProjectId,
		DatasetID: t.TableReference.DatasetId,
		TableID:   t.TableReference.TableId,
	}
}

// patchTableConf contains fields to be patched.
type patchTableConf struct {
	// These fields are omitted from the patch operation if nil.
	Description *string
	Name        *string
	Schema      Schema
}

func (s *bigqueryService) patchTable(ctx context.Context, projectID, datasetID, tableID string, conf *patchTableConf) (*TableMetadata, error) {
	t := &bq.Table{}
	forceSend := func(field string) {
		t.ForceSendFields = append(t.ForceSendFields, field)
	}

	if conf.Description != nil {
		t.Description = *conf.Description
		forceSend("Description")
	}
	if conf.Name != nil {
		t.FriendlyName = *conf.Name
		forceSend("FriendlyName")
	}
	if conf.Schema != nil {
		t.Schema = conf.Schema.asTableSchema()
		forceSend("Schema")
	}
	table, err := s.s.Tables.Patch(projectID, datasetID, tableID, t).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}
	return bqTableToMetadata(table), nil
}

func (s *bigqueryService) insertDataset(ctx context.Context, datasetID, projectID string) error {
	ds := &bq.Dataset{
		DatasetReference: &bq.DatasetReference{DatasetId: datasetID},
	}
	_, err := s.s.Datasets.Insert(projectID, ds).Context(ctx).Do()
	return err
}

func (s *bigqueryService) deleteDataset(ctx context.Context, datasetID, projectID string) error {
	return s.s.Datasets.Delete(projectID, datasetID).Context(ctx).Do()
}

func (s *bigqueryService) getDatasetMetadata(ctx context.Context, projectID, datasetID string) (*DatasetMetadata, error) {
	table, err := s.s.Datasets.Get(projectID, datasetID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return bqDatasetToMetadata(table), nil
}

func (s *bigqueryService) listDatasets(ctx context.Context, projectID string, maxResults int, pageToken string, all bool, filter string) ([]*Dataset, string, error) {
	req := s.s.Datasets.List(projectID).
		Context(ctx).
		PageToken(pageToken).
		All(all)
	if maxResults > 0 {
		req.MaxResults(int64(maxResults))
	}
	if filter != "" {
		req.Filter(filter)
	}
	res, err := req.Do()
	if err != nil {
		return nil, "", err
	}
	var datasets []*Dataset
	for _, d := range res.Datasets {
		datasets = append(datasets, s.convertListedDataset(d))
	}
	return datasets, res.NextPageToken, nil
}

func (s *bigqueryService) convertListedDataset(d *bq.DatasetListDatasets) *Dataset {
	return &Dataset{
		ProjectID: d.DatasetReference.ProjectId,
		DatasetID: d.DatasetReference.DatasetId,
	}
}

// runWithRetry calls the function until it returns nil or a non-retryable error, or
// the context is done.
// See the similar function in ../storage/invoke.go. The main difference is the
// reason for retrying.
func runWithRetry(ctx context.Context, call func() error) error {
	backoff := gax.Backoff{
		Initial:    2 * time.Second,
		Max:        32 * time.Second,
		Multiplier: 2,
	}
	return internal.Retry(ctx, backoff, func() (stop bool, err error) {
		err = call()
		if err == nil {
			return true, nil
		}
		e, ok := err.(*googleapi.Error)
		if !ok {
			return true, err
		}
		var reason string
		if len(e.Errors) > 0 {
			reason = e.Errors[0].Reason
		}
		// Retry using the criteria in
		// https://cloud.google.com/bigquery/troubleshooting-errors
		if reason == "backendError" && (e.Code == 500 || e.Code == 503) {
			return false, nil
		}
		return true, err
	})
}
