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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("ssl-ciphers", func() {
	f := framework.NewDefaultFramework("sslciphers")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should change ssl ciphers", func() {
		host := "ciphers.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/ssl-ciphers":               "ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP",
			"nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers": "false",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "ssl_ciphers ALL:!aNULL:!EXPORT56:RC4+RSA:+HIGH:+MEDIUM:+LOW:+SSLv2:+EXP;") &&
					strings.Contains(server, "ssl_prefer_server_ciphers off;")
			})
		f.HTTPTestClient().
			GET("/something").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should keep ssl ciphers", func() {
		host := "ciphers.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/ssl-ciphers":               "ALL:!aNULL:!EXPORT56:RC4+RSA@STRENGTH:+HIGH@SECLEVEL=0:+MEDIUM:+LOW:+SSLv2:+EXP",
			"nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers": "true",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "ssl_ciphers ALL:!aNULL:!EXPORT56:RC4+RSA@STRENGTH:+HIGH@SECLEVEL=0:+MEDIUM:+LOW:+SSLv2:+EXP;") &&
					strings.Contains(server, "ssl_prefer_server_ciphers on;")
			})
		f.HTTPTestClient().
			GET("/something").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
