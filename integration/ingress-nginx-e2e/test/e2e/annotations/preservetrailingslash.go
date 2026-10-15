/*
Copyright 2021 The Kubernetes Authors.

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

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("preserve-trailing-slash", func() {
	f := framework.NewDefaultFramework("preservetrailingslash")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should allow preservation of trailing slashes", func() {
		host := "forcesslredirect.bar.com"
		tlsHosts := []string{host}

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/ssl-redirect":            "true",
			"nginx.ingress.kubernetes.io/preserve-trailing-slash": "true",
		}

		ing := framework.NewSingleIngressWithTLS(host, "/", host, tlsHosts, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusPermanentRedirect).
			Header("Location").Equal("https://forcesslredirect.bar.com/")
	})
})
