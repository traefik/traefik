package goinwx

import (
	"errors"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

const (
	methodNameserverCheck        = "nameserver.check"
	methodNameserverCreate       = "nameserver.create"
	methodNameserverCreateRecord = "nameserver.createRecord"
	methodNameserverDelete       = "nameserver.delete"
	methodNameserverDeleteRecord = "nameserver.deleteRecord"
	methodNameserverInfo         = "nameserver.info"
	methodNameserverList         = "nameserver.list"
	methodNameserverUpdate       = "nameserver.update"
	methodNameserverUpdateRecord = "nameserver.updateRecord"
)

type NameserverService interface {
	Check(domain string, nameservers []string) (*NameserverCheckResponse, error)
	Create(*NameserverCreateRequest) (int, error)
	Info(*NameserverInfoRequest) (*NamserverInfoResponse, error)
	List(domain string) (*NamserverListResponse, error)
	CreateRecord(*NameserverRecordRequest) (int, error)
	UpdateRecord(recId int, request *NameserverRecordRequest) error
	DeleteRecord(recId int) error
	FindRecordById(recId int) (*NameserverRecord, *NameserverDomain, error)
}

type NameserverServiceOp struct {
	client *Client
}

var _ NameserverService = &NameserverServiceOp{}

type NameserverCheckResponse struct {
	Details []string
	Status  string
}

type NameserverRecordRequest struct {
	RoId                   int    `structs:"roId,omitempty"`
	Domain                 string `structs:"domain,omitempty"`
	Type                   string `structs:"type"`
	Content                string `structs:"content"`
	Name                   string `structs:"name,omitempty"`
	Ttl                    int    `structs:"ttl,omitempty"`
	Priority               int    `structs:"prio,omitempty"`
	UrlRedirectType        string `structs:"urlRedirectType,omitempty"`
	UrlRedirectTitle       string `structs:"urlRedirectTitle,omitempty"`
	UrlRedirectDescription string `structs:"urlRedirectDescription,omitempty"`
	UrlRedirectFavIcon     string `structs:"urlRedirectFavIcon,omitempty"`
	UrlRedirectKeywords    string `structs:"urlRedirectKeywords,omitempty"`
}

type NameserverCreateRequest struct {
	Domain                 string   `structs:"domain"`
	Type                   string   `structs:"type"`
	Nameservers            []string `structs:"ns,omitempty"`
	MasterIp               string   `structs:"masterIp,omitempty"`
	Web                    string   `structs:"web,omitempty"`
	Mail                   string   `structs:"mail,omitempty"`
	SoaEmail               string   `structs:"soaEmail,omitempty"`
	UrlRedirectType        string   `structs:"urlRedirectType,omitempty"`
	UrlRedirectTitle       string   `structs:"urlRedirectTitle,omitempty"`
	UrlRedirectDescription string   `structs:"urlRedirectDescription,omitempty"`
	UrlRedirectFavIcon     string   `structs:"urlRedirectFavIcon,omitempty"`
	UrlRedirectKeywords    string   `structs:"urlRedirectKeywords,omitempty"`
	Testing                bool     `structs:"testing,omitempty"`
}

type NameserverInfoRequest struct {
	Domain   string `structs:"domain,omitempty"`
	RoId     int    `structs:"roId,omitempty"`
	RecordId int    `structs:"recordId,omitempty"`
	Type     string `structs:"type,omitempty"`
	Name     string `structs:"name,omitempty"`
	Content  string `structs:"content,omitempty"`
	Ttl      int    `structs:"ttl,omitempty"`
	Prio     int    `structs:"prio,omitempty"`
}

type NamserverInfoResponse struct {
	RoId          int
	Domain        string
	Type          string
	MasterIp      string
	LastZoneCheck time.Time
	SlaveDns      interface{}
	SOAserial     string
	Count         int
	Records       []NameserverRecord `mapstructure:"record"`
}

type NameserverRecord struct {
	Id                     int
	Name                   string
	Type                   string
	Content                string
	Ttl                    int
	Prio                   int
	UrlRedirectType        string
	UrlRedirectTitle       string
	UrlRedirectDescription string
	UrlRedirectKeywords    string
	UrlRedirectFavIcon     string
}

type NamserverListResponse struct {
	Count   int
	Domains []NameserverDomain `mapstructure:"domains"`
}

type NameserverDomain struct {
	RoId     int    `mapstructure:"roId"`
	Domain   string `mapstructure:"domain"`
	Type     string `mapstructure:"type"`
	MasterIp string `mapstructure:"masterIp"`
	Mail     string `mapstructure:"mail"`
	Web      string `mapstructure:"web"`
	Url      string `mapstructure:"url"`
	Ipv4     string `mapstructure:"ipv4"`
	Ipv6     string `mapstructure:"ipv6"`
}

func (s *NameserverServiceOp) Check(domain string, nameservers []string) (*NameserverCheckResponse, error) {
	req := s.client.NewRequest(methodNameserverCheck, map[string]interface{}{
		"domain": domain,
		"ns":     nameservers,
	})

	resp, err := s.client.Do(*req)
	if err != nil {
		return nil, err
	}

	var result NameserverCheckResponse
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *NameserverServiceOp) Info(request *NameserverInfoRequest) (*NamserverInfoResponse, error) {
	req := s.client.NewRequest(methodNameserverInfo, structs.Map(request))

	resp, err := s.client.Do(*req)
	if err != nil {
		return nil, err
	}
	var result NamserverInfoResponse
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *NameserverServiceOp) List(domain string) (*NamserverListResponse, error) {
	requestMap := map[string]interface{}{
		"domain": "*",
		"wide":   2,
	}
	if domain != "" {
		requestMap["domain"] = domain
	}
	req := s.client.NewRequest(methodNameserverList, requestMap)

	resp, err := s.client.Do(*req)
	if err != nil {
		return nil, err
	}
	var result NamserverListResponse
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *NameserverServiceOp) Create(request *NameserverCreateRequest) (int, error) {
	req := s.client.NewRequest(methodNameserverCreate, structs.Map(request))

	resp, err := s.client.Do(*req)
	if err != nil {
		return 0, err
	}

	var result map[string]int
	err = mapstructure.Decode(*resp, &result)
	if err != nil {
		return 0, err
	}

	return result["roId"], nil
}

func (s *NameserverServiceOp) CreateRecord(request *NameserverRecordRequest) (int, error) {
	req := s.client.NewRequest(methodNameserverCreateRecord, structs.Map(request))

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

func (s *NameserverServiceOp) UpdateRecord(recId int, request *NameserverRecordRequest) error {
	if request == nil {
		return errors.New("Request can't be nil")
	}
	requestMap := structs.Map(request)
	requestMap["id"] = recId

	req := s.client.NewRequest(methodNameserverUpdateRecord, requestMap)

	_, err := s.client.Do(*req)
	if err != nil {
		return err
	}

	return nil
}

func (s *NameserverServiceOp) DeleteRecord(recId int) error {
	req := s.client.NewRequest(methodNameserverDeleteRecord, map[string]interface{}{
		"id": recId,
	})

	_, err := s.client.Do(*req)
	if err != nil {
		return err
	}

	return nil
}

func (s *NameserverServiceOp) FindRecordById(recId int) (*NameserverRecord, *NameserverDomain, error) {
	listResp, err := s.client.Nameservers.List("")
	if err != nil {
		return nil, nil, err
	}

	for _, domainItem := range listResp.Domains {
		resp, err := s.client.Nameservers.Info(&NameserverInfoRequest{RoId: domainItem.RoId})
		if err != nil {
			return nil, nil, err
		}

		for _, record := range resp.Records {
			if record.Id == recId {
				return &record, &domainItem, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("couldn't find INWX Record for id %d", recId)

}
