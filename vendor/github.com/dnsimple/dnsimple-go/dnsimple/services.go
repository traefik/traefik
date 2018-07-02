package dnsimple

import (
	"fmt"
)

// ServicesService handles communication with the service related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/services/
type ServicesService struct {
	client *Client
}

// Service represents a Service in DNSimple.
type Service struct {
	ID               int64            `json:"id,omitempty"`
	SID              string           `json:"sid,omitempty"`
	Name             string           `json:"name,omitempty"`
	Description      string           `json:"description,omitempty"`
	SetupDescription string           `json:"setup_description,omitempty"`
	RequiresSetup    bool             `json:"requires_setup,omitempty"`
	DefaultSubdomain string           `json:"default_subdomain,omitempty"`
	CreatedAt        string           `json:"created_at,omitempty"`
	UpdatedAt        string           `json:"updated_at,omitempty"`
	Settings         []ServiceSetting `json:"settings,omitempty"`
}

// ServiceSetting represents a single group of settings for a DNSimple Service.
type ServiceSetting struct {
	Name        string `json:"name,omitempty"`
	Label       string `json:"label,omitempty"`
	Append      string `json:"append,omitempty"`
	Description string `json:"description,omitempty"`
	Example     string `json:"example,omitempty"`
	Password    bool   `json:"password,omitempty"`
}

func servicePath(serviceIdentifier string) (path string) {
	path = "/services"
	if serviceIdentifier != "" {
		path += fmt.Sprintf("/%v", serviceIdentifier)
	}
	return
}

// serviceResponse represents a response from an API method that returns a Service struct.
type serviceResponse struct {
	Response
	Data *Service `json:"data"`
}

// servicesResponse represents a response from an API method that returns a collection of Service struct.
type servicesResponse struct {
	Response
	Data []Service `json:"data"`
}

// ListServices lists the one-click services available in DNSimple.
//
// See https://developer.dnsimple.com/v2/services/#list
func (s *ServicesService) ListServices(options *ListOptions) (*servicesResponse, error) {
	path := versioned(servicePath(""))
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

// GetService fetches a one-click service.
//
// See https://developer.dnsimple.com/v2/services/#get
func (s *ServicesService) GetService(serviceIdentifier string) (*serviceResponse, error) {
	path := versioned(servicePath(serviceIdentifier))
	serviceResponse := &serviceResponse{}

	resp, err := s.client.get(path, serviceResponse)
	if err != nil {
		return nil, err
	}

	serviceResponse.HttpResponse = resp
	return serviceResponse, nil
}
