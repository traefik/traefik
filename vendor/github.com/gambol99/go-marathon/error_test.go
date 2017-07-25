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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		httpCode   int
		nameSuffix string
		errCode    int
		errText    string
		content    string
	}{
		// 400
		{
			httpCode: http.StatusBadRequest,
			errCode:  ErrCodeBadRequest,
			errText:  "Invalid JSON (path: '/id' errors: error.expected.jsstring, error.something.else; path: '/name' errors: error.not.inventive)",
			content:  content400(),
		},
		// 401
		{
			httpCode: http.StatusUnauthorized,
			errCode:  ErrCodeUnauthorized,
			errText:  "invalid username or password.",
			content:  `{"message": "invalid username or password."}`,
		},
		// 403
		{
			httpCode: http.StatusForbidden,
			errCode:  ErrCodeForbidden,
			errText:  "Not Authorized to perform this action!",
			content:  `{"message": "Not Authorized to perform this action!"}`,
		},
		// 404
		{
			httpCode: http.StatusNotFound,
			errCode:  ErrCodeNotFound,
			errText:  "App '/not_existent' does not exist",
			content:  `{"message": "App '/not_existent' does not exist"}`,
		},
		// 409 POST
		{
			httpCode:   http.StatusConflict,
			nameSuffix: "POST",
			errCode:    ErrCodeDuplicateID,
			errText:    "An app with id [/existing_app] already exists.",
			content:    `{"message": "An app with id [/existing_app] already exists."}`,
		},
		// 409 PUT
		{
			httpCode:   http.StatusConflict,
			nameSuffix: "PUT",
			errCode:    ErrCodeAppLocked,
			errText:    "App is locked (locking deployment IDs: 97c136bf-5a28-4821-9d94-480d9fbb01c8)",
			content:    `{"message":"App is locked", "deployments": [ { "id": "97c136bf-5a28-4821-9d94-480d9fbb01c8" } ] }`,
		},
		// 422 pre-1.0 "details" key
		{
			httpCode:   422,
			nameSuffix: "pre-1.0 details key",
			errCode:    ErrCodeInvalidBean,
			errText:    "Something is not valid (attribute 'upgradeStrategy.minimumHealthCapacity': is greater than 1; attribute 'foobar': foo does not have enough bar)",
			content:    content422("details"),
		},
		// 422 pre-1.0 "errors" key
		{
			httpCode:   422,
			nameSuffix: "pre-1.0 errors key",
			errCode:    ErrCodeInvalidBean,
			errText:    "Something is not valid (attribute 'upgradeStrategy.minimumHealthCapacity': is greater than 1; attribute 'foobar': foo does not have enough bar)",
			content:    content422("errors"),
		},
		// 422 1.0 "invalid object"
		{
			httpCode:   422,
			nameSuffix: "invalid object",
			errCode:    ErrCodeInvalidBean,
			errText:    "Object is not valid (path: 'upgradeStrategy.minimumHealthCapacity' errors: is greater than 1; path: '/value' errors: service port conflict app /app1, service port conflict app /app2)",
			content:    content422V1(),
		},
		// 499 unknown error
		{
			httpCode:   499,
			nameSuffix: "unknown error",
			errCode:    ErrCodeUnknown,
			errText:    "unknown error",
			content:    `{"message": "unknown error"}`,
		},
		// 500
		{
			httpCode: http.StatusInternalServerError,
			errCode:  ErrCodeServer,
			errText:  "internal server error",
			content:  `{"message": "internal server error"}`,
		},
		// 503 (no JSON)
		{
			httpCode:   http.StatusServiceUnavailable,
			nameSuffix: "no JSON",
			errCode:    ErrCodeServer,
			errText:    "No server is available to handle this request.",
			content:    `No server is available to handle this request.`,
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf("%d", test.httpCode)
		if len(test.nameSuffix) > 0 {
			name = fmt.Sprintf("%s (%s)", name, test.nameSuffix)
		}
		apiErr := NewAPIError(test.httpCode, []byte(test.content))
		gotErrCode := apiErr.(*APIError).ErrCode
		assert.Equal(t, test.errCode, gotErrCode, fmt.Sprintf("HTTP code %s (error code): got %d, want %d", name, gotErrCode, test.errCode))
		pureErrText := strings.TrimPrefix(apiErr.Error(), "Marathon API error: ")
		assert.Equal(t, pureErrText, test.errText, fmt.Sprintf("HTTP code %s (error text)", name))
	}
}

func content400() string {
	return `{
	"message": "Invalid JSON",
	"details": [
		{
			"path": "/id",
			"errors": ["error.expected.jsstring", "error.something.else"]
		},
		{
			"path": "/name",
			"errors": ["error.not.inventive"]
		}
	]
}`
}

func content422(detailsPropKey string) string {
	return fmt.Sprintf(`{
	"message": "Something is not valid",
	"%s": [
		{
			"attribute": "upgradeStrategy.minimumHealthCapacity",
			"error": "is greater than 1"
		},
		{
			"attribute": "foobar",
			"error": "foo does not have enough bar"
		}
	]
}`, detailsPropKey)
}

func content422V1() string {
	return `{
	"message": "Object is not valid",
	"details": [
		{
			"path": "upgradeStrategy.minimumHealthCapacity",
			"errors": ["is greater than 1"]
		},
		{
			"path": "/value",
			"errors": ["service port conflict app /app1", "service port conflict app /app2"]
		}
	]
}`
}
