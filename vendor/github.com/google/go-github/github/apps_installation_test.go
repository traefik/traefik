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

func TestAppsService_ListRepos(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/installation/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeIntegrationPreview)
		testFormValues(t, r, values{
			"page":     "1",
			"per_page": "2",
		})
		fmt.Fprint(w, `{"repositories": [{"id":1}]}`)
	})

	opt := &ListOptions{Page: 1, PerPage: 2}
	repositories, _, err := client.Apps.ListRepos(context.Background(), opt)
	if err != nil {
		t.Errorf("Apps.ListRepos returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(repositories, want) {
		t.Errorf("Apps.ListRepos returned %+v, want %+v", repositories, want)
	}
}
