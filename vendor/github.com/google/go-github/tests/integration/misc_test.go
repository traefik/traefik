// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"
	"time"
)

func TestEmojis(t *testing.T) {
	emoji, _, err := client.ListEmojis()
	if err != nil {
		t.Fatalf("ListEmojis returned error: %v", err)
	}

	if len(emoji) == 0 {
		t.Errorf("ListEmojis returned no emojis")
	}

	if _, ok := emoji["+1"]; !ok {
		t.Errorf("ListEmojis missing '+1' emoji")
	}
}

func TestAPIMeta(t *testing.T) {
	meta, _, err := client.APIMeta()
	if err != nil {
		t.Fatalf("APIMeta returned error: %v", err)
	}

	if len(meta.Hooks) == 0 {
		t.Errorf("APIMeta returned no hook addresses")
	}

	if len(meta.Git) == 0 {
		t.Errorf("APIMeta returned no git addresses")
	}

	if !*meta.VerifiablePasswordAuthentication {
		t.Errorf("APIMeta VerifiablePasswordAuthentication is false")
	}
}

func TestRateLimits(t *testing.T) {
	limits, _, err := client.RateLimits()
	if err != nil {
		t.Fatalf("RateLimits returned error: %v", err)
	}

	// do some sanity checks
	if limits.Core.Limit == 0 {
		t.Errorf("RateLimits returned 0 core limit")
	}

	if limits.Core.Limit < limits.Core.Remaining {
		t.Errorf("Core.Limits is less than Core.Remaining.")
	}

	if limits.Core.Reset.Time.Before(time.Now().Add(-1 * time.Minute)) {
		t.Errorf("Core.Reset is more than 1 minute in the past; that doesn't seem right.")
	}
}
