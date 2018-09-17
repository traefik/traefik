package glesys

// types for JSON method calls, parameters, and responses

type addRecordRequest struct {
	DomainName string `json:"domainname"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	TTL        int    `json:"ttl,omitempty"`
}

type deleteRecordRequest struct {
	RecordID int `json:"recordid"`
}

type responseStruct struct {
	Response struct {
		Status struct {
			Code int `json:"code"`
		} `json:"status"`
		Record deleteRecordRequest `json:"record"`
	} `json:"response"`
}
