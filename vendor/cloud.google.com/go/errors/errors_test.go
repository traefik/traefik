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

package errors

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	erpb "google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1"
)

const testProjectID = "testproject"

type fakeReportErrorsClient struct {
	req  *erpb.ReportErrorEventRequest
	fail bool
}

func (c *fakeReportErrorsClient) ReportErrorEvent(ctx context.Context, req *erpb.ReportErrorEventRequest) (*erpb.ReportErrorEventResponse, error) {
	if c.fail {
		return nil, errors.New("request failed")
	}
	c.req = req
	return &erpb.ReportErrorEventResponse{}, nil
}

func newTestClient(c *fakeReportErrorsClient) *Client {
	newApiInterface = func(ctx context.Context, opts ...option.ClientOption) (apiInterface, error) {
		return c, nil
	}
	t, err := NewClient(context.Background(), testProjectID, "myservice", "v1.000", false)
	if err != nil {
		panic(err)
	}
	t.RepanicDefault = false
	return t
}

var ctx context.Context

func init() {
	ctx = context.Background()
}

func TestCatchNothing(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	defer func() {
		r := fc.req
		if r != nil {
			t.Errorf("got error report, expected none")
		}
	}()
	defer c.Catch(ctx)
}

func commonChecks(t *testing.T, req *erpb.ReportErrorEventRequest, panickingFunction string) {
	if req.Event.ServiceContext.Service != "myservice" {
		t.Errorf("error report didn't contain service name")
	}
	if req.Event.ServiceContext.Version != "v1.000" {
		t.Errorf("error report didn't contain version name")
	}
	if !strings.Contains(req.Event.Message, "hello, error") {
		t.Errorf("error report didn't contain message")
	}
	if !strings.Contains(req.Event.Message, panickingFunction) {
		t.Errorf("error report didn't contain stack trace")
	}
}

func TestCatchPanic(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	defer func() {
		r := fc.req
		if r == nil {
			t.Fatalf("got no error report, expected one")
		}
		commonChecks(t, r, "errors.TestCatchPanic")
		if !strings.Contains(r.Event.Message, "divide by zero") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	defer c.Catch(ctx, WithMessage("hello, error"))
	var x int
	x = x / x
}

func TestCatchPanicNilClient(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		recover()
		body := buf.String()
		if !strings.Contains(body, "divide by zero") {
			t.Errorf("error report didn't contain recovered value")
		}
		if !strings.Contains(body, "hello, error") {
			t.Errorf("error report didn't contain message")
		}
		if !strings.Contains(body, "TestCatchPanicNilClient") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	var c *Client
	defer c.Catch(ctx, WithMessage("hello, error"))
	var x int
	x = x / x
}

func TestLogFailedReports(t *testing.T) {
	fc := &fakeReportErrorsClient{fail: true}
	c := newTestClient(fc)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	defer func() {
		recover()
		body := buf.String()
		if !strings.Contains(body, "hello, error") {
			t.Errorf("error report didn't contain message")
		}
		if !strings.Contains(body, "errors.TestLogFailedReports") {
			t.Errorf("error report didn't contain stack trace")
		}
		if !strings.Contains(body, "divide by zero") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	defer c.Catch(ctx, WithMessage("hello, error"))
	var x int
	x = x / x
}

func TestCatchNilPanic(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	defer func() {
		r := fc.req
		if r == nil {
			t.Fatalf("got no error report, expected one")
		}
		commonChecks(t, r, "errors.TestCatchNilPanic")
		if !strings.Contains(r.Event.Message, "nil") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	b := true
	defer c.Catch(ctx, WithMessage("hello, error"), PanicFlag(&b))
	panic(nil)
}

func TestNotCatchNilPanic(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	defer func() {
		r := fc.req
		if r != nil {
			t.Errorf("got error report, expected none")
		}
	}()
	defer c.Catch(ctx, WithMessage("hello, error"))
	panic(nil)
}

func TestReport(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	c.Report(ctx, nil, "hello, ", "error")
	r := fc.req
	if r == nil {
		t.Fatalf("got no error report, expected one")
	}
	commonChecks(t, r, "errors.TestReport")
}

func TestReportf(t *testing.T) {
	fc := &fakeReportErrorsClient{}
	c := newTestClient(fc)
	c.Reportf(ctx, nil, "hello, error 2+%d=%d", 2, 2+2)
	r := fc.req
	if r == nil {
		t.Fatalf("got no error report, expected one")
	}
	commonChecks(t, r, "errors.TestReportf")
	if !strings.Contains(r.Event.Message, "2+2=4") {
		t.Errorf("error report didn't contain formatted message")
	}
}
