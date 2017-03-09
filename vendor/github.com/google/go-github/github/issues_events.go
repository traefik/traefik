// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"fmt"
	"time"
)

// IssueEvent represents an event that occurred around an Issue or Pull Request.
type IssueEvent struct {
	ID  *int    `json:"id,omitempty"`
	URL *string `json:"url,omitempty"`

	// The User that generated this event.
	Actor *User `json:"actor,omitempty"`

	// Event identifies the actual type of Event that occurred.  Possible
	// values are:
	//
	//     closed
	//       The Actor closed the issue.
	//       If the issue was closed by commit message, CommitID holds the SHA1 hash of the commit.
	//
	//     merged
	//       The Actor merged into master a branch containing a commit mentioning the issue.
	//       CommitID holds the SHA1 of the merge commit.
	//
	//     referenced
	//       The Actor committed to master a commit mentioning the issue in its commit message.
	//       CommitID holds the SHA1 of the commit.
	//
	//     reopened, locked, unlocked
	//       The Actor did that to the issue.
	//
	//     renamed
	//       The Actor changed the issue title from Rename.From to Rename.To.
	//
	//     mentioned
	//       Someone unspecified @mentioned the Actor [sic] in an issue comment body.
	//
	//     assigned, unassigned
	//       The Actor assigned the issue to or removed the assignment from the Assignee.
	//
	//     labeled, unlabeled
	//       The Actor added or removed the Label from the issue.
	//
	//     milestoned, demilestoned
	//       The Actor added or removed the issue from the Milestone.
	//
	//     subscribed, unsubscribed
	//       The Actor subscribed to or unsubscribed from notifications for an issue.
	//
	//     head_ref_deleted, head_ref_restored
	//       The pull request’s branch was deleted or restored.
	//
	Event *string `json:"event,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	Issue     *Issue     `json:"issue,omitempty"`

	// Only present on certain events; see above.
	Assignee  *User      `json:"assignee,omitempty"`
	CommitID  *string    `json:"commit_id,omitempty"`
	Milestone *Milestone `json:"milestone,omitempty"`
	Label     *Label     `json:"label,omitempty"`
	Rename    *Rename    `json:"rename,omitempty"`
}

// ListIssueEvents lists events for the specified issue.
//
// GitHub API docs: https://developer.github.com/v3/issues/events/#list-events-for-an-issue
func (s *IssuesService) ListIssueEvents(owner, repo string, number int, opt *ListOptions) ([]*IssueEvent, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/%v/events", owner, repo, number)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var events []*IssueEvent
	resp, err := s.client.Do(req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, err
}

// ListRepositoryEvents lists events for the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/issues/events/#list-events-for-a-repository
func (s *IssuesService) ListRepositoryEvents(owner, repo string, opt *ListOptions) ([]*IssueEvent, *Response, error) {
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
	resp, err := s.client.Do(req, &events)
	if err != nil {
		return nil, resp, err
	}

	return events, resp, err
}

// GetEvent returns the specified issue event.
//
// GitHub API docs: https://developer.github.com/v3/issues/events/#get-a-single-event
func (s *IssuesService) GetEvent(owner, repo string, id int) (*IssueEvent, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/issues/events/%v", owner, repo, id)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	event := new(IssueEvent)
	resp, err := s.client.Do(req, event)
	if err != nil {
		return nil, resp, err
	}

	return event, resp, err
}

// Rename contains details for 'renamed' events.
type Rename struct {
	From *string `json:"from,omitempty"`
	To   *string `json:"to,omitempty"`
}

func (r Rename) String() string {
	return Stringify(r)
}
