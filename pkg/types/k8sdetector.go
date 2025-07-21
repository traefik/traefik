package types

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8sAttributesDetector detects the metadata of the Traefik pod running in a Kubernetes cluster.
// It reads the pod name from the hostname file and the namespace from the service account namespace file and queries the Kubernetes API to get the pod's UID.
type K8sAttributesDetector struct{}

func (K8sAttributesDetector) Detect(ctx context.Context) (*resource.Resource, error) {
	attrs := os.Getenv("OTEL_RESOURCE_ATTRIBUTES")
	if strings.Contains(attrs, string(semconv.K8SPodNameKey)) || strings.Contains(attrs, string(semconv.K8SPodUIDKey)) {
		return resource.Empty(), nil
	}

	// The InClusterConfig function returns a config for the Kubernetes API server
	// when it is running inside a Kubernetes cluster.
	config, err := rest.InClusterConfig()
	if err != nil && errors.Is(err, rest.ErrNotInCluster) {
		return resource.Empty(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("creating in cluster config: %w", err)
	}

	client, err := kclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	podName, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("getting pod name: %w", err)
	}

	podNamespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, fmt.Errorf("getting pod namespace: %w", err)
	}
	podNamespace := string(podNamespaceBytes)

	pod, err := client.CoreV1().Pods(podNamespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil && kerror.IsForbidden(err) {
		log.Error().Err(err).Msg("Unable to build K8s resource attributes for Traefik pod")
		return resource.Empty(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting pod metadata: %w", err)
	}

	// To avoid version conflicts with other detectors, we use a Schemaless resource.
	return resource.NewSchemaless(
		semconv.K8SPodUID(string(pod.UID)),
		semconv.K8SPodName(pod.Name),
		semconv.K8SNamespaceName(podNamespace),
	), nil
}
