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
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

const sslRedirectValue = "false"

var _ = framework.DescribeAnnotation("affinitymode", func() {
	f := framework.NewDefaultFramework("affinity")

	ginkgo.It("Balanced affinity mode should balance", func() {
		deploymentName := "affinitybalanceecho"
		replicas := 5
		f.NewEchoDeployment(
			framework.WithDeploymentName(deploymentName),
			framework.WithDeploymentReplicas(replicas),
		)

		host := "affinity-mode-balance.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = "cookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "hello-cookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-expires"] = "172800"
		annotations["nginx.ingress.kubernetes.io/session-cookie-max-age"] = "172800"
		annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = sslRedirectValue
		annotations["nginx.ingress.kubernetes.io/affinity-mode"] = "balanced"
		annotations["nginx.ingress.kubernetes.io/session-cookie-hash"] = "sha1"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, deploymentName, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		// Check configuration
		ingress := f.GetIngress(f.Namespace, host)
		returnedAnnotations := ingress.GetAnnotations()
		isItEqual := reflect.DeepEqual(annotations, returnedAnnotations)
		assert.Equal(ginkgo.GinkgoT(), isItEqual, true)
	})

	ginkgo.It("Check persistent affinity mode", func() {
		deploymentName := "affinitypersistentecho"
		replicas := 5
		f.NewEchoDeployment(
			framework.WithDeploymentName(deploymentName),
			framework.WithDeploymentReplicas(replicas),
		)

		host := "affinity-mode-persistent.com"
		annotations := make(map[string]string)
		annotations["nginx.ingress.kubernetes.io/affinity"] = "cookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-name"] = "hello-cookie"
		annotations["nginx.ingress.kubernetes.io/session-cookie-expires"] = "172800"
		annotations["nginx.ingress.kubernetes.io/session-cookie-max-age"] = "172800"
		annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = sslRedirectValue
		annotations["nginx.ingress.kubernetes.io/affinity-mode"] = "persistent"
		annotations["nginx.ingress.kubernetes.io/session-cookie-hash"] = "sha1"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, deploymentName, 80, annotations)
		f.EnsureIngress(ing)

		f.WaitForNginxServer(host,
			func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
			})

		// Check configuration
		ingress := f.GetIngress(f.Namespace, host)
		returnedAnnotations := ingress.GetAnnotations()
		isItEqual := reflect.DeepEqual(annotations, returnedAnnotations)
		assert.Equal(ginkgo.GinkgoT(), isItEqual, true)

		// Make a request
		request := f.HTTPTestClient().GET("/").WithHeader("Host", host)
		response := request.Expect()

		// Get the responder host name
		originalHostName := getHostnameFromResponseBody(response.Body().Raw())

		// Send new requests and add new backends. Check which backend responded to the sent request
		cookies := getCookiesFromHeader(response.Header("Set-Cookie").Raw())
		for sendRequestNumber := 0; sendRequestNumber < 10; sendRequestNumber++ {
			replicas++
			err := framework.UpdateDeployment(f.KubeClientSet, f.Namespace, deploymentName, replicas, nil)
			assert.Nil(ginkgo.GinkgoT(), err)
			framework.Sleep()

			response = request.WithCookies(cookies).Expect()
			newHostName := getHostnameFromResponseBody(response.Body().Raw())
			assert.Equal(ginkgo.GinkgoT(), originalHostName, newHostName,
				fmt.Sprintf("Response number %v is not from the same host. Original host: %s, response returned: %s", sendRequestNumber, originalHostName, newHostName))

		}

		// remove all backends
		replicas = 0
		err := framework.UpdateDeployment(f.KubeClientSet, f.Namespace, deploymentName, replicas, nil)
		assert.Nil(ginkgo.GinkgoT(), err)
		framework.Sleep()

		// validate, there is no backend to serve the request
		request.WithCookies(cookies).Expect().Status(http.StatusServiceUnavailable)

		// create brand new backends
		replicas = 2
		err = framework.UpdateDeployment(f.KubeClientSet, f.Namespace, deploymentName, replicas, nil)
		assert.Nil(ginkgo.GinkgoT(), err)
		framework.Sleep()

		// wait brand new backends to spawn
		response = request.WithCookies(cookies).Expect()
		try := 0
		for (response.Raw().StatusCode == http.StatusServiceUnavailable) && (try < 30) {
			framework.Sleep()
			response = request.WithCookies(cookies).Expect()
			try++
		}
		assert.LessOrEqual(ginkgo.GinkgoT(), try, 29, "Tries reached it's maximum, backends did not deployed in time")

		// brand new backends equals new hostname
		newHostName := getHostnameFromResponseBody(response.Body().Raw())
		assert.NotEqual(ginkgo.GinkgoT(), originalHostName, newHostName,
			fmt.Sprintf("Response is from the same host (That should not be possible). Original host: %s, response returned: %s", originalHostName, newHostName))
	})
})

func getHostnameFromResponseBody(rawResponseBody string) string {
	lines := strings.Split(strings.TrimSpace(rawResponseBody), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Hostname") {
			hostnameParts := strings.Split(strings.TrimSpace(line), ":")
			if len(hostnameParts) == 2 {
				return strings.TrimSpace(hostnameParts[1])
			}
			return ""
		}
	}
	return ""
}

func getCookiesFromHeader(rawheader string) map[string]string {
	cookies := make(map[string]string)
	parts := strings.Split(strings.TrimSpace(rawheader), ";")
	for _, part := range parts {
		subparts := strings.Split(strings.TrimSpace(part), "=")
		if len(subparts) == 2 {
			cookies[subparts[0]] = subparts[1]
		} else {
			cookies[subparts[0]] = ""
		}
	}
	return cookies
}
