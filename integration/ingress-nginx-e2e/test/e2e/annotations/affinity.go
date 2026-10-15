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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const (
	affinityAnnotation = "cookie"
	cookieName         = "SERVERID"
	enableAnnotation   = "true"
	disableAnnotation  = "false"
	defaultHost        = "foo.com"
)

var _ = framework.DescribeAnnotation("affinity session-cookie-name", func() {
	f := framework.NewDefaultFramework("affinity")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(2))
	})

	ginkgo.It("should set sticky cookie SERVERID", func() {
		host := "sticky.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("SERVERID=")
	})

	ginkgo.It("should change cookie name on ingress definition change", func() {
		host := "change.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains(cookieName)

		ing.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "OTHERCOOKIENAME"

		_, err := f.KubeClientSet.NetworkingV1().Ingresses(f.Namespace).Update(context.TODO(), ing, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "updating ingress")
		framework.Sleep()

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("OTHERCOOKIENAME")
	})

	ginkgo.It("should set the path to /something on the generated cookie", func() {
		host := "path.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName

		ing := framework.NewSingleIngress(host, "/something", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/something").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Path=/something")
	})

	ginkgo.It("does not set the path to / on the generated cookie if there's more than one rule referring to the same backend", func() {
		pathtype := networking.PathTypePrefix
		host := "morethanonerule.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName

		f.EnsureIngress(&networking.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        host,
				Namespace:   f.Namespace,
				Annotations: annotations,
			},
			Spec: networking.IngressSpec{
				IngressClassName: &f.IngressClass,
				Rules: []networking.IngressRule{
					{
						Host: host,
						IngressRuleValue: networking.IngressRuleValue{
							HTTP: &networking.HTTPIngressRuleValue{
								Paths: []networking.HTTPIngressPath{
									{
										Path:     "/something",
										PathType: &pathtype,
										Backend: networking.IngressBackend{
											Service: &networking.IngressServiceBackend{
												Name: framework.EchoService,
												Port: networking.ServiceBackendPort{
													Number: int32(80),
												},
											},
										},
									},
									{
										Path:     "/somewhereelse",
										PathType: &pathtype,
										Backend: networking.IngressBackend{
											Service: &networking.IngressServiceBackend{
												Name: framework.EchoService,
												Port: networking.ServiceBackendPort{
													Number: int32(80),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/something").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Path=/something")

		f.HTTPTestClient().
			GET("/somewhereelse").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Path=/somewhereelse")
	})

	ginkgo.It("should set cookie with expires", func() {
		host := "cookieexpires.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "ExpiresCookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-expires"] = "172800"
		annotations["nginx.ingress.kubernetes.io/session-cookie-max-age"] = "259200"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		local, err := time.LoadLocation("GMT")
		assert.Nil(ginkgo.GinkgoT(), err, "loading GMT location")
		assert.NotNil(ginkgo.GinkgoT(), local, "expected a location but none returned")

		duration, err := time.ParseDuration("48h")
		assert.Nil(ginkgo.GinkgoT(), err, "parsing duration")
		expected := time.Now().In(local).Add(duration).Format("Mon, 02-Jan-06 15:04")

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains(fmt.Sprintf("Expires=%s", expected)).Contains("Max-Age=259200")
	})

	ginkgo.It("should set cookie with domain", func() {
		host := "cookiedomain.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "DomainCookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-domain"] = "foo.bar"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Domain=foo.bar")
	})

	ginkgo.It("should not set cookie without domain annotation", func() {
		host := "cookienodomain.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "NoDomainCookie"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").NotContains("; Domain")
	})

	ginkgo.It("should work with use-regex annotation and session-cookie-path", func() {
		host := "useregex.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/use-regex"] = enableAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-path"] = "/foo/bar"

		ing := framework.NewSingleIngress(host, "/foo/.*", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Path=/foo/bar").Contains("SERVERID=")
	})

	ginkgo.It("should warn user when use-regex is true and session-cookie-path is not set", func() {
		host := "useregexwarn.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/use-regex"] = enableAnnotation

		ing := framework.NewSingleIngress(host, "/foo/.*", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK)

		logs, err := f.NginxLogs()
		assert.Nil(ginkgo.GinkgoT(), err, "obtaining nginx logs")
		assert.Contains(ginkgo.GinkgoT(), logs, `session-cookie-path should be set when use-regex is true`)
	})

	ginkgo.It("should not set affinity across all server locations when using separate ingresses", func() {
		host := "separate.foo.com"

		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation

		ing1 := framework.NewSingleIngress("ingress1", "/foo/bar", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing1)

		ing2 := framework.NewSingleIngress("ingress2", "/foo", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing2)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, `location "/foo/bar/"`) && strings.Contains(server, `location "/foo/"`)
			})

		f.HTTPTestClient().
			GET("/foo").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Empty()

		f.HTTPTestClient().
			GET("/foo/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("Path=/foo/bar")
	})

	ginkgo.It("should set sticky cookie without host", func() {
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName

		ing := framework.NewSingleIngress("default-no-host", "/", "", f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer("_",
			func(server string) bool {
				return strings.Contains(server, `server_name "_"`)
			})

		f.HTTPTestClient().
			GET("/").
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("SERVERID=")
	})

	ginkgo.It("should work with server-alias annotation", func() {
		host := defaultHost
		alias1 := "a1.foo.com"
		alias2 := "a2.foo.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/server-alias"] = fmt.Sprintf("%s,%s", alias1, alias2)

		ing := framework.NewSingleIngress(host, "/bar", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				// server alias sort by sort.Strings(), see: internal/ingress/annotations/alias/main.go:60
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" "%s" "%s" ;`, host, alias1, alias2))
			})

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("SERVERID=")

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", alias1).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("SERVERID=")

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", alias2).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("SERVERID=")
	})

	ginkgo.It("should set secure in cookie with provided true annotation on http", func() {
		host := defaultHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/session-cookie-secure"] = enableAnnotation

		ing := framework.NewSingleIngress(host, "/bar", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("; Secure")
	})

	ginkgo.It("should not set secure in cookie with provided false annotation on http", func() {
		host := defaultHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/session-cookie-secure"] = disableAnnotation

		ing := framework.NewSingleIngress(host, "/bar", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		f.HTTPTestClient().
			GET("/bar").
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").NotContains("; Secure")
	})

	ginkgo.It("should set secure in cookie with provided false annotation on https", func() {
		host := defaultHost
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = affinityAnnotation
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = cookieName
		annotations["nginx.ingress.kubernetes.io/session-cookie-secure"] = disableAnnotation

		f.EnsureIngress(framework.NewSingleIngressWithTLS(host, "/", host, []string{host}, f.Namespace, framework.EchoService, 80, annotations))

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host)) &&
					strings.Contains(server, "listen 443")
			})

		//nolint:gosec // Ignore the gosec error in testing
		f.HTTPTestClientWithTLSConfig(&tls.Config{ServerName: host, InsecureSkipVerify: true}).
			GET("/").
			WithURL(f.GetURL(framework.HTTPS)).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Header("Set-Cookie").Contains("; Secure")
	})
})
