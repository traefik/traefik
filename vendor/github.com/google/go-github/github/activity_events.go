// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a GitHub event.
type Event struct {
	Type       *string          `json:"type,omitempty"`
	Public     *bool            `json:"public"`
	RawPayload *json.RawMessage `json:"payload,omitempty"`
	Repo       *Repository      `json:"repo,omitempty"`
	Actor      *User            `json:"actor,omitempty"`
	Org        *Organization    `json:"org,omitempty"`
	CreatedAt  *time.Time       `json:"created_at,omitempty"`
	ID         *string          `json:"id,omitempty"`
}

func (e Event) String() string {
	return Stringify(e)
}

// ParsePayload parses the event payload. For recognized event types,
// a value of the corresponding struct type will be returned.
func (e *Event) ParsePayload() (payload interface{}, err error) {
	switch *e.Type {
	case "CommitCommentEvent":
		payload = &CommitCommentEvent{}
	case "CreateEvent":
		payload = &CreateEvent{}
	case "DeleteEvent":
		payload = &DeleteEvent{}
	case "DeploymentEvent":
		payload = &DeploymentEvent{}
	case "DeploymentStatusEvent":
		payload = &DeploymentStatusEvent{}
	case "ForkEvent":
		payload = &ForkEvent{}
	case "GollumEvent":
		payload = &GollumEvent{}
	case "InstallationEvent":
		payload = &InstallationEvent{}
	case "InstallationRepositoriesEvent":
		payload = &InstallationRepositoriesEvent{}
	case "IssueCommentEvent":
		payload = &IssueCommentEvent{}
	case "IssuesEvent":
		payload = &IssuesEvent{}
	case "LabelEvent":
		payload = &LabelEvent{}
	case "MemberEvent":
		payload = &MemberEvent{}
	case "MembershipEvent":
		payload = &MembershipEvent{}
	case "MilestoneEvent":
		payload = &MilestoneEvent{}
	case "OrganizationEvent":
		payload = &OrganizationEvent{}
	case "OrgBlockEvent":
		payload = &OrgBlockEvent{}
	case "PageBuildEvent":
		payload = &PageBuildEvent{}
	case "PingEvent":
		payload = &PingEvent{}
	case "ProjectEvent":
		payload = &ProjectEvent{}
	case "ProjectCardEvent":
		payload = &ProjectCardEvent{}
	case "ProjectColumnEvent":
		payload = &ProjectColumnEvent{}
	case "PublicEvent":
		payload = &PublicEvent{}
	case "PullRequestEvent":
		payload = &PullRequestEvent{}
	case "PullRequestReviewEvent":
		payload = &PullRequestReviewEvent{}
	case "PullRequestReviewCommentEvent":
		payload = &PullRequestReviewCommentEvent{}
	case "PushEvent":
		payload = &PushEvent{}
	case "ReleaseEvent":
		payload = &ReleaseEvent{}
	case "RepositoryEvent":
		payload = &RepositoryEvent{}
	case "StatusEvent":
		payload = &StatusEvent{}
	case "TeamEvent":
		payload = &TeamEvent{}
	case "TeamAddEvent":
		payload = &TeamAddEvent{}
	case "WatchEvent":
		payload = &WatchEvent{}
	}
	err = json.Unmarshal(*e.RawPayload, &payload)
	return payload, err
}

// Payload returns the parsed event payload. For recognized event types,
// a value of the corresponding struct type will be returned.
//
// Deprecated: Use ParsePayload instead, which returns an error
// rather than panics if JSON unmarshaling raw payload fails.
func (e *Event) Payload() (payload interface{}) {
	var err error
	payload, err = e.ParsePayload()
	if err != nil {
		panic(err)
	}
	return payload
}

// ListEvents drinks from the firehose of all public events across GitHub.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-public-events
func (s *ActivityService) ListEvents(ctx context.Context, opt *ListOptions) ([]*Event, *Response, error) {
	u, err := addOptions("events", opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListRepositoryEvents lists events for a repository.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-repository-events
func (s *ActivityService) ListRepositoryEvents(ctx context.Context, owner, repo string, opt *ListOptions) ([]*Event, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/events", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListIssueEventsForRepository lists issue events for a repository.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-issue-events-for-a-repository
func (s *ActivityService) ListIssueEventsForRepository(ctx context.Context, owner, repo string, opt *ListOptions) ([]*IssueEvent, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/events", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*IssueEvent
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListEventsForRepoNetwork lists public events for a network of repositories.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-public-events-for-a-network-of-repositories
func (s *ActivityService) ListEventsForRepoNetwork(ctx context.Context, owner, repo string, opt *ListOptions) ([]*Event, *Response, error) {
	u := fmt.Sprintf("networks/%v/%v/events", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListEventsForOrganization lists public events for an organization.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-public-events-for-an-organization
func (s *ActivityService) ListEventsForOrganization(ctx context.Context, org string, opt *ListOptions) ([]*Event, *Response, error) {
	u := fmt.Sprintf("orgs/%v/events", org)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListEventsPerformedByUser lists the events performed by a user. If publicOnly is
// true, only public events will be returned.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-events-performed-by-a-user
func (s *ActivityService) ListEventsPerformedByUser(ctx context.Context, user string, publicOnly bool, opt *ListOptions) ([]*Event, *Response, error) {
	var u string
	if publicOnly {
		u = fmt.Sprintf("users/%v/events/public", user)
	} else {
		u = fmt.Sprintf("users/%v/events", user)
	}
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListEventsReceivedByUser lists the events received by a user. If publicOnly is
// true, only public events will be returned.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-events-that-a-user-has-received
func (s *ActivityService) ListEventsReceivedByUser(ctx context.Context, user string, publicOnly bool, opt *ListOptions) ([]*Event, *Response, error) {
	var u string
	if publicOnly {
		u = fmt.Sprintf("users/%v/received_events/public", user)
	} else {
		u = fmt.Sprintf("users/%v/received_events", user)
	}
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}

// ListUserEventsForOrganization provides the user’s organization dashboard. You
// must be authenticated as the user to view this.
//
// GitHub API docs: https://developer.github.com/v3/activity/events/#list-events-for-an-organization
func (s *ActivityService) ListUserEventsForOrganization(ctx context.Context, org, user string, opt *ListOptions) ([]*Event, *Response, error) {
	u := fmt.Sprintf("users/%v/events/orgs/%v", user, org)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*Event
	resp, err := s.client.Do(ctx, req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, nil
}
