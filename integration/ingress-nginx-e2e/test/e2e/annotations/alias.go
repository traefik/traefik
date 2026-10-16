/*
Copyright 2017 The Kubernetes Authors.

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

const fooHost = "foo"

var _ = framework.DescribeAnnotation("server-alias", func() {
	f := framework.NewDefaultFramework("alias")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should return status code 200 for host 'foo' and 404 for 'bar'", func() {
		host := fooHost

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%v"`, host))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(fmt.Sprintf("host=%v", host))

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", "bar").
			Expect().
			Status(http.StatusNotFound).
			Body().Contains("404 Not Found")
	})

	ginkgo.It("should return status code 200 for host 'foo' and 'bar'", func() {
		host := fooHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-alias": "bar",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%v"`, host))
			})

		hosts := []string{fooHost, "bar"}
		for _, host := range hosts {
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Body().Contains(fmt.Sprintf("host=%v", host))
		}
	})

	ginkgo.It("should return status code 200 for hosts defined in two ingresses, different path with one alias", func() {
		host := fooHost

		ing := framework.NewSingleIngress("app-a", "/app-a", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-alias": "bar",
		}
		ing = framework.NewSingleIngress("app-b", "/app-b", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%v" "bar"`, host))
			})

		hosts := []string{fooHost, "bar"}
		for _, host := range hosts {
			f.HTTPTestClient().
				GET("/app-a").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Body().Contains(fmt.Sprintf("host=%v", host))
		}
	})
})
