// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// ListBlockedUsers lists all the blocked users by the authenticated user.
//
// GitHub API docs: https://developer.github.com/v3/users/blocking/#list-blocked-users
func (s *UsersService) ListBlockedUsers(ctx context.Context, opt *ListOptions) ([]*User, *Response, error) {
	u := "user/blocks"
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeBlockUsersPreview)

	var blockedUsers []*User
	resp, err := s.client.Do(ctx, req, &blockedUsers)
	if err != nil {
		return nil, resp, err
	}

	return blockedUsers, resp, nil
}

// IsBlocked reports whether specified user is blocked by the authenticated user.
//
// GitHub API docs: https://developer.github.com/v3/users/blocking/#check-whether-youve-blocked-a-user
func (s *UsersService) IsBlocked(ctx context.Context, user string) (bool, *Response, error) {
	u := fmt.Sprintf("user/blocks/%v", user)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return false, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeBlockUsersPreview)

	resp, err := s.client.Do(ctx, req, nil)
	isBlocked, err := parseBoolResponse(err)
	return isBlocked, resp, err
}

// BlockUser blocks specified user for the authenticated user.
//
// GitHub API docs: https://developer.github.com/v3/users/blocking/#block-a-user
func (s *UsersService) BlockUser(ctx context.Context, user string) (*Response, error) {
	u := fmt.Sprintf("user/blocks/%v", user)

	req, err := s.client.NewRequest("PUT", u, nil)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeBlockUsersPreview)

	return s.client.Do(ctx, req, nil)
}

// UnblockUser unblocks specified user for the authenticated user.
//
// GitHub API docs: https://developer.github.com/v3/users/blocking/#unblock-a-user
func (s *UsersService) UnblockUser(ctx context.Context, user string) (*Response, error) {
	u := fmt.Sprintf("user/blocks/%v", user)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeBlockUsersPreview)

	return s.client.Do(ctx, req, nil)
}
