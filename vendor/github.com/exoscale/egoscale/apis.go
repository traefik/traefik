package egoscale

// API represents an API service
type API struct {
	Description string        `json:"description"`
	IsAsync     bool          `json:"isasync"`
	Name        string        `json:"name"`
	Related     string        `json:"related"` // comma separated
	Since       string        `json:"since"`
	Type        string        `json:"type"`
	Params      []APIParam    `json:"params"`
	Response    []APIResponse `json:"responses"`
}

// APIParam represents an API parameter field
type APIParam struct {
	Description string `json:"description"`
	Length      int64  `json:"length"`
	Name        string `json:"name"`
	Related     string `json:"related"` // comma separated
	Since       string `json:"since"`
	Type        string `json:"type"`
}

// APIResponse represents an API response field
type APIResponse struct {
	Description string        `json:"description"`
	Name        string        `json:"name"`
	Response    []APIResponse `json:"response"`
	Type        string        `json:"type"`
}

// ListAPIs represents a query to list the api
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listApis.html
type ListAPIs struct {
	Name string `json:"name,omitempty"`
}

func (*ListAPIs) name() string {
	return "listApis"
}

func (*ListAPIs) response() interface{} {
	return new(ListAPIsResponse)
}

// ListAPIsResponse represents a list of API
type ListAPIsResponse struct {
	Count int   `json:"count"`
	API   []API `json:"api"`
}
