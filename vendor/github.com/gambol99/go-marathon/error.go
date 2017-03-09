/*
Copyright 2015 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	// ErrCodeBadRequest specifies a 400 Bad Request error.
	ErrCodeBadRequest = iota
	// ErrCodeUnauthorized specifies a 401 Unauthorized error.
	ErrCodeUnauthorized
	// ErrCodeForbidden specifies a 403 Forbidden error.
	ErrCodeForbidden
	// ErrCodeNotFound specifies a 404 Not Found error.
	ErrCodeNotFound
	// ErrCodeDuplicateID specifies a PUT 409 Conflict error.
	ErrCodeDuplicateID
	// ErrCodeAppLocked specifies a POST 409 Conflict error.
	ErrCodeAppLocked
	// ErrCodeInvalidBean specifies a 422 UnprocessableEntity error.
	ErrCodeInvalidBean
	// ErrCodeServer specifies a 500+ Server error.
	ErrCodeServer
	// ErrCodeUnknown specifies an unknown error.
	ErrCodeUnknown
)

// InvalidEndpointError indicates a endpoint error in the marathon urls
type InvalidEndpointError struct {
	message string
}

// Error returns the string message
func (e *InvalidEndpointError) Error() string {
	return e.message
}

// newInvalidEndpointError creates a new error
func newInvalidEndpointError(message string, args ...interface{}) error {
	return &InvalidEndpointError{message: fmt.Sprintf(message, args)}
}

// APIError represents a generic API error.
type APIError struct {
	// ErrCode specifies the nature of the error.
	ErrCode int
	message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Marathon API error: %s", e.message)
}

// NewAPIError creates a new APIError instance from the given response code and content.
func NewAPIError(code int, content []byte) error {
	var errDef errorDefinition
	switch {
	case code == http.StatusBadRequest:
		errDef = &badRequestDef{}
	case code == http.StatusUnauthorized:
		errDef = &simpleErrDef{code: ErrCodeUnauthorized}
	case code == http.StatusForbidden:
		errDef = &simpleErrDef{code: ErrCodeForbidden}
	case code == http.StatusNotFound:
		errDef = &simpleErrDef{code: ErrCodeNotFound}
	case code == http.StatusConflict:
		errDef = &conflictDef{}
	case code == 422:
		errDef = &unprocessableEntityDef{}
	case code >= http.StatusInternalServerError:
		errDef = &simpleErrDef{code: ErrCodeServer}
	default:
		errDef = &simpleErrDef{code: ErrCodeUnknown}
	}

	return parseContent(errDef, content)
}

type errorDefinition interface {
	message() string
	errCode() int
}

func parseContent(errDef errorDefinition, content []byte) error {
	// If the content cannot be JSON-unmarshalled, we assume that it's not JSON
	// and encode it into the APIError instance as-is.
	errMessage := string(content)
	if err := json.Unmarshal(content, errDef); err == nil {
		errMessage = errDef.message()
	}

	return &APIError{message: errMessage, ErrCode: errDef.errCode()}
}

type simpleErrDef struct {
	Message string `json:"message"`
	code    int
}

func (def *simpleErrDef) message() string {
	return def.Message
}

func (def *simpleErrDef) errCode() int {
	return def.code
}

type detailDescription struct {
	Path   string   `json:"path"`
	Errors []string `json:"errors"`
}

func (d detailDescription) String() string {
	return fmt.Sprintf("path: '%s' errors: %s", d.Path, strings.Join(d.Errors, ", "))
}

type badRequestDef struct {
	Message string              `json:"message"`
	Details []detailDescription `json:"details"`
}

func (def *badRequestDef) message() string {
	var details []string
	for _, detail := range def.Details {
		details = append(details, detail.String())
	}

	return fmt.Sprintf("%s (%s)", def.Message, strings.Join(details, "; "))
}

func (def *badRequestDef) errCode() int {
	return ErrCodeBadRequest
}

type conflictDef struct {
	Message     string `json:"message"`
	Deployments []struct {
		ID string `json:"id"`
	} `json:"deployments"`
}

func (def *conflictDef) message() string {
	if len(def.Deployments) == 0 {
		// 409 Conflict response to "POST /v2/apps".
		return def.Message
	}

	// 409 Conflict response to "PUT /v2/apps/{appId}".
	var ids []string
	for _, deployment := range def.Deployments {
		ids = append(ids, deployment.ID)
	}
	return fmt.Sprintf("%s (locking deployment IDs: %s)", def.Message, strings.Join(ids, ", "))
}

func (def *conflictDef) errCode() int {
	if len(def.Deployments) == 0 {
		return ErrCodeDuplicateID
	}

	return ErrCodeAppLocked
}

type unprocessableEntityDetails []struct {
	// Used in Marathon >= 1.0.0-RC1.
	detailDescription
	// Used in Marathon < 1.0.0-RC1.
	Attribute string `json:"attribute"`
	Error     string `json:"error"`
}

type unprocessableEntityDef struct {
	Message string `json:"message"`
	// Name used in Marathon >= 0.15.0.
	Details unprocessableEntityDetails `json:"details"`
	// Name used in Marathon < 0.15.0.
	Errors unprocessableEntityDetails `json:"errors"`
}

func (def *unprocessableEntityDef) message() string {
	joinDetails := func(details unprocessableEntityDetails) []string {
		var res []string
		for _, detail := range details {
			res = append(res, fmt.Sprintf("attribute '%s': %s", detail.Attribute, detail.Error))
		}
		return res
	}

	var details []string
	switch {
	case len(def.Errors) > 0:
		details = joinDetails(def.Errors)
	case len(def.Details) > 0 && len(def.Details[0].Attribute) > 0:
		details = joinDetails(def.Details)
	default:
		for _, detail := range def.Details {
			details = append(details, detail.detailDescription.String())
		}
	}

	return fmt.Sprintf("%s (%s)", def.Message, strings.Join(details, "; "))
}

func (def *unprocessableEntityDef) errCode() int {
	return ErrCodeInvalidBean
}
