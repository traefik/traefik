package dnsimple

import (
	"fmt"
)

// ApplyTemplate applies a template to the given domain.
//
// See https://developer.dnsimple.com/v2/templates/domains/#apply
func (s *TemplatesService) ApplyTemplate(accountID string, templateIdentifier string, domainIdentifier string) (*templateResponse, error) {
	path := versioned(fmt.Sprintf("%v/templates/%v", domainPath(accountID, domainIdentifier), templateIdentifier))
	templateResponse := &templateResponse{}

	resp, err := s.client.post(path, nil, nil)
	if err != nil {
		return nil, err
	}

	templateResponse.HttpResponse = resp
	return templateResponse, nil
}
