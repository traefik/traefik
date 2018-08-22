package dnsimple

import (
	"fmt"
)

// TemplatesService handles communication with the template related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/templates/
type TemplatesService struct {
	client *Client
}

// Template represents a Template in DNSimple.
type Template struct {
	ID          int64  `json:"id,omitempty"`
	SID         string `json:"sid,omitempty"`
	AccountID   int64  `json:"account_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

func templatePath(accountID string, templateIdentifier string) (path string) {
	path = fmt.Sprintf("/%v/templates", accountID)
	if templateIdentifier != "" {
		path += fmt.Sprintf("/%v", templateIdentifier)
	}
	return
}

// templateResponse represents a response from an API method that returns a Template struct.
type templateResponse struct {
	Response
	Data *Template `json:"data"`
}

// templatesResponse represents a response from an API method that returns a collection of Template struct.
type templatesResponse struct {
	Response
	Data []Template `json:"data"`
}

// ListTemplates list the templates for an account.
//
// See https://developer.dnsimple.com/v2/templates/#list
func (s *TemplatesService) ListTemplates(accountID string, options *ListOptions) (*templatesResponse, error) {
	path := versioned(templatePath(accountID, ""))
	templatesResponse := &templatesResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, templatesResponse)
	if err != nil {
		return templatesResponse, err
	}

	templatesResponse.HttpResponse = resp
	return templatesResponse, nil
}

// CreateTemplate creates a new template.
//
// See https://developer.dnsimple.com/v2/templates/#create
func (s *TemplatesService) CreateTemplate(accountID string, templateAttributes Template) (*templateResponse, error) {
	path := versioned(templatePath(accountID, ""))
	templateResponse := &templateResponse{}

	resp, err := s.client.post(path, templateAttributes, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HttpResponse = resp
	return templateResponse, nil
}

// GetTemplate fetches a template.
//
// See https://developer.dnsimple.com/v2/templates/#get
func (s *TemplatesService) GetTemplate(accountID string, templateIdentifier string) (*templateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &templateResponse{}

	resp, err := s.client.get(path, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HttpResponse = resp
	return templateResponse, nil
}

// UpdateTemplate updates a template.
//
// See https://developer.dnsimple.com/v2/templates/#update
func (s *TemplatesService) UpdateTemplate(accountID string, templateIdentifier string, templateAttributes Template) (*templateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &templateResponse{}

	resp, err := s.client.patch(path, templateAttributes, templateResponse)
	if err != nil {
		return nil, err
	}

	templateResponse.HttpResponse = resp
	return templateResponse, nil
}

// DeleteTemplate deletes a template.
//
// See https://developer.dnsimple.com/v2/templates/#delete
func (s *TemplatesService) DeleteTemplate(accountID string, templateIdentifier string) (*templateResponse, error) {
	path := versioned(templatePath(accountID, templateIdentifier))
	templateResponse := &templateResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	templateResponse.HttpResponse = resp
	return templateResponse, nil
}
