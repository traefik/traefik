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

//nolint:dupl // Ignore dupl errors for similar test case
package framework

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewFastCGIHelloServerDeployment creates a new single replica
// deployment of the fortune teller image in a particular namespace
func (f *Framework) NewFastCGIHelloServerDeployment() {
	f.NewNewFastCGIHelloServerDeploymentWithReplicas(1)
}

// NewNewFastCGIHelloServerDeploymentWithReplicas creates a new deployment of the
// fortune teller image in a particular namespace. Number of replicas is configurable
func (f *Framework) NewNewFastCGIHelloServerDeploymentWithReplicas(replicas int32) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fastcgi-helloserver",
			Namespace: f.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: NewInt32(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "fastcgi-helloserver",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "fastcgi-helloserver",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{
						{
							Name:  "fastcgi-helloserver",
							Image: "registry.k8s.io/ingress-nginx/fastcgi-helloserver:v1.2.9@sha256:0e3d0f312967bec687cc335accb989a6b72e3faa20c1d71e4e77ab842320b9c6",
							Env:   []corev1.EnvVar{},
							Ports: []corev1.ContainerPort{
								{
									Name:          "fastcgi",
									ContainerPort: 9000,
								},
							},
						},
					},
				},
			},
		},
	}

	d := f.EnsureDeployment(deployment)

	err := waitForPodsReady(f.KubeClientSet, DefaultTimeout, int(replicas), f.Namespace, &metav1.ListOptions{
		LabelSelector: fields.SelectorFromSet(fields.Set(d.Spec.Template.ObjectMeta.Labels)).String(),
	})
	assert.Nil(ginkgo.GinkgoT(), err, "failed to wait for to become ready")

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fastcgi-helloserver",
			Namespace: f.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "fastcgi",
					Port:       9000,
					TargetPort: intstr.FromInt(9000),
					Protocol:   "TCP",
				},
			},
			Selector: map[string]string{
				"app": "fastcgi-helloserver",
			},
		},
	}

	f.EnsureService(service)
}
