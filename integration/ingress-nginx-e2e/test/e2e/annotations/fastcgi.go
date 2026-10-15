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
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("backend-protocol - FastCGI", func() {
	f := framework.NewDefaultFramework("fastcgi")

	ginkgo.BeforeEach(func() {
		f.NewFastCGIHelloServerDeployment()
	})

	ginkgo.It("should use fastcgi_pass in the configuration file", func() {
		host := "fastcgi"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "include /etc/nginx/fastcgi_params;") &&
					strings.Contains(server, "fastcgi_pass")
			})
	})

	ginkgo.It("should add fastcgi_index in the configuration file", func() {
		host := "fastcgi-index"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
			"nginx.ingress.kubernetes.io/fastcgi-index":    "index.php",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "fastcgi_index \"index.php\";")
			})
	})

	ginkgo.It("should add fastcgi_param in the configuration file", func() {
		configuration := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fastcgi-configmap",
				Namespace: f.Namespace,
			},
			Data: map[string]string{
				"SCRIPT_FILENAME": "$fastcgi_script_name",
				"REDIRECT_STATUS": "200",
			},
		}

		f.EnsureConfigMap(configuration)

		host := "fastcgi-params-configmap" //#nosec G101

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol":         "FCGI",
			"nginx.ingress.kubernetes.io/fastcgi-params-configmap": "fastcgi-configmap",
		}

		ing := framework.NewSingleIngress(host, "/hello", host, f.Namespace, "fastcgi-helloserver", 9000, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "fastcgi_param SCRIPT_FILENAME \"$fastcgi_script_name\";") &&
					strings.Contains(server, "fastcgi_param REDIRECT_STATUS \"200\";")
			})
	})

	ginkgo.It("should return OK for service with backend protocol FastCGI", func() {
		host := "fastcgi-helloserver"
		path := "/hello"

		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/backend-protocol": "FCGI",
		}

		ing := framework.NewSingleIngress(host, path, host, f.Namespace, "fastcgi-helloserver", 9000, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, "fastcgi_pass")
			})

		f.HTTPTestClient().
			GET(path).
			WithHeader("Host", host).
			Expect().
			Status(http.StatusOK).
			Body().Contains("Hello world!")
	})
})
