// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestRepositoriesService_ListProjects(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProjectsPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ProjectListOptions{ListOptions: ListOptions{Page: 2}}
	projects, _, err := client.Repositories.ListProjects(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListProjects returned error: %v", err)
	}

	want := []*Project{{ID: Int(1)}}
	if !reflect.DeepEqual(projects, want) {
		t.Errorf("Repositories.ListProjects returned %+v, want %+v", projects, want)
	}
}

func TestRepositoriesService_CreateProject(t *testing.T) {
	setup()
	defer teardown()

	input := &ProjectOptions{Name: "Project Name", Body: "Project body."}

	mux.HandleFunc("/repos/o/r/projects", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeProjectsPreview)

		v := &ProjectOptions{}
		json.NewDecoder(r.Body).Decode(v)
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	project, _, err := client.Repositories.CreateProject(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Repositories.CreateProject returned error: %v", err)
	}

	want := &Project{ID: Int(1)}
	if !reflect.DeepEqual(project, want) {
		t.Errorf("Repositories.CreateProject returned %+v, want %+v", project, want)
	}
}
