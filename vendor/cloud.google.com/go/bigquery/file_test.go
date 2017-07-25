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

package bigquery

import (
	"reflect"
	"testing"

	"cloud.google.com/go/internal/pretty"
	bq "google.golang.org/api/bigquery/v2"
)

func TestQuote(t *testing.T) {
	ptr := func(s string) *string { return &s }

	for _, test := range []struct {
		quote string
		force bool
		want  *string
	}{
		{"", false, nil},
		{"", true, ptr("")},
		{"-", false, ptr("-")},
		{"-", true, ptr("")},
	} {
		fc := FileConfig{
			Quote:          test.quote,
			ForceZeroQuote: test.force,
		}
		got := fc.quote()
		if (got == nil) != (test.want == nil) {
			t.Errorf("%+v\ngot %v\nwant %v", test, pretty.Value(got), pretty.Value(test.want))
		}
		if got != nil && test.want != nil && *got != *test.want {
			t.Errorf("%+v: got %q, want %q", test, *got, *test.want)
		}
	}
}

func TestPopulateLoadConfig(t *testing.T) {
	hyphen := "-"
	fc := FileConfig{
		SourceFormat:        CSV,
		FieldDelimiter:      "\t",
		SkipLeadingRows:     8,
		AllowJaggedRows:     true,
		AllowQuotedNewlines: true,
		Encoding:            UTF_8,
		MaxBadRecords:       7,
		IgnoreUnknownValues: true,
		Schema: Schema{
			stringFieldSchema(),
			nestedFieldSchema(),
		},
		Quote: hyphen,
	}
	want := &bq.JobConfigurationLoad{
		SourceFormat:        "CSV",
		FieldDelimiter:      "\t",
		SkipLeadingRows:     8,
		AllowJaggedRows:     true,
		AllowQuotedNewlines: true,
		Encoding:            "UTF-8",
		MaxBadRecords:       7,
		IgnoreUnknownValues: true,
		Schema: &bq.TableSchema{
			Fields: []*bq.TableFieldSchema{
				bqStringFieldSchema(),
				bqNestedFieldSchema(),
			}},
		Quote: &hyphen,
	}
	got := &bq.JobConfigurationLoad{}
	fc.populateLoadConfig(got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%v\nwant:\n%v", pretty.Value(got), pretty.Value(want))
	}
}
