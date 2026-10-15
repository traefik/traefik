/*
Copyright 2022 The Kubernetes Authors.

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

package httpexpect

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// StatusRange is enum for response status ranges.
type StatusRange int

const (
	// Status1xx defines "Informational" status codes.
	Status1xx StatusRange = 100

	// Status2xx defines "Success" status codes.
	Status2xx StatusRange = 200

	// Status3xx defines "Redirection" status codes.
	Status3xx StatusRange = 300

	// Status4xx defines "Client Error" status codes.
	Status4xx StatusRange = 400

	// Status5xx defines "Server Error" status codes.
	Status5xx StatusRange = 500
)

type HTTPResponse struct {
	chain    chain
	content  []byte
	cookies  []*http.Cookie
	Response *http.Response
}

// NewResponse returns an empty HTTPResponse object.
func NewResponse(reporter Reporter) *HTTPResponse {
	return &HTTPResponse{
		chain: makeChain(reporter),
	}
}

// Body returns the body of the response.
func (r *HTTPResponse) Body() *String {
	// The chain is required to report a failed assertion instead of panicking.
	return &String{chain: r.chain, value: string(r.content)}
}

// Raw returns the raw http response.
func (r *HTTPResponse) Raw() *http.Response {
	return r.Response
}

// Status compare the actual http response with the expected one raising and error
// if they don't match.
func (r *HTTPResponse) Status(status int) *HTTPResponse {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("status", statusCodeText(status), statusCodeText(r.Response.StatusCode))
	return r
}

// ContentEncoding succeeds if response has exactly given Content-Encoding
func (r *HTTPResponse) ContentEncoding(encoding ...string) *HTTPResponse {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("\"Content-Encoding\" header", encoding, r.Response.Header["Content-Encoding"])
	return r
}

// ContentType succeeds if response contains Content-Type header with given
// media type and charset.
func (r *HTTPResponse) ContentType(mediaType string, charset ...string) *HTTPResponse {
	r.checkContentType(mediaType, charset...)
	return r
}

// Cookies returns a new Array object with all cookie names set by this response.
// Returned Array contains a String value for every cookie name.
func (r *HTTPResponse) Cookies() *Array {
	if r.chain.failed() {
		return &Array{r.chain, nil}
	}
	names := []interface{}{}
	for _, c := range r.cookies {
		names = append(names, c.Name)
	}
	return &Array{r.chain, names}
}

// Cookie returns a new Cookie object that may be used to inspect given cookie
// set by this response.
func (r *HTTPResponse) Cookie(name string) *Cookie {
	if r.chain.failed() {
		return &Cookie{r.chain, nil}
	}
	names := []string{}
	for _, c := range r.cookies {
		if c.Name == name {
			return &Cookie{r.chain, c}
		}
		names = append(names, c.Name)
	}
	r.chain.fail("\nexpected response with cookie:\n %q\n\nbut got only cookies:\n%s", name, dumpValue(names))
	return &Cookie{r.chain, nil}
}

// Headers returns a new Object that may be used to inspect header map.
func (r *HTTPResponse) Headers() *Object {
	var value map[string]interface{}
	if !r.chain.failed() {
		value, _ = canonMap(&r.chain, r.Response.Header)
	}
	return &Object{r.chain, value}
}

// Header returns a new String object that may be used to inspect given header.
func (r *HTTPResponse) Header(header string) *String {
	return &String{chain: r.chain, value: r.Response.Header.Get(header)}
}

func canonMap(chain *chain, in interface{}) (map[string]interface{}, bool) {
	var out map[string]interface{}
	data, ok := canonValue(chain, in)
	if ok {
		out, ok = data.(map[string]interface{})
		if !ok {
			chain.fail("expected map, got %v", out)
		}
	}
	return out, ok
}

func canonValue(chain *chain, in interface{}) (interface{}, bool) {
	b, err := json.Marshal(in)
	if err != nil {
		chain.fail(err.Error())
		return nil, false
	}

	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		chain.fail(err.Error())
		return nil, false
	}

	return out, true
}

// StatusRange succeeds if response status belongs to given range.
func (r *HTTPResponse) StatusRange(rn StatusRange) *HTTPResponse {
	if r.chain.failed() {
		return r
	}
	status := statusCodeText(r.Response.StatusCode)

	actual := statusRangeText(r.Response.StatusCode)
	expected := statusRangeText(int(rn))

	if actual == "" || actual != expected {
		if actual == "" {
			r.chain.fail("\nexpected status from range:\n %q\n\nbut got:\n %q",
				expected, status)
		} else {
			r.chain.fail("\nexpected status from range:\n %q\n\nbut got:\n %q (%q)",
				expected, actual, status)
		}
	}
	return r
}

func statusCodeText(code int) string {
	if s := http.StatusText(code); s != "" {
		return strconv.Itoa(code) + " " + s
	}
	return strconv.Itoa(code)
}

func statusRangeText(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx Informational"
	case code >= 200 && code < 300:
		return "2xx Success"
	case code >= 300 && code < 400:
		return "3xx Redirection"
	case code >= 400 && code < 500:
		return "4xx Client Error"
	case code >= 500 && code < 600:
		return "5xx Server Error"
	default:
		return ""
	}
}

func (r *HTTPResponse) checkContentType(expectedType string, expectedCharset ...string) bool {
	if r.chain.failed() {
		return false
	}

	contentType := r.Response.Header.Get("Content-Type")

	if expectedType == "" && len(expectedCharset) == 0 {
		if contentType == "" {
			return true
		}
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		r.chain.fail("\ngot invalid \"Content-Type\" header %q", contentType)
		return false
	}

	if mediaType != expectedType {
		//nolint:goconst //string interpolation
		r.chain.fail("\nexpected \"Content-Type\" header with %q media type,"+
			"\nbut got %q", expectedType, mediaType)
		return false
	}

	charset := params["charset"]

	if len(expectedCharset) == 0 {
		if charset != "" && !strings.EqualFold(charset, "utf-8") {
			r.chain.fail("\nexpected \"Content-Type\" header with \"utf-8\" or empty charset,"+
				"\nbut got %q", charset)
			return false
		}
	} else {
		if !strings.EqualFold(charset, expectedCharset[0]) {
			r.chain.fail("\nexpected \"Content-Type\" header with %q charset,"+
				"\nbut got %q", expectedCharset[0], charset)
			return false
		}
	}
	return true
}

func (r *HTTPResponse) checkEqual(what string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		r.chain.fail("\nexpected %s equal to:\n%s\n\nbut got:\n%s",
			what, dumpValue(expected), dumpValue(actual))
	}
}

func dumpValue(value interface{}) string {
	b, err := json.MarshalIndent(value, " ", "  ")
	if err != nil {
		return " " + fmt.Sprintf("%#v", value)
	}
	return " " + string(b)
}
