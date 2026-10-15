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

var _ = framework.DescribeAnnotation("server-snippet", func() {
	f := framework.NewDefaultFramework("serversnippet")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It(`add valid directives to server via server snippet`, func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := "serversnippet.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-snippet": `
				more_set_headers "Foo: Bar";
				more_set_headers "Xpto: Lalala";`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Foo: Bar`) &&
					strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Foo", []string{"Bar"}).
			ValueEqual("Xpto", []string{"Lalala"})
	})

	ginkgo.It(`drops server snippet if disabled by the administrator`, func() {
		f.UpdateNginxConfigMapData("annotations-risk-level", "Critical") // To enable snippet configurations
		defer f.UpdateNginxConfigMapData("annotations-risk-level", "High")

		host := "noserversnippet.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/server-snippet": `
				more_set_headers "Foo: Bar";
				more_set_headers "Xpto: Lalala";`,
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, `more_set_headers "Foo: Bar`) &&
					!strings.Contains(server, `more_set_headers "Xpto: Lalala";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			NotContainsKey("Foo").
			NotContainsKey("Xpto")
	})
})
