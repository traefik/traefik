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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const proxyRedirectToHost = "goodbye.com"

var _ = framework.DescribeAnnotation("proxy-*", func() {
	f := framework.NewDefaultFramework("proxy")
	host := "proxy.foo.com"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set proxy_redirect to off", func() {
		proxyRedirectFrom := "off"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-from"] = proxyRedirectFrom
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-to"] = proxyRedirectToHost

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_redirect %s;", proxyRedirectFrom))
			})
	})

	ginkgo.It("should set proxy_redirect to default", func() {
		proxyRedirectFrom := "default"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-from"] = proxyRedirectFrom
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-to"] = proxyRedirectToHost

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_redirect %s;", proxyRedirectFrom))
			})
	})

	ginkgo.It("should set proxy_redirect to hello.com goodbye.com", func() {
		proxyRedirectFrom := "hello.com"
		proxyRedirectTo := proxyRedirectToHost

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-from"] = proxyRedirectFrom
		annotations["nginx.ingress.kubernetes.io/proxy-redirect-to"] = proxyRedirectTo

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_redirect %s %s;", proxyRedirectFrom, proxyRedirectTo))
			})
	})

	ginkgo.It("should set proxy client-max-body-size to 8m", func() {
		proxyBodySize := "8m"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = proxyBodySize

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("client_max_body_size %s;", proxyBodySize))
			})
	})

	ginkgo.It("should not set proxy client-max-body-size to incorrect value", func() {
		proxyBodySize := "15r"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-body-size"] = proxyBodySize

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("client_max_body_size %s;", proxyBodySize))
			})
	})

	ginkgo.It("should set valid proxy timeouts", func() {
		proxyConnectTimeout := "50"
		proxySendTimeout := "20"
		proxyReadtimeout := "20"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-connect-timeout"] = proxyConnectTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = proxySendTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = proxyReadtimeout

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_connect_timeout %ss;", proxyConnectTimeout)) &&
					strings.Contains(server, fmt.Sprintf("proxy_send_timeout %ss;", proxySendTimeout)) &&
					strings.Contains(server, fmt.Sprintf("proxy_read_timeout %ss;", proxyReadtimeout))
			})
	})

	ginkgo.It("should not set invalid proxy timeouts", func() {
		proxyConnectTimeout := "50k"
		proxySendTimeout := "20k"
		proxyReadtimeout := "20k"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-connect-timeout"] = proxyConnectTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = proxySendTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = proxyReadtimeout

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, fmt.Sprintf("proxy_connect_timeout %ss;", proxyConnectTimeout)) &&
					!strings.Contains(server, fmt.Sprintf("proxy_send_timeout %ss;", proxySendTimeout)) &&
					!strings.Contains(server, fmt.Sprintf("proxy_read_timeout %ss;", proxyReadtimeout))
			})
	})

	ginkgo.It("should turn on proxy-buffering", func() {
		proxyBuffering := "on"
		proxyBuffersNumber := "8"
		proxyBufferSize := "8k"
		proxyBusyBuffersSize := "16k"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-buffering"] = proxyBuffering
		annotations["nginx.ingress.kubernetes.io/proxy-buffers-number"] = proxyBuffersNumber
		annotations["nginx.ingress.kubernetes.io/proxy-buffer-size"] = proxyBufferSize
		annotations["nginx.ingress.kubernetes.io/proxy-busy-buffers-size"] = proxyBusyBuffersSize

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_buffering %s;", proxyBuffering)) &&
					strings.Contains(server, fmt.Sprintf("proxy_buffer_size %s;", proxyBufferSize)) &&
					strings.Contains(server, fmt.Sprintf("proxy_buffers %s %s;", proxyBuffersNumber, proxyBufferSize)) &&
					strings.Contains(server, fmt.Sprintf("proxy_busy_buffers_size %s;", proxyBusyBuffersSize)) &&
					strings.Contains(server, fmt.Sprintf("proxy_request_buffering %s;", proxyBuffering))
			})
	})

	ginkgo.It("should turn off proxy-request-buffering", func() {
		proxyRequestBuffering := "off"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-request-buffering"] = proxyRequestBuffering

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_request_buffering %s;", proxyRequestBuffering))
			})
	})

	ginkgo.It("should build proxy next upstream", func() {
		proxyNextUpstream := "error timeout http_502"
		proxyNextUpstreamTimeout := "999999"
		proxyNextUpstreamTries := "888888"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-next-upstream"] = proxyNextUpstream
		annotations["nginx.ingress.kubernetes.io/proxy-next-upstream-timeout"] = proxyNextUpstreamTimeout
		annotations["nginx.ingress.kubernetes.io/proxy-next-upstream-tries"] = proxyNextUpstreamTries

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_next_upstream %s;", proxyNextUpstream)) &&
					strings.Contains(server, fmt.Sprintf("proxy_next_upstream_timeout %s;", proxyNextUpstreamTimeout)) &&
					strings.Contains(server, fmt.Sprintf("proxy_next_upstream_tries %s;", proxyNextUpstreamTries))
			})
	})

	ginkgo.It("should setup proxy cookies", func() {
		proxyCookieDomain := "localhost example.org"
		proxyCookiePath := "/one/ /"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-cookie-domain"] = proxyCookieDomain
		annotations["nginx.ingress.kubernetes.io/proxy-cookie-path"] = proxyCookiePath

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_cookie_domain %s;", proxyCookieDomain)) &&
					strings.Contains(server, fmt.Sprintf("proxy_cookie_path %s;", proxyCookiePath))
			})
	})

	ginkgo.It("should change the default proxy HTTP version", func() {
		proxyHTTPVersion := "1.0"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/proxy-http-version"] = proxyHTTPVersion

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf("proxy_http_version %s;", proxyHTTPVersion))
			})
	})
})
