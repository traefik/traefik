// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diff

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		desc   string
		A, B   []string
		chunks []Chunk
	}{
		{
			desc: "constitution",
			A: []string{
				"We the People of the United States, in Order to form a more perfect Union,",
				"establish Justice, insure domestic Tranquility, provide for the common defence,",
				"and secure the Blessings of Liberty to ourselves",
				"and our Posterity, do ordain and establish this Constitution for the United",
				"States of America.",
			},
			B: []string{
				"We the People of the United States, in Order to form a more perfect Union,",
				"establish Justice, insure domestic Tranquility, provide for the common defence,",
				"promote the general Welfare, and secure the Blessings of Liberty to ourselves",
				"and our Posterity, do ordain and establish this Constitution for the United",
				"States of America.",
			},
			chunks: []Chunk{
				0: {
					Equal: []string{
						"We the People of the United States, in Order to form a more perfect Union,",
						"establish Justice, insure domestic Tranquility, provide for the common defence,",
					},
				},
				1: {
					Deleted: []string{
						"and secure the Blessings of Liberty to ourselves",
					},
				},
				2: {
					Added: []string{
						"promote the general Welfare, and secure the Blessings of Liberty to ourselves",
					},
					Equal: []string{
						"and our Posterity, do ordain and establish this Constitution for the United",
						"States of America.",
					},
				},
			},
		},
	}

	for _, test := range tests {
		got := DiffChunks(test.A, test.B)
		if got, want := len(got), len(test.chunks); got != want {
			t.Errorf("%s: edit distance = %v, want %v", test.desc, got-1, want-1)
			continue
		}
		for i := range got {
			got, want := got[i], test.chunks[i]
			if got, want := got.Added, want.Added; !reflect.DeepEqual(got, want) {
				t.Errorf("%s[%d]: Added = %v, want %v", test.desc, i, got, want)
			}
			if got, want := got.Deleted, want.Deleted; !reflect.DeepEqual(got, want) {
				t.Errorf("%s[%d]: Deleted = %v, want %v", test.desc, i, got, want)
			}
			if got, want := got.Equal, want.Equal; !reflect.DeepEqual(got, want) {
				t.Errorf("%s[%d]: Equal = %v, want %v", test.desc, i, got, want)
			}
		}
	}
}

func ExampleDiff() {
	constitution := strings.TrimSpace(`
We the People of the United States, in Order to form a more perfect Union,
establish Justice, insure domestic Tranquility, provide for the common defence,
promote the general Welfare, and secure the Blessings of Liberty to ourselves
and our Posterity, do ordain and establish this Constitution for the United
States of America.
`)

	got := strings.TrimSpace(`
:wq
We the People of the United States, in Order to form a more perfect Union,
establish Justice, insure domestic Tranquility, provide for the common defence,
and secure the Blessings of Liberty to ourselves
and our Posterity, do ordain and establish this Constitution for the United
States of America.
`)

	fmt.Println(Diff(got, constitution))

	// Output:
	// -:wq
	//  We the People of the United States, in Order to form a more perfect Union,
	//  establish Justice, insure domestic Tranquility, provide for the common defence,
	// -and secure the Blessings of Liberty to ourselves
	// +promote the general Welfare, and secure the Blessings of Liberty to ourselves
	//  and our Posterity, do ordain and establish this Constitution for the United
	//  States of America.
}
