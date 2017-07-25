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

func TestOrganizationsService_ListMembers(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"filter": "2fa_disabled",
			"role":   "admin",
			"page":   "2",
		})
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListMembersOptions{
		PublicOnly:  false,
		Filter:      "2fa_disabled",
		Role:        "admin",
		ListOptions: ListOptions{Page: 2},
	}
	members, _, err := client.Organizations.ListMembers(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListMembers returned error: %v", err)
	}

	want := []*User{{ID: Int(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListMembers returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_ListMembers_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.ListMembers(context.Background(), "%", nil)
	testURLParseError(t, err)
}

func TestOrganizationsService_ListMembers_public(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &ListMembersOptions{PublicOnly: true}
	members, _, err := client.Organizations.ListMembers(context.Background(), "o", opt)
	if err != nil {
		t.Errorf("Organizations.ListMembers returned error: %v", err)
	}

	want := []*User{{ID: Int(1)}}
	if !reflect.DeepEqual(members, want) {
		t.Errorf("Organizations.ListMembers returned %+v, want %+v", members, want)
	}
}

func TestOrganizationsService_IsMember(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNoContent)
	})

	member, _, err := client.Organizations.IsMember(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.IsMember returned error: %v", err)
	}
	if want := true; member != want {
		t.Errorf("Organizations.IsMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 404 response is interpreted as "false" and not an error
func TestOrganizationsService_IsMember_notMember(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
	})

	member, _, err := client.Organizations.IsMember(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.IsMember returned error: %+v", err)
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 400 response is interpreted as an actual error, and not simply
// as "false" like the above case of a 404
func TestOrganizationsService_IsMember_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		http.Error(w, "BadRequest", http.StatusBadRequest)
	})

	member, _, err := client.Organizations.IsMember(context.Background(), "o", "u")
	if err == nil {
		t.Errorf("Expected HTTP 400 response")
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsMember returned %+v, want %+v", member, want)
	}
}

func TestOrganizationsService_IsMember_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.IsMember(context.Background(), "%", "u")
	testURLParseError(t, err)
}

func TestOrganizationsService_IsPublicMember(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNoContent)
	})

	member, _, err := client.Organizations.IsPublicMember(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.IsPublicMember returned error: %v", err)
	}
	if want := true; member != want {
		t.Errorf("Organizations.IsPublicMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 404 response is interpreted as "false" and not an error
func TestOrganizationsService_IsPublicMember_notMember(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
	})

	member, _, err := client.Organizations.IsPublicMember(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.IsPublicMember returned error: %v", err)
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsPublicMember returned %+v, want %+v", member, want)
	}
}

// ensure that a 400 response is interpreted as an actual error, and not simply
// as "false" like the above case of a 404
func TestOrganizationsService_IsPublicMember_error(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/public_members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		http.Error(w, "BadRequest", http.StatusBadRequest)
	})

	member, _, err := client.Organizations.IsPublicMember(context.Background(), "o", "u")
	if err == nil {
		t.Errorf("Expected HTTP 400 response")
	}
	if want := false; member != want {
		t.Errorf("Organizations.IsPublicMember returned %+v, want %+v", member, want)
	}
}

func TestOrganizationsService_IsPublicMember_invalidOrg(t *testing.T) {
	_, _, err := client.Organizations.IsPublicMember(context.Background(), "%", "u")
	testURLParseError(t, err)
}

func TestOrganizationsService_RemoveMember(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/members/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Organizations.RemoveMember(context.Background(), "o", "u")
	if err != nil {
		t.Errorf("Organizations.RemoveMember returned error: %v", err)
	}
}

func TestOrganizationsService_RemoveMember_invalidOrg(t *testing.T) {
	_, err := client.Organizations.RemoveMember(context.Background(), "%", "u")
	testURLParseError(t, err)
}

func TestOrganizationsService_ListOrgMemberships(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/memberships/orgs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"state": "active",
			"page":  "2",
		})
		fmt.Fprint(w, `[{"url":"u"}]`)
	})

	opt := &ListOrgMembershipsOptions{
		State:       "active",
		ListOptions: ListOptions{Page: 2},
	}
	memberships, _, err := client.Organizations.ListOrgMemberships(context.Background(), opt)
	if err != nil {
		t.Errorf("Organizations.ListOrgMemberships returned error: %v", err)
	}

	want := []*Membership{{URL: String("u")}}
	if !reflect.DeepEqual(memberships, want) {
		t.Errorf("Organizations.ListOrgMemberships returned %+v, want %+v", memberships, want)
	}
}

func TestOrganizationsService_GetOrgMembership_AuthenticatedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/memberships/orgs/o", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"url":"u"}`)
	})

	membership, _, err := client.Organizations.GetOrgMembership(context.Background(), "", "o")
	if err != nil {
		t.Errorf("Organizations.GetOrgMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.GetOrgMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_GetOrgMembership_SpecifiedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"url":"u"}`)
	})

	membership, _, err := client.Organizations.GetOrgMembership(context.Background(), "u", "o")
	if err != nil {
		t.Errorf("Organizations.GetOrgMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.GetOrgMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_EditOrgMembership_AuthenticatedUser(t *testing.T) {
	setup()
	defer teardown()

	input := &Membership{State: String("active")}

	mux.HandleFunc("/user/memberships/orgs/o", func(w http.ResponseWriter, r *http.Request) {
		v := new(Membership)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"url":"u"}`)
	})

	membership, _, err := client.Organizations.EditOrgMembership(context.Background(), "", "o", input)
	if err != nil {
		t.Errorf("Organizations.EditOrgMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.EditOrgMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_EditOrgMembership_SpecifiedUser(t *testing.T) {
	setup()
	defer teardown()

	input := &Membership{State: String("active")}

	mux.HandleFunc("/orgs/o/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		v := new(Membership)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"url":"u"}`)
	})

	membership, _, err := client.Organizations.EditOrgMembership(context.Background(), "u", "o", input)
	if err != nil {
		t.Errorf("Organizations.EditOrgMembership returned error: %v", err)
	}

	want := &Membership{URL: String("u")}
	if !reflect.DeepEqual(membership, want) {
		t.Errorf("Organizations.EditOrgMembership returned %+v, want %+v", membership, want)
	}
}

func TestOrganizationsService_RemoveOrgMembership(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/o/memberships/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Organizations.RemoveOrgMembership(context.Background(), "u", "o")
	if err != nil {
		t.Errorf("Organizations.RemoveOrgMembership returned error: %v", err)
	}
}

func TestOrganizationsService_ListPendingOrgInvitations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/orgs/1/invitations", func(w http.ResponseWriter, r *http.Request) {
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
	invitations, _, err := client.Organizations.ListPendingOrgInvitations(context.Background(), 1, opt)
	if err != nil {
		t.Errorf("Organizations.ListPendingOrgInvitations returned error: %v", err)
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
		t.Errorf("Organizations.ListPendingOrgInvitations returned %+v, want %+v", invitations, want)
	}
}
