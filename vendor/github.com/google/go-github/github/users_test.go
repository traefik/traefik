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
)

func TestUser_marshall(t *testing.T) {
	testJSONMarshal(t, &User{}, "{}")

	u := &User{
		Login:       String("l"),
		ID:          Int(1),
		URL:         String("u"),
		AvatarURL:   String("a"),
		GravatarID:  String("g"),
		Name:        String("n"),
		Company:     String("c"),
		Blog:        String("b"),
		Location:    String("l"),
		Email:       String("e"),
		Hireable:    Bool(true),
		PublicRepos: Int(1),
		Followers:   Int(1),
		Following:   Int(1),
		CreatedAt:   &Timestamp{referenceTime},
		SuspendedAt: &Timestamp{referenceTime},
	}
	want := `{
		"login": "l",
		"id": 1,
		"avatar_url": "a",
		"gravatar_id": "g",
		"name": "n",
		"company": "c",
		"blog": "b",
		"location": "l",
		"email": "e",
		"hireable": true,
		"public_repos": 1,
		"followers": 1,
		"following": 1,
		"created_at": ` + referenceTimeStr + `,
		"suspended_at": ` + referenceTimeStr + `,
		"url": "u"
	}`
	testJSONMarshal(t, u, want)
}

func TestUsersService_Get_authenticatedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		t.Errorf("Users.Get returned error: %v", err)
	}

	want := &User{ID: Int(1)}
	if !reflect.DeepEqual(user, want) {
		t.Errorf("Users.Get returned %+v, want %+v", user, want)
	}
}

func TestUsersService_Get_specifiedUser(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users/u", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	user, _, err := client.Users.Get(context.Background(), "u")
	if err != nil {
		t.Errorf("Users.Get returned error: %v", err)
	}

	want := &User{ID: Int(1)}
	if !reflect.DeepEqual(user, want) {
		t.Errorf("Users.Get returned %+v, want %+v", user, want)
	}
}

func TestUsersService_Get_invalidUser(t *testing.T) {
	_, _, err := client.Users.Get(context.Background(), "%")
	testURLParseError(t, err)
}

func TestUsersService_GetByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})

	user, _, err := client.Users.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Users.GetByID returned error: %v", err)
	}

	want := &User{ID: Int(1)}
	if !reflect.DeepEqual(user, want) {
		t.Errorf("Users.GetByID returned %+v, want %+v", user, want)
	}
}

func TestUsersService_Edit(t *testing.T) {
	setup()
	defer teardown()

	input := &User{Name: String("n")}

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		v := new(User)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "PATCH")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"id":1}`)
	})

	user, _, err := client.Users.Edit(context.Background(), input)
	if err != nil {
		t.Errorf("Users.Edit returned error: %v", err)
	}

	want := &User{ID: Int(1)}
	if !reflect.DeepEqual(user, want) {
		t.Errorf("Users.Edit returned %+v, want %+v", user, want)
	}
}

func TestUsersService_ListAll(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"since": "1", "page": "2"})
		fmt.Fprint(w, `[{"id":2}]`)
	})

	opt := &UserListOptions{1, ListOptions{Page: 2}}
	users, _, err := client.Users.ListAll(context.Background(), opt)
	if err != nil {
		t.Errorf("Users.Get returned error: %v", err)
	}

	want := []*User{{ID: Int(2)}}
	if !reflect.DeepEqual(users, want) {
		t.Errorf("Users.ListAll returned %+v, want %+v", users, want)
	}
}

func TestUsersService_ListInvitations(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repository_invitations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		fmt.Fprintf(w, `[{"id":1}, {"id":2}]`)
	})

	got, _, err := client.Users.ListInvitations(context.Background(), nil)
	if err != nil {
		t.Errorf("Users.ListInvitations returned error: %v", err)
	}

	want := []*RepositoryInvitation{{ID: Int(1)}, {ID: Int(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Users.ListInvitations = %+v, want %+v", got, want)
	}
}

func TestUsersService_ListInvitations_withOptions(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repository_invitations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"page": "2",
		})
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		fmt.Fprintf(w, `[{"id":1}, {"id":2}]`)
	})

	_, _, err := client.Users.ListInvitations(context.Background(), &ListOptions{Page: 2})
	if err != nil {
		t.Errorf("Users.ListInvitations returned error: %v", err)
	}
}
func TestUsersService_AcceptInvitation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repository_invitations/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Users.AcceptInvitation(context.Background(), 1); err != nil {
		t.Errorf("Users.AcceptInvitation returned error: %v", err)
	}
}

func TestUsersService_DeclineInvitation(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/user/repository_invitations/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		testHeader(t, r, "Accept", mediaTypeRepositoryInvitationsPreview)
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.Users.DeclineInvitation(context.Background(), 1); err != nil {
		t.Errorf("Users.DeclineInvitation returned error: %v", err)
	}
}
