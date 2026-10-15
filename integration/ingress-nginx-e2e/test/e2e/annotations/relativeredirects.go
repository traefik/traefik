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

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	relativeRedirectsHostname            = "rr.foo.com"
	relativeRedirectsRedirectPath        = "/something"
	relativeRedirectsRelativeRedirectURL = "/new-location"
)

var _ = framework.DescribeAnnotation("relative-redirects", func() {
	f := framework.NewDefaultFramework("relative-redirects")

	ginkgo.BeforeEach(func() {
		f.NewHttpbunDeployment()
		f.NewEchoDeployment()
	})

	ginkgo.It("configures Nginx correctly", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/relative-redirects": "true",
		}

		ing := framework.NewSingleIngress(relativeRedirectsHostname, "/", relativeRedirectsHostname, f.Namespace, framework.HTTPBunService, 80, annotations)
		f.EnsureIngress(ing)

		var serverConfig string
		f.WaitForNginxServer(relativeRedirectsHostname, func(srvCfg string) bool {
			serverConfig = srvCfg
			return strings.Contains(serverConfig, fmt.Sprintf(`server_name "%s"`, relativeRedirectsHostname))
		})

		ginkgo.By("turning off absolute_redirect directive")
		assert.Contains(ginkgo.GinkgoT(), serverConfig, "absolute_redirect off;")
	})

	ginkgo.It("should respond with absolute URL in Location", func() {
		absoluteRedirectURL := fmt.Sprintf("http://%s%s", relativeRedirectsHostname, relativeRedirectsRelativeRedirectURL)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/permanent-redirect": relativeRedirectsRelativeRedirectURL,
			"nginx.ingress.kubernetes.io/relative-redirects": "false",
		}

		ginkgo.By("setup ingress")
		ing := framework.NewSingleIngress(relativeRedirectsHostname, relativeRedirectsRedirectPath, relativeRedirectsHostname, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(relativeRedirectsHostname, func(srvCfg string) bool {
			return strings.Contains(srvCfg, fmt.Sprintf(`server_name "%s"`, relativeRedirectsHostname))
		})

		ginkgo.By("sending request to redirected URL path")
		f.HTTPTestClient().
			GET(relativeRedirectsRedirectPath).
			WithHeader("Host", relativeRedirectsHostname).
			Expect().
			Status(http.StatusMovedPermanently).
			Header("Location").Equal(absoluteRedirectURL)
	})

	ginkgo.It("should respond with relative URL in Location", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/permanent-redirect": relativeRedirectsRelativeRedirectURL,
			"nginx.ingress.kubernetes.io/relative-redirects": "true",
		}

		ginkgo.By("setup ingress")
		ing := framework.NewSingleIngress(relativeRedirectsHostname, relativeRedirectsRedirectPath, relativeRedirectsHostname, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(relativeRedirectsHostname, func(srvCfg string) bool {
			return strings.Contains(srvCfg, fmt.Sprintf(`server_name "%s"`, relativeRedirectsHostname))
		})

		ginkgo.By("sending request to redirected URL path")
		f.HTTPTestClient().
			GET(relativeRedirectsRedirectPath).
			WithHeader("Host", relativeRedirectsHostname).
			Expect().
			Status(http.StatusMovedPermanently).
			Header("Location").Equal(relativeRedirectsRelativeRedirectURL)
	})
})
