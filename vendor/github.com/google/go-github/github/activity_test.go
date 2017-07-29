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

func TestActivityService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/feeds", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")

		w.WriteHeader(http.StatusOK)
		w.Write(feedsJSON)
	})

	got, _, err := client.Activity.ListFeeds(context.Background())
	if err != nil {
		t.Errorf("Activity.ListFeeds returned error: %v", err)
	}
	if want := wantFeeds; !reflect.DeepEqual(got, want) {
		t.Errorf("Activity.ListFeeds = %+v, want %+v", got, want)
	}
}

var feedsJSON = []byte(`{
  "timeline_url": "https://github.com/timeline",
  "user_url": "https://github.com/{user}",
  "current_user_public_url": "https://github.com/defunkt",
  "current_user_url": "https://github.com/defunkt.private?token=abc123",
  "current_user_actor_url": "https://github.com/defunkt.private.actor?token=abc123",
  "current_user_organization_url": "",
  "current_user_organization_urls": [
    "https://github.com/organizations/github/defunkt.private.atom?token=abc123"
  ],
  "_links": {
    "timeline": {
      "href": "https://github.com/timeline",
      "type": "application/atom+xml"
    },
    "user": {
      "href": "https://github.com/{user}",
      "type": "application/atom+xml"
    },
    "current_user_public": {
      "href": "https://github.com/defunkt",
      "type": "application/atom+xml"
    },
    "current_user": {
      "href": "https://github.com/defunkt.private?token=abc123",
      "type": "application/atom+xml"
    },
    "current_user_actor": {
      "href": "https://github.com/defunkt.private.actor?token=abc123",
      "type": "application/atom+xml"
    },
    "current_user_organization": {
      "href": "",
      "type": ""
    },
    "current_user_organizations": [
      {
        "href": "https://github.com/organizations/github/defunkt.private.atom?token=abc123",
        "type": "application/atom+xml"
      }
    ]
  }
}`)

var wantFeeds = &Feeds{
	TimelineURL:                String("https://github.com/timeline"),
	UserURL:                    String("https://github.com/{user}"),
	CurrentUserPublicURL:       String("https://github.com/defunkt"),
	CurrentUserURL:             String("https://github.com/defunkt.private?token=abc123"),
	CurrentUserActorURL:        String("https://github.com/defunkt.private.actor?token=abc123"),
	CurrentUserOrganizationURL: String(""),
	CurrentUserOrganizationURLs: []string{
		"https://github.com/organizations/github/defunkt.private.atom?token=abc123",
	},
	Links: &struct {
		Timeline                 *FeedLink  `json:"timeline,omitempty"`
		User                     *FeedLink  `json:"user,omitempty"`
		CurrentUserPublic        *FeedLink  `json:"current_user_public,omitempty"`
		CurrentUser              *FeedLink  `json:"current_user,omitempty"`
		CurrentUserActor         *FeedLink  `json:"current_user_actor,omitempty"`
		CurrentUserOrganization  *FeedLink  `json:"current_user_organization,omitempty"`
		CurrentUserOrganizations []FeedLink `json:"current_user_organizations,omitempty"`
	}{
		Timeline: &FeedLink{
			HRef: String("https://github.com/timeline"),
			Type: String("application/atom+xml"),
		},
		User: &FeedLink{
			HRef: String("https://github.com/{user}"),
			Type: String("application/atom+xml"),
		},
		CurrentUserPublic: &FeedLink{
			HRef: String("https://github.com/defunkt"),
			Type: String("application/atom+xml"),
		},
		CurrentUser: &FeedLink{
			HRef: String("https://github.com/defunkt.private?token=abc123"),
			Type: String("application/atom+xml"),
		},
		CurrentUserActor: &FeedLink{
			HRef: String("https://github.com/defunkt.private.actor?token=abc123"),
			Type: String("application/atom+xml"),
		},
		CurrentUserOrganization: &FeedLink{
			HRef: String(""),
			Type: String(""),
		},
		CurrentUserOrganizations: []FeedLink{
			{
				HRef: String("https://github.com/organizations/github/defunkt.private.atom?token=abc123"),
				Type: String("application/atom+xml"),
			},
		},
	},
}
