// Copyright 2017 Google Inc. All Rights Reserved.
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

package pubsub

import (
	"reflect"
	"testing"

	pb "google.golang.org/genproto/googleapis/pubsub/v1"
)

func TestSplitRequest(t *testing.T) {
	split := func(a []string, i int) ([]string, []string) {
		if len(a) < i {
			return a, nil
		}
		return a[:i], a[i:]
	}
	ackIDs := []string{"aaaa", "bbbb", "cccc", "dddd", "eeee"}
	modDeadlines := []int32{1, 2, 3, 4, 5}
	for i, test := range []struct {
		ackIDs     []string
		modAckIDs  []string
		splitIndex int
	}{
		{ackIDs, ackIDs, 2},
		{nil, ackIDs, 3},
		{ackIDs, nil, 5},
		{nil, ackIDs[:1], 1},
	} {
		req := &pb.StreamingPullRequest{
			AckIds:                test.ackIDs,
			ModifyDeadlineAckIds:  test.modAckIDs,
			ModifyDeadlineSeconds: modDeadlines[:len(test.modAckIDs)],
		}
		a1, a2 := split(test.ackIDs, test.splitIndex)
		m1, m2 := split(test.modAckIDs, test.splitIndex)
		want1 := &pb.StreamingPullRequest{
			AckIds:                a1,
			ModifyDeadlineAckIds:  m1,
			ModifyDeadlineSeconds: modDeadlines[:len(m1)],
		}
		want2 := &pb.StreamingPullRequest{
			AckIds:                a2,
			ModifyDeadlineAckIds:  m2,
			ModifyDeadlineSeconds: modDeadlines[len(m1) : len(m1)+len(m2)],
		}
		got1, got2 := splitRequest(req, reqFixedOverhead+40)
		if !reflect.DeepEqual(got1, want1) {
			t.Errorf("#%d: first:\ngot  %+v\nwant %+v", i, got1, want1)
		}
		if !reflect.DeepEqual(got2, want2) {
			t.Errorf("#%d: second:\ngot  %+v\nwant %+v", i, got2, want2)
		}
	}
}
