package records

// CreateRecordRequest describes the json payload for creating a record
type CreateRecordRequest struct {
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
}

// CreateRecordResponse describes the json response for creating a record
type CreateRecordResponse struct {
	ID         string `json:"id"`
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
}

// GetRecordsResponse describes the json response of a single record
type GetRecordsResponse struct {
	ID         string `json:"id"`
	RecordType string `json:"type"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	TTL        int    `json:"ttl"`
}

// RemoveRecordResponse describes the json response for removing a record
type RemoveRecordResponse struct {
}
