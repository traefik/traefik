package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/jsonhooks-v1"
)

// APIError exposes an Akamai OPEN Edgegrid Error
type APIError struct {
	error
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Status      int              `json:"status"`
	Detail      string           `json:"detail"`
	Errors      []APIErrorDetail `json:"errors"`
	Problems    []APIErrorDetail `json:"problems"`
	Instance    string           `json:"instance"`
	Method      string           `json:"method"`
	ServerIP    string           `json:"serverIp"`
	ClientIP    string           `json:"clientIp"`
	RequestID   string           `json:"requestId"`
	RequestTime string           `json:"requestTime"`
	Response    *http.Response   `json:"-"`
	RawBody     string           `json:"-"`
}

type APIErrorDetail struct {
	Type          string `json:"type"`
	Title         string `json:"title"`
	Detail        string `json:"detail"`
	RejectedValue string `json:"rejectedValue"`
}

func (error APIError) Error() string {
	var errorDetails string
	if len(error.Errors) > 0 {
		for _, e := range error.Errors {
			errorDetails = fmt.Sprintf("%s \n %s", errorDetails, e)
		}
	}
	if len(error.Problems) > 0 {
		for _, e := range error.Problems {
			errorDetails = fmt.Sprintf("%s \n %s", errorDetails, e)
		}
	}
	return strings.TrimSpace(fmt.Sprintf("API Error: %d %s %s More Info %s\n %s", error.Status, error.Title, error.Detail, error.Type, errorDetails))
}

// NewAPIError creates a new API error based on a Response,
// or http.Response-like.
func NewAPIError(response *http.Response) APIError {
	// TODO: handle this error
	body, _ := ioutil.ReadAll(response.Body)

	return NewAPIErrorFromBody(response, body)
}

// NewAPIErrorFromBody creates a new API error, allowing you to pass in a body
//
// This function is intended to be used after the body has already been read for
// other purposes.
func NewAPIErrorFromBody(response *http.Response, body []byte) APIError {
	error := APIError{}
	if err := jsonhooks.Unmarshal(body, &error); err == nil {
		error.Status = response.StatusCode
		error.Title = response.Status
	}

	error.Response = response
	error.RawBody = string(body)

	return error
}

// IsInformational determines if a response was informational (1XX status)
func IsInformational(r *http.Response) bool {
	return r.StatusCode > 99 && r.StatusCode < 200
}

// IsSuccess determines if a response was successful (2XX status)
func IsSuccess(r *http.Response) bool {
	return r.StatusCode > 199 && r.StatusCode < 300
}

// IsRedirection determines if a response was a redirect (3XX status)
func IsRedirection(r *http.Response) bool {
	return r.StatusCode > 299 && r.StatusCode < 400
}

// IsClientError determines if a response was a client error (4XX status)
func IsClientError(r *http.Response) bool {
	return r.StatusCode > 399 && r.StatusCode < 500
}

// IsServerError determines if a response was a server error (5XX status)
func IsServerError(r *http.Response) bool {
	return r.StatusCode > 499 && r.StatusCode < 600
}

// IsError determines if the response was a client or server error (4XX or 5XX status)
func IsError(r *http.Response) bool {
	return r.StatusCode > 399 && r.StatusCode < 600
}
