/*
Copyright 2018 The Kubernetes Authors.

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

package annotations

import (
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("allowlist-source-range", func() {
	f := framework.NewDefaultFramework("ipallowlist")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set valid ip allowlist range", func() {
		host := "ipallowlist.foo.com"
		nameSpace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/allowlist-source-range": "18.0.0.0/8, 56.0.0.0/8",
		}

		ing := framework.NewSingleIngress(host, "/", host, nameSpace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "allow 18.0.0.0/8;") &&
					strings.Contains(server, "allow 56.0.0.0/8;") &&
					strings.Contains(server, "deny all;")
			})
	})
})
