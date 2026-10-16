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
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/ingress-nginx/test/e2e/framework"
)

func startIngress(f *framework.Framework, annotations map[string]string) map[string]bool {
	host := "upstream-hash-by.foo.com"

	ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotations)
	f.EnsureIngress(ing)
	f.WaitForNginxServer(host,
		func(server string) bool {
			return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
		})

	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	err := wait.Poll(framework.Poll, framework.DefaultTimeout, func() (bool, error) {
		resp := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().Raw()
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true, nil
		}

		return false, nil
	})

	assert.Nil(ginkgo.GinkgoT(), err)

	re, err := regexp.Compile(fmt.Sprintf(`Hostname: %v.*`, framework.EchoService))
	assert.Nil(ginkgo.GinkgoT(), err, "error compiling regex")

	podMap := map[string]bool{}

	for i := 0; i < 100; i++ {
		data := f.HTTPTestClient().
			GET("/").
			WithHeader("Host", host).
			Expect().
			Body().Raw()

		podName := re.FindString(data)
		assert.NotEmpty(ginkgo.GinkgoT(), podName, "expected a pod name")
		podMap[podName] = true
	}

	return podMap
}

var _ = framework.DescribeAnnotation("upstream-hash-by-*", func() {
	f := framework.NewDefaultFramework("upstream-hash-by")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment(framework.WithDeploymentReplicas(6))
	})

	ginkgo.It("should connect to the same pod", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-hash-by": "$request_uri",
		}

		podMap := startIngress(f, annotations)
		assert.Equal(ginkgo.GinkgoT(), len(podMap), 1)
	})

	ginkgo.It("should connect to the same subset of pods", func() {
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/upstream-hash-by":             "$request_uri",
			"nginx.ingress.kubernetes.io/upstream-hash-by-subset":      "true",
			"nginx.ingress.kubernetes.io/upstream-hash-by-subset-size": "3",
		}

		podMap := startIngress(f, annotations)
		assert.Equal(ginkgo.GinkgoT(), len(podMap), 3)
	})
})
