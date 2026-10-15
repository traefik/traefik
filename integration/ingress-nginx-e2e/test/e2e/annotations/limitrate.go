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

	"github.com/onsi/ginkgo/v2"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeAnnotation("limit-rate", func() {
	f := framework.NewDefaultFramework("limit-rate-annotation")
	host := "limit-rate-annotation"

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("Check limit-rate annotation", func() {
		annotation := make(map[string]string)
		annotation["nginx.ingress.kubernetes.io/proxy-buffering"] = "on"
		limitRate := 1
		annotation["nginx.ingress.kubernetes.io/limit-rate"] = strconv.Itoa(limitRate)

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, annotation)
		f.EnsureIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, fmt.Sprintf("limit_rate %vk;", limitRate))
		})

		limitRate = 90
		annotation["nginx.ingress.kubernetes.io/limit-rate"] = strconv.Itoa(limitRate)

		ing.SetAnnotations(annotation)
		f.UpdateIngress(ing)

		f.WaitForNginxConfiguration(func(cfg string) bool {
			return strings.Contains(cfg, fmt.Sprintf("limit_rate %vk;", limitRate))
		})
	})
})
