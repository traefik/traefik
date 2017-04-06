package ovh

import (
	"fmt"
	"strings"
)

// Map user friendly access level names to corresponding HTTP verbs
var (
	ReadOnly      = []string{"GET"}
	ReadWrite     = []string{"GET", "POST", "PUT", "DELETE"}
	ReadWriteSafe = []string{"GET", "POST", "PUT"}
)

// AccessRule represents a method allowed for a path
type AccessRule struct {
	// Allowed HTTP Method for the requested AccessRule.
	// Can be set to GET/POST/PUT/DELETE.
	Method string `json:"method"`
	// Allowed path.
	// Can be an exact string or a string with '*' char.
	// Example :
	// 		/me : only /me is authorized
	//		/* : all calls are authorized
	Path string `json:"path"`
}

// CkValidationState represents the response when asking a new consumerKey.
type CkValidationState struct {
	// Consumer key, which need to be validated by customer.
	ConsumerKey string `json:"consumerKey"`
	// Current status, should be always "pendingValidation".
	State string `json:"state"`
	// URL to redirect user in order to log in.
	ValidationURL string `json:"validationUrl"`
}

// CkRequest represents the parameters to fill in order to ask a new
// consumerKey.
type CkRequest struct {
	client      *Client
	AccessRules []AccessRule `json:"accessRules"`
	Redirection string       `json:"redirection,omitempty"`
}

func (ck *CkValidationState) String() string {
	return fmt.Sprintf("CK: %q\nStatus: %q\nValidation URL: %q\n",
		ck.ConsumerKey,
		ck.State,
		ck.ValidationURL,
	)
}

// NewCkRequest helps create a new ck request
func (c *Client) NewCkRequest() *CkRequest {
	return &CkRequest{
		client:      c,
		AccessRules: []AccessRule{},
	}
}

// NewCkRequestWithRedirection helps create a new ck request with a redirect URL
func (c *Client) NewCkRequestWithRedirection(redirection string) *CkRequest {
	return &CkRequest{
		client:      c,
		AccessRules: []AccessRule{},
		Redirection: redirection,
	}
}

// AddRule adds a new rule to the ckRequest
func (ck *CkRequest) AddRule(method, path string) {
	ck.AccessRules = append(ck.AccessRules, AccessRule{
		Method: method,
		Path:   path,
	})
}

// AddRules adds grant requests on "path" for all methods. "ReadOnly",
// "ReadWrite" and "ReadWriteSafe" should be used for "methods" unless
// specific access are required.
func (ck *CkRequest) AddRules(methods []string, path string) {
	for _, method := range methods {
		ck.AddRule(method, path)
	}

}

// AddRecursiveRules adds grant requests on "path" and "path/*", for all
// methods "ReadOnly", "ReadWrite" and "ReadWriteSafe" should be used for
// "methods" unless specific access are required.
func (ck *CkRequest) AddRecursiveRules(methods []string, path string) {
	path = strings.TrimRight(path, "/")

	// Add rules. Skip base rule when requesting access to "/"
	if path != "" {
		ck.AddRules(methods, path)
	}
	ck.AddRules(methods, path+"/*")
}

// Do executes the request. On success, set the consumer key in the client
// and return the URL the user needs to visit to validate the key
func (ck *CkRequest) Do() (*CkValidationState, error) {
	state := CkValidationState{}
	err := ck.client.PostUnAuth("/auth/credential", ck, &state)

	if err == nil {
		ck.client.ConsumerKey = state.ConsumerKey
	}

	return &state, err
}
