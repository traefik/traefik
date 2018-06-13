package kubernetes

import (
	"fmt"
	"sort"
	"strings"

	"github.com/containous/traefik/provider/label"
	"gopkg.in/yaml.v2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type weightAllocator interface {
	getWeight(host, path, serviceName string) int
}

var _ weightAllocator = &defaultWeightAllocator{}
var _ weightAllocator = &fractionalWeightAllocator{}

type defaultWeightAllocator struct{}

func (d *defaultWeightAllocator) getWeight(host, path, serviceName string) int {
	return label.DefaultWeight
}

type ingressService struct {
	host    string
	path    string
	service string
}

type fractionalWeightAllocator struct {
	serviceWeights map[ingressService]int
}

func (f *fractionalWeightAllocator) String() string {
	var sorted []ingressService
	for ingServ := range f.serviceWeights {
		sorted = append(sorted, ingServ)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].service < sorted[j].service
	})

	var res []string
	for _, ingServ := range sorted {
		res = append(res, fmt.Sprintf("%s: %s", ingServ.service, percentageValue(f.serviceWeights[ingServ])))
	}
	return fmt.Sprintf("[%s]", strings.Join(res, " "))
}

func newFractionalWeightAllocator(ingress *extensionsv1beta1.Ingress, client Client) (*fractionalWeightAllocator, error) {
	servicePercentageWeights, err := getServicesPercentageWeights(ingress)
	if err != nil {
		return nil, err
	}

	allocator := &fractionalWeightAllocator{
		serviceWeights: map[ingressService]int{},
	}

	serviceInstanceCounts, err := getServiceInstanceCounts(ingress, client)
	if err != nil {
		return nil, err
	}

	serviceWeights := map[ingressService]int{}
	for _, rule := range ingress.Spec.Rules {
		// key: rule path string
		// value: service names
		fractionalPathServices := map[string][]string{}

		// key: rule path string
		// value: fractional percentage weight
		fractionalPathWeights := map[string]percentageValue{}

		for _, pa := range rule.HTTP.Paths {
			if _, ok := fractionalPathWeights[pa.Path]; !ok {
				fractionalPathWeights[pa.Path] = newPercentageValueFromFloat64(1)
			}

			if weight, ok := servicePercentageWeights[pa.Backend.ServiceName]; ok {
				ingSvc := ingressService{
					host:    rule.Host,
					path:    pa.Path,
					service: pa.Backend.ServiceName,
				}

				serviceWeights[ingSvc] = weight.computeWeight(serviceInstanceCounts[ingSvc])

				fractionalPathWeights[pa.Path] = fractionalPathWeights[pa.Path].sub(weight)
				if fractionalPathWeights[pa.Path].toFloat64() < 0 {
					return nil, fmt.Errorf("percentage value %s is out of range", fractionalPathWeights[pa.Path].String())
				}
			} else {
				fractionalPathServices[pa.Path] = append(fractionalPathServices[pa.Path], pa.Backend.ServiceName)
			}
		}

		for pa, svcs := range fractionalPathServices {
			totalFractionalInstanceCount := 0

			for _, svc := range svcs {
				totalFractionalInstanceCount += serviceInstanceCounts[ingressService{
					host:    rule.Host,
					path:    pa,
					service: svc,
				}]
			}

			for _, svc := range svcs {
				ingSvc := ingressService{
					host:    rule.Host,
					path:    pa,
					service: svc,
				}
				serviceWeights[ingSvc] = fractionalPathWeights[pa].computeWeight(totalFractionalInstanceCount)
			}
		}
	}

	allocator.serviceWeights = serviceWeights
	return allocator, nil
}

func (f *fractionalWeightAllocator) getWeight(host, path, serviceName string) int {
	return f.serviceWeights[ingressService{
		host:    host,
		path:    path,
		service: serviceName,
	}]
}

func getServicesPercentageWeights(ingress *extensionsv1beta1.Ingress) (map[string]percentageValue, error) {
	percentageWeight := make(map[string]string)

	annotationPercentageWeights := getAnnotationName(ingress.Annotations, annotationKubernetesServiceWeights)
	if err := yaml.Unmarshal([]byte(ingress.Annotations[annotationPercentageWeights]), percentageWeight); err != nil {
		return nil, err
	}

	servicesPercentageWeights := make(map[string]percentageValue)
	for serviceName, percentageStr := range percentageWeight {
		percentageValue, err := newPercentageValueFromString(percentageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid percentage value %q", percentageStr)
		}

		servicesPercentageWeights[serviceName] = percentageValue
	}
	return servicesPercentageWeights, nil
}

func getServiceInstanceCounts(ingress *extensionsv1beta1.Ingress, client Client) (map[ingressService]int, error) {
	serviceInstanceCounts := map[ingressService]int{}

	for _, rule := range ingress.Spec.Rules {
		for _, pa := range rule.HTTP.Paths {
			count := 0
			endpoints, exists, err := client.GetEndpoints(ingress.Namespace, pa.Backend.ServiceName)
			if err != nil {
				return nil, fmt.Errorf("failed to get endpoints %s/%s: %v", ingress.Namespace, pa.Backend.ServiceName, err)
			}
			if !exists {
				return nil, fmt.Errorf("endpoints not found for %s/%s", ingress.Namespace, pa.Backend.ServiceName)
			}

			for _, subset := range endpoints.Subsets {
				count += len(subset.Addresses)
			}

			serviceInstanceCounts[ingressService{
				host:    rule.Host,
				path:    pa.Path,
				service: pa.Backend.ServiceName,
			}] += count
		}
	}

	return serviceInstanceCounts, nil
}
