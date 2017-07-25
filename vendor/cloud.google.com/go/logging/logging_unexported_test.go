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

// Tests that require access to unexported names of the logging package.

package logging

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	durpb "github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/api/support/bundler"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	logtypepb "google.golang.org/genproto/googleapis/logging/type"
)

func TestLoggerCreation(t *testing.T) {
	const logID = "testing"
	c := &Client{projectID: "PROJECT_ID"}
	defaultResource := &mrpb.MonitoredResource{Type: "global"}
	defaultBundler := &bundler.Bundler{
		DelayThreshold:       DefaultDelayThreshold,
		BundleCountThreshold: DefaultEntryCountThreshold,
		BundleByteThreshold:  DefaultEntryByteThreshold,
		BundleByteLimit:      0,
		BufferedByteLimit:    DefaultBufferedByteLimit,
	}
	for _, test := range []struct {
		options     []LoggerOption
		wantLogger  *Logger
		wantBundler *bundler.Bundler
	}{
		{nil, &Logger{commonResource: defaultResource}, defaultBundler},
		{
			[]LoggerOption{CommonResource(nil), CommonLabels(map[string]string{"a": "1"})},
			&Logger{commonResource: nil, commonLabels: map[string]string{"a": "1"}},
			defaultBundler,
		},
		{
			[]LoggerOption{DelayThreshold(time.Minute), EntryCountThreshold(99),
				EntryByteThreshold(17), EntryByteLimit(18), BufferedByteLimit(19)},
			&Logger{commonResource: defaultResource},
			&bundler.Bundler{
				DelayThreshold:       time.Minute,
				BundleCountThreshold: 99,
				BundleByteThreshold:  17,
				BundleByteLimit:      18,
				BufferedByteLimit:    19,
			},
		},
	} {
		gotLogger := c.Logger(logID, test.options...)
		if got, want := gotLogger.commonResource, test.wantLogger.commonResource; !reflect.DeepEqual(got, want) {
			t.Errorf("%v: resource: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.commonLabels, test.wantLogger.commonLabels; !reflect.DeepEqual(got, want) {
			t.Errorf("%v: commonLabels: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.bundler.DelayThreshold, test.wantBundler.DelayThreshold; got != want {
			t.Errorf("%v: DelayThreshold: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.bundler.BundleCountThreshold, test.wantBundler.BundleCountThreshold; got != want {
			t.Errorf("%v: BundleCountThreshold: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.bundler.BundleByteThreshold, test.wantBundler.BundleByteThreshold; got != want {
			t.Errorf("%v: BundleByteThreshold: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.bundler.BundleByteLimit, test.wantBundler.BundleByteLimit; got != want {
			t.Errorf("%v: BundleByteLimit: got %v, want %v", test.options, got, want)
		}
		if got, want := gotLogger.bundler.BufferedByteLimit, test.wantBundler.BufferedByteLimit; got != want {
			t.Errorf("%v: BufferedByteLimit: got %v, want %v", test.options, got, want)
		}
	}
}

func TestToProtoStruct(t *testing.T) {
	v := struct {
		Foo string                 `json:"foo"`
		Bar int                    `json:"bar,omitempty"`
		Baz []float64              `json:"baz"`
		Moo map[string]interface{} `json:"moo"`
	}{
		Foo: "foovalue",
		Baz: []float64{1.1},
		Moo: map[string]interface{}{
			"a": 1,
			"b": "two",
			"c": true,
		},
	}

	got, err := toProtoStruct(v)
	if err != nil {
		t.Fatal(err)
	}
	want := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"foo": {Kind: &structpb.Value_StringValue{v.Foo}},
			"baz": {Kind: &structpb.Value_ListValue{&structpb.ListValue{
				[]*structpb.Value{{Kind: &structpb.Value_NumberValue{1.1}}}}}},
			"moo": {Kind: &structpb.Value_StructValue{
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"a": {Kind: &structpb.Value_NumberValue{1}},
						"b": {Kind: &structpb.Value_StringValue{"two"}},
						"c": {Kind: &structpb.Value_BoolValue{true}},
					},
				},
			}},
		},
	}
	if !proto.Equal(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}

	// Non-structs should fail to convert.
	for v := range []interface{}{3, "foo", []int{1, 2, 3}} {
		_, err := toProtoStruct(v)
		if err == nil {
			t.Errorf("%v: got nil, want error", v)
		}
	}

	// Test fast path.
	got, err = toProtoStruct(want)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("got and want should be identical, but are not")
	}
}

func TestFromHTTPRequest(t *testing.T) {
	const testURL = "http:://example.com/path?q=1"
	u, err := url.Parse(testURL)
	if err != nil {
		t.Fatal(err)
	}
	req := &HTTPRequest{
		Request: &http.Request{
			Method: "GET",
			URL:    u,
			Header: map[string][]string{
				"User-Agent": []string{"user-agent"},
				"Referer":    []string{"referer"},
			},
		},
		RequestSize:                    100,
		Status:                         200,
		ResponseSize:                   25,
		Latency:                        100 * time.Second,
		RemoteIP:                       "127.0.0.1",
		CacheHit:                       true,
		CacheValidatedWithOriginServer: true,
	}
	got := fromHTTPRequest(req)
	want := &logtypepb.HttpRequest{
		RequestMethod:                  "GET",
		RequestUrl:                     testURL,
		RequestSize:                    100,
		Status:                         200,
		ResponseSize:                   25,
		Latency:                        &durpb.Duration{Seconds: 100},
		UserAgent:                      "user-agent",
		RemoteIp:                       "127.0.0.1",
		Referer:                        "referer",
		CacheHit:                       true,
		CacheValidatedWithOriginServer: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

// Used by the tests in logging_test.
func SetNow(f func() time.Time) {
	now = f
}
