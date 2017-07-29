// Copyright 2016 The go-github AUTHORS. All rights reserved.
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

func TestPullRequestsService_ListReviews(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListOptions{Page: 2}
	reviews, _, err := client.PullRequests.ListReviews(context.Background(), "o", "r", 1, opt)
	if err != nil {
		t.Errorf("PullRequests.ListReviews returned error: %v", err)
	}

	want := []*PullRequestReview{
		{ID: Int(1)},
		{ID: Int(2)},
	}
	if !reflect.DeepEqual(reviews, want) {
		t.Errorf("PullRequests.ListReviews returned %+v, want %+v", reviews, want)
	}
}

func TestPullRequestsService_ListReviews_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.ListReviews(context.Background(), "%", "r", 1, nil)
	testURLParseError(t, err)
}

func TestPullRequestsService_GetReview(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	review, _, err := client.PullRequests.GetReview(context.Background(), "o", "r", 1, 1)
	if err != nil {
		t.Errorf("PullRequests.GetReview returned error: %v", err)
	}

	want := &PullRequestReview{ID: Int(1)}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("PullRequests.GetReview returned %+v, want %+v", review, want)
	}
}

func TestPullRequestsService_GetReview_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.GetReview(context.Background(), "%", "r", 1, 1)
	testURLParseError(t, err)
}

func TestPullRequestsService_DeletePendingReview(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"id":1}`)
	})

	review, _, err := client.PullRequests.DeletePendingReview(context.Background(), "o", "r", 1, 1)
	if err != nil {
		t.Errorf("PullRequests.DeletePendingReview returned error: %v", err)
	}

	want := &PullRequestReview{ID: Int(1)}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("PullRequests.DeletePendingReview returned %+v, want %+v", review, want)
	}
}

func TestPullRequestsService_DeletePendingReview_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.DeletePendingReview(context.Background(), "%", "r", 1, 1)
	testURLParseError(t, err)
}

func TestPullRequestsService_ListReviewComments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1/comments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	comments, _, err := client.PullRequests.ListReviewComments(context.Background(), "o", "r", 1, 1, nil)
	if err != nil {
		t.Errorf("PullRequests.ListReviewComments returned error: %v", err)
	}

	want := []*PullRequestComment{
		{ID: Int(1)},
		{ID: Int(2)},
	}
	if !reflect.DeepEqual(comments, want) {
		t.Errorf("PullRequests.ListReviewComments returned %+v, want %+v", comments, want)
	}
}

func TestPullRequestsService_ListReviewComments_withOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1/comments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		fmt.Fprint(w, `[]`)
	})

	_, _, err := client.PullRequests.ListReviewComments(context.Background(), "o", "r", 1, 1, &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("PullRequests.ListReviewComments returned error: %v", err)
	}
}

func TestPullRequestsService_ListReviewComments_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.ListReviewComments(context.Background(), "%", "r", 1, 1, nil)
	testURLParseError(t, err)
}

func TestPullRequestsService_CreateReview(t *testing.T) {
	setup()
	defer teardown()

	input := &PullRequestReviewRequest{
		Body:  String("b"),
		Event: String("APPROVE"),
	}

	mux.HandleFunc("/repos/o/r/pulls/1/reviews", func(w http.ResponseWriter, r *http.Request) {
		v := new(PullRequestReviewRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	review, _, err := client.PullRequests.CreateReview(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("PullRequests.CreateReview returned error: %v", err)
	}

	want := &PullRequestReview{ID: Int(1)}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("PullRequests.CreateReview returned %+v, want %+v", review, want)
	}
}

func TestPullRequestsService_CreateReview_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.CreateReview(context.Background(), "%", "r", 1, &PullRequestReviewRequest{})
	testURLParseError(t, err)
}

func TestPullRequestsService_SubmitReview(t *testing.T) {
	setup()
	defer teardown()

	input := &PullRequestReviewRequest{
		Body:  String("b"),
		Event: String("APPROVE"),
	}

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1/events", func(w http.ResponseWriter, r *http.Request) {
		v := new(PullRequestReviewRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	review, _, err := client.PullRequests.SubmitReview(context.Background(), "o", "r", 1, 1, input)
	if err != nil {
		t.Errorf("PullRequests.SubmitReview returned error: %v", err)
	}

	want := &PullRequestReview{ID: Int(1)}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("PullRequests.SubmitReview returned %+v, want %+v", review, want)
	}
}

func TestPullRequestsService_SubmitReview_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.SubmitReview(context.Background(), "%", "r", 1, 1, &PullRequestReviewRequest{})
	testURLParseError(t, err)
}

func TestPullRequestsService_DismissReview(t *testing.T) {
	setup()
	defer teardown()

	input := &PullRequestReviewDismissalRequest{Message: String("m")}

	mux.HandleFunc("/repos/o/r/pulls/1/reviews/1/dismissals", func(w http.ResponseWriter, r *http.Request) {
		v := new(PullRequestReviewDismissalRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	review, _, err := client.PullRequests.DismissReview(context.Background(), "o", "r", 1, 1, input)
	if err != nil {
		t.Errorf("PullRequests.DismissReview returned error: %v", err)
	}

	want := &PullRequestReview{ID: Int(1)}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("PullRequests.DismissReview returned %+v, want %+v", review, want)
	}
}

func TestPullRequestsService_DismissReview_invalidOwner(t *testing.T) {
	_, _, err := client.PullRequests.DismissReview(context.Background(), "%", "r", 1, 1, &PullRequestReviewDismissalRequest{})
	testURLParseError(t, err)
}
