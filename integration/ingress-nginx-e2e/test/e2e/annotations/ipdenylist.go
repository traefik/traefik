/*
Copyright 2023 The Kubernetes Authors.

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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("denylist-source-range", func() {
	f := framework.NewDefaultFramework("ipdenylist")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("only deny explicitly denied IPs, allow all others", func() {
		host := "ipdenylist.foo.com"
		namespace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/denylist-source-range": "18.0.0.0/8, 56.0.0.1",
		}

		ing := framework.NewSingleIngress(host, "/", host, namespace, framework.EchoService, 80, annotations)

		// Temporarily trust forwarded headers so we can test IP based access control
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")
		defer func() {
			// Return to the original value
			f.UpdateNginxConfigMapData("use-forwarded-headers", "false")
		}()

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "deny 18.0.0.0/8;") &&
					strings.Contains(server, "deny 56.0.0.1;") &&
					!strings.Contains(server, "deny all;")
			})

		ginkgo.By("sending request from an explicitly denied IP range")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "18.0.0.1").
			Expect().
			Status(http.StatusForbidden)

		ginkgo.By("sending request from an explicitly denied IP address")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "56.0.0.1").
			Expect().
			Status(http.StatusForbidden)

		ginkgo.By("sending request from an implicitly allowed IP range")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "56.0.0.2").
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("only allow explicitly allowed IPs, deny all others", func() {
		host := "ipdenylist.foo.com"
		namespace := f.Namespace

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/denylist-source-range":  "18.1.0.0/16, 56.0.0.0/8",
			"nginx.ingress.kubernetes.io/whitelist-source-range": "18.0.0.0/8, 55.0.0.0/8",
		}

		ing := framework.NewSingleIngress(host, "/", host, namespace, framework.EchoService, 80, annotations)

		// Temporarily trust forwarded headers so we can test IP based access control
		f.UpdateNginxConfigMapData("use-forwarded-headers", "true")
		defer func() {
			// Return to the original value
			f.UpdateNginxConfigMapData("use-forwarded-headers", "false")
		}()

		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "deny 18.1.0.0/16;") &&
					strings.Contains(server, "deny 56.0.0.0/8;") &&
					strings.Contains(server, "allow 18.0.0.0/8;") &&
					strings.Contains(server, "allow 55.0.0.0/8;") &&
					strings.Contains(server, "deny all;")
			})

		ginkgo.By("sending request from an explicitly denied IP range")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "18.1.0.1").
			Expect().
			Status(http.StatusForbidden)

		ginkgo.By("sending request from an implicitly denied IP")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "10.10.10.10").
			Expect().
			Status(http.StatusForbidden)

		ginkgo.By("sending request from an explicitly allowed IP range")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "18.4.0.1").
			Expect().
			Status(http.StatusOK)

		ginkgo.By("sending request from an explicitly allowed IP range")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("X-Forwarded-For", "55.55.55.55").
			Expect().
			Status(http.StatusOK)
	})
})
