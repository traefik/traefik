/*
Copyright 2014 The Kubernetes Authors.

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

package framework

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/ingress-nginx/internal/k8s"
)

const (
	// Poll how often to poll for conditions
	Poll = 2 * time.Second

	// DefaultTimeout time to wait for operations to complete
	DefaultTimeout = 90 * time.Second
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func logf(level, format string, args ...interface{}) {
	fmt.Fprintf(ginkgo.GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

// Logf logs to the INFO logs.
func Logf(format string, args ...interface{}) {
	logf("INFO", format, args...)
}

// Failf logs to the INFO logs and fails the test.
func Failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logf("INFO", msg)
	ginkgo.Fail(nowStamp()+": "+msg, 1)
}

// RestclientConfig deserializes the contents of a kubeconfig file into a Config object.
func RestclientConfig(config, newContext string) (*api.Config, error) {
	Logf(">>> config: %s\n", config)
	if config == "" {
		return nil, fmt.Errorf("config file must be specified to load client config")
	}
	c, err := clientcmd.LoadFromFile(config)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err.Error())
	}
	if newContext != "" {
		Logf(">>> context: %s\n", newContext)
		c.CurrentContext = newContext
	}
	return c, nil
}

// RunID unique identifier of the e2e run
var RunID = uuid.NewUUID()

func createNamespace(baseName string, labels map[string]string, c kubernetes.Interface) (string, error) {
	ts := time.Now().UnixNano()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("e2e-tests-%v-%v-", baseName, ts),
			Labels:       labels,
		},
	}

	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	var err error

	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	err = wait.Poll(Poll, DefaultTimeout, func() (bool, error) {
		got, err = c.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return "", err
	}
	return got.Name, nil
}

// CreateKubeNamespace creates a new namespace in the cluster
func CreateKubeNamespace(baseName string, c kubernetes.Interface) (string, error) {
	return createNamespace(baseName, nil, c)
}

// CreateKubeNamespaceWithLabel creates a new namespace with given labels in the cluster
func CreateKubeNamespaceWithLabel(baseName string, labels map[string]string, c kubernetes.Interface) (string, error) {
	return createNamespace(baseName, labels, c)
}

// DeleteKubeNamespace deletes a namespace and all the objects inside
func DeleteKubeNamespace(c kubernetes.Interface, namespace string) error {
	grace := int64(0)
	pb := metav1.DeletePropagationBackground
	return c.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	})
}

// CreateIngressClass creates a new IngressClass related to a test/namespace and also
// the required ClusterRole/ClusterRoleBinding
func CreateIngressClass(namespace string, c kubernetes.Interface) (string, error) {
	icname := fmt.Sprintf("ic-%s", namespace)
	var err error

	ic, err := c.NetworkingV1().IngressClasses().
		Create(context.TODO(), &networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: icname,
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: k8s.IngressNGINXController,
			},
		}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("unexpected error creating IngressClass %s: %v", icname, err)
	}

	_, err = c.RbacV1().ClusterRoles().Create(context.TODO(), &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: icname},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingressclasses"},
			Verbs:     []string{"get", "list", "watch"},
		}},
	}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("unexpected error creating IngressClass ClusterRole %s: %v", icname, err)
	}

	_, err = c.RbacV1().ClusterRoleBindings().Create(context.TODO(), &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: icname,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     icname,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup:  "",
				Kind:      "ServiceAccount",
				Namespace: namespace,
				Name:      "nginx-ingress",
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("unexpected error creating IngressClass ClusterRoleBinding %s: %v", icname, err)
	}
	return ic.Name, nil
}

// deleteIngressClass deletes an IngressClass and its related ClusterRole* objects
func deleteIngressClass(c kubernetes.Interface, ingressclass string) error {
	var err error
	grace := int64(0)
	pb := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
		PropagationPolicy:  &pb,
	}
	err = c.NetworkingV1().IngressClasses().Delete(context.TODO(), ingressclass, deleteOptions)
	if err != nil {
		return fmt.Errorf("unexpected error deleting IngressClass %s: %v", ingressclass, err)
	}

	err = c.RbacV1().ClusterRoleBindings().Delete(context.TODO(), ingressclass, deleteOptions)
	if err != nil {
		return fmt.Errorf("unexpected error deleting IngressClass ClusterRoleBinding %s: %v", ingressclass, err)
	}
	err = c.RbacV1().ClusterRoles().Delete(context.TODO(), ingressclass, deleteOptions)
	if err != nil {
		return fmt.Errorf("unexpected error deleting IngressClass ClusterRole %s: %v", ingressclass, err)
	}

	return nil
}

// GetIngressClassName returns the default IngressClassName given a namespace
func GetIngressClassName(namespace string) *string {
	icname := fmt.Sprintf("ic-%s", namespace)
	return &icname
}

// WaitForKubeNamespaceNotExist waits until a namespaces is not present in the cluster
func WaitForKubeNamespaceNotExist(c kubernetes.Interface, namespace string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, namespaceNotExist(c, namespace))
}

func namespaceNotExist(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}
}

// WaitForNoPodsInNamespace waits until there are no pods running in a namespace
func WaitForNoPodsInNamespace(c kubernetes.Interface, namespace string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, noPodsInNamespace(c, namespace))
}

func noPodsInNamespace(c kubernetes.Interface, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		items, err := c.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		if len(items.Items) == 0 {
			return true, nil
		}
		return false, nil
	}
}

// WaitForPodRunningInNamespace waits a default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodRunningInNamespace(c kubernetes.Interface, pod *corev1.Pod) error {
	if pod.Status.Phase == corev1.PodRunning {
		return nil
	}
	return waitTimeoutForPodRunningInNamespace(c, pod.Name, pod.Namespace)
}

func waitTimeoutForPodRunningInNamespace(c kubernetes.Interface, podName, namespace string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, podRunning(c, podName, namespace))
}

// WaitForSecretInNamespace waits a default amount of time for the specified secret is present in a particular namespace
func WaitForSecretInNamespace(c kubernetes.Interface, namespace, name string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, secretInNamespace(c, namespace, name))
}

func secretInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		s, err := c.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if s != nil {
			return true, nil
		}
		return false, nil
	}
}

// WaitForFileInFS waits a default amount of time for the specified file is present in the filesystem
func WaitForFileInFS(file string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, fileInFS(file))
}

func fileInFS(file string) wait.ConditionFunc {
	return func() (bool, error) {
		stat, err := os.Stat(file)
		if err != nil {
			return false, err
		}

		if stat == nil {
			return false, fmt.Errorf("file %v does not exist", file)
		}

		if stat.Size() > 0 {
			return true, nil
		}

		return false, fmt.Errorf("the file %v exists but it is empty", file)
	}
}

// WaitForNoIngressInNamespace waits until there is no ingress object in a particular namespace
func WaitForNoIngressInNamespace(c kubernetes.Interface, namespace, name string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, noIngressInNamespace(c, namespace, name))
}

func noIngressInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		ing, err := c.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		if ing == nil {
			return true, nil
		}
		return false, nil
	}
}

// WaitForIngressInNamespace waits until a particular ingress object exists namespace
func WaitForIngressInNamespace(c kubernetes.Interface, namespace, name string) error {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	return wait.Poll(Poll, DefaultTimeout, ingressInNamespace(c, namespace, name))
}

func ingressInNamespace(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		ing, err := c.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		if ing != nil {
			return true, nil
		}
		return false, nil
	}
}

func podRunning(c kubernetes.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod ran to completion")
		}
		return false, nil
	}
}

func labelSelectorToString(labels map[string]string) string {
	var str string
	for k, v := range labels {
		str += fmt.Sprintf("%s=%s,", k, v)
	}
	return str[:len(str)-1]
}

// NewInt32 converts int32 to a pointer
func NewInt32(val int32) *int32 {
	p := new(int32)
	*p = val
	return p
}

// NewInt64 converts int64 to a pointer
func NewInt64(val int64) *int64 {
	p := new(int64)
	*p = val
	return p
}
