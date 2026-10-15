/*
Copyright 2016 The Kubernetes Authors.

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

package framework

import (
	"flag"
)

// TestContextType describes the client context to use in communications with the Kubernetes API.
type TestContextType struct {
	KubeHost string
	// KubeConfig  string
	KubeContext string
}

// TestContext is the global client context for tests.
var TestContext TestContextType

// registerCommonFlags registers flags common to all e2e test suites.
func registerCommonFlags() {
	flag.StringVar(&TestContext.KubeHost, "kubernetes-host", "http://127.0.0.1:8080", "The kubernetes host, or apiserver, to connect to")
	flag.StringVar(&TestContext.KubeContext, "kubernetes-context", "", "config context to use for kubernetes. If unset, will use value from 'current-context'")
}

// RegisterParseFlags registers and parses flags for the test binary.
func RegisterParseFlags() {
	registerCommonFlags()
	flag.Parse()
}
