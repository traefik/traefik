package consulcatalog

import (
	"fmt"
)

func validateProtocol(protocol string) error {
	if protocol != "http" && protocol != "tcp" {
		return fmt.Errorf("wrong protocol '%s', allowed 'http' or 'tcp'", protocol)
	}

	return nil
}

func (p *Provider) validateConfig() error {

	if err := validateProtocol(p.Protocol); err != nil {
		return err
	}

	return nil
}
