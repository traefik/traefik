package egoscale

import (
	"encoding/json"
)

// AsyncJobResult represents an asynchronous job result
type AsyncJobResult struct {
	AccountID       string           `json:"accountid"`
	Cmd             string           `json:"cmd"`
	Created         string           `json:"created"`
	JobInstanceID   string           `json:"jobinstanceid"`
	JobInstanceType string           `json:"jobinstancetype"`
	JobProcStatus   int              `json:"jobprocstatus"`
	JobResult       *json.RawMessage `json:"jobresult"`
	JobResultCode   int              `json:"jobresultcode"`
	JobResultType   string           `json:"jobresulttype"`
	JobStatus       JobStatusType    `json:"jobstatus"`
	UserID          string           `json:"userid"`
	JobID           string           `json:"jobid"`
}

// QueryAsyncJobResult represents a query to fetch the status of async job
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/queryAsyncJobResult.html
type QueryAsyncJobResult struct {
	JobID string `json:"jobid"`
}

func (*QueryAsyncJobResult) name() string {
	return "queryAsyncJobResult"
}

func (*QueryAsyncJobResult) response() interface{} {
	return new(QueryAsyncJobResultResponse)
}

// QueryAsyncJobResultResponse represents the current status of an asynchronous job
type QueryAsyncJobResultResponse AsyncJobResult

// ListAsyncJobs list the asynchronous jobs
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listAsyncJobs.html
type ListAsyncJobs struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	IsRecursive bool   `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	StartDate   string `json:"startdate,omitempty"`
}

func (*ListAsyncJobs) name() string {
	return "listAsyncJobs"
}

func (*ListAsyncJobs) response() interface{} {
	return new(ListAsyncJobsResponse)
}

// ListAsyncJobsResponse represents a list of job results
type ListAsyncJobsResponse struct {
	Count     int              `json:"count"`
	AsyncJobs []AsyncJobResult `json:"asyncjobs"`
}
