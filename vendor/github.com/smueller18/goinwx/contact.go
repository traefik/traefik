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

type ContactService interface {
	Create(*ContactCreateRequest) (int, error)
	Update(*ContactUpdateRequest) error
	Delete(int) error
	Info(int) (*ContactInfoResponse, error)
	List(string) (*ContactListResponse, error)
}

type ContactServiceOp struct {
	client *Client
}

var _ ContactService = &ContactServiceOp{}

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

type ContactUpdateRequest struct {
	Id            int    `structs:"id"`
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

type ContactInfoResponse struct {
	Contact Contact `mapstructure:"contact"`
}

type ContactListResponse struct {
	Count    int
	Contacts []Contact `mapstructure:"contact"`
}

func (s *ContactServiceOp) Create(request *ContactCreateRequest) (int, error) {
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

func (s *ContactServiceOp) Delete(roId int) error {
	req := s.client.NewRequest(methodContactDelete, map[string]interface{}{
		"id": roId,
	})

	_, err := s.client.Do(*req)
	return err
}

func (s *ContactServiceOp) Update(request *ContactUpdateRequest) error {
	req := s.client.NewRequest(methodContactUpdate, structs.Map(request))

	_, err := s.client.Do(*req)
	return err
}

func (s *ContactServiceOp) Info(contactId int) (*ContactInfoResponse, error) {
	var requestMap = make(map[string]interface{})
	requestMap["wide"] = 1

	if contactId != 0 {
		requestMap["id"] = contactId
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

func (s *ContactServiceOp) List(search string) (*ContactListResponse, error) {
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
