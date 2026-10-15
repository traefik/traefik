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

const (
	originHost = "http://origin.com:8080"
	corsHost   = "cors.foo.com"
)

var _ = framework.DescribeAnnotation("cors-*", func() {
	f := framework.NewDefaultFramework("cors")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2))
	})

	ginkgo.It("should enable cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "more_set_headers 'Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS';") &&
					strings.Contains(server, "more_set_headers 'Access-Control-Allow-Origin: $http_origin';") &&
					strings.Contains(server, "more_set_headers 'Access-Control-Allow-Headers: DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization';") &&
					strings.Contains(server, "more_set_headers 'Access-Control-Max-Age: 1728000';") &&
					strings.Contains(server, "more_set_headers 'Access-Control-Allow-Credentials: true';") &&
					strings.Contains(server, "set $http_origin *;") &&
					strings.Contains(server, "$cors 'true';")
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should set cors methods to only allow POST, GET", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-methods": "POST, GET",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "more_set_headers 'Access-Control-Allow-Methods: POST, GET';")
			})
	})

	ginkgo.It("should set cors max-age", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":  "true",
			"nginx.ingress.kubernetes.io/cors-max-age": "200",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "more_set_headers 'Access-Control-Max-Age: 200';")
			})
	})

	ginkgo.It("should disable cors allow credentials", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":            "true",
			"nginx.ingress.kubernetes.io/cors-allow-credentials": "false",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return !strings.Contains(server, "more_set_headers 'Access-Control-Allow-Credentials: true';")
			})
	})

	ginkgo.It("should allow origin for cors", func() {
		host := corsHost
		origin := "https://origin.cors.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin.cors.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})
	})

	ginkgo.It("should allow headers for cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":        "true",
			"nginx.ingress.kubernetes.io/cors-allow-headers": "DNT, User-Agent",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "more_set_headers 'Access-Control-Allow-Headers: DNT, User-Agent';")
			})
	})

	ginkgo.It("should expose headers for cors", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":         "true",
			"nginx.ingress.kubernetes.io/cors-expose-headers": "X-CustomResponseHeader, X-CustomSecondHeader",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "more_set_headers 'Access-Control-Expose-Headers: X-CustomResponseHeader, X-CustomSecondHeader';")
			})
	})

	ginkgo.It("should allow - single origin for multiple cors values", func() {
		host := corsHost
		origin := "https://origin.cors.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin.cors.com:8080, https://origin2.cors.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})
	})

	ginkgo.It("should not allow - single origin for multiple cors values", func() {
		host := corsHost
		origin := "http://no.origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin2.cors.com, https://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow correct origins - single origin for multiple cors values", func() {
		host := corsHost
		badOrigin := "origin.cors.com:8080"
		origin1 := "https://origin2.cors.com"
		origin2 := "https://origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "origin.cors.com:8080, https://origin2.cors.com, https://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", badOrigin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin1).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin1).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin1})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2})
	})

	ginkgo.It("should not break functionality", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "*",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})
	})

	ginkgo.It("should not break functionality - without `*`", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors": "true",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})
	})

	ginkgo.It("should not break functionality with extra domain", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "*, foo.bar.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})
	})

	ginkgo.It("should not match", func() {
		host := corsHost
		origin := "https://fooxbar.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://foo.bar.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow - single origin with required port", func() {
		host := corsHost
		origin := originHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})
	})

	ginkgo.It("should not allow - single origin with port and origin without port", func() {
		host := corsHost
		origin := originHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin2.cors.com, http://origin.com",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should not allow - single origin without port and origin with required port", func() {
		host := corsHost
		origin := "http://origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow - matching origin with wildcard origin (2 subdomains)", func() {
		host := corsHost
		origin := "http://foo.origin.cors.com"
		origin2 := "http://bar-foo.origin.cors.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://*.origin.cors.com, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2})
	})

	ginkgo.It("should not allow - unmatching origin with wildcard origin (2 subdomains)", func() {
		host := corsHost
		origin := "http://bar.foo.origin.cors.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://*.origin.cors.com, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow - matching origin+port with wildcard origin", func() {
		host := corsHost
		origin := "http://abc.origin.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})
	})

	ginkgo.It("should not allow - portless origin with wildcard origin", func() {
		host := corsHost
		origin := "http://abc.origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow correct origins - missing subdomain + origin with wildcard origin and correct origin", func() {
		host := corsHost
		badOrigin := originHost
		origin := "http://bar.origin.com:8080"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "http://origin.cors.com:8080, http://*.origin.com:8080",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", badOrigin).
			Expect().
			Headers().NotContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin})
	})

	ginkgo.It("should allow - missing origins (should allow all origins)", func() {
		host := corsHost
		origin := "http://origin.com"
		origin2 := "http://book.origin.com"
		origin3 := "test.origin.com"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "      ",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		// the client should still receive a response but browsers should block the request
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin3).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin3).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"*"})
	})

	ginkgo.It("should allow correct origin but not others - cors allow origin annotations contain trailing comma", func() {
		host := corsHost
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "https://origin-123.cors.com:8080,   ,https://origin-321.cors.com:8080,",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		origin1 := "https://origin-123.cors.com:8080"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin1).
			Expect().
			Headers().ContainsKey("Access-Control-Allow-Origin")

		origin2 := "https://origin-321.cors.com:8080"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{origin2})

		origin3 := "https://unknown.cors.com:8080"
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin3).
			Expect().
			Headers().
			NotContainsKey("Access-Control-Allow-Origin")
	})

	ginkgo.It("should allow - origins with non-http[s] protocols", func() {
		host := corsHost
		origin := "test://localhost"
		origin2 := "tauri://localhost:3000"
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/enable-cors":       "true",
			"nginx.ingress.kubernetes.io/cors-allow-origin": "test://localhost, tauri://localhost:3000",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"test://localhost"})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithHeader("Origin", origin2).
			Expect().
			Status(http.StatusOK).Headers().
			ValueEqual("Access-Control-Allow-Origin", []string{"tauri://localhost:3000"})
	})
})
