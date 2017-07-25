// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build integration

package tests

import "testing"

func TestPullRequests_ListCommits(t *testing.T) {
	commits, _, err := client.PullRequests.ListCommits("google", "go-github", 2, nil)
	if err != nil {
		t.Fatalf("PullRequests.ListCommits() returned error: %v", err)
	}

	if got, want := len(commits), 3; got != want {
		t.Fatalf("PullRequests.ListCommits() returned %d commits, want %d", got, want)
	}

	if got, want := *commits[0].Author.Login, "sqs"; got != want {
		t.Fatalf("PullRequests.ListCommits()[0].Author.Login returned %v, want %v", got, want)
	}
}
