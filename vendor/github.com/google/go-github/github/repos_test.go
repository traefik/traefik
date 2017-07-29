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
	"strings"
	"testing"
)

func TestRepositoriesService_List_authenticatedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeLicensesPreview)
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	repos, _, err := client.Repositories.List(context.Background(), "", nil)
	if err != nil {
		t.Errorf("Repositories.List returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}, {ID: Int(2)}}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.List returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_List_specifiedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeLicensesPreview)
		testFormValues(t, r, values{
			"visibility":  "public",
			"affiliation": "owner,collaborator",
			"sort":        "created",
			"direction":   "asc",
			"page":        "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &RepositoryListOptions{
		Visibility:  "public",
		Affiliation: "owner,collaborator",
		Sort:        "created",
		Direction:   "asc",
		ListOptions: ListOptions{Page: 2},
	}
	repos, _, err := client.Repositories.List(context.Background(), "u", opt)
	if err != nil {
		t.Errorf("Repositories.List returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.List returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_List_specifiedUser_type(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeLicensesPreview)
		testFormValues(t, r, values{
			"type": "owner",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &RepositoryListOptions{
		Type: "owner",
	}
	repos, _, err := client.Repositories.List(context.Background(), "u", opt)
	if err != nil {
		t.Errorf("Repositories.List returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.List returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_List_invalidUser(t *testing.T) {
	_, _, err := client.Repositories.List(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestRepositoriesService_ListByOrg(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeLicensesPreview)
		testFormValues(t, r, values{
			"type": "forks",
			"page": "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &RepositoryListByOrgOptions{"forks", ListOptions{Page: 2}}
	repos, _, err := client.Repositories.ListByOrg(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Repositories.ListByOrg returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.ListByOrg returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_ListByOrg_invalidOrg(t *testing.T) {
	_, _, err := client.Repositories.ListByOrg(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestRepositoriesService_ListAll(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"since":    "1",
			"page":     "2",
			"per_page": "3",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &RepositoryListAllOptions{1, ListOptions{2, 3}}
	repos, _, err := client.Repositories.ListAll(context.Background(), opt)
	if err != nil {
		t.Errorf("Repositories.ListAll returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("Repositories.ListAll returned %+v, want %+v", repos, want)
	}
}

func TestRepositoriesService_Create_user(t *testing.T) {
	setup()
	defer teardown()

	input := &Repository{Name: String("n")}

	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		v := new(Repository)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	repo, _, err := client.Repositories.Create(context.Background(), "", input)
	if err != nil {
		t.Errorf("Repositories.Create returned error: %v", err)
	}

	want := &Repository{ID: Int(1)}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Repositories.Create returned %+v, want %+v", repo, want)
	}
}

func TestRepositoriesService_Create_org(t *testing.T) {
	setup()
	defer teardown()

	input := &Repository{Name: String("n")}

	mux.HandleFunc("/orgs/o/repos", func(w http.ResponseWriter, r *http.Request) {
		v := new(Repository)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	repo, _, err := client.Repositories.Create(context.Background(), "o", input)
	if err != nil {
		t.Errorf("Repositories.Create returned error: %v", err)
	}

	want := &Repository{ID: Int(1)}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Repositories.Create returned %+v, want %+v", repo, want)
	}
}

func TestRepositoriesService_Create_invalidOrg(t *testing.T) {
	_, _, err := client.Repositories.Create(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestRepositoriesService_Get(t *testing.T) {
	setup()
	defer teardown()

	acceptHeader := []string{mediaTypeLicensesPreview, mediaTypeSquashPreview, mediaTypeCodesOfConductPreview}
	mux.HandleFunc("/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", strings.Join(acceptHeader, ", "))
		fmt.Fprint(w, `{"id":1,"name":"n","description":"d","owner":{"login":"l"},"license":{"key":"mit"}}`)
	})

	repo, _, err := client.Repositories.Get(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.Get returned error: %v", err)
	}

	want := &Repository{ID: Int(1), Name: String("n"), Description: String("d"), Owner: &User{Login: String("l")}, License: &License{Key: String("mit")}}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Repositories.Get returned %+v, want %+v", repo, want)
	}
}

func TestRepositoriesService_GetCodeOfConduct(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/community/code_of_conduct", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeCodesOfConductPreview)
		fmt.Fprint(w, `{
						"key": "key",
						"name": "name",
						"url": "url",
						"body": "body"}`,
		)
	})

	coc, _, err := client.Repositories.GetCodeOfConduct(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.GetCodeOfConduct returned error: %v", err)
	}

	want := &CodeOfConduct{
		Key:  String("key"),
		Name: String("name"),
		URL:  String("url"),
		Body: String("body"),
	}

	if !reflect.DeepEqual(coc, want) {
		t.Errorf("Repositories.Get returned %+v, want %+v", coc, want)
	}
}

func TestRepositoriesService_GetByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repositories/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeLicensesPreview)
		fmt.Fprint(w, `{"id":1,"name":"n","description":"d","owner":{"login":"l"},"license":{"key":"mit"}}`)
	})

	repo, _, err := client.Repositories.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Repositories.GetByID returned error: %v", err)
	}

	want := &Repository{ID: Int(1), Name: String("n"), Description: String("d"), Owner: &User{Login: String("l")}, License: &License{Key: String("mit")}}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Repositories.GetByID returned %+v, want %+v", repo, want)
	}
}

func TestRepositoriesService_Edit(t *testing.T) {
	setup()
	defer teardown()

	i := true
	input := &Repository{HasIssues: &i}

	mux.HandleFunc("/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		v := new(Repository)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	repo, _, err := client.Repositories.Edit(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Repositories.Edit returned error: %v", err)
	}

	want := &Repository{ID: Int(1)}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Repositories.Edit returned %+v, want %+v", repo, want)
	}
}

func TestRepositoriesService_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Repositories.Delete(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.Delete returned error: %v", err)
	}
}

func TestRepositoriesService_Get_invalidOwner(t *testing.T) {
	_, _, err := client.Repositories.Get(context.Background(), "%", "r")
	testURLParseError(t, err)
}

func TestRepositoriesService_Edit_invalidOwner(t *testing.T) {
	_, _, err := client.Repositories.Edit(context.Background(), "%", "r", nil)
	testURLParseError(t, err)
}

func TestRepositoriesService_ListContributors(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/contributors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"anon": "true",
			"page": "2",
		})
		fmt.Fprint(w, `[{"contributions":42}]`)
	})

	opts := &ListContributorsOptions{Anon: "true", ListOptions: ListOptions{Page: 2}}
	contributors, _, err := client.Repositories.ListContributors(context.Background(), "o", "r", opts)
	if err != nil {
		t.Errorf("Repositories.ListContributors returned error: %v", err)
	}

	want := []*Contributor{{Contributions: Int(42)}}
	if !reflect.DeepEqual(contributors, want) {
		t.Errorf("Repositories.ListContributors returned %+v, want %+v", contributors, want)
	}
}

func TestRepositoriesService_ListLanguages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/languages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"go":1}`)
	})

	languages, _, err := client.Repositories.ListLanguages(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.ListLanguages returned error: %v", err)
	}

	want := map[string]int{"go": 1}
	if !reflect.DeepEqual(languages, want) {
		t.Errorf("Repositories.ListLanguages returned %+v, want %+v", languages, want)
	}
}

func TestRepositoriesService_ListTeams(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 2}
	teams, _, err := client.Repositories.ListTeams(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListTeams returned error: %v", err)
	}

	want := []*Team{{ID: Int(1)}}
	if !reflect.DeepEqual(teams, want) {
		t.Errorf("Repositories.ListTeams returned %+v, want %+v", teams, want)
	}
}

func TestRepositoriesService_ListTags(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"name":"n", "commit" : {"sha" : "s", "url" : "u"}, "zipball_url": "z", "tarball_url": "t"}]`)
	})

	opt := &ListOptions{Page: 2}
	tags, _, err := client.Repositories.ListTags(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListTags returned error: %v", err)
	}

	want := []*RepositoryTag{
		{
			Name: String("n"),
			Commit: &Commit{
				SHA: String("s"),
				URL: String("u"),
			},
			ZipballURL: String("z"),
			TarballURL: String("t"),
		},
	}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("Repositories.ListTags returned %+v, want %+v", tags, want)
	}
}

func TestRepositoriesService_ListBranches(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"name":"master", "commit" : {"sha" : "a57781", "url" : "https://api.github.com/repos/o/r/commits/a57781"}}]`)
	})

	opt := &ListOptions{Page: 2}
	branches, _, err := client.Repositories.ListBranches(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListBranches returned error: %v", err)
	}

	want := []*Branch{{Name: String("master"), Commit: &RepositoryCommit{SHA: String("a57781"), URL: String("https://api.github.com/repos/o/r/commits/a57781")}}}
	if !reflect.DeepEqual(branches, want) {
		t.Errorf("Repositories.ListBranches returned %+v, want %+v", branches, want)
	}
}

func TestRepositoriesService_GetBranch(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprint(w, `{"name":"n", "commit":{"sha":"s","commit":{"message":"m"}}, "protected":true}`)
	})

	branch, _, err := client.Repositories.GetBranch(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.GetBranch returned error: %v", err)
	}

	want := &Branch{
		Name: String("n"),
		Commit: &RepositoryCommit{
			SHA: String("s"),
			Commit: &Commit{
				Message: String("m"),
			},
		},
		Protected: Bool(true),
	}

	if !reflect.DeepEqual(branch, want) {
		t.Errorf("Repositories.GetBranch returned %+v, want %+v", branch, want)
	}
}

func TestRepositoriesService_GetBranchProtection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection", func(w http.ResponseWriter, r *http.Request) {
		v := new(ProtectionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"required_status_checks":{"strict":true,"contexts":["continuous-integration"]},"required_pull_request_reviews":{"dismissal_restrictions":{"users":[{"id":3,"login":"u"}],"teams":[{"id":4,"slug":"t"}]},"dismiss_stale_reviews":true},"enforce_admins":{"url":"/repos/o/r/branches/b/protection/enforce_admins","enabled":true},"restrictions":{"users":[{"id":1,"login":"u"}],"teams":[{"id":2,"slug":"t"}]}}`)
	})

	protection, _, err := client.Repositories.GetBranchProtection(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.GetBranchProtection returned error: %v", err)
	}

	want := &Protection{
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"continuous-integration"},
		},
		RequiredPullRequestReviews: &PullRequestReviewsEnforcement{
			DismissStaleReviews: true,
			DismissalRestrictions: DismissalRestrictions{
				Users: []*User{
					{Login: String("u"), ID: Int(3)},
				},
				Teams: []*Team{
					{Slug: String("t"), ID: Int(4)},
				},
			},
		},
		EnforceAdmins: &AdminEnforcement{
			URL:     String("/repos/o/r/branches/b/protection/enforce_admins"),
			Enabled: true,
		},
		Restrictions: &BranchRestrictions{
			Users: []*User{
				{Login: String("u"), ID: Int(1)},
			},
			Teams: []*Team{
				{Slug: String("t"), ID: Int(2)},
			},
		},
	}
	if !reflect.DeepEqual(protection, want) {
		t.Errorf("Repositories.GetBranchProtection returned %+v, want %+v", protection, want)
	}
}

func TestRepositoriesService_UpdateBranchProtection(t *testing.T) {
	setup()
	defer teardown()

	input := &ProtectionRequest{
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"continuous-integration"},
		},
		RequiredPullRequestReviews: &PullRequestReviewsEnforcementRequest{
			DismissStaleReviews: true,
			DismissalRestrictionsRequest: &DismissalRestrictionsRequest{
				Users: []string{"uu"},
				Teams: []string{"tt"},
			},
		},
		Restrictions: &BranchRestrictionsRequest{
			Users: []string{"u"},
			Teams: []string{"t"},
		},
	}

	mux.HandleFunc("/repos/o/r/branches/b/protection", func(w http.ResponseWriter, r *http.Request) {
		v := new(ProtectionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"required_status_checks":{"strict":true,"contexts":["continuous-integration"]},"required_pull_request_reviews":{"dismissal_restrictions":{"users":[{"id":3,"login":"uu"}],"teams":[{"id":4,"slug":"tt"}]},"dismiss_stale_reviews":true},"restrictions":{"users":[{"id":1,"login":"u"}],"teams":[{"id":2,"slug":"t"}]}}`)
	})

	protection, _, err := client.Repositories.UpdateBranchProtection(context.Background(), "o", "r", "b", input)
	if err != nil {
		t.Errorf("Repositories.UpdateBranchProtection returned error: %v", err)
	}

	want := &Protection{
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"continuous-integration"},
		},
		RequiredPullRequestReviews: &PullRequestReviewsEnforcement{
			DismissStaleReviews: true,
			DismissalRestrictions: DismissalRestrictions{
				Users: []*User{
					{Login: String("uu"), ID: Int(3)},
				},
				Teams: []*Team{
					{Slug: String("tt"), ID: Int(4)},
				},
			},
		},
		Restrictions: &BranchRestrictions{
			Users: []*User{
				{Login: String("u"), ID: Int(1)},
			},
			Teams: []*Team{
				{Slug: String("t"), ID: Int(2)},
			},
		},
	}
	if !reflect.DeepEqual(protection, want) {
		t.Errorf("Repositories.UpdateBranchProtection returned %+v, want %+v", protection, want)
	}
}

func TestRepositoriesService_RemoveBranchProtection(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Repositories.RemoveBranchProtection(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.RemoveBranchProtection returned error: %v", err)
	}
}

func TestRepositoriesService_ListLanguages_invalidOwner(t *testing.T) {
	_, _, err := client.Repositories.ListLanguages(context.Background(), "%", "%")
	testURLParseError(t, err)
}

func TestRepositoriesService_License(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/license", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"name": "LICENSE", "path": "LICENSE", "license":{"key":"mit","name":"MIT License","spdx_id":"MIT","url":"https://api.github.com/licenses/mit","featured":true}}`)
	})

	got, _, err := client.Repositories.License(context.Background(), "o", "r")
	if err != nil {
		t.Errorf("Repositories.License returned error: %v", err)
	}

	want := &RepositoryLicense{
		Name: String("LICENSE"),
		Path: String("LICENSE"),
		License: &License{
			Name:     String("MIT License"),
			Key:      String("mit"),
			SPDXID:   String("MIT"),
			URL:      String("https://api.github.com/licenses/mit"),
			Featured: Bool(true),
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Repositories.License returned %+v, want %+v", got, want)
	}
}

func TestRepositoriesService_GetRequiredStatusChecks(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_status_checks", func(w http.ResponseWriter, r *http.Request) {
		v := new(ProtectionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprint(w, `{"strict": true,"contexts": ["x","y","z"]}`)
	})

	checks, _, err := client.Repositories.GetRequiredStatusChecks(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.GetRequiredStatusChecks returned error: %v", err)
	}

	want := &RequiredStatusChecks{
		Strict:   true,
		Contexts: []string{"x", "y", "z"},
	}
	if !reflect.DeepEqual(checks, want) {
		t.Errorf("Repositories.GetRequiredStatusChecks returned %+v, want %+v", checks, want)
	}
}

func TestRepositoriesService_ListRequiredStatusChecksContexts(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_status_checks/contexts", func(w http.ResponseWriter, r *http.Request) {
		v := new(ProtectionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprint(w, `["x", "y", "z"]`)
	})

	contexts, _, err := client.Repositories.ListRequiredStatusChecksContexts(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.ListRequiredStatusChecksContexts returned error: %v", err)
	}

	want := []string{"x", "y", "z"}
	if !reflect.DeepEqual(contexts, want) {
		t.Errorf("Repositories.ListRequiredStatusChecksContexts returned %+v, want %+v", contexts, want)
	}
}

func TestRepositoriesService_GetPullRequestReviewEnforcement(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_pull_request_reviews", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"dismissal_restrictions":{"users":[{"id":1,"login":"u"}],"teams":[{"id":2,"slug":"t"}]},"dismiss_stale_reviews":true}`)
	})

	enforcement, _, err := client.Repositories.GetPullRequestReviewEnforcement(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.GetPullRequestReviewEnforcement returned error: %v", err)
	}

	want := &PullRequestReviewsEnforcement{
		DismissStaleReviews: true,
		DismissalRestrictions: DismissalRestrictions{
			Users: []*User{
				{Login: String("u"), ID: Int(1)},
			},
			Teams: []*Team{
				{Slug: String("t"), ID: Int(2)},
			},
		},
	}

	if !reflect.DeepEqual(enforcement, want) {
		t.Errorf("Repositories.GetPullRequestReviewEnforcement returned %+v, want %+v", enforcement, want)
	}
}

func TestRepositoriesService_UpdatePullRequestReviewEnforcement(t *testing.T) {
	setup()
	defer teardown()

	input := &PullRequestReviewsEnforcementUpdate{
		DismissalRestrictionsRequest: &DismissalRestrictionsRequest{
			Users: []string{"u"},
			Teams: []string{"t"},
		},
	}

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_pull_request_reviews", func(w http.ResponseWriter, r *http.Request) {
		v := new(PullRequestReviewsEnforcementUpdate)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"dismissal_restrictions":{"users":[{"id":1,"login":"u"}],"teams":[{"id":2,"slug":"t"}]},"dismiss_stale_reviews":true}`)
	})

	enforcement, _, err := client.Repositories.UpdatePullRequestReviewEnforcement(context.Background(), "o", "r", "b", input)
	if err != nil {
		t.Errorf("Repositories.UpdatePullRequestReviewEnforcement returned error: %v", err)
	}

	want := &PullRequestReviewsEnforcement{
		DismissStaleReviews: true,
		DismissalRestrictions: DismissalRestrictions{
			Users: []*User{
				{Login: String("u"), ID: Int(1)},
			},
			Teams: []*Team{
				{Slug: String("t"), ID: Int(2)},
			},
		},
	}
	if !reflect.DeepEqual(enforcement, want) {
		t.Errorf("Repositories.UpdatePullRequestReviewEnforcement returned %+v, want %+v", enforcement, want)
	}
}

func TestRepositoriesService_DisableDismissalRestrictions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_pull_request_reviews", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		testBody(t, r, `{"dismissal_restrictions":[]}`+"\n")
		fmt.Fprintf(w, `{"dismissal_restrictions":{"users":[],"teams":[]},"dismiss_stale_reviews":true}`)
	})

	enforcement, _, err := client.Repositories.DisableDismissalRestrictions(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.DisableDismissalRestrictions returned error: %v", err)
	}

	want := &PullRequestReviewsEnforcement{
		DismissStaleReviews: true,
		DismissalRestrictions: DismissalRestrictions{
			Users: []*User{},
			Teams: []*Team{},
		},
	}
	if !reflect.DeepEqual(enforcement, want) {
		t.Errorf("Repositories.DisableDismissalRestrictions returned %+v, want %+v", enforcement, want)
	}
}

func TestRepositoriesService_RemovePullRequestReviewEnforcement(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/required_pull_request_reviews", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Repositories.RemovePullRequestReviewEnforcement(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.RemovePullRequestReviewEnforcement returned error: %v", err)
	}
}

func TestRepositoriesService_GetAdminEnforcement(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/enforce_admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"url":"/repos/o/r/branches/b/protection/enforce_admins","enabled":true}`)
	})

	enforcement, _, err := client.Repositories.GetAdminEnforcement(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.GetAdminEnforcement returned error: %v", err)
	}

	want := &AdminEnforcement{
		URL:     String("/repos/o/r/branches/b/protection/enforce_admins"),
		Enabled: true,
	}

	if !reflect.DeepEqual(enforcement, want) {
		t.Errorf("Repositories.GetAdminEnforcement returned %+v, want %+v", enforcement, want)
	}
}

func TestRepositoriesService_AddAdminEnforcement(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/enforce_admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		fmt.Fprintf(w, `{"url":"/repos/o/r/branches/b/protection/enforce_admins","enabled":true}`)
	})

	enforcement, _, err := client.Repositories.AddAdminEnforcement(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.AddAdminEnforcement returned error: %v", err)
	}

	want := &AdminEnforcement{
		URL:     String("/repos/o/r/branches/b/protection/enforce_admins"),
		Enabled: true,
	}
	if !reflect.DeepEqual(enforcement, want) {
		t.Errorf("Repositories.AddAdminEnforcement returned %+v, want %+v", enforcement, want)
	}
}

func TestRepositoriesService_RemoveAdminEnforcement(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/branches/b/protection/enforce_admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeProtectedBranchesPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Repositories.RemoveAdminEnforcement(context.Background(), "o", "r", "b")
	if err != nil {
		t.Errorf("Repositories.RemoveAdminEnforcement returned error: %v", err)
	}
}

func TestPullRequestReviewsEnforcementRequest_MarshalJSON_nilDismissalRestirctions(t *testing.T) {
	req := PullRequestReviewsEnforcementRequest{}

	json, err := json.Marshal(req)
	if err != nil {
		t.Errorf("PullRequestReviewsEnforcementRequest.MarshalJSON returned error: %v", err)
	}

	want := `{"dismissal_restrictions":[],"dismiss_stale_reviews":false}`
	if want != string(json) {
		t.Errorf("PullRequestReviewsEnforcementRequest.MarshalJSON returned %+v, want %+v", string(json), want)
	}
}
