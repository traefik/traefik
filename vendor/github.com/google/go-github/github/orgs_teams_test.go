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

func TestOrganizationsService_ListTeams(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 2}
	teams, _, err := client.Organizations.ListTeams(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListTeams returned error: %v", err)
	}

	want := []*Team{{ID: Int(1)}}
	if !reflect.DeepEqual(teams, want) {
		t.Errorf("Organizations.ListTeams returned %+v, want %+v", teams, want)
	}
}

func TestOrganizationsService_ListTeams_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.ListTeams(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestOrganizationsService_GetTeam(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "name":"n", "description": "d", "url":"u", "slug": "s", "permission":"p", "ldap_dn":"cn=n,ou=groups,dc=example,dc=com"}`)
	})

	team, _, err := client.Organizations.GetTeam(context.Background(), 1)
	if err != nil {
		t.Errorf("Organizations.GetTeam returned error: %v", err)
	}

	want := &Team{ID: Int(1), Name: String("n"), Description: String("d"), URL: String("u"), Slug: String("s"), Permission: String("p"), LDAPDN: String("cn=n,ou=groups,dc=example,dc=com")}
	if !reflect.DeepEqual(team, want) {
		t.Errorf("Organizations.GetTeam returned %+v, want %+v", team, want)
	}
}

func TestOrganizationsService_CreateTeam(t *testing.T) {
	setup()
	defer teardown()

	input := &Team{Name: String("n"), Privacy: String("closed")}

	mux.HandleFunc("/orgs/o/teams", func(w http.ResponseWriter, r *http.Request) {
		v := new(Team)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	team, _, err := client.Organizations.CreateTeam(context.Background(), "o", input)
	if err != nil {
		t.Errorf("Organizations.CreateTeam returned error: %v", err)
	}

	want := &Team{ID: Int(1)}
	if !reflect.DeepEqual(team, want) {
		t.Errorf("Organizations.CreateTeam returned %+v, want %+v", team, want)
	}
}

func TestOrganizationsService_CreateTeam_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.CreateTeam(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestOrganizationsService_EditTeam(t *testing.T) {
	setup()
	defer teardown()

	input := &Team{Name: String("n"), Privacy: String("closed")}

	mux.HandleFunc("/teams/1", func(w http.ResponseWriter, r *http.Request) {
		v := new(Team)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	team, _, err := client.Organizations.EditTeam(context.Background(), 1, input)
	if err != nil {
		t.Errorf("Organizations.EditTeam returned error: %v", err)
	}

	want := &Team{ID: Int(1)}
	if !reflect.DeepEqual(team, want) {
		t.Errorf("Organizations.EditTeam returned %+v, want %+v", team, want)
	}
}

func TestOrganizationsService_DeleteTeam(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Organizations.DeleteTeam(context.Background(), 1)
	if err != nil {
		t.Errorf("Organizations.DeleteTeam returned error: %v", err)
	}
}

func TestOrganizationsService_ListTeamMembers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"role": "member", "page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &OrganizationListTeamMembersOptions{Role: "member", ListOptions: ListOptions{Page: 2}}
	members, _, err := client.Organizations.ListTeamMembers(context.Background(), 1, opt)
	if err != nil {
		t.Errorf("Organizations.ListTeamMembers returned error: %v", err)
	}

	want := []*User{{ID: Int(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListTeamMembers returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_IsTeamMember_true(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
	})

	member, _, err := client.Organizations.IsTeamMember(context.Background(), 1, "u")
	if err != nil {
		t.Errorf("Organizations.IsTeamMember returned error: %v", err)
	}
	if want := true; member != want {
		t.Errorf("Organizations.IsTeamMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 404 response is interpreted as "false" and not an error
func TestOrganizationsService_IsTeamMember_false(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
	})

	member, _, err := client.Organizations.IsTeamMember(context.Background(), 1, "u")
	if err != nil {
		t.Errorf("Organizations.IsTeamMember returned error: %+v", err)
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsTeamMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 400 response is interpreted as an actual error, and not simply
// as "false" like the above case of a 404
func TestOrganizationsService_IsTeamMember_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		http.Error(w, "BadRequest", http.StatusBadRequest)
	})

	member, _, err := client.Organizations.IsTeamMember(context.Background(), 1, "u")
	if err == nil {
		t.Errorf("Expected HTTP 400 response")
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsTeamMember returned %+v, want %+v", member, want)
	}
}

func TestOrganizationsService_IsTeamMember_invalidUser(t *testing.T) {
	_, _, err := client.Organizations.IsTeamMember(context.Background(), 1, "%")
	testURLParseError(t, err)
}

func TestOrganizationsService_PublicizeMembership(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.PublicizeMembership(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.PublicizeMembership returned error: %v", err)
	}
}

func TestOrganizationsService_PublicizeMembership_invalidOrg(t *testing.T) {
	_, err := client.Organizations.PublicizeMembership(context.Background(), "%", "u")
	testURLParseError(t, err)
}

func TestOrganizationsService_ConcealMembership(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.ConcealMembership(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.ConcealMembership returned error: %v", err)
	}
}

func TestOrganizationsService_ConcealMembership_invalidOrg(t *testing.T) {
	_, err := client.Organizations.ConcealMembership(context.Background(), "%", "u")
	testURLParseError(t, err)
}

func TestOrganizationsService_ListTeamRepos(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 2}
	members, _, err := client.Organizations.ListTeamRepos(context.Background(), 1, opt)
	if err != nil {
		t.Errorf("Organizations.ListTeamRepos returned error: %v", err)
	}

	want := []*Repository{{ID: Int(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListTeamRepos returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_IsTeamRepo_true(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeOrgPermissionRepo)
		fmt.Fprint(w, `{"id":1}`)
	})

	repo, _, err := client.Organizations.IsTeamRepo(context.Background(), 1, "o", "r")
	if err != nil {
		t.Errorf("Organizations.IsTeamRepo returned error: %v", err)
	}

	want := &Repository{ID: Int(1)}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("Organizations.IsTeamRepo returned %+v, want %+v", repo, want)
	}
}

func TestOrganizationsService_IsTeamRepo_false(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
	})

	repo, resp, err := client.Organizations.IsTeamRepo(context.Background(), 1, "o", "r")
	if err == nil {
		t.Errorf("Expected HTTP 404 response")
	}
	if got, want := resp.Response.StatusCode, http.StatusNotFound; got != want {
		t.Errorf("Organizations.IsTeamRepo returned status %d, want %d", got, want)
	}
	if repo != nil {
		t.Errorf("Organizations.IsTeamRepo returned %+v, want nil", repo)
	}
}

func TestOrganizationsService_IsTeamRepo_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		http.Error(w, "BadRequest", http.StatusBadRequest)
	})

	repo, resp, err := client.Organizations.IsTeamRepo(context.Background(), 1, "o", "r")
	if err == nil {
		t.Errorf("Expected HTTP 400 response")
	}
	if got, want := resp.Response.StatusCode, http.StatusBadRequest; got != want {
		t.Errorf("Organizations.IsTeamRepo returned status %d, want %d", got, want)
	}
	if repo != nil {
		t.Errorf("Organizations.IsTeamRepo returned %+v, want nil", repo)
	}
}

func TestOrganizationsService_IsTeamRepo_invalidOwner(t *testing.T) {
	_, _, err := client.Organizations.IsTeamRepo(context.Background(), 1, "%", "r")
	testURLParseError(t, err)
}

func TestOrganizationsService_AddTeamRepo(t *testing.T) {
	setup()
	defer teardown()

	opt := &OrganizationAddTeamRepoOptions{Permission: "admin"}

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		v := new(OrganizationAddTeamRepoOptions)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, opt) {
			t.Errorf("Request body = %+v, want %+v", v, opt)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.AddTeamRepo(context.Background(), 1, "o", "r", opt)
	if err != nil {
		t.Errorf("Organizations.AddTeamRepo returned error: %v", err)
	}
}

func TestOrganizationsService_AddTeamRepo_noAccess(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusUnprocessableEntity)
	})

	_, err := client.Organizations.AddTeamRepo(context.Background(), 1, "o", "r", nil)
	if err == nil {
		t.Errorf("Expcted error to be returned")
	}
}

func TestOrganizationsService_AddTeamRepo_invalidOwner(t *testing.T) {
	_, err := client.Organizations.AddTeamRepo(context.Background(), 1, "%", "r", nil)
	testURLParseError(t, err)
}

func TestOrganizationsService_RemoveTeamRepo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/repos/o/r", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.RemoveTeamRepo(context.Background(), 1, "o", "r")
	if err != nil {
		t.Errorf("Organizations.RemoveTeamRepo returned error: %v", err)
	}
}

func TestOrganizationsService_RemoveTeamRepo_invalidOwner(t *testing.T) {
	_, err := client.Organizations.RemoveTeamRepo(context.Background(), 1, "%", "r")
	testURLParseError(t, err)
}

func TestOrganizationsService_GetTeamMembership(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"url":"u", "state":"active"}`)
	})

	membership, _, err := client.Organizations.GetTeamMembership(context.Background(), 1, "u")
	if err != nil {
		t.Errorf("Organizations.GetTeamMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u"), State: String("active")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.GetTeamMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_AddTeamMembership(t *testing.T) {
	setup()
	defer teardown()

	opt := &OrganizationAddTeamMembershipOptions{Role: "maintainer"}

	mux.HandleFunc("/teams/1/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		v := new(OrganizationAddTeamMembershipOptions)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, opt) {
			t.Errorf("Request body = %+v, want %+v", v, opt)
		}

		fmt.Fprint(w, `{"url":"u", "state":"pending"}`)
	})

	membership, _, err := client.Organizations.AddTeamMembership(context.Background(), 1, "u", opt)
	if err != nil {
		t.Errorf("Organizations.AddTeamMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u"), State: String("pending")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.AddTeamMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_RemoveTeamMembership(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.RemoveTeamMembership(context.Background(), 1, "u")
	if err != nil {
		t.Errorf("Organizations.RemoveTeamMembership returned error: %v", err)
	}
}

func TestOrganizationsService_ListUserTeams(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "1"})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListOptions{Page: 1}
	teams, _, err := client.Organizations.ListUserTeams(context.Background(), opt)
	if err != nil {
		t.Errorf("Organizations.ListUserTeams returned error: %v", err)
	}

	want := []*Team{{ID: Int(1)}}
	if !reflect.DeepEqual(teams, want) {
		t.Errorf("Organizations.ListUserTeams returned %+v, want %+v", teams, want)
	}
}

func TestOrganizationsService_ListPendingTeamInvitations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/teams/1/invitations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "1"})
		fmt.Fprint(w, `[
				{
    					"id": 1,
    					"login": "monalisa",
    					"email": "octocat@github.com",
    					"role": "direct_member",
					"created_at": "2017-01-21T00:00:00Z",
    					"inviter": {
      						"login": "other_user",
      						"id": 1,
      						"avatar_url": "https://github.com/images/error/other_user_happy.gif",
      						"gravatar_id": "",
      						"url": "https://api.github.com/users/other_user",
      						"html_url": "https://github.com/other_user",
      						"followers_url": "https://api.github.com/users/other_user/followers",
      						"following_url": "https://api.github.com/users/other_user/following/other_user",
      						"gists_url": "https://api.github.com/users/other_user/gists/gist_id",
      						"starred_url": "https://api.github.com/users/other_user/starred/owner/repo",
      						"subscriptions_url": "https://api.github.com/users/other_user/subscriptions",
      						"organizations_url": "https://api.github.com/users/other_user/orgs",
      						"repos_url": "https://api.github.com/users/other_user/repos",
      						"events_url": "https://api.github.com/users/other_user/events/privacy",
      						"received_events_url": "https://api.github.com/users/other_user/received_events/privacy",
      						"type": "User",
      						"site_admin": false
    					}
  				}
			]`)
	})

	opt := &ListOptions{Page: 1}
	invitations, _, err := client.Organizations.ListPendingTeamInvitations(context.Background(), 1, opt)
	if err != nil {
		t.Errorf("Organizations.ListPendingTeamInvitations returned error: %v", err)
	}

	createdAt := time.Date(2017, 01, 21, 0, 0, 0, 0, time.UTC)
	want := []*Invitation{
		{
			ID:        Int(1),
			Login:     String("monalisa"),
			Email:     String("octocat@github.com"),
			Role:      String("direct_member"),
			CreatedAt: &createdAt,
			Inviter: &User{
				Login:             String("other_user"),
				ID:                Int(1),
				AvatarURL:         String("https://github.com/images/error/other_user_happy.gif"),
				GravatarID:        String(""),
				URL:               String("https://api.github.com/users/other_user"),
				HTMLURL:           String("https://github.com/other_user"),
				FollowersURL:      String("https://api.github.com/users/other_user/followers"),
				FollowingURL:      String("https://api.github.com/users/other_user/following/other_user"),
				GistsURL:          String("https://api.github.com/users/other_user/gists/gist_id"),
				StarredURL:        String("https://api.github.com/users/other_user/starred/owner/repo"),
				SubscriptionsURL:  String("https://api.github.com/users/other_user/subscriptions"),
				OrganizationsURL:  String("https://api.github.com/users/other_user/orgs"),
				ReposURL:          String("https://api.github.com/users/other_user/repos"),
				EventsURL:         String("https://api.github.com/users/other_user/events/privacy"),
				ReceivedEventsURL: String("https://api.github.com/users/other_user/received_events/privacy"),
				Type:              String("User"),
				SiteAdmin:         Bool(false),
			},
		}}

	if !reflect.DeepEqual(invitations, want) {
		t.Errorf("Organizations.ListPendingTeamInvitations returned %+v, want %+v", invitations, want)
	}
}
