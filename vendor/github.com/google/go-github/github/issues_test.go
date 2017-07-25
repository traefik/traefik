// Copyright 2013 The go-github AUTHORS. All rights reserved.
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
	"time"
)

func TestIssuesService_List_all(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		testFormValues(t, r, values{
			"filter":    "all",
			"state":     "closed",
			"labels":    "a,b",
			"sort":      "updated",
			"direction": "asc",
			"since":     "2002-02-10T15:30:00Z",
			"page":      "1",
			"per_page":  "2",
		})
		fmt.Fprint(w, `[{"number":1}]`)
	})

	opt := &IssueListOptions{
		"all", "closed", []string{"a", "b"}, "updated", "asc",
		time.Date(2002, time.February, 10, 15, 30, 0, 0, time.UTC),
		ListOptions{Page: 1, PerPage: 2},
	}
	issues, _, err := client.Issues.List(context.Background(), true, opt)
	if err != nil {
		t.Errorf("Issues.List returned error: %v", err)
	}

	want := []*Issue{{Number: Int(1)}}
	if !reflect.DeepEqual(issues, want) {
		t.Errorf("Issues.List returned %+v, want %+v", issues, want)
	}
}

func TestIssuesService_List_owned(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		fmt.Fprint(w, `[{"number":1}]`)
	})

	issues, _, err := client.Issues.List(context.Background(), false, nil)
	if err != nil {
		t.Errorf("Issues.List returned error: %v", err)
	}

	want := []*Issue{{Number: Int(1)}}
	if !reflect.DeepEqual(issues, want) {
		t.Errorf("Issues.List returned %+v, want %+v", issues, want)
	}
}

func TestIssuesService_ListByOrg(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		fmt.Fprint(w, `[{"number":1}]`)
	})

	issues, _, err := client.Issues.ListByOrg(context.Background(), "o", nil)
	if err != nil {
		t.Errorf("Issues.ListByOrg returned error: %v", err)
	}

	want := []*Issue{{Number: Int(1)}}
	if !reflect.DeepEqual(issues, want) {
		t.Errorf("Issues.List returned %+v, want %+v", issues, want)
	}
}

func TestIssuesService_ListByOrg_invalidOrg(t *testing.T) {
	_, _, err := client.Issues.ListByOrg(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestIssuesService_ListByRepo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		testFormValues(t, r, values{
			"milestone": "*",
			"state":     "closed",
			"assignee":  "a",
			"creator":   "c",
			"mentioned": "m",
			"labels":    "a,b",
			"sort":      "updated",
			"direction": "asc",
			"since":     "2002-02-10T15:30:00Z",
		})
		fmt.Fprint(w, `[{"number":1}]`)
	})

	opt := &IssueListByRepoOptions{
		"*", "closed", "a", "c", "m", []string{"a", "b"}, "updated", "asc",
		time.Date(2002, time.February, 10, 15, 30, 0, 0, time.UTC),
		ListOptions{0, 0},
	}
	issues, _, err := client.Issues.ListByRepo(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Issues.ListByOrg returned error: %v", err)
	}

	want := []*Issue{{Number: Int(1)}}
	if !reflect.DeepEqual(issues, want) {
		t.Errorf("Issues.List returned %+v, want %+v", issues, want)
	}
}

func TestIssuesService_ListByRepo_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.ListByRepo(context.Background(), "%", "r", nil)
	testURLParseError(t, err)
}

func TestIssuesService_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		fmt.Fprint(w, `{"number":1, "labels": [{"url": "u", "name": "n", "color": "c"}]}`)
	})

	issue, _, err := client.Issues.Get(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Issues.Get returned error: %v", err)
	}

	want := &Issue{
		Number: Int(1),
		Labels: []Label{{
			URL:   String("u"),
			Name:  String("n"),
			Color: String("c"),
		}},
	}
	if !reflect.DeepEqual(issue, want) {
		t.Errorf("Issues.Get returned %+v, want %+v", issue, want)
	}
}

func TestIssuesService_Get_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.Get(context.Background(), "%", "r", 1)
	testURLParseError(t, err)
}

func TestIssuesService_Create(t *testing.T) {
	setup()
	defer teardown()

	input := &IssueRequest{
		Title:    String("t"),
		Body:     String("b"),
		Assignee: String("a"),
		Labels:   &[]string{"l1", "l2"},
	}

	mux.HandleFunc("/repos/o/r/issues", func(w http.ResponseWriter, r *http.Request) {
		v := new(IssueRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1}`)
	})

	issue, _, err := client.Issues.Create(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Issues.Create returned error: %v", err)
	}

	want := &Issue{Number: Int(1)}
	if !reflect.DeepEqual(issue, want) {
		t.Errorf("Issues.Create returned %+v, want %+v", issue, want)
	}
}

func TestIssuesService_Create_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.Create(context.Background(), "%", "r", nil)
	testURLParseError(t, err)
}

func TestIssuesService_Edit(t *testing.T) {
	setup()
	defer teardown()

	input := &IssueRequest{Title: String("t")}

	mux.HandleFunc("/repos/o/r/issues/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(IssueRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1}`)
	})

	issue, _, err := client.Issues.Edit(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Issues.Edit returned error: %v", err)
	}

	want := &Issue{Number: Int(1)}
	if !reflect.DeepEqual(issue, want) {
		t.Errorf("Issues.Edit returned %+v, want %+v", issue, want)
	}
}

func TestIssuesService_Edit_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.Edit(context.Background(), "%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestIssuesService_Lock(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/lock", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")

		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Issues.Lock(context.Background(), "o", "r", 1); err != nil {
		t.Errorf("Issues.Lock returned error: %v", err)
	}
}

func TestIssuesService_Unlock(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/lock", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Issues.Unlock(context.Background(), "o", "r", 1); err != nil {
		t.Errorf("Issues.Unlock returned error: %v", err)
	}
}
