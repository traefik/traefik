/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("mirror-*", func() {
	f := framework.NewDefaultFramework("mirror")
	host := "mirror.foo.com"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set mirror-target to http://localhost/mirror", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/mirror-target": "http://localhost/mirror",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("mirror \"/_mirror-%v\";", ing.UID)) &&
					strings.Contains(server, "mirror_request_body on;")
			})
	})

	ginkgo.It("should set mirror-target to https://test.env.com/$request_uri", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/mirror-target": "https://test.env.com/$request_uri",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("mirror \"/_mirror-%v\";", ing.UID)) &&
					strings.Contains(server, "mirror_request_body on;") &&
					strings.Contains(server, `proxy_pass "https://test.env.com/$request_uri";`)
			})
	})

	ginkgo.It("should disable mirror-request-body", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/mirror-target":       "http://localhost/mirror",
			"nginx.ingress.kubernetes.io/mirror-request-body": "off",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		ing = f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("mirror \"/_mirror-%v\";", ing.UID)) &&
					strings.Contains(server, "mirror_request_body off;")
			})
	})
})
