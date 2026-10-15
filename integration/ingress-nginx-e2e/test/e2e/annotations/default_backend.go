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
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("default-backend", func() {
	f := framework.NewDefaultFramework("default-backend")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.Context("when default backend annotation is enabled", func() {
		ginkgo.It("should use a custom default backend as upstream", func() {
			host := "default-backend"
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/default-backend": framework.EchoService,
			}

			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, "invalid", 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, fmt.Sprintf(`server_name "%v"`, host))
				})

			requestID := "something-unique"

			f.HTTPTestClient().
				GET("/alma/armud").
				WithHeader("Host", host).
				WithHeader("x-request-id", requestID).
				Expect().
				Status(http.StatusOK).
				Body().Contains("x-code=503").
				Contains(fmt.Sprintf("x-ingress-name=%s", host)).
				Contains("x-service-name=invalid").
				Contains(fmt.Sprintf("x-request-id=%s", requestID))
		})
	})
})
