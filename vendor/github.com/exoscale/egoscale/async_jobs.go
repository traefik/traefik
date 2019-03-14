package egoscale

import (
	"encoding/json"
	"errors"
)

// AsyncJobResult represents an asynchronous job result
type AsyncJobResult struct {
	AccountID       *UUID            `json:"accountid,omitempty" doc:"the account that executed the async command"`
	Cmd             string           `json:"cmd,omitempty" doc:"the async command executed"`
	Created         string           `json:"created,omitempty" doc:"the created date of the job"`
	JobID           *UUID            `json:"jobid" doc:"extra field for the initial async call"`
	JobInstanceID   *UUID            `json:"jobinstanceid,omitempty" doc:"the unique ID of the instance/entity object related to the job"`
	JobInstanceType string           `json:"jobinstancetype,omitempty" doc:"the instance/entity object related to the job"`
	JobProcStatus   int              `json:"jobprocstatus,omitempty" doc:"the progress information of the PENDING job"`
	JobResult       *json.RawMessage `json:"jobresult,omitempty" doc:"the result reason"`
	JobResultCode   int              `json:"jobresultcode,omitempty" doc:"the result code for the job"`
	JobResultType   string           `json:"jobresulttype,omitempty" doc:"the result type"`
	JobStatus       JobStatusType    `json:"jobstatus,omitempty" doc:"the current job status-should be 0 for PENDING"`
	UserID          *UUID            `json:"userid,omitempty" doc:"the user that executed the async command"`
}

// ListRequest buils the (empty) ListAsyncJobs request
func (a AsyncJobResult) ListRequest() (ListCommand, error) {
	req := &ListAsyncJobs{
		StartDate: a.Created,
	}

	return req, nil
}

// Error builds an error message from the result
func (a AsyncJobResult) Error() error {
	r := new(ErrorResponse)
	if e := json.Unmarshal(*a.JobResult, r); e != nil {
		return e
	}
	return r
}

// QueryAsyncJobResult represents a query to fetch the status of async job
type QueryAsyncJobResult struct {
	JobID *UUID `json:"jobid" doc:"the ID of the asynchronous job"`
	_     bool  `name:"queryAsyncJobResult" description:"Retrieves the current status of asynchronous job."`
}

// Response returns the struct to unmarshal
func (QueryAsyncJobResult) Response() interface{} {
	return new(AsyncJobResult)
}

//go:generate go run generate/main.go -interface=Listable ListAsyncJobs

// ListAsyncJobs list the asynchronous jobs
type ListAsyncJobs struct {
	Keyword   string `json:"keyword,omitempty" doc:"List by keyword"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"pagesize,omitempty"`
	StartDate string `json:"startdate,omitempty" doc:"the start date of the async job"`
	_         bool   `name:"listAsyncJobs" description:"Lists all pending asynchronous jobs for the account."`
}

// ListAsyncJobsResponse represents a list of job results
type ListAsyncJobsResponse struct {
	Count    int              `json:"count"`
	AsyncJob []AsyncJobResult `json:"asyncjobs"`
}

// Result unmarshals the result of an AsyncJobResult into the given interface
func (a AsyncJobResult) Result(i interface{}) error {
	if a.JobStatus == Failure {
		return a.Error()
	}

	if a.JobStatus == Success {
		m := map[string]json.RawMessage{}
		err := json.Unmarshal(*(a.JobResult), &m)

		if err == nil {
			if len(m) >= 1 {
				if _, ok := m["success"]; ok {
					return json.Unmarshal(*(a.JobResult), i)
				}

				// otherwise, pick the first key
				for k := range m {
					return json.Unmarshal(m[k], i)
				}
			}
			return errors.New("empty response")
		}
	}

	return nil
}
