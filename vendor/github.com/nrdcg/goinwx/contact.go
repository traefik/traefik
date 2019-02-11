package goinwx

import (
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

const (
	methodContactInfo   = "contact.info"
	methodContactList   = "contact.list"
	methodContactCreate = "contact.create"
	methodContactDelete = "contact.delete"
	methodContactUpdate = "contact.update"
)

// ContactService API access to Contact.
type ContactService service

// Create Creates a contact.
func (s *ContactService) Create(request *ContactCreateRequest) (int, error) {
	req := s.client.NewRequest(methodContactCreate, structs.Map(request))

	resp, err := s.client.Do(*req)
	if err != nil {
		return 0, err
	}

	var result map[string]int
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return 0, err
	}

	return result["id"], nil
}

// Delete Deletes a contact.
func (s *ContactService) Delete(roID int) error {
	req := s.client.NewRequest(methodContactDelete, map[string]interface{}{
		"id": roID,
	})

	_, err := s.client.Do(*req)
	return err
}

// Update Updates a contact.
func (s *ContactService) Update(request *ContactUpdateRequest) error {
	req := s.client.NewRequest(methodContactUpdate, structs.Map(request))

	_, err := s.client.Do(*req)
	return err
}

// Info Get information about a contact.
func (s *ContactService) Info(contactID int) (*ContactInfoResponse, error) {
	var requestMap = make(map[string]interface{})
	requestMap["wide"] = 1

	if contactID != 0 {
		requestMap["id"] = contactID
	}

	req := s.client.NewRequest(methodContactInfo, requestMap)

	resp, err := s.client.Do(*req)
	if err != nil {
		return nil, err
	}

	var result ContactInfoResponse
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// List Search contacts.
func (s *ContactService) List(search string) (*ContactListResponse, error) {
	var requestMap = make(map[string]interface{})

	if search != "" {
		requestMap["search"] = search
	}
	req := s.client.NewRequest(methodContactList, requestMap)

	resp, err := s.client.Do(*req)
	if err != nil {
		return nil, err
	}

	var result ContactListResponse
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ContactCreateRequest API model.
type ContactCreateRequest struct {
	Type          string `structs:"type"`
	Name          string `structs:"name"`
	Org           string `structs:"org,omitempty"`
	Street        string `structs:"street"`
	City          string `structs:"city"`
	PostalCode    string `structs:"pc"`
	StateProvince string `structs:"sp,omitempty"`
	CountryCode   string `structs:"cc"`
	Voice         string `structs:"voice"`
	Fax           string `structs:"fax,omitempty"`
	Email         string `structs:"email"`
	Remarks       string `structs:"remarks,omitempty"`
	Protection    bool   `structs:"protection,omitempty"`
	Testing       bool   `structs:"testing,omitempty"`
}

// ContactUpdateRequest API model.
type ContactUpdateRequest struct {
	ID            int    `structs:"id"`
	Name          string `structs:"name,omitempty"`
	Org           string `structs:"org,omitempty"`
	Street        string `structs:"street,omitempty"`
	City          string `structs:"city,omitempty"`
	PostalCode    string `structs:"pc,omitempty"`
	StateProvince string `structs:"sp,omitempty"`
	CountryCode   string `structs:"cc,omitempty"`
	Voice         string `structs:"voice,omitempty"`
	Fax           string `structs:"fax,omitempty"`
	Email         string `structs:"email,omitempty"`
	Remarks       string `structs:"remarks,omitempty"`
	Protection    bool   `structs:"protection,omitempty"`
	Testing       bool   `structs:"testing,omitempty"`
}

// ContactInfoResponse API model.
type ContactInfoResponse struct {
	Contact Contact `mapstructure:"contact"`
}

// ContactListResponse API model.
type ContactListResponse struct {
	Count    int
	Contacts []Contact `mapstructure:"contact"`
}
