/*
Copyright 2021 The Kubernetes Authors.

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

const dbgCmd = "/dbg backends all"

var _ = framework.DescribeAnnotation("service-upstream", func() {
	f := framework.NewDefaultFramework("serviceupstream")
	host := "serviceupstream"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.Context("when using the default value (false) and enabling in the annotations", func() {
		ginkgo.It("should use the Service Cluster IP and Port ", func() {
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/service-upstream": "true",
			}

			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s"`, host))
			})

			ginkgo.By("checking if the service is reached")
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

			ginkgo.By("checking if the Service Cluster IP and Port are used")
			s := f.GetService(f.Namespace, framework.EchoService)
			output, err := f.ExecIngressPod(dbgCmd)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.Contains(ginkgo.GinkgoT(), output, fmt.Sprintf(`"address": %q`, s.Spec.ClusterIP))
		})
	})

	ginkgo.Context("when enabling in the configmap", func() {
		ginkgo.It("should use the Service Cluster IP and Port ", func() {
			annotations := map[string]string{}

			f.UpdateNginxConfigMapData("service-upstream", "true")

			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s"`, host))
			})

			ginkgo.By("checking if the service is reached")
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

			ginkgo.By("checking if the Service Cluster IP and Port are used")
			s := f.GetService(f.Namespace, framework.EchoService)
			output, err := f.ExecIngressPod(dbgCmd)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.Contains(ginkgo.GinkgoT(), output, fmt.Sprintf(`"address": %q`, s.Spec.ClusterIP))
		})
	})

	ginkgo.Context("when enabling in the configmap and disabling in the annotations", func() {
		ginkgo.It("should not use the Service Cluster IP and Port", func() {
			annotations := map[string]string{
				"nginx.ingress.kubernetes.io/service-upstream": "false",
			}

			f.UpdateNginxConfigMapData("service-upstream", "true")

			ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
			f.EnsureIngress(ing)

			f.WaitForNginxServer(host, func(server string) bool {
				return strings.Contains(server, fmt.Sprintf(`server_name "%s"`, host))
			})

			ginkgo.By("checking if the service is reached")
			f.HTTPTestClient().
				GET("/").
				WithHeader("Host", host).
				Expect().
				Status(http.StatusOK)

			ginkgo.By("checking if the Service Cluster IP and Port are not used")
			s := f.GetService(f.Namespace, framework.EchoService)
			output, err := f.ExecIngressPod(dbgCmd)
			assert.Nil(ginkgo.GinkgoT(), err)
			assert.NotContains(ginkgo.GinkgoT(), output, fmt.Sprintf(`"address": %q`, s.Spec.ClusterIP))
		})
	})
})
