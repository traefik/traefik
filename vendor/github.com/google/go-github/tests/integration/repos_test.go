// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

func TestRepositories_CRUD(t *testing.T) {
	if !checkAuth("TestRepositories_CRUD") {
		return
	}

	// get authenticated user
	me, _, err := client.Users.Get("")
	if err != nil {
		t.Fatalf("Users.Get('') returned error: %v", err)
	}

	repo, err := createRandomTestRepository(*me.Login, false)
	if err != nil {
		t.Fatalf("createRandomTestRepository returned error: %v", err)
	}

	// update the repository description
	repo.Description = github.String("description")
	repo.DefaultBranch = nil // FIXME: this shouldn't be necessary
	_, _, err = client.Repositories.Edit(*repo.Owner.Login, *repo.Name, repo)
	if err != nil {
		t.Fatalf("Repositories.Edit() returned error: %v", err)
	}

	// delete the repository
	_, err = client.Repositories.Delete(*repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Fatalf("Repositories.Delete() returned error: %v", err)
	}

	// verify that the repository was deleted
	_, resp, err := client.Repositories.Get(*repo.Owner.Login, *repo.Name)
	if err == nil {
		t.Fatalf("Test repository still exists after deleting it.")
	}
	if err != nil && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Repositories.Get() returned error: %v", err)
	}
}

func TestRepositories_BranchesTags(t *testing.T) {
	// branches
	branches, _, err := client.Repositories.ListBranches("git", "git", nil)
	if err != nil {
		t.Fatalf("Repositories.ListBranches() returned error: %v", err)
	}

	if len(branches) == 0 {
		t.Fatalf("Repositories.ListBranches('git', 'git') returned no branches")
	}

	_, _, err = client.Repositories.GetBranch("git", "git", *branches[0].Name)
	if err != nil {
		t.Fatalf("Repositories.GetBranch() returned error: %v", err)
	}

	// tags
	tags, _, err := client.Repositories.ListTags("git", "git", nil)
	if err != nil {
		t.Fatalf("Repositories.ListTags() returned error: %v", err)
	}

	if len(tags) == 0 {
		t.Fatalf("Repositories.ListTags('git', 'git') returned no tags")
	}
}

func TestRepositories_ServiceHooks(t *testing.T) {
	hooks, _, err := client.Repositories.ListServiceHooks()
	if err != nil {
		t.Fatalf("Repositories.ListServiceHooks() returned error: %v", err)
	}

	if len(hooks) == 0 {
		t.Fatalf("Repositories.ListServiceHooks() returned no hooks")
	}
}

func TestRepositories_EditBranches(t *testing.T) {
	if !checkAuth("TestRepositories_EditBranches") {
		return
	}

	// get authenticated user
	me, _, err := client.Users.Get("")
	if err != nil {
		t.Fatalf("Users.Get('') returned error: %v", err)
	}

	repo, err := createRandomTestRepository(*me.Login, true)
	if err != nil {
		t.Fatalf("createRandomTestRepository returned error: %v", err)
	}

	branch, _, err := client.Repositories.GetBranch(*repo.Owner.Login, *repo.Name, "master")
	if err != nil {
		t.Fatalf("Repositories.GetBranch() returned error: %v", err)
	}

	if *branch.Protected {
		t.Fatalf("Branch %v of repo %v is already protected", "master", *repo.Name)
	}

	protectionRequest := &github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			IncludeAdmins: true,
			Strict:        true,
			Contexts:      []string{"continuous-integration"},
		},
		RequiredPullRequestReviews: &github.RequiredPullRequestReviews{
			IncludeAdmins: true,
		},
		// TODO: Only organization repositories can have users and team restrictions.
		//       In order to be able to test these Restrictions, need to add support
		//       for creating temporary organization repositories.
		Restrictions: nil,
	}

	protection, _, err := client.Repositories.UpdateBranchProtection(*repo.Owner.Login, *repo.Name, "master", protectionRequest)
	if err != nil {
		t.Fatalf("Repositories.UpdateBranchProtection() returned error: %v", err)
	}

	want := &github.Protection{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			IncludeAdmins: true,
			Strict:        true,
			Contexts:      []string{"continuous-integration"},
		},
		RequiredPullRequestReviews: &github.RequiredPullRequestReviews{
			IncludeAdmins: true,
		},
		Restrictions: nil,
	}
	if !reflect.DeepEqual(protection, want) {
		t.Errorf("Repositories.UpdateBranchProtection() returned %+v, want %+v", protection, want)
	}

	_, err = client.Repositories.Delete(*repo.Owner.Login, *repo.Name)
	if err != nil {
		t.Fatalf("Repositories.Delete() returned error: %v", err)
	}
}

func TestRepositories_List(t *testing.T) {
	if !checkAuth("TestRepositories_List") {
		return
	}

	_, _, err := client.Repositories.List("", nil)
	if err != nil {
		t.Fatalf("Repositories.List('') returned error: %v", err)
	}

	_, _, err = client.Repositories.List("google", nil)
	if err != nil {
		t.Fatalf("Repositories.List('google') returned error: %v", err)
	}

	opt := github.RepositoryListOptions{Sort: "created"}
	repos, _, err := client.Repositories.List("google", &opt)
	if err != nil {
		t.Fatalf("Repositories.List('google') with Sort opt returned error: %v", err)
	}
	for i, repo := range repos {
		if i > 0 && (*repos[i-1].CreatedAt).Time.Before((*repo.CreatedAt).Time) {
			t.Fatalf("Repositories.List('google') with default descending Sort returned incorrect order")
		}
	}
}
