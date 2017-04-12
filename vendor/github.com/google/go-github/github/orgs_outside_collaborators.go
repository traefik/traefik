// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// ListOutsideCollaboratorsOptions specifies optional parameters to the
// OrganizationsService.ListOutsideCollaborators method.
type ListOutsideCollaboratorsOptions struct {
	// Filter outside collaborators returned in the list. Possible values are:
	// 2fa_disabled, all.  Default is "all".
	Filter string `url:"filter,omitempty"`

	ListOptions
}

// ListOutsideCollaborators lists outside collaborators of organization's repositories.
// This will only work if the authenticated
// user is an owner of the organization.
//
// Warning: The API may change without advance notice during the preview period.
// Preview features are not supported for production use.
//
// GitHub API docs: https://developer.github.com/v3/orgs/outside_collaborators/#list-outside-collaborators
func (s *OrganizationsService) ListOutsideCollaborators(ctx context.Context, org string, opt *ListOutsideCollaboratorsOptions) ([]*User, *Response, error) {
	u := fmt.Sprintf("orgs/%v/outside_collaborators", org)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var members []*User
	resp, err := s.client.Do(ctx, req, &members)
	if err != nil {
		return nil, resp, err
	}

	return members, resp, nil
}
