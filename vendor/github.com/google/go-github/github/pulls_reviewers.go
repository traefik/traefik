// Copyright 2017 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

// RequestReviewers creates a review request for the provided GitHub users for the specified pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/review_requests/#create-a-review-request
func (s *PullRequestsService) RequestReviewers(ctx context.Context, owner, repo string, number int, logins []string) (*PullRequest, *Response, error) {
	u := fmt.Sprintf("repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, number)

	reviewers := struct {
		Reviewers []string `json:"reviewers,omitempty"`
	}{
		Reviewers: logins,
	}
	req, err := s.client.NewRequest("POST", u, &reviewers)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	r := new(PullRequest)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// ListReviewers lists users whose reviews have been requested on the specified pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/review_requests/#list-review-requests
func (s *PullRequestsService) ListReviewers(ctx context.Context, owner, repo string, number int) ([]*User, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/requested_reviewers", owner, repo, number)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	var users []*User
	resp, err := s.client.Do(ctx, req, &users)
	if err != nil {
		return nil, resp, err
	}

	return users, resp, nil
}

// RemoveReviewers removes the review request for the provided GitHub users for the specified pull request.
//
// GitHub API docs: https://developer.github.com/v3/pulls/review_requests/#delete-a-review-request
func (s *PullRequestsService) RemoveReviewers(ctx context.Context, owner, repo string, number int, logins []string) (*Response, error) {
	u := fmt.Sprintf("repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, number)

	reviewers := struct {
		Reviewers []string `json:"reviewers,omitempty"`
	}{
		Reviewers: logins,
	}
	req, err := s.client.NewRequest("DELETE", u, &reviewers)
	if err != nil {
		return nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	return s.client.Do(ctx, req, reviewers)
}
