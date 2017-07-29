// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestReactionsService_ListCommentReactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"user":{"login":"l","id":2},"content":"+1"}]`))
	})

	got, _, err := client.Reactions.ListCommentReactions(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("ListCommentReactions returned error: %v", err)
	}
	if want := []*Reaction{{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListCommentReactions = %+v, want %+v", got, want)
	}
}

func TestReactionsService_CreateCommentReaction(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"user":{"login":"l","id":2},"content":"+1"}`))
	})

	got, _, err := client.Reactions.CreateCommentReaction(context.Background(), "o", "r", 1, "+1")
	if err != nil {
		t.Errorf("CreateCommentReaction returned error: %v", err)
	}
	want := &Reaction{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateCommentReaction = %+v, want %+v", got, want)
	}
}

func TestReactionsService_ListIssueReactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"user":{"login":"l","id":2},"content":"+1"}]`))
	})

	got, _, err := client.Reactions.ListIssueReactions(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("ListIssueReactions returned error: %v", err)
	}
	if want := []*Reaction{{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListIssueReactions = %+v, want %+v", got, want)
	}
}

func TestReactionsService_CreateIssueReaction(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"user":{"login":"l","id":2},"content":"+1"}`))
	})

	got, _, err := client.Reactions.CreateIssueReaction(context.Background(), "o", "r", 1, "+1")
	if err != nil {
		t.Errorf("CreateIssueReaction returned error: %v", err)
	}
	want := &Reaction{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateIssueReaction = %+v, want %+v", got, want)
	}
}

func TestReactionsService_ListIssueCommentReactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"user":{"login":"l","id":2},"content":"+1"}]`))
	})

	got, _, err := client.Reactions.ListIssueCommentReactions(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("ListIssueCommentReactions returned error: %v", err)
	}
	if want := []*Reaction{{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListIssueCommentReactions = %+v, want %+v", got, want)
	}
}

func TestReactionsService_CreateIssueCommentReaction(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/issues/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"user":{"login":"l","id":2},"content":"+1"}`))
	})

	got, _, err := client.Reactions.CreateIssueCommentReaction(context.Background(), "o", "r", 1, "+1")
	if err != nil {
		t.Errorf("CreateIssueCommentReaction returned error: %v", err)
	}
	want := &Reaction{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateIssueCommentReaction = %+v, want %+v", got, want)
	}
}

func TestReactionsService_ListPullRequestCommentReactions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"user":{"login":"l","id":2},"content":"+1"}]`))
	})

	got, _, err := client.Reactions.ListPullRequestCommentReactions(context.Background(), "o", "r", 1, nil)
	if err != nil {
		t.Errorf("ListPullRequestCommentReactions returned error: %v", err)
	}
	if want := []*Reaction{{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListPullRequestCommentReactions = %+v, want %+v", got, want)
	}
}

func TestReactionsService_CreatePullRequestCommentReaction(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/pulls/comments/1/reactions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"user":{"login":"l","id":2},"content":"+1"}`))
	})

	got, _, err := client.Reactions.CreatePullRequestCommentReaction(context.Background(), "o", "r", 1, "+1")
	if err != nil {
		t.Errorf("CreatePullRequestCommentReaction returned error: %v", err)
	}
	want := &Reaction{ID: Int(1), User: &User{Login: String("l"), ID: Int(2)}, Content: String("+1")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreatePullRequestCommentReaction = %+v, want %+v", got, want)
	}
}

func TestReactionsService_DeleteReaction(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/reactions/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeReactionsPreview)

		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Reactions.DeleteReaction(context.Background(), 1); err != nil {
		t.Errorf("DeleteReaction returned error: %v", err)
	}
}
