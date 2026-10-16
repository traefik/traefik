package framework

// This file is not part of upstream ingress-nginx.
// It replaces the ingress-nginx controller deployed by the e2e framework with Traefik,
// running the kubernetesIngressNGINX provider, so that the upstream test cases can be reused verbatim.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/ingress-nginx/test/e2e/framework/httpexpect"
)

const (
	// The names below are the ones the upstream framework uses to look up the controller.
	controllerName           = "nginx-ingress-controller"
	controllerServiceAccount = "nginx-ingress"
	controllerContainer      = "controller"

	traefikClusterRole  = "traefik-ingress-nginx-e2e"
	traefikAPIPort      = 9000
	allowSnippetArg     = "--providers.kubernetesingressnginx.allowSnippetAnnotations=true"
	defaultTraefikImage = "traefik/traefik:latest"
)

// traefikImage is the image of the Traefik controller under test, it must already be present on the nodes.
var traefikImage = envOrDefault("TRAEFIK_IMAGE", defaultTraefikImage)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadKubeconfig falls back to the KUBECONFIG file because the suite runs on a k3s node, not in a pod.
func loadKubeconfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, err
		}
	}

	config.UserAgent = "ingress-nginx-e2e"
	return config, nil
}

// skipNginxOnly skips the current spec because it relies on something only ingress-nginx provides.
func skipNginxOnly(helper string) {
	ginkgo.Skip(fmt.Sprintf("%s: inspects the ingress-nginx controller directly", helper))
}

// deployTraefik deploys Traefik in place of the ingress-nginx helm chart.
func (f *Framework) deployTraefik(namespace string) error {
	ctx := context.TODO()

	_, err := f.KubeClientSet.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: traefikClusterRole},
		// Mirrors docs/content/reference/dynamic-configuration/kubernetes-ingress-nginx-rbac.yml in Traefik.
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"services", "secrets", "configmaps"}, Verbs: []string{"get", "list", "watch"}},
			{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"list", "watch"}},
			{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"}},
			{APIGroups: []string{"discovery.k8s.io"}, Resources: []string{"endpointslices"}, Verbs: []string{"list", "watch"}},
			{APIGroups: []string{"networking.k8s.io"}, Resources: []string{"ingresses", "ingressclasses"}, Verbs: []string{"list", "watch"}},
			{APIGroups: []string{"networking.k8s.io"}, Resources: []string{"ingresses/status"}, Verbs: []string{"update"}},
		},
	}, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating ClusterRole: %w", err)
	}

	_, err = f.KubeClientSet.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: traefikClusterRoleBinding(namespace)},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     traefikClusterRole,
		},
		Subjects: []rbacv1.Subject{{Kind: "ServiceAccount", Namespace: namespace, Name: controllerServiceAccount}},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating ClusterRoleBinding: %w", err)
	}

	_, err = f.KubeClientSet.CoreV1().ServiceAccounts(namespace).Create(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: controllerServiceAccount},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating ServiceAccount: %w", err)
	}

	labels := map[string]string{
		"app.kubernetes.io/name":      "ingress-nginx",
		"app.kubernetes.io/component": "controller",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: controllerName, Namespace: namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: NewInt32(1),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			// Recreate guarantees that, after an update, the only running pod has the new arguments.
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ServiceAccountName:            controllerServiceAccount,
					TerminationGracePeriodSeconds: NewInt64(0),
					Containers: []corev1.Container{{
						Name:            controllerContainer,
						Image:           traefikImage,
						ImagePullPolicy: corev1.PullNever,
						Args:            traefikArgs(namespace),
						Ports: []corev1.ContainerPort{
							{Name: "http", ContainerPort: 80},
							{Name: "https", ContainerPort: 443},
							{Name: "traefik", ContainerPort: traefikAPIPort},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{Path: "/ping", Port: intstr.FromInt32(traefikAPIPort)},
							},
							PeriodSeconds: 1,
						},
					}},
				},
			},
		},
	}
	if _, err = f.KubeClientSet.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("creating Deployment: %w", err)
	}

	_, err = f.KubeClientSet.CoreV1().Services(namespace).Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: controllerName, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, TargetPort: intstr.FromString("http")},
				{Name: "https", Port: 443, TargetPort: intstr.FromString("https")},
				{Name: "traefik", Port: traefikAPIPort, TargetPort: intstr.FromString("traefik")},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating Service: %w", err)
	}

	return waitForDeploymentRollout(f.KubeClientSet, deployment)
}

func traefikArgs(namespace string) []string {
	return []string{
		"--log.level=DEBUG",
		"--accesslog=true",
		"--api.insecure=true",
		"--ping=true",
		"--entrypoints.web.address=:80",
		"--entrypoints.websecure.address=:443",
		fmt.Sprintf("--entrypoints.traefik.address=:%d", traefikAPIPort),
		"--providers.kubernetesingressnginx=true",
		"--providers.kubernetesingressnginx.watchNamespace=" + namespace,
		"--providers.kubernetesingressnginx.httpEntryPoint=web",
		"--providers.kubernetesingressnginx.httpsEntryPoint=websecure",
	}
}

func traefikClusterRoleBinding(namespace string) string {
	return "traefik-" + namespace
}

// deleteTraefik removes the cluster scoped resources of deployTraefik, the namespaced ones go away with the namespace.
func (f *Framework) deleteTraefik() error {
	err := f.KubeClientSet.RbacV1().ClusterRoleBindings().Delete(context.TODO(), traefikClusterRoleBinding(f.Namespace), metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

// updateTraefikArgs restarts Traefik with the arguments returned by update.
func (f *Framework) updateTraefikArgs(update func(args []string) []string) {
	err := UpdateDeployment(f.KubeClientSet, f.Namespace, controllerName, 1, func(d *appsv1.Deployment) error {
		container := &d.Spec.Template.Spec.Containers[0]
		container.Args = update(container.Args)
		_, err := f.KubeClientSet.AppsV1().Deployments(f.Namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
		return err
	})
	assert.Nil(ginkgo.GinkgoT(), err, "updating Traefik arguments")

	err = f.updateIngressNGINXPod()
	assert.Nil(ginkgo.GinkgoT(), err, "updating ingress controller pod information")

	f.WaitForNginxListening(80)
	f.waitForStableTraefikConfig()
}

// allowSnippetAnnotations maps the allow-snippet-annotations ConfigMap key to the equivalent provider option.
func (f *Framework) allowSnippetAnnotations() func() {
	f.updateTraefikArgs(func(args []string) []string {
		return append(args, allowSnippetArg)
	})

	return func() {
		f.updateTraefikArgs(func(args []string) []string {
			return slices.DeleteFunc(args, func(arg string) bool { return arg == allowSnippetArg })
		})
	}
}

func (f *Framework) traefikRawData() (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/api/rawdata", f.GetNginxIP(), traefikAPIPort))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

const (
	traefikPoll = 500 * time.Millisecond
	// traefikStableFor must be longer than the default providers throttle duration (2s).
	traefikStableFor = 3 * time.Second
	// traefikNoChangeTimeout bounds the wait when an action does not change the configuration,
	// like upstream does when no reload happens.
	traefikNoChangeTimeout = 10 * time.Second
)

// waitForTraefikReload replaces the wait for an ingress-nginx RELOAD event:
// it calls fn, then waits until the Traefik configuration has changed and stopped changing.
func (f *Framework) waitForTraefikReload(fn func()) {
	before, _ := f.traefikRawData()

	fn()

	start := time.Now()
	f.waitForTraefikConfig(func(current string) bool {
		return current != before || time.Since(start) > traefikNoChangeTimeout
	})
}

// waitForStableTraefikConfig waits until the Traefik configuration stopped changing.
func (f *Framework) waitForStableTraefikConfig() {
	f.waitForTraefikConfig(func(string) bool { return true })
}

func (f *Framework) waitForTraefikConfig(ready func(current string) bool) {
	var last string
	var stableSince time.Time

	err := wait.PollUntilContextTimeout(context.Background(), traefikPoll, DefaultTimeout, true, func(context.Context) (bool, error) {
		current, err := f.traefikRawData()
		if err != nil {
			return false, nil
		}

		if current != last || stableSince.IsZero() {
			last = current
			stableSince = time.Now()
			return false, nil
		}

		return ready(current) && time.Since(stableSince) >= traefikStableFor, nil
	})
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for Traefik configuration")
}

// waitForNginxConfigAssertion replaces the nginx.conf matchers: they cannot be evaluated against Traefik,
// so they only act as a synchronization barrier, and the spec is skipped if it makes no HTTP assertion at all.
func (f *Framework) waitForNginxConfigAssertion() {
	f.nginxConfigAsserted = true
	f.waitForStableTraefikConfig()
}

func (f *Framework) resetSpecTracking() {
	f.nginxConfigAsserted = false
	f.requestsAtStart = httpexpect.RequestCount.Load()
}

// skipIfOnlyNginxConfigAsserted runs after the spec to avoid reporting as passed a spec whose only assertions
// were nginx.conf matchers. Skipping from an AfterEach turns a passed spec into a skipped one.
func (f *Framework) skipIfOnlyNginxConfigAsserted() {
	if ginkgo.CurrentSpecReport().State != ginkgotypes.SpecStatePassed {
		return
	}

	if f.nginxConfigAsserted && httpexpect.RequestCount.Load() == f.requestsAtStart {
		ginkgo.Skip("only asserts on the nginx configuration")
	}
}

// dumpTraefikState replaces the nginx.conf and logs dump done after a failure.
func (f *Framework) dumpTraefikState() {
	if f.pod != nil {
		logs, err := Logs(f.KubeClientSet, f.Namespace, f.pod.Name)
		if err != nil {
			Logf("Unexpected error obtaining Traefik logs: %v", err)
		} else {
			ginkgo.By("Dumping Traefik logs")
			Logf("%v", logs)
		}
	}

	rawData, err := f.traefikRawData()
	if err != nil {
		Logf("Unexpected error obtaining Traefik configuration: %v", err)
		return
	}

	ginkgo.By("Dumping Traefik configuration after failure")
	Logf("%v", rawData)
}
