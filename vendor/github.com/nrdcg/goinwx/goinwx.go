package goinwx

import (
	"net/url"

	"github.com/kolo/xmlrpc"
)

// API information.
const (
	APIBaseURL        = "https://api.domrobot.com/xmlrpc/"
	APISandboxBaseURL = "https://api.ote.domrobot.com/xmlrpc/"
	APILanguage       = "eng"
)

// Client manages communication with INWX API.
type Client struct {
	// HTTP client used to communicate with the INWX API.
	RPCClient *xmlrpc.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// API username and password
	username string
	password string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for communicating with the API
	Account     *AccountService
	Domains     *DomainService
	Nameservers *NameserverService
	Contacts    *ContactService
}

type service struct {
	client *Client
}

// ClientOptions Options of the API client.
type ClientOptions struct {
	Sandbox bool
}

// Request The representation of an API request.
type Request struct {
	ServiceMethod string
	Args          map[string]interface{}
}

// NewClient returns a new INWX API client.
func NewClient(username, password string, opts *ClientOptions) *Client {
	var useSandbox bool
	if opts != nil {
		useSandbox = opts.Sandbox
	}

	var baseURL *url.URL

	if useSandbox {
		baseURL, _ = url.Parse(APISandboxBaseURL)
	} else {
		baseURL, _ = url.Parse(APIBaseURL)
	}

	rpcClient, _ := xmlrpc.NewClient(baseURL.String(), nil)

	client := &Client{
		RPCClient: rpcClient,
		BaseURL:   baseURL,
		username:  username,
		password:  password,
	}

	client.common.client = client
	client.Account = (*AccountService)(&client.common)
	client.Domains = (*DomainService)(&client.common)
	client.Nameservers = (*NameserverService)(&client.common)
	client.Contacts = (*ContactService)(&client.common)

	return client
}

// NewRequest creates an API request.
func (c *Client) NewRequest(serviceMethod string, args map[string]interface{}) *Request {
	if args != nil {
		args["lang"] = APILanguage
	}

	return &Request{ServiceMethod: serviceMethod, Args: args}
}

// Do sends an API request and returns the API response.
func (c *Client) Do(req Request) (*map[string]interface{}, error) {
	var resp Response
	err := c.RPCClient.Call(req.ServiceMethod, req.Args, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.ResponseData, checkResponse(&resp)
}

// checkResponse checks the API response for errors, and returns them if present.
func checkResponse(r *Response) error {
	if c := r.Code; c >= 1000 && c <= 1500 {
		return nil
	}

	return &ErrorResponse{Code: r.Code, Message: r.Message, Reason: r.Reason, ReasonCode: r.ReasonCode}
}
