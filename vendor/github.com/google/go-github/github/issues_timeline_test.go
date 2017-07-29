// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestIssuesService_ListIssueTimeline(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/timeline", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeTimelinePreview)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 1, PerPage: 2}
	events, _, err := client.Issues.ListIssueTimeline(context.Background(), "o", "r", 1, opt)
	if err != nil {
		t.Errorf("Issues.ListIssueTimeline returned error: %v", err)
	}

	want := []*Timeline{{ID: Int(1)}}
	if !reflect.DeepEqual(events, want) {
		t.Errorf("Issues.ListIssueTimeline = %+v, want %+v", events, want)
	}
}
