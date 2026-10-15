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

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	customHeaderHost = "custom-headers"
)

var _ = framework.DescribeAnnotation("custom-headers-*", func() {
	f := framework.NewDefaultFramework("custom-headers")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should return status code 200 when no custom-headers is configured", func() {
		ing := framework.NewSingleIngress(customHeaderHost, "/", customHeaderHost, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(customHeaderHost,
			func(server string) bool {
				return strings.Contains(server, `server_name "custom-headers"`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", customHeaderHost).
			Expect().
			Status(http.StatusOK).
			Body().Contains(fmt.Sprintf("host=%v", customHeaderHost))
	})

	ginkgo.It("should return status code 503 when custom-headers is configured with an invalid secret", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers",
		}

		ing := framework.NewSingleIngress(customHeaderHost, "/", customHeaderHost, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(customHeaderHost,
			func(server string) bool {
				return strings.Contains(server, `server_name "custom-headers"`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", customHeaderHost).
			Expect().
			Status(http.StatusServiceUnavailable).
			Body().Contains("503 Service Temporarily Unavailable")
	})

	ginkgo.It(`should set "more_set_headers 'My-Custom-Header' '42';" when custom-headers are set`, func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/custom-headers": f.Namespace + "/custom-headers",
		}

		f.CreateConfigMap("custom-headers", map[string]string{
			"My-Custom-Header":        "42",
			"My-Custom-Header-Dollar": "$remote_addr",
		})
		f.UpdateNginxConfigMapData("global-allowed-response-headers", "My-Custom-Header,My-Custom-Header-Dollar")

		ing := framework.NewSingleIngress(customHeaderHost, "/", customHeaderHost, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(customHeaderHost,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "My-Custom-Header: 42";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", customHeaderHost).
			Expect().
			Status(http.StatusOK).
			Header("My-Custom-Header").Contains("42")
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", customHeaderHost).
			Expect().
			Status(http.StatusOK).
			Header("My-Custom-Header-Dollar").Contains("$remote_addr")
	})
})
