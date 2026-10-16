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
	"strings"

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const backendProtocolHost = "backendprotocol.foo.com"

var _ = framework.DescribeAnnotation("backend-protocol", func() {
	f := framework.NewDefaultFramework("backendprotocol")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should set backend protocol to https:// and use proxy_pass", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "HTTPS",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass https://upstream_balancer;")
			})
	})

	ginkgo.It("should set backend protocol to https:// and use proxy_pass with lowercase annotation", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "https",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass https://upstream_balancer;")
			})
	})

	ginkgo.It("should set backend protocol to $scheme:// and use proxy_pass", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "AUTO_HTTP",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "proxy_pass $scheme://upstream_balancer;")
			})
	})

	ginkgo.It("should set backend protocol to grpc:// and use grpc_pass", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPC",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpc://upstream_balancer;")
			})
	})

	ginkgo.It("should set backend protocol to grpcs:// and use grpc_pass", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "GRPCS",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "grpc_pass grpcs://upstream_balancer;")
			})
	})

	ginkgo.It("should set backend protocol to '' and use fastcgi_pass", func() {
		host := backendProtocolHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "fastcgi_pass upstream_balancer;")
			})
	})
})
