// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file supports generating unique IDs so that multiple test executions
// don't interfere with each other, and cleaning up old entities that may
// remain if tests exit early.

package testing

import (
	"reflect"
	"testing"
	"time"
)

func TestExtractTime(t *testing.T) {
	uid := UniqueID("unique-ID")
	got, ok := extractTime(uid, "unique-ID")
	if !ok {
		t.Fatal("got ok = false, want true")
	}
	if !startTime.Equal(got) {
		t.Errorf("got %s, want %s", got, startTime)
	}

	got, ok = extractTime("p-t0-doesnotmatter", "p")
	if !ok {
		t.Fatal("got false, want true")
	}
	if want := time.Unix(0, 0); !want.Equal(got) {
		t.Errorf("got %s, want %s", got, want)
	}
	if _, ok = extractTime("invalid-time-1234", "invalid"); ok {
		t.Error("got true, want false")
	}
}

func TestExpiredUniqueIDs(t *testing.T) {
	const prefix = "uid"
	// The freshly unique IDs will have startTime as their timestamp.
	uids := []string{UniqueID(prefix), "uid-tinvalid-1234", UniqueID(prefix), "uid-t0-1111"}

	// This test hasn't been running for very long, so only the last ID is expired.
	got := ExpiredUniqueIDs(uids, prefix)
	want := []string{uids[3]}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	time.Sleep(100 * time.Millisecond)

	prev := expiredAge
	expiredAge = 10 * time.Millisecond
	defer func() { expiredAge = prev }()
	// This test has been running for at least 10ms, so all but the invalid ID have expired.
	got = ExpiredUniqueIDs(uids, prefix)
	want = []string{uids[0], uids[2], uids[3]}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
