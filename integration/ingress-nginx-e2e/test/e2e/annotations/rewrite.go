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
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const rewriteHost = "rewrite.bar.com"

var _ = framework.DescribeAnnotation("rewrite-target use-regex enable-rewrite-log", func() {
	f := framework.NewDefaultFramework("rewrite")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should write rewrite logs", func() {
		ginkgo.By("setting enable-rewrite-log annotation")

		host := rewriteHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target":     "/",
			"nginx.ingress.kubernetes.io/enable-rewrite-log": "true",
		}

		ing := framework.NewSingleIngress(host, "/something", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "rewrite_log on;")
			})

		f.HTTPTestClient().
			GET("/something").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Contains(ginkgo.GinkgoT(), logs, `"(?i)/something" matches "/something", client:`)
		assert.Contains(ginkgo.GinkgoT(), logs, `rewritten data: "/", args: "",`)
	})

	ginkgo.It("should use correct longest path match", func() {
		host := rewriteHost

		ginkgo.By("creating a regular ingress definition")
		ing := framework.NewSingleIngress("kube-lego", "/.well-known/acme/challenge", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "/.well-known/acme/challenge")
			})

		ginkgo.By("making a request to the non-rewritten location")
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:80/.well-known/acme/challenge", host)

		f.HTTPTestClient().
			GET("/.well-known/acme/challenge").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)

		ginkgo.By(`creating an ingress definition with the rewrite-target annotation set on the "/" location`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		rewriteIng := framework.NewSingleIngress("rewrite-index", "/", host, f.Namespace, framework.EchoService, 80, annotations)

		f.EnsureIngress(rewriteIng)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/" {`) &&
					strings.Contains(server, `location ~* "^/.well-known/acme/challenge" {`)
			})

		ginkgo.By("making a second request to the non-rewritten location")
		f.HTTPTestClient().
			GET("/.well-known/acme/challenge").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)
	})

	ginkgo.It("should use ~* location modifier if regex annotation is present", func() {
		host := rewriteHost

		ginkgo.By("creating a regular ingress definition")
		ing := framework.NewSingleIngress(fooHost, "/foo", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location "/foo/" {`)
			})

		ginkgo.By(`creating an ingress definition with the use-regex amd rewrite-target annotation`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		ing = framework.NewSingleIngress("regex", "/foo.+", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/foo" {`) &&
					strings.Contains(server, `location ~* "^/foo.+" {`)
			})

		ginkgo.By("ensuring '/foo' matches '~* ^/foo'")
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:80/foo", host)

		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)

		ginkgo.By("ensuring '/foo/bar' matches '~* ^/foo.+'")
		expectBodyRequestURI = fmt.Sprintf("request_uri=http://%v:80/new/backend", host)

		f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)
	})

	ginkgo.It("should fail to use longest match for documented warning", func() {
		host := rewriteHost

		ginkgo.By("creating a regular ingress definition")
		ing := framework.NewSingleIngress(fooHost, "/foo/bar/bar", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		ginkgo.By(`creating an ingress definition with the use-regex annotation`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend",
		}
		ing = framework.NewSingleIngress("regex", "/foo/bar/[a-z]{3}", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/foo/bar/bar" {`) &&
					strings.Contains(server, `location ~* "^/foo/bar/[a-z]{3}" {`)
			})

		ginkgo.By("check that '/foo/bar/bar' does not match the longest exact path")
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:80/new/backend", host)

		f.HTTPTestClient().
			GET("/foo/bar/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)
	})

	ginkgo.It("should allow for custom rewrite parameters", func() {
		host := rewriteHost

		ginkgo.By(`creating an ingress definition with the use-regex annotation`)
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/new/backend/$1",
		}
		ing := framework.NewSingleIngress("regex", "/foo/bar/(.+)", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location ~* "^/foo/bar/(.+)" {`)
			})

		ginkgo.By("check that '/foo/bar/bar' redirects to custom rewrite")
		expectBodyRequestURI := fmt.Sprintf("request_uri=http://%v:80/new/backend/bar", host)

		f.HTTPTestClient().
			GET("/foo/bar/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains(expectBodyRequestURI)
	})
})
