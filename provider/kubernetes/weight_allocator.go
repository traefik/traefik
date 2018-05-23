package kubernetes

import (
	"fmt"
	"gopkg.in/yaml.v2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

type weightAllocator interface {
	allocWeight(host, path, serviceName string) int
}

var _ weightAllocator = &defaultWeightAllocator{}
var _ weightAllocator = &fractionalWeightAllocator{}

type defaultWeightAllocator struct{}

func (allocator *defaultWeightAllocator) allocWeight(host, path, serviceName string) int {
	return 1
}

type ingressService struct {
	host    string
	path    string
	service string
}

type fractionalWeightAllocator struct {
	serviceWeights map[ingressService]int
}

func newFractionalWeightAllocator(ingress *extensionsv1beta1.Ingress, client Client) (*fractionalWeightAllocator, error) {
	allocator := &fractionalWeightAllocator{
		serviceWeights: map[ingressService]int{},
	}
	serviceInstanceCounts := map[ingressService]int{}
	serviceWeights := map[ingressService]int{}

	servicePercentageWeights, err := getServicesPercentageWeights(ingress.Annotations[annotationKubernetesPercentageWeights])
	if err != nil {
		return nil, err
	}

	for _, rule := range ingress.Spec.Rules {
		for _, pa := range rule.HTTP.Paths {
			count := 0
			endpoints, exists, err := client.GetEndpoints(ingress.Namespace, pa.Backend.ServiceName)
			if !exists || err != nil {
				return nil, fmt.Errorf("fail to get endpoints %s/%s", ingress.Namespace, pa.Backend.ServiceName)
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
					return nil, fmt.Errorf("percentage value %s out of range", fractionalPathWeights[pa.Path].String())
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

func (allocator *fractionalWeightAllocator) allocWeight(host, path, serviceName string) int {
	return allocator.serviceWeights[ingressService{
		host:    host,
		path:    path,
		service: serviceName,
	}]
}

func getServicesPercentageWeights(percentageAnnotation string) (map[string]percentageValue, error) {
	percentageWeight := make(map[string]string)

	if err := yaml.Unmarshal([]byte(percentageAnnotation), percentageWeight); err != nil {
		return nil, err
	}

	servicesPercentageWeights := make(map[string]percentageValue)
	for serviceName, percentageStr := range percentageWeight {
		percentageValue, err := newPercentageValueFromString(percentageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid percentage value %q in ingress: %v", percentageStr, err)
		}
		servicesPercentageWeights[serviceName] = percentageValue
	}
	return servicesPercentageWeights, nil
}
