// Copyright (c) 2015-2018 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response is an object represents executed request and its values.
type Response struct {
	Request     *Request
	RawResponse *http.Response

	body       []byte
	size       int64
	receivedAt time.Time
}

// Body method returns HTTP response as []byte array for the executed request.
// Note: `Response.Body` might be nil, if `Request.SetOutput` is used.
func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.body
}

// Status method returns the HTTP status string for the executed request.
//	Example: 200 OK
func (r *Response) Status() string {
	if r.RawResponse == nil {
		return ""
	}

	return r.RawResponse.Status
}

// StatusCode method returns the HTTP status code for the executed request.
//	Example: 200
func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}

	return r.RawResponse.StatusCode
}

// Result method returns the response value as an object if it has one
func (r *Response) Result() interface{} {
	return r.Request.Result
}

// Error method returns the error object if it has one
func (r *Response) Error() interface{} {
	return r.Request.Error
}

// Header method returns the response headers
func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}

	return r.RawResponse.Header
}

// Cookies method to access all the response cookies
func (r *Response) Cookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}

	return r.RawResponse.Cookies()
}

// String method returns the body of the server response as String.
func (r *Response) String() string {
	if r.body == nil {
		return ""
	}

	return strings.TrimSpace(string(r.body))
}

// Time method returns the time of HTTP response time that from request we sent and received a request.
// See `response.ReceivedAt` to know when client recevied response and see `response.Request.Time` to know
// when client sent a request.
func (r *Response) Time() time.Duration {
	return r.receivedAt.Sub(r.Request.Time)
}

// ReceivedAt method returns when response got recevied from server for the request.
func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

// Size method returns the HTTP response size in bytes. Ya, you can relay on HTTP `Content-Length` header,
// however it won't be good for chucked transfer/compressed response. Since Resty calculates response size
// at the client end. You will get actual size of the http response.
func (r *Response) Size() int64 {
	return r.size
}

// RawBody method exposes the HTTP raw response body. Use this method in-conjunction with `SetDoNotParseResponse`
// option otherwise you get an error as `read err: http: read on closed response body`.
//
// Do not forget to close the body, otherwise you might get into connection leaks, no connection reuse.
// Basically you have taken over the control of response parsing from `Resty`.
func (r *Response) RawBody() io.ReadCloser {
	if r.RawResponse == nil {
		return nil
	}
	return r.RawResponse.Body
}

// IsSuccess method returns true if HTTP status code >= 200 and <= 299 otherwise false.
func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

// IsError method returns true if HTTP status code >= 400 otherwise false.
func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}

func (r *Response) fmtBodyString(sl int64) string {
	if r.body != nil {
		if int64(len(r.body)) > sl {
			return fmt.Sprintf("***** RESPONSE TOO LARGE (size - %d) *****", len(r.body))
		}
		ct := r.Header().Get(hdrContentTypeKey)
		if IsJSONType(ct) {
			out := acquireBuffer()
			defer releaseBuffer(out)
			if err := json.Indent(out, r.body, "", "   "); err == nil {
				return out.String()
			}
		}
		return r.String()
	}

	return "***** NO CONTENT *****"
}
