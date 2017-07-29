// Copyright 2017 The go-github AUTHORS. All rights reserved.
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

func TestRequestReviewers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"reviewers":["octocat","googlebot"]}`+"\n")
		fmt.Fprint(w, `{"number":1}`)
	})

	// This returns a PR, unmarshalling of which is tested elsewhere
	pull, _, err := client.PullRequests.RequestReviewers(context.Background(), "o", "r", 1, []string{"octocat", "googlebot"})
	if err != nil {
		t.Errorf("PullRequests.RequestReviewers returned error: %v", err)
	}
	want := &PullRequest{Number: Int(1)}
	if !reflect.DeepEqual(pull, want) {
		t.Errorf("PullRequests.RequestReviewers returned %+v, want %+v", pull, want)
	}
}

func TestRemoveReviewers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testBody(t, r, `{"reviewers":["octocat","googlebot"]}`+"\n")
	})

	_, err := client.PullRequests.RemoveReviewers(context.Background(), "o", "r", 1, []string{"octocat", "googlebot"})
	if err != nil {
		t.Errorf("PullRequests.RemoveReviewers returned error: %v", err)
	}
}

func TestListReviewers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"login":"octocat","id":1}]`)
	})

	reviewers, _, err := client.PullRequests.ListReviewers(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("PullRequests.ListReviewers returned error: %v", err)
	}

	want := []*User{
		{
			Login: String("octocat"),
			ID:    Int(1),
		},
	}
	if !reflect.DeepEqual(reviewers, want) {
		t.Errorf("PullRequests.ListReviewers returned %+v, want %+v", reviewers, want)
	}
}

func TestListReviewers_withOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/requested_reviewers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		fmt.Fprint(w, `[]`)
	})

	_, _, err := client.PullRequests.ListReviewers(context.Background(), "o", "r", 1, &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("PullRequests.ListReviewers returned error: %v", err)
	}
}
