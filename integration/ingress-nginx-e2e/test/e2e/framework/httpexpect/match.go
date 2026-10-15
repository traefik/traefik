/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httpexpect

// Match provides methods to inspect attached regexp match results.
type Match struct {
	chain      chain
	submatches []string
	names      map[string]int
}

func makeMatch(chain chain, submatches, names []string) *Match {
	if submatches == nil {
		submatches = []string{}
	}
	namemap := map[string]int{}
	for n, name := range names {
		if name != "" {
			namemap[name] = n
		}
	}
	return &Match{chain, submatches, namemap}
}
