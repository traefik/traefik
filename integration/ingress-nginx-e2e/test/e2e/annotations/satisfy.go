/*
Copyright 2019 The Kubernetes Authors.

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
	"net/url"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("satisfy", func() {
	f := framework.NewDefaultFramework("satisfy", framework.WithHTTPBunEnabled())

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should configure satisfy directive correctly", func() {
		host := "satisfy"
		annotationKey := "nginx.ingress.kubernetes.io/satisfy"

		annotations := map[string]string{
			"any": "any",
			"all": "all",
		}

		results := map[string]string{
			"any": "satisfy any",
			"all": "satisfy all",
		}

		initAnnotations := map[string]string{
			annotationKey: "all",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, initAnnotations)
		f.EnsureIngress(ing)

		for key, result := range results {
			err := framework.UpdateIngress(f.KubeClientSet, f.Namespace, host, func(ingress *networking.Ingress) error {
				ingress.ObjectMeta.Annotations[annotationKey] = annotations[key]
				return nil
			})
			assert.Nil(ginkgo.GinkgoT(), err)

			f.WaitForNginxServer(host,
				func(server string) bool {
					return strings.Contains(server, result)
				})

			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK).
				Body().Contains(fmt.Sprintf("host=%v", host))
		}
	})

	ginkgo.It("should allow multiple auth with satisfy any", func() {
		host := "auth"

		// create basic auth secret at ingress
		s := f.EnsureSecret(buildSecret("uname", "pwd", "basic-secret", f.Namespace))

		annotations := map[string]string{
			// annotations for basic auth at ingress
			"nginx.ingress.kubernetes.io/auth-type":   "basic",
			"nginx.ingress.kubernetes.io/auth-secret": s.Name,
			"nginx.ingress.kubernetes.io/auth-realm":  "test basic auth",

			// annotations for external auth
			"nginx.ingress.kubernetes.io/auth-url":    fmt.Sprintf("http://%s/basic-auth/user/password", f.HTTPBunIP),
			"nginx.ingress.kubernetes.io/auth-signin": "http://$host/auth/start",

			// set satisfy any
			"nginx.ingress.kubernetes.io/satisfy": "any",
		}

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, `server_name "auth"`)
		})

		// with basic auth cred
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithBasicAuth("uname", "pwd").
			Expect().
			Status(http.StatusOK)

		// reroute to signin if without basic cred
		f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			WithQuery("a", "b").
			WithQuery("c", "d").
			Expect().
			Status(http.StatusFound).
			Header("Location").Equal(fmt.Sprintf("http://%s/auth/start?rd=http://%s%s", host, host, url.QueryEscape("/?a=b&c=d")))
	})
})
