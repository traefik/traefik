package hostingde

import "encoding/json"

// APIError represents an error in an API response.
// https://www.hosting.de/api/?json#warnings-and-errors
type APIError struct {
	Code          int      `json:"code"`
	ContextObject string   `json:"contextObject"`
	ContextPath   string   `json:"contextPath"`
	Details       []string `json:"details"`
	Text          string   `json:"text"`
	Value         string   `json:"value"`
}

// Filter is used to filter FindRequests to the API.
// https://www.hosting.de/api/?json#filter-object
type Filter struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Sort is used to sort FindRequests from the API.
// https://www.hosting.de/api/?json#filtering-and-sorting
type Sort struct {
	Field string `json:"zoneName"`
	Order string `json:"order"`
}

// Metadata represents the metadata in an API response.
// https://www.hosting.de/api/?json#metadata-object
type Metadata struct {
	ClientTransactionID string `json:"clientTransactionId"`
	ServerTransactionID string `json:"serverTransactionId"`
}

// ZoneConfig The ZoneConfig object defines a zone.
// https://www.hosting.de/api/?json#the-zoneconfig-object
type ZoneConfig struct {
	ID                    string          `json:"id"`
	AccountID             string          `json:"accountId"`
	Status                string          `json:"status"`
	Name                  string          `json:"name"`
	NameUnicode           string          `json:"nameUnicode"`
	MasterIP              string          `json:"masterIp"`
	Type                  string          `json:"type"`
	EMailAddress          string          `json:"emailAddress"`
	ZoneTransferWhitelist []string        `json:"zoneTransferWhitelist"`
	LastChangeDate        string          `json:"lastChangeDate"`
	DNSServerGroupID      string          `json:"dnsServerGroupId"`
	DNSSecMode            string          `json:"dnsSecMode"`
	SOAValues             *SOAValues      `json:"soaValues,omitempty"`
	TemplateValues        json.RawMessage `json:"templateValues,omitempty"`
}

// SOAValues The SOA values object contains the time (seconds) used in a zone’s SOA record.
// https://www.hosting.de/api/?json#the-soa-values-object
type SOAValues struct {
	Refresh     int `json:"refresh"`
	Retry       int `json:"retry"`
	Expire      int `json:"expire"`
	TTL         int `json:"ttl"`
	NegativeTTL int `json:"negativeTtl"`
}

// DNSRecord The DNS Record object is part of a zone. It is used to manage DNS resource records.
// https://www.hosting.de/api/?json#the-record-object
type DNSRecord struct {
	ID               string `json:"id,omitempty"`
	ZoneID           string `json:"zoneId,omitempty"`
	RecordTemplateID string `json:"recordTemplateId,omitempty"`
	Name             string `json:"name,omitempty"`
	Type             string `json:"type,omitempty"`
	Content          string `json:"content,omitempty"`
	TTL              int    `json:"ttl,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	LastChangeDate   string `json:"lastChangeDate,omitempty"`
}

// Zone The Zone Object.
// https://www.hosting.de/api/?json#the-zone-object
type Zone struct {
	Records    []DNSRecord `json:"records"`
	ZoneConfig ZoneConfig  `json:"zoneConfig"`
}

// ZoneUpdateRequest represents a API ZoneUpdate request.
// https://www.hosting.de/api/?json#updating-zones
type ZoneUpdateRequest struct {
	BaseRequest
	ZoneConfig      `json:"zoneConfig"`
	RecordsToAdd    []DNSRecord `json:"recordsToAdd"`
	RecordsToDelete []DNSRecord `json:"recordsToDelete"`
}

// ZoneUpdateResponse represents a response from the API.
// https://www.hosting.de/api/?json#updating-zones
type ZoneUpdateResponse struct {
	BaseResponse
	Response Zone `json:"response"`
}

// ZoneConfigsFindRequest represents a API ZonesFind request.
// https://www.hosting.de/api/?json#list-zoneconfigs
type ZoneConfigsFindRequest struct {
	BaseRequest
	Filter Filter `json:"filter"`
	Limit  int    `json:"limit"`
	Page   int    `json:"page"`
	Sort   *Sort  `json:"sort,omitempty"`
}

// ZoneConfigsFindResponse represents the API response for ZoneConfigsFind.
// https://www.hosting.de/api/?json#list-zoneconfigs
type ZoneConfigsFindResponse struct {
	BaseResponse
	Response struct {
		Limit        int          `json:"limit"`
		Page         int          `json:"page"`
		TotalEntries int          `json:"totalEntries"`
		TotalPages   int          `json:"totalPages"`
		Type         string       `json:"type"`
		Data         []ZoneConfig `json:"data"`
	} `json:"response"`
}

// BaseResponse Common response struct.
// https://www.hosting.de/api/?json#responses
type BaseResponse struct {
	Errors   []APIError `json:"errors"`
	Metadata Metadata   `json:"metadata"`
	Warnings []string   `json:"warnings"`
	Status   string     `json:"status"`
}

// BaseRequest Common request struct.
type BaseRequest struct {
	AuthToken string `json:"authToken"`
}
