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

	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

type fakeLogger struct {
	entry *logging.Entry
	fail  bool
}

func (c *fakeLogger) LogSync(ctx context.Context, e logging.Entry) error {
	if c.fail {
		return errors.New("request failed")
	}
	c.entry = &e
	return nil
}

func newTestClientUsingLogging(c *fakeLogger) *Client {
	newLoggerInterface = func(ctx context.Context, project string, opts ...option.ClientOption) (loggerInterface, error) {
		return c, nil
	}
	t, err := NewClient(context.Background(), testProjectID, "myservice", "v1.000", true)
	if err != nil {
		panic(err)
	}
	t.RepanicDefault = false
	return t
}

func TestCatchNothingUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	defer func() {
		e := fl.entry
		if e != nil {
			t.Errorf("got error report, expected none")
		}
	}()
	defer c.Catch(ctx)
}

func entryMessage(e *logging.Entry) string {
	return e.Payload.(map[string]interface{})["message"].(string)
}

func commonLoggingChecks(t *testing.T, e *logging.Entry, panickingFunction string) {
	if e.Payload.(map[string]interface{})["serviceContext"].(map[string]string)["service"] != "myservice" {
		t.Errorf("error report didn't contain service name")
	}
	if e.Payload.(map[string]interface{})["serviceContext"].(map[string]string)["version"] != "v1.000" {
		t.Errorf("error report didn't contain version name")
	}
	if !strings.Contains(entryMessage(e), "hello, error") {
		t.Errorf("error report didn't contain message")
	}
	if !strings.Contains(entryMessage(e), panickingFunction) {
		t.Errorf("error report didn't contain stack trace")
	}
}

func TestCatchPanicUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	defer func() {
		e := fl.entry
		if e == nil {
			t.Fatalf("got no error report, expected one")
		}
		commonLoggingChecks(t, e, "TestCatchPanic")
		if !strings.Contains(entryMessage(e), "divide by zero") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	defer c.Catch(ctx, WithMessage("hello, error"))
	var x int
	x = x / x
}

func TestCatchPanicNilClientUsingLogging(t *testing.T) {
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

func TestLogFailedReportsUsingLogging(t *testing.T) {
	fl := &fakeLogger{fail: true}
	c := newTestClientUsingLogging(fl)
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

func TestCatchNilPanicUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	defer func() {
		e := fl.entry
		if e == nil {
			t.Fatalf("got no error report, expected one")
		}
		commonLoggingChecks(t, e, "TestCatchNilPanic")
		if !strings.Contains(entryMessage(e), "nil") {
			t.Errorf("error report didn't contain recovered value")
		}
	}()
	b := true
	defer c.Catch(ctx, WithMessage("hello, error"), PanicFlag(&b))
	panic(nil)
}

func TestNotCatchNilPanicUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	defer func() {
		e := fl.entry
		if e != nil {
			t.Errorf("got error report, expected none")
		}
	}()
	defer c.Catch(ctx, WithMessage("hello, error"))
	panic(nil)
}

func TestReportUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	c.Report(ctx, nil, "hello, ", "error")
	e := fl.entry
	if e == nil {
		t.Fatalf("got no error report, expected one")
	}
	commonLoggingChecks(t, e, "TestReport")
}

func TestReportfUsingLogging(t *testing.T) {
	fl := &fakeLogger{}
	c := newTestClientUsingLogging(fl)
	c.Reportf(ctx, nil, "hello, error 2+%d=%d", 2, 2+2)
	e := fl.entry
	if e == nil {
		t.Fatalf("got no error report, expected one")
	}
	commonLoggingChecks(t, e, "TestReportf")
	if !strings.Contains(entryMessage(e), "2+2=4") {
		t.Errorf("error report didn't contain formatted message")
	}
}
