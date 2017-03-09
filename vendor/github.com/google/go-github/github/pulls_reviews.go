// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import "time"

// PullRequestReview represents a review of a pull request.
type PullRequestReview struct {
	ID          *int       `json:"id,omitempty"`
	User        *User      `json:"user,omitempty"`
	Body        *string    `json:"body,omitempty"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`

	// State can be "approved", "rejected", or "commented".
	State *string `json:"state,omitempty"`
}
