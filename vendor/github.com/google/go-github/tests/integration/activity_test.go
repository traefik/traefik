// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"

	"github.com/google/go-github/github"
)

const (
	owner = "google"
	repo  = "go-github"
)

func TestActivity_Starring(t *testing.T) {
	stargazers, _, err := client.Activity.ListStargazers(owner, repo, nil)
	if err != nil {
		t.Fatalf("Activity.ListStargazers returned error: %v", err)
	}

	if len(stargazers) == 0 {
		t.Errorf("Activity.ListStargazers(%q, %q) returned no stargazers", owner, repo)
	}

	// the rest of the tests requires auth
	if !checkAuth("TestActivity_Starring") {
		return
	}

	// first, check if already starred the target repository
	star, _, err := client.Activity.IsStarred(owner, repo)
	if err != nil {
		t.Fatalf("Activity.IsStarred returned error: %v", err)
	}
	if star {
		t.Fatalf("Already starring %v/%v. Please manually unstar it first.", owner, repo)
	}

	// star the target repository
	_, err = client.Activity.Star(owner, repo)
	if err != nil {
		t.Fatalf("Activity.Star returned error: %v", err)
	}

	// check again and verify starred
	star, _, err = client.Activity.IsStarred(owner, repo)
	if err != nil {
		t.Fatalf("Activity.IsStarred returned error: %v", err)
	}
	if !star {
		t.Fatalf("Not starred %v/%v after starring it.", owner, repo)
	}

	// unstar
	_, err = client.Activity.Unstar(owner, repo)
	if err != nil {
		t.Fatalf("Activity.Unstar returned error: %v", err)
	}

	// check again and verify not watching
	star, _, err = client.Activity.IsStarred(owner, repo)
	if err != nil {
		t.Fatalf("Activity.IsStarred returned error: %v", err)
	}
	if star {
		t.Fatalf("Still starred %v/%v after unstarring it.", owner, repo)
	}
}

func deleteSubscription(t *testing.T) {
	// delete subscription
	_, err := client.Activity.DeleteRepositorySubscription(owner, repo)
	if err != nil {
		t.Fatalf("Activity.DeleteRepositorySubscription returned error: %v", err)
	}

	// check again and verify not watching
	sub, _, err := client.Activity.GetRepositorySubscription(owner, repo)
	if err != nil {
		t.Fatalf("Activity.GetRepositorySubscription returned error: %v", err)
	}
	if sub != nil {
		t.Fatalf("Still watching %v/%v after deleting subscription.", owner, repo)
	}
}

func createSubscription(t *testing.T) {
	// watch the target repository
	sub := &github.Subscription{Subscribed: github.Bool(true)}
	_, _, err := client.Activity.SetRepositorySubscription(owner, repo, sub)
	if err != nil {
		t.Fatalf("Activity.SetRepositorySubscription returned error: %v", err)
	}

	// check again and verify watching
	sub, _, err = client.Activity.GetRepositorySubscription(owner, repo)
	if err != nil {
		t.Fatalf("Activity.GetRepositorySubscription returned error: %v", err)
	}
	if sub == nil || !*sub.Subscribed {
		t.Fatalf("Not watching %v/%v after setting subscription.", owner, repo)
	}
}

func TestActivity_Watching(t *testing.T) {
	watchers, _, err := client.Activity.ListWatchers(owner, repo, nil)
	if err != nil {
		t.Fatalf("Activity.ListWatchers returned error: %v", err)
	}

	if len(watchers) == 0 {
		t.Errorf("Activity.ListWatchers(%q, %q) returned no watchers", owner, repo)
	}

	// the rest of the tests requires auth
	if !checkAuth("TestActivity_Watching") {
		return
	}

	// first, check if already watching the target repository
	sub, _, err := client.Activity.GetRepositorySubscription(owner, repo)
	if err != nil {
		t.Fatalf("Activity.GetRepositorySubscription returned error: %v", err)
	}

	switch {
	case sub != nil: // If already subscribing, delete then recreate subscription.
		deleteSubscription(t)
		createSubscription(t)
	case sub == nil: // Otherwise, create subscription and then delete it.
		createSubscription(t)
		deleteSubscription(t)
	}
}
