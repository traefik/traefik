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

var _ = framework.DescribeAnnotation("configuration-snippet", func() {
	f := framework.NewDefaultFramework(
		"configurationsnippet",
		framework.WithHTTPBunEnabled(),
	)

	ginkgo.It("set snippet more_set_headers in all locations", func() {
		host := "configurationsnippet.foo.com"
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Foo1: Bar1";`,
		}

		f.EnsureIngress(framework.NewSingleIngress(
			host,
			"/",
			host,
			f.Namespace,
			framework.HTTPBunService,
			80,
			annotations))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `more_set_headers "Foo1: Bar1";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			ValueEqual("Foo1", []string{"Bar1"})
	})

	ginkgo.It("drops snippet more_set_header in all locations if disabled by admin", func() {
		f.UpdateNginxConfigMapData("annotations-risk-level", "Critical") // To enable snippet configurations
		defer f.UpdateNginxConfigMapData("annotations-risk-level", "High")
		host := "noconfigurationsnippet.foo.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/configuration-snippet": `more_set_headers "Foo1: Bar1";`,
		}

		ing := framework.NewSingleIngress(
			host,
			"/",
			host,
			f.Namespace,
			framework.HTTPBunService,
			80,
			annotations)

		f.UpdateNginxConfigMapData("allow-snippet-annotations", "false")

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, `more_set_headers "Foo1: Bar1";`)
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Headers().
			NotContainsKey("Foo1")
	})
})
