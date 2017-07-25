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

// LoadConfig holds the configuration for a load job.
type LoadConfig struct {
	// JobID is the ID to use for the load job. If unset, a job ID will be automatically created.
	JobID string

	// Src is the source from which data will be loaded.
	Src LoadSource

	// Dst is the table into which the data will be loaded.
	Dst *Table

	// CreateDisposition specifies the circumstances under which the destination table will be created.
	// The default is CreateIfNeeded.
	CreateDisposition TableCreateDisposition

	// WriteDisposition specifies how existing data in the destination table is treated.
	// The default is WriteAppend.
	WriteDisposition TableWriteDisposition
}

// A Loader loads data from Google Cloud Storage into a BigQuery table.
type Loader struct {
	LoadConfig
	c *Client
}

// A LoadSource represents a source of data that can be loaded into
// a BigQuery table.
//
// This package defines two LoadSources: GCSReference, for Google Cloud Storage
// objects, and ReaderSource, for data read from an io.Reader.
type LoadSource interface {
	populateInsertJobConfForLoad(conf *insertJobConf)
}

// LoaderFrom returns a Loader which can be used to load data into a BigQuery table.
// The returned Loader may optionally be further configured before its Run method is called.
func (t *Table) LoaderFrom(src LoadSource) *Loader {
	return &Loader{
		c: t.c,
		LoadConfig: LoadConfig{
			Src: src,
			Dst: t,
		},
	}
}

// Run initiates a load job.
func (l *Loader) Run(ctx context.Context) (*Job, error) {
	job := &bq.Job{
		Configuration: &bq.JobConfiguration{
			Load: &bq.JobConfigurationLoad{
				CreateDisposition: string(l.CreateDisposition),
				WriteDisposition:  string(l.WriteDisposition),
			},
		},
	}
	conf := &insertJobConf{job: job}
	l.Src.populateInsertJobConfForLoad(conf)
	setJobRef(job, l.JobID, l.c.projectID)

	job.Configuration.Load.DestinationTable = l.Dst.tableRefProto()

	return l.c.service.insertJob(ctx, l.c.projectID, conf)
}
