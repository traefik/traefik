package project

import (
	"fmt"

	"golang.org/x/net/context"
)

// Port returns the public port for a port binding of the specified service.
func (p *Project) Port(ctx context.Context, index int, protocol, serviceName, privatePort string) (string, error) {
	service, err := p.CreateService(serviceName)
	if err != nil {
		return "", err
	}

	containers, err := service.Containers(ctx)
	if err != nil {
		return "", err
	}

	if index < 1 || index > len(containers) {
		return "", fmt.Errorf("Invalid index %d", index)
	}

	return containers[index-1].Port(ctx, fmt.Sprintf("%s/%s", privatePort, protocol))
}
