// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// ListCollaborators lists the GitHub users that have access to the repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/collaborators/#list
func (s *RepositoriesService) ListCollaborators(ctx context.Context, owner, repo string, opt *ListOptions) ([]*User, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/collaborators", owner, repo)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var users []*User
	resp, err := s.client.Do(ctx, req, &users)
	if err != nil {
		return nil, resp, err
	}

	return users, resp, nil
}

// IsCollaborator checks whether the specified GitHub user has collaborator
// access to the given repo.
// Note: This will return false if the user is not a collaborator OR the user
// is not a GitHub user.
//
// GitHub API docs: https://developer.github.com/v3/repos/collaborators/#get
func (s *RepositoriesService) IsCollaborator(ctx context.Context, owner, repo, user string) (bool, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/collaborators/%v", owner, repo, user)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return false, nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	isCollab, err := parseBoolResponse(err)
	return isCollab, resp, err
}

// RepositoryPermissionLevel represents the permission level an organization
// member has for a given repository.
type RepositoryPermissionLevel struct {
	// Possible values: "admin", "write", "read", "none"
	Permission *string `json:"permission,omitempty"`

	User *User `json:"user,omitempty"`
}

// GetPermissionLevel retrieves the specific permission level a collaborator has for a given repository.
// GitHub API docs: https://developer.github.com/v3/repos/collaborators/#review-a-users-permission-level
func (s *RepositoriesService) GetPermissionLevel(ctx context.Context, owner, repo, user string) (*RepositoryPermissionLevel, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/collaborators/%v/permission", owner, repo, user)
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	rpl := new(RepositoryPermissionLevel)
	resp, err := s.client.Do(ctx, req, rpl)
	if err != nil {
		return nil, resp, err
	}
	return rpl, resp, nil
}

// RepositoryAddCollaboratorOptions specifies the optional parameters to the
// RepositoriesService.AddCollaborator method.
type RepositoryAddCollaboratorOptions struct {
	// Permission specifies the permission to grant the user on this repository.
	// Possible values are:
	//     pull - team members can pull, but not push to or administer this repository
	//     push - team members can pull and push, but not administer this repository
	//     admin - team members can pull, push and administer this repository
	//
	// Default value is "push". This option is only valid for organization-owned repositories.
	Permission string `json:"permission,omitempty"`
}

// AddCollaborator adds the specified GitHub user as collaborator to the given repo.
//
// GitHub API docs: https://developer.github.com/v3/repos/collaborators/#add-user-as-a-collaborator
func (s *RepositoriesService) AddCollaborator(ctx context.Context, owner, repo, user string, opt *RepositoryAddCollaboratorOptions) (*Response, error) {
	u := fmt.Sprintf("repos/%v/%v/collaborators/%v", owner, repo, user)
	req, err := s.client.NewRequest("PUT", u, opt)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeRepositoryInvitationsPreview)

	return s.client.Do(ctx, req, nil)
}

// RemoveCollaborator removes the specified GitHub user as collaborator from the given repo.
// Note: Does not return error if a valid user that is not a collaborator is removed.
//
// GitHub API docs: https://developer.github.com/v3/repos/collaborators/#remove-collaborator
func (s *RepositoriesService) RemoveCollaborator(ctx context.Context, owner, repo, user string) (*Response, error) {
	u := fmt.Sprintf("repos/%v/%v/collaborators/%v", owner, repo, user)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
