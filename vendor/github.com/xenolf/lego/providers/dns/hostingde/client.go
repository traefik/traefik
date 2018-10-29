package hostingde

// RecordsAddRequest represents a DNS record to add
type RecordsAddRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// RecordsDeleteRequest represents a DNS record to remove
type RecordsDeleteRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
}

// ZoneConfigObject represents the ZoneConfig-section of a hosting.de API response.
type ZoneConfigObject struct {
	AccountID      string `json:"accountId"`
	EmailAddress   string `json:"emailAddress"`
	ID             string `json:"id"`
	LastChangeDate string `json:"lastChangeDate"`
	MasterIP       string `json:"masterIp"`
	Name           string `json:"name"`
	NameUnicode    string `json:"nameUnicode"`
	SOAValues      struct {
		Expire      int    `json:"expire"`
		NegativeTTL int    `json:"negativeTtl"`
		Refresh     int    `json:"refresh"`
		Retry       int    `json:"retry"`
		Serial      string `json:"serial"`
		TTL         int    `json:"ttl"`
	} `json:"soaValues"`
	Status                string   `json:"status"`
	TemplateValues        string   `json:"templateValues"`
	Type                  string   `json:"type"`
	ZoneTransferWhitelist []string `json:"zoneTransferWhitelist"`
}

// ZoneUpdateError represents an error in a ZoneUpdateResponse
type ZoneUpdateError struct {
	Code          int      `json:"code"`
	ContextObject string   `json:"contextObject"`
	ContextPath   string   `json:"contextPath"`
	Details       []string `json:"details"`
	Text          string   `json:"text"`
	Value         string   `json:"value"`
}

// ZoneUpdateMetadata represents the metadata in a ZoneUpdateResponse
type ZoneUpdateMetadata struct {
	ClientTransactionID string `json:"clientTransactionId"`
	ServerTransactionID string `json:"serverTransactionId"`
}

// ZoneUpdateResponse represents a response from hosting.de API
type ZoneUpdateResponse struct {
	Errors   []ZoneUpdateError  `json:"errors"`
	Metadata ZoneUpdateMetadata `json:"metadata"`
	Warnings []string           `json:"warnings"`
	Status   string             `json:"status"`
	Response struct {
		Records []struct {
			Content          string `json:"content"`
			Type             string `json:"type"`
			ID               string `json:"id"`
			Name             string `json:"name"`
			LastChangeDate   string `json:"lastChangeDate"`
			Priority         int    `json:"priority"`
			RecordTemplateID string `json:"recordTemplateId"`
			ZoneConfigID     string `json:"zoneConfigId"`
			TTL              int    `json:"ttl"`
		} `json:"records"`
		ZoneConfig ZoneConfigObject `json:"zoneConfig"`
	} `json:"response"`
}

// ZoneConfigSelector represents a "minimal" ZoneConfig object used in hosting.de API requests
type ZoneConfigSelector struct {
	Name string `json:"name"`
}

// ZoneUpdateRequest represents a hosting.de API ZoneUpdate request
type ZoneUpdateRequest struct {
	AuthToken          string `json:"authToken"`
	ZoneConfigSelector `json:"zoneConfig"`
	RecordsToAdd       []RecordsAddRequest    `json:"recordsToAdd"`
	RecordsToDelete    []RecordsDeleteRequest `json:"recordsToDelete"`
}
