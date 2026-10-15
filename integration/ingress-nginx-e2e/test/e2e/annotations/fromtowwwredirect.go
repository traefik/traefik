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
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("from-to-www-redirect", func() {
	f := framework.NewDefaultFramework("fromtowwwredirect")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should redirect from www HTTP to HTTP", func() {
		ginkgo.By("setting up server for redirect from www")
		host := "fromtowwwredirect.bar.com"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/from-to-www-redirect": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, `server_name "www.fromtowwwredirect.bar.com";`) &&
					strings.Contains(cfg, `return 308 $redirect_to;`)
			})

		ginkgo.By("sending request to www.fromtowwwredirect.bar.com")
		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", fmt.Sprintf("%s.%s", "www", host)).
			Expect().
			Status(http.StatusPermanentRedirect).
			Header("Location").Equal("http://fromtowwwredirect.bar.com:80/foo")
	})

	ginkgo.It("should redirect from www HTTPS to HTTPS", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		ginkgo.By("setting up server for redirect from www")

		fromHost := fmt.Sprintf("%s.nip.io", f.GetNginxIP())
		toHost := fmt.Sprintf("www.%s", fromHost)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/from-to-www-redirect":  "true",
			"nginx.ingress.kubernetes.io/configuration-snippet": "more_set_headers \"ExpectedHost: $http_host\";",
		}

		ing := framework.NewSingleIngressWithTLS(fromHost, "/", fromHost, []string{fromHost, toHost}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		_, err := framework.CreateIngressTLSSecret(f.KubeClientSet,
			ing.Spec.TLS[0].Hosts,
			ing.Spec.TLS[0].SecretName,
			ing.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)
		framework.Sleep()

		f.WaitForNginxServer(toHost,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%v";`, toHost)) &&
					strings.Contains(server, `return 308 $redirect_to;`)
			})

		ginkgo.By("sending request to www should redirect to domain")
		f.HTTPTestClientWithTLSConfig(&tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Ignore the gosec error in testing
			ServerName:         toHost,
		}).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", toHost).
			Expect().
			Status(http.StatusPermanentRedirect).
			Header("Location").Equal(fmt.Sprintf("https://%v:443", fromHost))

		ginkgo.By("sending request to domain should not redirect to www")
		f.HTTPTestClientWithTLSConfig(&tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Ignore the gosec error in testing
			ServerName:         fromHost,
		}).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", fromHost).
			Expect().
			Status(http.StatusOK).
			Header("ExpectedHost").Equal(fromHost)
	})
})
