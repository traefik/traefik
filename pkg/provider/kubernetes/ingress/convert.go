package ingress

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

type marshaler interface {
	Marshal() ([]byte, error)
}

type unmarshaler interface {
	Unmarshal(data []byte) error
}

type LoadBalancerIngress interface {
	corev1.LoadBalancerIngress | netv1.IngressLoadBalancerIngress
}

// convertSlice converts slice of LoadBalancerIngress to slice of LoadBalancerIngress.
// O (Bar), I (Foo) => []Bar.
func convertSlice[O LoadBalancerIngress, I LoadBalancerIngress](loadBalancerIngresses []I) ([]O, error) {
	var results []O

	for _, loadBalancerIngress := range loadBalancerIngresses {
		mar, ok := any(&loadBalancerIngress).(marshaler)
		if !ok {
			// All the pointer of types related to the interface LoadBalancerIngress are compatible with the interface marshaler.
			continue
		}

		um, err := convert[O](mar)
		if err != nil {
			return nil, err
		}

		v, ok := any(*um).(O)
		if !ok {
			continue
		}

		results = append(results, v)
	}

	return results, nil
}

// convert must only be used with unmarshaler and marshaler compatible types.
func convert[T any](input marshaler) (*T, error) {
	data, err := input.Marshal()
	if err != nil {
		return nil, err
	}

	var output T
	um, ok := any(&output).(unmarshaler)
	if !ok {
		return nil, errors.New("the output type doesn't implement unmarshaler interface")
	}

	err = um.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
