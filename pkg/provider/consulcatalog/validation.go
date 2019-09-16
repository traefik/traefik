package consulcatalog

import "fmt"

func (p *Provider) validateConfig() error {

	if p.Protocol != "http" && p.Protocol != "tcp" {
		return fmt.Errorf("wrong protocol specified, allowed values are http and tcp")
	}

	if len(p.Entrypoints) == 0 {
		return fmt.Errorf("default entrypoints must be specified")
	}

	return nil
}
