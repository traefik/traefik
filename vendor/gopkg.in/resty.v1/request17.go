// +build go1.7 go1.8

// Copyright (c) 2015-2019 Jeevanandam M (jeeva@myjeeva.com)
// 2016 Andrew Grigorev (https://github.com/ei-grad)
// All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// Request type is used to compose and send individual request from client
// go-resty is provide option override client level settings such as
// Auth Token, Basic Auth credentials, Header, Query Param, Form Data, Error object
// and also you can add more options for that particular request
type Request struct {
	URL        string
	Method     string
	Token      string
	QueryParam url.Values
	FormData   url.Values
	Header     http.Header
	Time       time.Time
	Body       interface{}
	Result     interface{}
	Error      interface{}
	RawRequest *http.Request
	SRV        *SRVRecord
	UserInfo   *User

	isMultiPart         bool
	isFormData          bool
	setContentLength    bool
	isSaveResponse      bool
	notParseResponse    bool
	jsonEscapeHTML      bool
	outputFile          string
	fallbackContentType string
	ctx                 context.Context
	pathParams          map[string]string
	client              *Client
	bodyBuf             *bytes.Buffer
	multipartFiles      []*File
	multipartFields     []*MultipartField
}

// Context method returns the Context if its already set in request
// otherwise it creates new one using `context.Background()`.
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

// SetContext method sets the context.Context for current Request. It allows
// to interrupt the request execution if ctx.Done() channel is closed.
// See https://blog.golang.org/context article and the "context" package
// documentation.
func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) addContextIfAvailable() {
	if r.ctx != nil {
		r.RawRequest = r.RawRequest.WithContext(r.ctx)
	}
}

func (r *Request) isContextCancelledIfAvailable() bool {
	if r.ctx != nil {
		if r.ctx.Err() != nil {
			return true
		}
	}
	return false
}

// for go1.7+
var noescapeJSONMarshal = func(v interface{}) ([]byte, error) {
	buf := acquireBuffer()
	defer releaseBuffer(buf)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}
