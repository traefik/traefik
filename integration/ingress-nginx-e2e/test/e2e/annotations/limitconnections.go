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
	"strconv"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("Annotation - limit-connections", func() {
	f := framework.NewDefaultFramework("limit-connections")

	ginkgo.BeforeEach(func() {
		f.NewSlowEchoDeployment()
	})

	ginkgo.It("should limit-connections", func() {
		host := "limit-connections"

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.SlowEchoService, 80, nil)
		f.EnsureIngress(ing)
		f.WaitForNginxServer(host, func(server string) bool {
			return strings.Contains(server, fmt.Sprintf(`server_name "%s" ;`, host))
		})

		// limit connections
		connectionLimit := 8
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/limit-connections": strconv.Itoa(connectionLimit),
		}

		ing.SetAnnotations(annotations)
		f.UpdateIngress(ing)
		f.WaitForNginxConfiguration(func(server string) bool {
			return strings.Contains(server, "limit_conn") && strings.Contains(server, fmt.Sprintf("conn %v;", connectionLimit))
		})

		requests := 20
		results := make(chan int, requests)

		worker := func(wg *sync.WaitGroup) {
			defer wg.Done()
			defer ginkgo.GinkgoRecover()

			resp := f.HTTPTestClient().
				GET("/sleep/10").
				WithHeader("Host", host).
				Expect().
				Raw()

			if resp != nil {
				results <- resp.StatusCode
			}
		}

		var wg sync.WaitGroup
		for currentRequest := 0; currentRequest < requests; currentRequest++ {
			wg.Add(1)
			go worker(&wg)
		}

		wg.Wait()

		ok := 0
		failed := 0
		errors := 0

		close(results)
		for status := range results {
			switch status {
			case 200:
				ok++
			case 503:
				failed++
			default:
				errors++
			}
		}

		assert.Equal(ginkgo.GinkgoT(), connectionLimit, ok, "expecting the ok (200) requests to be equal to the connection limit")
		assert.Equal(ginkgo.GinkgoT(), requests-connectionLimit, failed, "expecting the failed (503) requests to be the total requests - connection limit")
		assert.Equal(ginkgo.GinkgoT(), 0, errors, "expecting failed (other than 503) requests to ber zero")
	})
})
