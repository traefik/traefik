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
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/ingress-nginx/test/e2e/framework"
)

var _ = framework.DescribeSetting("stream-snippet", func() {
	f := framework.NewDefaultFramework("stream-snippet")

	ginkgo.BeforeEach(func() {
		f.NewEchoDeployment()
	})

	ginkgo.It("should add value of stream-snippet to nginx config", func() {
		disableSnippet := f.AllowSnippetConfiguration()
		defer disableSnippet()

		host := "foo.com"

		snippet := `server {listen 8000; proxy_pass 127.0.0.1:80;}`

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, map[string]string{
			"nginx.ingress.kubernetes.io/stream-snippet": snippet,
		})
		f.EnsureIngress(ing)

		svc, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Get(context.TODO(), "nginx-ingress-controller", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining ingress-nginx service")
		assert.NotNil(ginkgo.GinkgoT(), svc, "expected a service but none returned")

		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       framework.EchoService,
			Port:       8000,
			TargetPort: intstr.FromInt(8000),
		})

		_, err = f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, snippet)
			})

		f.HTTPTestClient().
			GET("/healthz").
			WithURL(fmt.Sprintf("http://%v:8000/healthz", f.GetNginxIP())).
			Expect().
			Status(http.StatusOK)
	})

	ginkgo.It("should add stream-snippet and drop annotations per admin config", func() {
		host := "cm.foo.com"
		hostAnnot := "annot.foo.com"

		cmSnippet := `server {listen 8000; proxy_pass 127.0.0.1:80;}`
		annotSnippet := `server {listen 8001; proxy_pass 127.0.0.1:80;}`

		f.SetNginxConfigMapData(map[string]string{
			"allow-snippet-annotations": "false",
			"stream-snippet":            cmSnippet,
		})

		ing := framework.NewSingleIngress(host, "/", host, f.Namespace, framework.EchoService, 80, nil)
		f.EnsureIngress(ing)

		ing1 := framework.NewSingleIngress(hostAnnot, "/", hostAnnot, f.Namespace, framework.EchoService, 80, map[string]string{
			"nginx.ingress.kubernetes.io/stream-snippet": annotSnippet,
		})
		f.EnsureIngress(ing1)

		svc, err := f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Get(context.TODO(), "nginx-ingress-controller", metav1.GetOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error obtaining ingress-nginx service")
		assert.NotNil(ginkgo.GinkgoT(), svc, "expected a service but none returned")

		svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
			Name:       framework.EchoService,
			Port:       8000,
			TargetPort: intstr.FromInt(8000),
		})

		_, err = f.KubeClientSet.
			CoreV1().
			Services(f.Namespace).
			Update(context.TODO(), svc, metav1.UpdateOptions{})
		assert.Nil(ginkgo.GinkgoT(), err, "unexpected error updating service")

		// Sleep a while just to guarantee that the configmap is applied
		framework.Sleep()

		f.WaitForNginxConfiguration(
			func(cfg string) bool {
				return strings.Contains(cfg, cmSnippet) && !strings.Contains(cfg, annotSnippet)
			})

		f.HTTPTestClient().
			GET("/healthz").
			WithURL(fmt.Sprintf("http://%v:8000/healthz", f.GetNginxIP())).
			Expect().
			Status(http.StatusOK)
	})
})
