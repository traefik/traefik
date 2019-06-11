package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labbsr0x/bindman-dns-webhook/src/types"
	"github.com/labbsr0x/goh/gohclient"
	"net/http"
	"strings"
)

const recordsPath = "/records"

// DNSWebhookClient defines the basic structure of a DNS Listener
type DNSWebhookClient struct {
	ClientAPI gohclient.API
}

// New builds the client to communicate with the dns manager
func New(managerAddress string, httpClient *http.Client) (*DNSWebhookClient, error) {
	if strings.TrimSpace(managerAddress) == "" {
		return nil, errors.New("managerAddress parameter must be a non-empty string")
	}
	client, err := gohclient.New(httpClient, managerAddress)
	if err != nil {
		return nil, err
	}
	client.Accept = "application/json"
	client.ContentType = "application/json"
	client.UserAgent = "bindman-dns-webhook-client"

	return &DNSWebhookClient{
		ClientAPI: client,
	}, nil
}

// GetRecords communicates with the dns manager and gets the DNS Records
func (l *DNSWebhookClient) GetRecords() (result []types.DNSRecord, err error) {
	resp, data, err := l.ClientAPI.Get(recordsPath)
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(data, &result)
	} else {
		err = parseResponseBodyToError(data)
	}
	return
}

// GetRecord communicates with the dns manager and gets a DNS Record
func (l *DNSWebhookClient) GetRecord(name, recordType string) (result types.DNSRecord, err error) {
	resp, data, err := l.ClientAPI.Get(fmt.Sprintf(recordsPath+"/%s/%s", name, recordType))
	if err != nil {
		return
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(data, &result)
	} else {
		err = parseResponseBodyToError(data)
	}
	return
}

// AddRecord adds a DNS record
func (l *DNSWebhookClient) AddRecord(name string, recordType string, value string) error {
	return l.addOrUpdateRecord(&types.DNSRecord{Value: value, Name: name, Type: recordType}, l.ClientAPI.Post)
}

// UpdateRecord is a function that calls the defined webhook to update a specific dns record
func (l *DNSWebhookClient) UpdateRecord(record *types.DNSRecord) error {
	return l.addOrUpdateRecord(record, l.ClientAPI.Put)
}

// addOrUpdateRecord .
func (l *DNSWebhookClient) addOrUpdateRecord(record *types.DNSRecord, action func(url string, body []byte) (*http.Response, []byte, error)) error {
	if errs := record.Check(); errs != nil {
		return fmt.Errorf("invalid DNS Record: %v", strings.Join(errs, ", "))
	}
	mr, err := json.Marshal(record)
	if err != nil {
		return err
	}
	resp, data, err := action(recordsPath, mr)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return parseResponseBodyToError(data)
	}
	return nil
}

// RemoveRecord is a function that calls the defined webhook to remove a specific dns record
func (l *DNSWebhookClient) RemoveRecord(name, recordType string) error {
	resp, data, err := l.ClientAPI.Delete(fmt.Sprintf(recordsPath+"/%s/%s", name, recordType))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return parseResponseBodyToError(data)
	}
	return err
}

func parseResponseBodyToError(data []byte) error {
	var err types.Error
	if errUnmarshal := json.Unmarshal(data, &err); errUnmarshal != nil {
		return errUnmarshal
	}
	return &err
}
