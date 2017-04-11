// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"time"
)

// PullRequestReview represents a review of a pull request.
type PullRequestReview struct {
	ID             *int       `json:"id,omitempty"`
	User           *User      `json:"user,omitempty"`
	Body           *string    `json:"body,omitempty"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	CommitID       *string    `json:"commit_id,omitempty"`
	HTMLURL        *string    `json:"html_url,omitempty"`
	PullRequestURL *string    `json:"pull_request_url,omitempty"`
	State          *string    `json:"state,omitempty"`
}

func (p PullRequestReview) String() string {
	return Stringify(p)
}

// DraftReviewComment represents a comment part of the review.
type DraftReviewComment struct {
	Path     *string `json:"path,omitempty"`
	Position *int    `json:"position,omitempty"`
	Body     *string `json:"body,omitempty"`
}

func (c DraftReviewComment) String() string {
	return Stringify(c)
}

// PullRequestReviewRequest represents a request to create a review.
type PullRequestReviewRequest struct {
	Body     *string               `json:"body,omitempty"`
	Event    *string               `json:"event,omitempty"`
	Comments []*DraftReviewComment `json:"comments,omitempty"`
}

func (r PullRequestReviewRequest) String() string {
	return Stringify(r)
}

// PullRequestReviewDismissalRequest represents a request to dismiss a review.
type PullRequestReviewDismissalRequest struct {
	Message *string `json:"message,omitempty"`
}

func (r PullRequestReviewDismissalRequest) String() string {
	return Stringify(r)
}

// ListReviews lists all reviews on the specified pull request.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#list-reviews-on-a-pull-request
func (s *PullRequestsService) ListReviews(ctx context.Context, owner, repo string, number int) ([]*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews", owner, repo, number)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	var reviews []*PullRequestReview
	resp, err := s.client.Do(ctx, req, &reviews)
	if err != nil {
		return nil, resp, err
	}

	return reviews, resp, nil
}

// GetReview fetches the specified pull request review.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#get-a-single-review
func (s *PullRequestsService) GetReview(ctx context.Context, owner, repo string, number, reviewID int) (*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews/%d", owner, repo, number, reviewID)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	review := new(PullRequestReview)
	resp, err := s.client.Do(ctx, req, review)
	if err != nil {
		return nil, resp, err
	}

	return review, resp, nil
}

// DeletePendingReview deletes the specified pull request pending review.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#delete-a-pending-review
func (s *PullRequestsService) DeletePendingReview(ctx context.Context, owner, repo string, number, reviewID int) (*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews/%d", owner, repo, number, reviewID)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	review := new(PullRequestReview)
	resp, err := s.client.Do(ctx, req, review)
	if err != nil {
		return nil, resp, err
	}

	return review, resp, nil
}

// ListReviewComments lists all the comments for the specified review.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#get-a-single-reviews-comments
func (s *PullRequestsService) ListReviewComments(ctx context.Context, owner, repo string, number, reviewID int) ([]*PullRequestComment, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews/%d/comments", owner, repo, number, reviewID)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	var comments []*PullRequestComment
	resp, err := s.client.Do(ctx, req, &comments)
	if err != nil {
		return nil, resp, err
	}

	return comments, resp, nil
}

// CreateReview creates a new review on the specified pull request.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#create-a-pull-request-review
func (s *PullRequestsService) CreateReview(ctx context.Context, owner, repo string, number int, review *PullRequestReviewRequest) (*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews", owner, repo, number)

	req, err := s.client.NewRequest("POST", u, review)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	r := new(PullRequestReview)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// SubmitReview submits a specified review on the specified pull request.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#submit-a-pull-request-review
func (s *PullRequestsService) SubmitReview(ctx context.Context, owner, repo string, number, reviewID int, review *PullRequestReviewRequest) (*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews/%d/events", owner, repo, number, reviewID)

	req, err := s.client.NewRequest("POST", u, review)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	r := new(PullRequestReview)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}

// DismissReview dismisses a specified review on the specified pull request.
//
// TODO: Follow up with GitHub support about an issue with this method's
// returned error format and remove this comment once it's fixed.
// Read more about it here - https://github.com/google/go-github/issues/540
//
// GitHub API docs: https://developer.github.com/v3/pulls/reviews/#dismiss-a-pull-request-review
func (s *PullRequestsService) DismissReview(ctx context.Context, owner, repo string, number, reviewID int, review *PullRequestReviewDismissalRequest) (*PullRequestReview, *Response, error) {
	u := fmt.Sprintf("repos/%v/%v/pulls/%d/reviews/%d/dismissals", owner, repo, number, reviewID)

	req, err := s.client.NewRequest("PUT", u, review)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches
	req.Header.Set("Accept", mediaTypePullRequestReviewsPreview)

	r := new(PullRequestReview)
	resp, err := s.client.Do(ctx, req, r)
	if err != nil {
		return nil, resp, err
	}

	return r, resp, nil
}
