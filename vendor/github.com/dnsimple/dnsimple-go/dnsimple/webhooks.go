package dnsimple

import (
	"fmt"
)

// WebhooksService handles communication with the webhook related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/webhooks
type WebhooksService struct {
	client *Client
}

// Webhook represents a DNSimple webhook.
type Webhook struct {
	ID  int64  `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

func webhookPath(accountID string, webhookID int64) (path string) {
	path = fmt.Sprintf("/%v/webhooks", accountID)
	if webhookID != 0 {
		path = fmt.Sprintf("%v/%v", path, webhookID)
	}
	return
}

// webhookResponse represents a response from an API method that returns a Webhook struct.
type webhookResponse struct {
	Response
	Data *Webhook `json:"data"`
}

// webhookResponse represents a response from an API method that returns a collection of Webhook struct.
type webhooksResponse struct {
	Response
	Data []Webhook `json:"data"`
}

// ListWebhooks lists the webhooks for an account.
//
// See https://developer.dnsimple.com/v2/webhooks#list
func (s *WebhooksService) ListWebhooks(accountID string, _ *ListOptions) (*webhooksResponse, error) {
	path := versioned(webhookPath(accountID, 0))
	webhooksResponse := &webhooksResponse{}

	resp, err := s.client.get(path, webhooksResponse)
	if err != nil {
		return webhooksResponse, err
	}

	webhooksResponse.HttpResponse = resp
	return webhooksResponse, nil
}

// CreateWebhook creates a new webhook.
//
// See https://developer.dnsimple.com/v2/webhooks#create
func (s *WebhooksService) CreateWebhook(accountID string, webhookAttributes Webhook) (*webhookResponse, error) {
	path := versioned(webhookPath(accountID, 0))
	webhookResponse := &webhookResponse{}

	resp, err := s.client.post(path, webhookAttributes, webhookResponse)
	if err != nil {
		return nil, err
	}

	webhookResponse.HttpResponse = resp
	return webhookResponse, nil
}

// GetWebhook fetches a webhook.
//
// See https://developer.dnsimple.com/v2/webhooks#get
func (s *WebhooksService) GetWebhook(accountID string, webhookID int64) (*webhookResponse, error) {
	path := versioned(webhookPath(accountID, webhookID))
	webhookResponse := &webhookResponse{}

	resp, err := s.client.get(path, webhookResponse)
	if err != nil {
		return nil, err
	}

	webhookResponse.HttpResponse = resp
	return webhookResponse, nil
}

// DeleteWebhook PERMANENTLY deletes a webhook from the account.
//
// See https://developer.dnsimple.com/v2/webhooks#delete
func (s *WebhooksService) DeleteWebhook(accountID string, webhookID int64) (*webhookResponse, error) {
	path := versioned(webhookPath(accountID, webhookID))
	webhookResponse := &webhookResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	webhookResponse.HttpResponse = resp
	return webhookResponse, nil
}
