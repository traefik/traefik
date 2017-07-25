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

func TestIssuesService_ListComments_allIssues(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/comments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		testFormValues(t, r, values{
			"sort":      "updated",
			"direction": "desc",
			"since":     "2002-02-10T15:30:00Z",
			"page":      "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &IssueListCommentsOptions{
		Sort:        "updated",
		Direction:   "desc",
		Since:       time.Date(2002, time.February, 10, 15, 30, 0, 0, time.UTC),
		ListOptions: ListOptions{Page: 2},
	}
	comments, _, err := client.Issues.ListComments(context.Background(), "o", "r", 0, opt)
	if err != nil {
		t.Errorf("Issues.ListComments returned error: %v", err)
	}

	want := []*IssueComment{{ID: Int(1)}}
	if !reflect.DeepEqual(comments, want) {
		t.Errorf("Issues.ListComments returned %+v, want %+v", comments, want)
	}
}

func TestIssuesService_ListComments_specificIssue(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		fmt.Fprint(w, `[{"id":1}]`)
	})

	comments, _, err := client.Issues.ListComments(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("Issues.ListComments returned error: %v", err)
	}

	want := []*IssueComment{{ID: Int(1)}}
	if !reflect.DeepEqual(comments, want) {
		t.Errorf("Issues.ListComments returned %+v, want %+v", comments, want)
	}
}

func TestIssuesService_ListComments_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.ListComments(context.Background(), "%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestIssuesService_GetComment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/comments/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)
		fmt.Fprint(w, `{"id":1}`)
	})

	comment, _, err := client.Issues.GetComment(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Issues.GetComment returned error: %v", err)
	}

	want := &IssueComment{ID: Int(1)}
	if !reflect.DeepEqual(comment, want) {
		t.Errorf("Issues.GetComment returned %+v, want %+v", comment, want)
	}
}

func TestIssuesService_GetComment_invalidOrg(t *testing.T) {
	_, _, err := client.Issues.GetComment(context.Background(), "%", "r", 1)
	testURLParseError(t, err)
}

func TestIssuesService_CreateComment(t *testing.T) {
	setup()
	defer teardown()

	input := &IssueComment{Body: String("b")}

	mux.HandleFunc("/repos/o/r/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		v := new(IssueComment)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	comment, _, err := client.Issues.CreateComment(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Issues.CreateComment returned error: %v", err)
	}

	want := &IssueComment{ID: Int(1)}
	if !reflect.DeepEqual(comment, want) {
		t.Errorf("Issues.CreateComment returned %+v, want %+v", comment, want)
	}
}

func TestIssuesService_CreateComment_invalidOrg(t *testing.T) {
	_, _, err := client.Issues.CreateComment(context.Background(), "%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestIssuesService_EditComment(t *testing.T) {
	setup()
	defer teardown()

	input := &IssueComment{Body: String("b")}

	mux.HandleFunc("/repos/o/r/issues/comments/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(IssueComment)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	comment, _, err := client.Issues.EditComment(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Issues.EditComment returned error: %v", err)
	}

	want := &IssueComment{ID: Int(1)}
	if !reflect.DeepEqual(comment, want) {
		t.Errorf("Issues.EditComment returned %+v, want %+v", comment, want)
	}
}

func TestIssuesService_EditComment_invalidOwner(t *testing.T) {
	_, _, err := client.Issues.EditComment(context.Background(), "%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestIssuesService_DeleteComment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/comments/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Issues.DeleteComment(context.Background(), "o", "r", 1)
	if err != nil {
		t.Errorf("Issues.DeleteComments returned error: %v", err)
	}
}

func TestIssuesService_DeleteComment_invalidOwner(t *testing.T) {
	_, err := client.Issues.DeleteComment(context.Background(), "%", "r", 1)
	testURLParseError(t, err)
}
