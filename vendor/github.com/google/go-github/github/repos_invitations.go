// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "fmt"

// RepositoryInvitation represents an invitation to collaborate on a repo.
type RepositoryInvitation struct {
	ID      *int        `json:"id,omitempty"`
	Repo    *Repository `json:"repository,omitempty"`
	Invitee *User       `json:"invitee,omitempty"`
	Inviter *User       `json:"inviter,omitempty"`

	// Permissions represents the permissions that the associated user will have
	// on the repository. Possible values are: "read", "write", "admin".
	Permissions *string    `json:"permissions,omitempty"`
	CreatedAt   *Timestamp `json:"created_at,omitempty"`
	URL         *string    `json:"url,omitempty"`
	HTMLURL     *string    `json:"html_url,omitempty"`
}

// ListInvitations lists all currently-open repository invitations.
//
// GitHub API docs: https://developer.github.com/v3/repos/invitations/#list-invitations-for-a-repository
func (s *RepositoriesService) ListInvitations(repoID int, opt *ListOptions) ([]*RepositoryInvitation, *Response, error) {
	u := fmt.Sprintf("repositories/%v/invitations", repoID)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeRepositoryInvitationsPreview)

	invites := []*RepositoryInvitation{}
	resp, err := s.client.Do(req, &invites)
	if err != nil {
		return nil, resp, err
	}

	return invites, resp, err
}

// DeleteInvitation deletes a repository invitation.
//
// GitHub API docs: https://developer.github.com/v3/repos/invitations/#delete-a-repository-invitation
func (s *RepositoriesService) DeleteInvitation(repoID, invitationID int) (*Response, error) {
	u := fmt.Sprintf("repositories/%v/invitations/%v", repoID, invitationID)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeRepositoryInvitationsPreview)

	return s.client.Do(req, nil)
}

// UpdateInvitation updates the permissions associated with a repository
// invitation.
//
// permissions represents the permissions that the associated user will have
// on the repository. Possible values are: "read", "write", "admin".
//
// GitHub API docs: https://developer.github.com/v3/repos/invitations/#update-a-repository-invitation
func (s *RepositoriesService) UpdateInvitation(repoID, invitationID int, permissions string) (*RepositoryInvitation, *Response, error) {
	opts := &struct {
		Permissions string `json:"permissions"`
	}{Permissions: permissions}
	u := fmt.Sprintf("repositories/%v/invitations/%v", repoID, invitationID)
	req, err := s.client.NewRequest("PATCH", u, opts)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeRepositoryInvitationsPreview)

	invite := &RepositoryInvitation{}
	resp, err := s.client.Do(req, invite)
	return invite, resp, err
}
