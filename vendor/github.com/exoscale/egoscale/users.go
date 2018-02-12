package egoscale

// User represents a User
type User struct {
	Account             string `json:"account,omitempty"`
	AccountID           string `json:"accountid,omitempty"`
	AccountType         string `json:"accounttype,omitempty"`
	APIKey              string `json:"apikey,omitempty"`
	Created             string `json:"created,omitempty"`
	Domain              string `json:"domain,omitempty"`
	DomainID            string `json:"domainid,omitempty"`
	Email               string `json:"email,omitempty"`
	FirstName           string `json:"firstname,omitempty"`
	ID                  string `json:"id,omitempty"`
	IsCallerChildDomain bool   `json:"iscallerchilddomain,omitempty"`
	IsDefault           bool   `json:"isdefault,omitempty"`
	LastName            string `json:"lastname,omitempty"`
	SecretKey           string `json:"secretkey,omitempty"`
	State               string `json:"state,omitempty"`
	UserName            string `json:"username,omitempty"`
}

// RegisterUserKeys registers a new set of key of the given user
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerUserKeys.html
type RegisterUserKeys struct {
	ID string `json:"id"`
}

func (*RegisterUserKeys) name() string {
	return "registerUserKeys"
}

func (*RegisterUserKeys) response() interface{} {
	return new(RegisterUserKeysResponse)
}

// RegisterUserKeysResponse represents a new set of UserKeys
//
// NB: only the APIKey and SecretKey will be filled
type RegisterUserKeysResponse struct {
	UserKeys User `json:"userkeys"`
}
