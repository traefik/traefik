// Copyright 2014 The go-github AUTHORS. All rights reserved.
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

func TestRepositoriesService_GetPagesInfo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePagesPreview)
		fmt.Fprint(w, `{"url":"u","status":"s","cname":"c","custom_404":false,"html_url":"h"}`)
	})

	page, _, err := client.Repositories.GetPagesInfo(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.GetPagesInfo returned error: %v", err)
	}

	want := &Pages{URL: String("u"), Status: String("s"), CNAME: String("c"), Custom404: Bool(false), HTMLURL: String("h")}
	if !reflect.DeepEqual(page, want) {
		t.Errorf("Repositories.GetPagesInfo returned %+v, want %+v", page, want)
	}
}

func TestRepositoriesService_ListPagesBuilds(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"url":"u","status":"s","commit":"c"}]`)
	})

	pages, _, err := client.Repositories.ListPagesBuilds(context.Background(), "o", "r", nil)
	if err != nil {
		t.Errorf("Repositories.ListPagesBuilds returned error: %v", err)
	}

	want := []*PagesBuild{{URL: String("u"), Status: String("s"), Commit: String("c")}}
	if !reflect.DeepEqual(pages, want) {
		t.Errorf("Repositories.ListPagesBuilds returned %+v, want %+v", pages, want)
	}
}

func TestRepositoriesService_ListPagesBuilds_withOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		fmt.Fprint(w, `[]`)
	})

	_, _, err := client.Repositories.ListPagesBuilds(context.Background(), "o", "r", &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("Repositories.ListPagesBuilds returned error: %v", err)
	}
}

func TestRepositoriesService_GetLatestPagesBuild(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages/builds/latest", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"url":"u","status":"s","commit":"c"}`)
	})

	build, _, err := client.Repositories.GetLatestPagesBuild(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.GetLatestPagesBuild returned error: %v", err)
	}

	want := &PagesBuild{URL: String("u"), Status: String("s"), Commit: String("c")}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Repositories.GetLatestPagesBuild returned %+v, want %+v", build, want)
	}
}

func TestRepositoriesService_GetPageBuild(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages/builds/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"url":"u","status":"s","commit":"c"}`)
	})

	build, _, err := client.Repositories.GetPageBuild(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Repositories.GetPageBuild returned error: %v", err)
	}

	want := &PagesBuild{URL: String("u"), Status: String("s"), Commit: String("c")}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Repositories.GetPageBuild returned %+v, want %+v", build, want)
	}
}

func TestRepositoriesService_RequestPageBuild(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pages/builds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypePagesPreview)
		fmt.Fprint(w, `{"url":"u","status":"s"}`)
	})

	build, _, err := client.Repositories.RequestPageBuild(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.RequestPageBuild returned error: %v", err)
	}

	want := &PagesBuild{URL: String("u"), Status: String("s")}
	if !reflect.DeepEqual(build, want) {
		t.Errorf("Repositories.RequestPageBuild returned %+v, want %+v", build, want)
	}
}
