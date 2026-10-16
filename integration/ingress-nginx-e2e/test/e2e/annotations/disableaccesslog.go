/*
Copyright 2020 The Kubernetes Authors.

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

var _ = framework.DescribeAnnotation("disable-access-log disable-http-access-log disable-stream-access-log", func() {
	f := framework.NewDefaultFramework("disableaccesslog")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("disable-access-log set access_log off", func() {
		host := "disableaccesslog.foo.com"

		f.UpdateNginxConfigMapData("disable-access-log", "true")
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `access_log off;`)
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("disable-http-access-log set access_log off", func() {
		host := "disablehttpaccesslog.foo.com"

		f.UpdateNginxConfigMapData("disable-http-access-log", "true")
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `access_log off;`)
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("disable-stream-access-log set access_log off", func() {
		host := "disablehttpaccesslog.foo.com"

		f.UpdateNginxConfigMapData("disable-stream-access-log", "true")
		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(ngx string) bool {
			return strings.Contains(ngx, `access_log off;`)
		})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})
})
