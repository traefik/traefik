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
	"fmt"
	"net/http"
	"strings"

	networking "k8s.io/api/networking/v1"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("disable-proxy-intercept-errors", func() {
	f := framework.NewDefaultFramework("disable-proxy-intercept-errors")

	ginkgo.BeforeEach(func() {
		f.NewHttpbunDeployment()
		f.NewEchoDeployment()
	})

	ginkgo.It("configures Nginx correctly", func() {
		host := "pie.foo.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-http-errors":             "404",
			"nginx.ingress.kubernetes.io/disable-proxy-intercept-errors": "true",
			"nginx.ingress.kubernetes.io/default-backend":                framework.EchoService,
		}

		ingHTTPBunService := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.HTTPBunService, 80, annotations)
		f.EnsureIngress(ingHTTPBunService)

		var serverConfig string
		f.WaitForNginxServer(host, func(sc string) bool {
			serverConfig = sc
			return strings.Contains(serverConfig, fmt.Sprintf(`server_name "%s"`, host))
		})

		ginkgo.By("turning off proxy_intercept_errors directive")
		assert.NotContains(ginkgo.GinkgoT(), serverConfig, "proxy_intercept_errors on;")

		// the plan for client side testing
		// create ingress where we disable intercept for code 404 - that error should get to the client
		// the same ingress should intercept any other error (>300 and not 404) where we will get intercepted error
		ginkgo.By("client test to check response - with intercept disabled")
		requestID := "proxy_intercept_errors"

		f.HTTPTestClient().
			GET("/status/404").
			WithHeader("Host", host).
			WithHeader("x-request-id", requestID).
			Expect().
			Status(http.StatusNotFound).
			Body().Empty()

		ginkgo.By("client test to check response - with intercept enabled")
		err := framework.UpdateIngress(f.KubeClientSet, f.Namespace, host, func(ingress *networking.Ingress) error {
			ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/disable-proxy-intercept-errors"] = "false"
			return nil
		})
		assert.Nil(ginkgo.GinkgoT(), err)

		f.WaitForNginxServer(host, func(sc string) bool {
			if serverConfig != sc {
				serverConfig = sc
				return true
			}
			return false
		})

		f.HTTPTestClient().
			GET("/status/404").
			WithHeader("Host", host).
			WithHeader("x-request-id", requestID).
			Expect().
			Status(http.StatusOK).
			Body().Contains("x-code=404").
			Contains(fmt.Sprintf("x-ingress-name=%s", host)).
			Contains(fmt.Sprintf("x-service-name=%s", framework.HTTPBunService)).
			Contains(fmt.Sprintf("x-request-id=%s", requestID))
	})
})
