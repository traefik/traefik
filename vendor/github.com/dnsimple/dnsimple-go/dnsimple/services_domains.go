package dnsimple

import (
	"fmt"
)

func domainServicesPath(accountID string, domainIdentifier string, serviceIdentifier string) string {
	if serviceIdentifier != "" {
		return fmt.Sprintf("/%v/domains/%v/services/%v", accountID, domainIdentifier, serviceIdentifier)
	}
	return fmt.Sprintf("/%v/domains/%v/services", accountID, domainIdentifier)
}

// DomainServiceSettings represents optional settings when applying a DNSimple one-click service to a domain.
type DomainServiceSettings struct {
	Settings map[string]string `url:"settings,omitempty"`
}

// AppliedServices lists the applied one-click services for a domain.
//
// See https://developer.dnsimple.com/v2/services/domains/#applied
func (s *ServicesService) AppliedServices(accountID string, domainIdentifier string, options *ListOptions) (*servicesResponse, error) {
	path := versioned(domainServicesPath(accountID, domainIdentifier, ""))
	servicesResponse := &servicesResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, servicesResponse)
	if err != nil {
		return servicesResponse, err
	}

	servicesResponse.HttpResponse = resp
	return servicesResponse, nil
}

// ApplyService applies a one-click services to a domain.
//
// See https://developer.dnsimple.com/v2/services/domains/#apply
func (s *ServicesService) ApplyService(accountID string, serviceIdentifier string, domainIdentifier string, settings DomainServiceSettings) (*serviceResponse, error) {
	path := versioned(domainServicesPath(accountID, domainIdentifier, serviceIdentifier))
	serviceResponse := &serviceResponse{}

	resp, err := s.client.post(path, settings, nil)
	if err != nil {
		return nil, err
	}

	serviceResponse.HttpResponse = resp
	return serviceResponse, nil
}

// UnapplyService unapplies a one-click services from a domain.
//
// See https://developer.dnsimple.com/v2/services/domains/#unapply
func (s *ServicesService) UnapplyService(accountID string, serviceIdentifier string, domainIdentifier string) (*serviceResponse, error) {
	path := versioned(domainServicesPath(accountID, domainIdentifier, serviceIdentifier))
	serviceResponse := &serviceResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	serviceResponse.HttpResponse = resp
	return serviceResponse, nil
}
