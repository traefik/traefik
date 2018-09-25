package gandiv5

// types for JSON method calls and parameters

type addFieldRequest struct {
	RRSetTTL    int      `json:"rrset_ttl"`
	RRSetValues []string `json:"rrset_values"`
}

type deleteFieldRequest struct {
	Delete bool `json:"delete"`
}

// types for JSON responses

type responseStruct struct {
	Message string `json:"message"`
}
