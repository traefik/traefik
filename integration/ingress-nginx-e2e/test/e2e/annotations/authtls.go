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
	"github.com/stretchr/testify/assert"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

const authTLSFooHost = "authtls.foo.com"

var _ = framework.DescribeAnnotation("auth-tls-*", func() {
	f := framework.NewDefaultFramework("authtls")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2))
	})

	ginkgo.It("should set sslClientCertificate, sslVerifyClient and sslVerifyDepth with auth-tls-secret", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{}

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "ssl_client_certificate") &&
					!strings.Contains(server, "ssl_verify_client") &&
					!strings.Contains(server, "ssl_verify_depth")
			})

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret": nameSpace + "/" + host,
		}

		ing.SetAnnotations(annotations)
		f.UpdateIngress(ing)

		assertSslClientCertificateConfig(f, host, "on", "1")

		// Send Request without Client Certs
		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusBadRequest)

		// Send Request Passing the Client Certs
		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should set valid auth-tls-secret, sslVerify to off, and sslVerifyDepth to 2", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		_, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "off",
			"nginx.ingress.kubernetes.io/auth-tls-verify-depth":  "2",
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "off", "2")

		// Send Request without Client Certs
		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should 302 redirect to error page instead of 400 when auth-tls-error-page is set", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		errorPath := "/error"

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":     nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-error-page": f.GetURL(framework.HTTP) + errorPath,
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		sslErrorPage := fmt.Sprintf("error_page 495 496 = %s;", f.GetURL(framework.HTTP)+errorPath)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, sslErrorPage)
			})

		// Send Request without Client Certs
		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusFound).
			Header("Location").Equal(f.GetURL(framework.HTTP) + errorPath)

		// Send Request Passing the Client Certs
		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should pass URL-encoded certificate to upstream", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":                       nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream": "true",
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		sslUpstreamClientCert := "proxy_set_header ssl-client-cert $ssl_client_escaped_cert;"

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, sslUpstreamClientCert)
			})

		// Send Request without Client Certs
		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusBadRequest)

		// Send Request Passing the Client Certs
		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains("ssl-client-cert=-----BEGIN%20CERTIFICATE-----%0A")
	})

	ginkgo.It("should validate auth-tls-verify-client", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "on",
		}

		ing := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusBadRequest)

		annotations = map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "off",
		}

		ing.SetAnnotations(annotations)
		f.UpdateIngress(ing)

		assertSslClientCertificateConfig(f, host, "off", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		f.HTTPTestClient().
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 403 using auth-tls-match-cn with no matching CN from client", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "on",
			"nginx.ingress.kubernetes.io/auth-tls-match-cn":      "CN=notgonnamatch",
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)
	})

	ginkgo.It("should return 200 using auth-tls-match-cn with matching CN from client", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "on",
			"nginx.ingress.kubernetes.io/auth-tls-match-cn":      "CN=authtls",
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should reload the nginx config when auth-tls-match-cn is updated", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		// First add an annotation that forbids our connection
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "on",
			"nginx.ingress.kubernetes.io/auth-tls-match-cn":      "CN=notvalid",
		}

		ingress := f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusForbidden)

		// Update the annotation to something that allows the connection
		ingress.Annotations["nginx.ingress.kubernetes.io/auth-tls-match-cn"] = "CN=authtls"
		f.UpdateIngress(ingress)

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should return 200 using auth-tls-match-cn where atleast one of the regex options matches CN from client", func() {
		host := authTLSFooHost
		nameSpace := f.Namespace

		clientConfig, err := framework.CreateIngressMASecret(
			f.KubeClientSet,
			host,
			host,
			nameSpace)
		assert.Nil(ginkgo.GinkgoT(), err)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/auth-tls-secret":        nameSpace + "/" + host,
			"nginx.ingress.kubernetes.io/auth-tls-verify-client": "on",
			"nginx.ingress.kubernetes.io/auth-tls-match-cn":      "CN=(itwillmatch|withthenextoption|authtls)",
		}

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, nameSpace, framework.EchoService, 80, annotations))

		assertSslClientCertificateConfig(f, host, "on", "1")

		f.HTTPTestClientWithTLSConfig(clientConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})

//nolint:unparam // Ignore the invariant param: host
func assertSslClientCertificateConfig(f *framework.Framework, host, verifyClient, verifyDepth string) {
	sslClientCertDirective := fmt.Sprintf("ssl_client_certificate /etc/ingress-controller/ssl/%s-%s.pem;", f.Namespace, host)
	sslVerify := fmt.Sprintf("ssl_verify_client %s;", verifyClient)
	sslVerifyDepth := fmt.Sprintf("ssl_verify_depth %s;", verifyDepth)

	f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, sslClientCertDirective) &&
				strings.Contains(server, sslVerify) &&
				strings.Contains(server, sslVerifyDepth)
		})
}
