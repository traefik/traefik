package dnsimple

import (
	"fmt"
)

// TemplateRecord represents a DNS record for a template in DNSimple.
type TemplateRecord struct {
	ID         int    `json:"id,omitempty"`
	TemplateID int    `json:"template_id,omitempty"`
	Name       string `json:"name"`
	Content    string `json:"content,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	Type       string `json:"type,omitempty"`
	Priority   int    `json:"priority,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

func templateRecordPath(accountID string, templateIdentifier string, templateRecordID string) string {
	if templateRecordID != "" {
		return fmt.Sprintf("%v/records/%v", templatePath(accountID, templateIdentifier), templateRecordID)
	}

	return templatePath(accountID, templateIdentifier) + "/records"
}

// templateRecordResponse represents a response from an API method that returns a TemplateRecord struct.
type templateRecordResponse struct {
	Response
	Data *TemplateRecord `json:"data"`
}

// templateRecordsResponse represents a response from an API method that returns a collection of TemplateRecord struct.
type templateRecordsResponse struct {
	Response
	Data []TemplateRecord `json:"data"`
}

// ListTemplateRecords list the templates for an account.
//
// See https://developer.dnsimple.com/v2/templates/records/#list
func (s *TemplatesService) ListTemplateRecords(accountID string, templateIdentifier string, options *ListOptions) (*templateRecordsResponse, error) {
	path := versioned(templateRecordPath(accountID, templateIdentifier, ""))
	templateRecordsResponse := &templateRecordsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, templateRecordsResponse)
	if err != nil {
		return templateRecordsResponse, err
	}

	templateRecordsResponse.HttpResponse = resp
	return templateRecordsResponse, nil
}

// CreateTemplateRecord creates a new template record.
//
// See https://developer.dnsimple.com/v2/templates/records/#create
func (s *TemplatesService) CreateTemplateRecord(accountID string, templateIdentifier string, templateRecordAttributes TemplateRecord) (*templateRecordResponse, error) {
	path := versioned(templateRecordPath(accountID, templateIdentifier, ""))
	templateRecordResponse := &templateRecordResponse{}

	resp, err := s.client.post(path, templateRecordAttributes, templateRecordResponse)
	if err != nil {
		return nil, err
	}

	templateRecordResponse.HttpResponse = resp
	return templateRecordResponse, nil
}

// GetTemplateRecord fetches a template record.
//
// See https://developer.dnsimple.com/v2/templates/records/#get
func (s *TemplatesService) GetTemplateRecord(accountID string, templateIdentifier string, templateRecordID string) (*templateRecordResponse, error) {
	path := versioned(templateRecordPath(accountID, templateIdentifier, templateRecordID))
	templateRecordResponse := &templateRecordResponse{}

	resp, err := s.client.get(path, templateRecordResponse)
	if err != nil {
		return nil, err
	}

	templateRecordResponse.HttpResponse = resp
	return templateRecordResponse, nil
}

// DeleteTemplateRecord deletes a template record.
//
// See https://developer.dnsimple.com/v2/templates/records/#delete
func (s *TemplatesService) DeleteTemplateRecord(accountID string, templateIdentifier string, templateRecordID string) (*templateRecordResponse, error) {
	path := versioned(templateRecordPath(accountID, templateIdentifier, templateRecordID))
	templateRecordResponse := &templateRecordResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	templateRecordResponse.HttpResponse = resp
	return templateRecordResponse, nil
}
