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

type chain struct {
	reporter Reporter
	failbit  bool
}

func makeChain(reporter Reporter) chain {
	return chain{reporter, false}
}

func (c *chain) failed() bool {
	return c.failbit
}

func (c *chain) fail(message string, args ...interface{}) {
	if c.failbit {
		return
	}
	c.failbit = true
	c.reporter.Errorf(message, args...)
}
