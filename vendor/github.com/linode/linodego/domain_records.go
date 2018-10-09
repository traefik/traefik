package linodego

import (
	"context"
	"encoding/json"
	"fmt"
)

// DomainRecord represents a DomainRecord object
type DomainRecord struct {
	ID       int              `json:"id"`
	Type     DomainRecordType `json:"type"`
	Name     string           `json:"name"`
	Target   string           `json:"target"`
	Priority int              `json:"priority"`
	Weight   int              `json:"weight"`
	Port     int              `json:"port"`
	Service  *string          `json:"service"`
	Protocol *string          `json:"protocol"`
	TTLSec   int              `json:"ttl_sec"`
	Tag      *string          `json:"tag"`
}

// DomainRecordCreateOptions fields are those accepted by CreateDomainRecord
type DomainRecordCreateOptions struct {
	Type     DomainRecordType `json:"type"`
	Name     string           `json:"name"`
	Target   string           `json:"target"`
	Priority *int             `json:"priority,omitempty"`
	Weight   *int             `json:"weight,omitempty"`
	Port     *int             `json:"port,omitempty"`
	Service  *string          `json:"service,omitempty"`
	Protocol *string          `json:"protocol,omitempty"`
	TTLSec   int              `json:"ttl_sec,omitempty"` // 0 is not accepted by Linode, so can be omitted
	Tag      *string          `json:"tag,omitempty"`
}

// DomainRecordUpdateOptions fields are those accepted by UpdateDomainRecord
type DomainRecordUpdateOptions struct {
	Type     DomainRecordType `json:"type,omitempty"`
	Name     string           `json:"name,omitempty"`
	Target   string           `json:"target,omitempty"`
	Priority *int             `json:"priority,omitempty"` // 0 is valid, so omit only nil values
	Weight   *int             `json:"weight,omitempty"`   // 0 is valid, so omit only nil values
	Port     *int             `json:"port,omitempty"`     // 0 is valid to spec, so omit only nil values
	Service  *string          `json:"service,omitempty"`
	Protocol *string          `json:"protocol,omitempty"`
	TTLSec   int              `json:"ttl_sec,omitempty"` // 0 is not accepted by Linode, so can be omitted
	Tag      *string          `json:"tag,omitempty"`
}

// DomainRecordType constants start with RecordType and include Linode API Domain Record Types
type DomainRecordType string

// DomainRecordType contants are the DNS record types a DomainRecord can assign
const (
	RecordTypeA     DomainRecordType = "A"
	RecordTypeAAAA  DomainRecordType = "AAAA"
	RecordTypeNS    DomainRecordType = "NS"
	RecordTypeMX    DomainRecordType = "MX"
	RecordTypeCNAME DomainRecordType = "CNAME"
	RecordTypeTXT   DomainRecordType = "TXT"
	RecordTypeSRV   DomainRecordType = "SRV"
	RecordTypePTR   DomainRecordType = "PTR"
	RecordTypeCAA   DomainRecordType = "CAA"
)

// GetUpdateOptions converts a DomainRecord to DomainRecordUpdateOptions for use in UpdateDomainRecord
func (d DomainRecord) GetUpdateOptions() (du DomainRecordUpdateOptions) {
	du.Type = d.Type
	du.Name = d.Name
	du.Target = d.Target
	du.Priority = copyInt(&d.Priority)
	du.Weight = copyInt(&d.Weight)
	du.Port = copyInt(&d.Port)
	du.Service = copyString(d.Service)
	du.Protocol = copyString(d.Protocol)
	du.TTLSec = d.TTLSec
	du.Tag = copyString(d.Tag)
	return
}

func copyInt(iPtr *int) *int {
	if iPtr == nil {
		return nil
	}
	var t = *iPtr
	return &t
}

func copyString(sPtr *string) *string {
	if sPtr == nil {
		return nil
	}
	var t = *sPtr
	return &t
}

// DomainRecordsPagedResponse represents a paginated DomainRecord API response
type DomainRecordsPagedResponse struct {
	*PageOptions
	Data []DomainRecord `json:"data"`
}

// endpoint gets the endpoint URL for InstanceConfig
func (DomainRecordsPagedResponse) endpointWithID(c *Client, id int) string {
	endpoint, err := c.DomainRecords.endpointWithID(id)
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends DomainRecords when processing paginated DomainRecord responses
func (resp *DomainRecordsPagedResponse) appendData(r *DomainRecordsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListDomainRecords lists DomainRecords
func (c *Client) ListDomainRecords(ctx context.Context, domainID int, opts *ListOptions) ([]DomainRecord, error) {
	response := DomainRecordsPagedResponse{}
	err := c.listHelperWithID(ctx, &response, domainID, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (d *DomainRecord) fixDates() *DomainRecord {
	return d
}

// GetDomainRecord gets the domainrecord with the provided ID
func (c *Client) GetDomainRecord(ctx context.Context, domainID int, id int) (*DomainRecord, error) {
	e, err := c.DomainRecords.endpointWithID(domainID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&DomainRecord{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*DomainRecord), nil
}

// CreateDomainRecord creates a DomainRecord
func (c *Client) CreateDomainRecord(ctx context.Context, domainID int, domainrecord DomainRecordCreateOptions) (*DomainRecord, error) {
	var body string
	e, err := c.DomainRecords.endpointWithID(domainID)
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&DomainRecord{})

	bodyData, err := json.Marshal(domainrecord)
	if err != nil {
		return nil, NewError(err)
	}
	body = string(bodyData)

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*DomainRecord).fixDates(), nil
}

// UpdateDomainRecord updates the DomainRecord with the specified id
func (c *Client) UpdateDomainRecord(ctx context.Context, domainID int, id int, domainrecord DomainRecordUpdateOptions) (*DomainRecord, error) {
	var body string
	e, err := c.DomainRecords.endpointWithID(domainID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&DomainRecord{})

	if bodyData, err := json.Marshal(domainrecord); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*DomainRecord).fixDates(), nil
}

// DeleteDomainRecord deletes the DomainRecord with the specified id
func (c *Client) DeleteDomainRecord(ctx context.Context, domainID int, id int) error {
	e, err := c.DomainRecords.endpointWithID(domainID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
