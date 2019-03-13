// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ServiceError models all potential errors generated the service call
type ServiceError interface {
	// The http status code of the error
	GetHTTPStatusCode() int

	// The human-readable error string as sent by the service
	GetMessage() string

	// A short error code that defines the error, meant for programmatic parsing.
	// See https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
	GetCode() string

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	GetOpcRequestID() string
}

type servicefailure struct {
	StatusCode   int
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
	OpcRequestID string `json:"opc-request-id"`
}

func newServiceFailureFromResponse(response *http.Response) error {
	var err error

	se := servicefailure{
		StatusCode:   response.StatusCode,
		Code:         "BadErrorResponse",
		OpcRequestID: response.Header.Get("opc-request-id")}

	//If there is an error consume the body, entirely
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		se.Message = fmt.Sprintf("The body of the response was not readable, due to :%s", err.Error())
		return se
	}

	err = json.Unmarshal(body, &se)
	if err != nil {
		Debugf("Error response could not be parsed due to: %s", err.Error())
		se.Message = fmt.Sprintf("Failed to parse json from response body due to: %s. With response body %s.", err.Error(), string(body[:]))
		return se
	}
	return se
}

func (se servicefailure) Error() string {
	return fmt.Sprintf("Service error:%s. %s. http status code: %d. Opc request id: %s",
		se.Code, se.Message, se.StatusCode, se.OpcRequestID)
}

func (se servicefailure) GetHTTPStatusCode() int {
	return se.StatusCode

}

func (se servicefailure) GetMessage() string {
	return se.Message
}

func (se servicefailure) GetCode() string {
	return se.Code
}

func (se servicefailure) GetOpcRequestID() string {
	return se.OpcRequestID
}

// IsServiceError returns false if the error is not service side, otherwise true
// additionally it returns an interface representing the ServiceError
func IsServiceError(err error) (failure ServiceError, ok bool) {
	failure, ok = err.(servicefailure)
	return
}

type deadlineExceededByBackoffError struct{}

func (deadlineExceededByBackoffError) Error() string {
	return "now() + computed backoff duration exceeds request deadline"
}

// DeadlineExceededByBackoff is the error returned by Call() when GetNextDuration() returns a time.Duration that would
// force the user to wait past the request deadline before re-issuing a request. This enables us to exit early, since
// we cannot succeed based on the configured retry policy.
var DeadlineExceededByBackoff error = deadlineExceededByBackoffError{}
