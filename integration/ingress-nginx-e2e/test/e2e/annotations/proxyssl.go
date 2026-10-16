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

const proxySSLHost = "proxyssl.foo.com"

var _ = framework.DescribeAnnotation("proxy-ssl-*", func() {
	f := framework.NewDefaultFramework("proxyssl")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set valid proxy-ssl-secret", func() {
		host := proxySSLHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-secret"] = f.Namespace + "/" + host

		tlsConfig, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "", "DEFAULT", "TLSv1.2", "off", 1, "")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusPermanentRedirect)

		f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should set valid proxy-ssl-secret, proxy-ssl-verify to on, proxy-ssl-verify-depth to 2, and proxy-ssl-server-name to on", func() {
		host := proxySSLHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-secret"] = f.Namespace + "/" + host
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-verify"] = "on"
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-verify-depth"] = "2"
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-server-name"] = "on"

		tlsConfig, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "", "DEFAULT", "TLSv1.2", "on", 2, "on")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusPermanentRedirect)

		f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should set valid proxy-ssl-secret, proxy-ssl-ciphers to HIGH:!AES", func() {
		host := proxySSLHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-secret"] = f.Namespace + "/" + host
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-ciphers"] = "HIGH:!AES"

		tlsConfig, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "", "HIGH:!AES", "TLSv1.2", "off", 1, "")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusPermanentRedirect)

		f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
	//nolint:dupl // Ignore dupl errors for similar test case
	ginkgo.It("should set valid proxy-ssl-secret, proxy-ssl-protocols", func() {
		host := proxySSLHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-secret"] = f.Namespace + "/" + host
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-protocols"] = "TLSv1.2 TLSv1.3"

		tlsConfig, err := framework.CreateIngressMASecret(f.KubeClientSet, host, host, f.Namespace)
		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		assertProxySSL(f, host, "", "DEFAULT", "TLSv1.2 TLSv1.3", "off", 1, "")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusPermanentRedirect)

		f.HTTPTestClientWithTLSConfig(tlsConfig).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("proxy-ssl-location-only flag should change the nginx config server part", func() {
		host := "proxyssl.com"

		f.NewEchoDeployment(framework.WithDeploymentName("echodeployment"))

		secretName := "secretone"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-secret"] = f.Namespace + "/" + secretName
		annotations["nginx.ingress.kubernetes.io/backend-protocol"] = "HTTPS"
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-verify"] = "on"
		annotations["nginx.ingress.kubernetes.io/proxy-ssl-server-name"] = "on"
		tlsConfig, err := framework.CreateIngressMASecret(f.KubeClientSet, host, secretName, f.Namespace)

		assert.Nil(ginkgo.GinkgoT(), err)

		ing := framework.NewSingleIngressWithTLS(host, "/bar", host, []string{tlsConfig.ServerName}, f.Namespace, "echodeployment", 80, annotations)
		f.EnsureIngress(ing)

		wlKey := "proxy-ssl-location-only"
		wlValue := "true"
		f.UpdateNginxConfigMapData(wlKey, wlValue)

		assertProxySSL(f, host, secretName, "DEFAULT", "TLSv1.2", "on", 1, "on")

		f.WaitForNginxCustomConfiguration("## start server proxyssl.com", "location ", func(server string) bool {
			return (!strings.Contains(server, "proxy_ssl_trusted_certificate") &&
				!strings.Contains(server, "proxy_ssl_ciphers") &&
				!strings.Contains(server, "proxy_ssl_protocols") &&
				!strings.Contains(server, "proxy_ssl_verify") &&
				!strings.Contains(server, "proxy_ssl_verify_depth") &&
				!strings.Contains(server, "proxy_ssl_certificate") &&
				!strings.Contains(server, "proxy_ssl_certificate_key"))
		})

		wlKey = "proxy-ssl-location-only"
		wlValue = "false"
		f.UpdateNginxConfigMapData(wlKey, wlValue)

		f.WaitForNginxCustomConfiguration("## start server proxyssl.com", "location ", func(server string) bool {
			return (strings.Contains(server, "proxy_ssl_trusted_certificate") &&
				strings.Contains(server, "proxy_ssl_ciphers") &&
				strings.Contains(server, "proxy_ssl_protocols") &&
				strings.Contains(server, "proxy_ssl_verify") &&
				strings.Contains(server, "proxy_ssl_verify_depth") &&
				strings.Contains(server, "proxy_ssl_certificate") &&
				strings.Contains(server, "proxy_ssl_certificate_key"))
		})
	})
})

func assertProxySSL(f *framework.Framework, host, sslName, ciphers, protocols, verify string, depth int, proxySSLServerName string) {
	certFile := fmt.Sprintf("/etc/ingress-controller/ssl/%s-%s.pem", f.Namespace, host)

	if sslName != "" {
		certFile = fmt.Sprintf("/etc/ingress-controller/ssl/%s-%s.pem", f.Namespace, sslName)
	}

	f.WaitForNginxServer(host,
		func(server string) bool {
			c := strings.Contains(server, fmt.Sprintf("proxy_ssl_certificate %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_certificate_key %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_trusted_certificate %s;", certFile)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_ciphers %s;", ciphers)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_protocols %s;", protocols)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_verify %s;", verify)) &&
				strings.Contains(server, fmt.Sprintf("proxy_ssl_verify_depth %d;", depth))

			if proxySSLServerName == "" {
				return c
			}

			return c && strings.Contains(server, fmt.Sprintf(`proxy_ssl_server_name %s;`, proxySSLServerName))
		})
}
