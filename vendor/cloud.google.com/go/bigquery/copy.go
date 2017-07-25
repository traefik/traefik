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

package bigquery

import (
	"golang.org/x/net/context"
	bq "google.golang.org/api/bigquery/v2"
)

// CopyConfig holds the configuration for a copy job.
type CopyConfig struct {
	// JobID is the ID to use for the copy job. If unset, a job ID will be automatically created.
	JobID string

	// Srcs are the tables from which data will be copied.
	Srcs []*Table

	// Dst is the table into which the data will be copied.
	Dst *Table

	// CreateDisposition specifies the circumstances under which the destination table will be created.
	// The default is CreateIfNeeded.
	CreateDisposition TableCreateDisposition

	// WriteDisposition specifies how existing data in the destination table is treated.
	// The default is WriteAppend.
	WriteDisposition TableWriteDisposition
}

// A Copier copies data into a BigQuery table from one or more BigQuery tables.
type Copier struct {
	CopyConfig
	c *Client
}

// CopierFrom returns a Copier which can be used to copy data into a
// BigQuery table from  one or more BigQuery tables.
// The returned Copier may optionally be further configured before its Run method is called.
func (t *Table) CopierFrom(srcs ...*Table) *Copier {
	return &Copier{
		c: t.c,
		CopyConfig: CopyConfig{
			Srcs: srcs,
			Dst:  t,
		},
	}
}

// Run initiates a copy job.
func (c *Copier) Run(ctx context.Context) (*Job, error) {
	conf := &bq.JobConfigurationTableCopy{
		CreateDisposition: string(c.CreateDisposition),
		WriteDisposition:  string(c.WriteDisposition),
		DestinationTable:  c.Dst.tableRefProto(),
	}
	for _, t := range c.Srcs {
		conf.SourceTables = append(conf.SourceTables, t.tableRefProto())
	}
	job := &bq.Job{Configuration: &bq.JobConfiguration{Copy: conf}}
	setJobRef(job, c.JobID, c.c.projectID)
	return c.c.service.insertJob(ctx, c.c.projectID, &insertJobConf{job: job})
}
