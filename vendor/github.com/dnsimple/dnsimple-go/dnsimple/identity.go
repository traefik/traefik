package dnsimple

// IdentityService handles communication with several authentication identity
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/identity/
type IdentityService struct {
	client *Client
}

// WhoamiData represents an authenticated context
// that contains information about the current logged User and/or Account.
type WhoamiData struct {
	User    *User    `json:"user,omitempty"`
	Account *Account `json:"account,omitempty"`
}

// whoamiResponse represents a response from an API method that returns a Whoami struct.
type whoamiResponse struct {
	Response
	Data *WhoamiData `json:"data"`
}

// Whoami gets the current authenticate context.
//
// See https://developer.dnsimple.com/v2/whoami
func (s *IdentityService) Whoami() (*whoamiResponse, error) {
	path := versioned("/whoami")
	whoamiResponse := &whoamiResponse{}

	resp, err := s.client.get(path, whoamiResponse)
	if err != nil {
		return nil, err
	}

	whoamiResponse.HttpResponse = resp
	return whoamiResponse, nil
}

// Whoami is a state-less shortcut to client.Whoami()
// that returns only the relevant Data.
func Whoami(c *Client) (data *WhoamiData, err error) {
	resp, err := c.Identity.Whoami()
	if resp != nil {
		data = resp.Data
	}
	return
}
