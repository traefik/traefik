package auroradnsclient

// AuroraDNSError describes the format of a generic AuroraDNS API error
type AuroraDNSError struct {
	ErrorCode string `json:"error"`
	Message   string `json:"errormsg"`
}

func (e AuroraDNSError) Error() string {
	return e.Message
}
