package dnsimple

import (
	"fmt"
)

// Collaborator represents a Collaborator in DNSimple.
type Collaborator struct {
	ID         int64  `json:"id,omitempty"`
	DomainID   int64  `json:"domain_id,omitempty"`
	DomainName string `json:"domain_name,omitempty"`
	UserID     int64  `json:"user_id,omitempty"`
	UserEmail  string `json:"user_email,omitempty"`
	Invitation bool   `json:"invitation,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	AcceptedAt string `json:"accepted_at,omitempty"`
}

func collaboratorPath(accountID, domainIdentifier string, collaboratorID int64) (path string) {
	path = fmt.Sprintf("%v/collaborators", domainPath(accountID, domainIdentifier))
	if collaboratorID != 0 {
		path += fmt.Sprintf("/%v", collaboratorID)
	}
	return
}

// CollaboratorAttributes represents Collaborator attributes for AddCollaborator operation.
type CollaboratorAttributes struct {
	Email string `json:"email,omitempty"`
}

// collaboratorResponse represents a response from an API method that returns a Collaborator struct.
type collaboratorResponse struct {
	Response
	Data *Collaborator `json:"data"`
}

// collaboratorsResponse represents a response from an API method that returns a collection of Collaborator struct.
type collaboratorsResponse struct {
	Response
	Data []Collaborator `json:"data"`
}

// ListCollaborators list the collaborators for a domain.
//
// See https://developer.dnsimple.com/v2/domains/collaborators#list
func (s *DomainsService) ListCollaborators(accountID, domainIdentifier string, options *ListOptions) (*collaboratorsResponse, error) {
	path := versioned(collaboratorPath(accountID, domainIdentifier, 0))
	collaboratorsResponse := &collaboratorsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, collaboratorsResponse)
	if err != nil {
		return collaboratorsResponse, err
	}

	collaboratorsResponse.HttpResponse = resp
	return collaboratorsResponse, nil
}

// AddCollaborator adds a new collaborator to the domain in the account.
//
// See https://developer.dnsimple.com/v2/domains/collaborators#add
func (s *DomainsService) AddCollaborator(accountID string, domainIdentifier string, attributes CollaboratorAttributes) (*collaboratorResponse, error) {
	path := versioned(collaboratorPath(accountID, domainIdentifier, 0))
	collaboratorResponse := &collaboratorResponse{}

	resp, err := s.client.post(path, attributes, collaboratorResponse)
	if err != nil {
		return nil, err
	}

	collaboratorResponse.HttpResponse = resp
	return collaboratorResponse, nil
}

// RemoveCollaborator PERMANENTLY deletes a domain from the account.
//
// See https://developer.dnsimple.com/v2/domains/collaborators#remove
func (s *DomainsService) RemoveCollaborator(accountID string, domainIdentifier string, collaboratorID int64) (*collaboratorResponse, error) {
	path := versioned(collaboratorPath(accountID, domainIdentifier, collaboratorID))
	collaboratorResponse := &collaboratorResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	collaboratorResponse.HttpResponse = resp
	return collaboratorResponse, nil
}
