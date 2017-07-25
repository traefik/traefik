// Package adexchangebuyer2 provides access to the Ad Exchange Buyer API II.
//
// See https://developers.google.com/ad-exchange/buyer-rest/guides/client-access/
//
// Usage example:
//
//   import "google.golang.org/api/adexchangebuyer2/v2beta1"
//   ...
//   adexchangebuyer2Service, err := adexchangebuyer2.New(oauthHttpClient)
package adexchangebuyer2 // import "google.golang.org/api/adexchangebuyer2/v2beta1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "adexchangebuyer2:v2beta1"
const apiName = "adexchangebuyer2"
const apiVersion = "v2beta1"
const basePath = "https://adexchangebuyer.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// Manage your Ad Exchange buyer account configuration
	AdexchangeBuyerScope = "https://www.googleapis.com/auth/adexchange.buyer"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Accounts = NewAccountsService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Accounts *AccountsService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewAccountsService(s *Service) *AccountsService {
	rs := &AccountsService{s: s}
	rs.Clients = NewAccountsClientsService(s)
	return rs
}

type AccountsService struct {
	s *Service

	Clients *AccountsClientsService
}

func NewAccountsClientsService(s *Service) *AccountsClientsService {
	rs := &AccountsClientsService{s: s}
	rs.Invitations = NewAccountsClientsInvitationsService(s)
	rs.Users = NewAccountsClientsUsersService(s)
	return rs
}

type AccountsClientsService struct {
	s *Service

	Invitations *AccountsClientsInvitationsService

	Users *AccountsClientsUsersService
}

func NewAccountsClientsInvitationsService(s *Service) *AccountsClientsInvitationsService {
	rs := &AccountsClientsInvitationsService{s: s}
	return rs
}

type AccountsClientsInvitationsService struct {
	s *Service
}

func NewAccountsClientsUsersService(s *Service) *AccountsClientsUsersService {
	rs := &AccountsClientsUsersService{s: s}
	return rs
}

type AccountsClientsUsersService struct {
	s *Service
}

// Client: A client resource represents a client buyer&mdash;an
// agency,
// a brand, or an advertiser customer of the sponsor buyer.
// Users associated with the client buyer have restricted access to
// the Ad Exchange Marketplace and certain other sections
// of the Ad Exchange Buyer UI based on the role
// granted to the client buyer.
// All fields are required unless otherwise specified.
type Client struct {
	// ClientAccountId: The globally-unique numerical ID of the client.
	// The value of this field is ignored in create and update operations.
	ClientAccountId int64 `json:"clientAccountId,omitempty,string"`

	// ClientName: Name used to represent this client to publishers.
	// You may have multiple clients that map to the same entity,
	// but for each client the combination of `clientName` and entity
	// must be unique.
	// You can specify this field as empty.
	ClientName string `json:"clientName,omitempty"`

	// EntityId: Numerical identifier of the client entity.
	// The entity can be an advertiser, a brand, or an agency.
	// This identifier is unique among all the entities with the same
	// type.
	//
	// A list of all known advertisers with their identifiers is available
	// in
	// the
	// [advertisers.txt](https://storage.googleapis.com/adx-rtb-dictionar
	// ies/advertisers.txt)
	// file.
	//
	// A list of all known brands with their identifiers is available in
	// the
	// [brands.txt](https://storage.googleapis.com/adx-rtb-dictionaries/b
	// rands.txt)
	// file.
	//
	// A list of all known agencies with their identifiers is available in
	// the
	// [agencies.txt](https://storage.googleapis.com/adx-rtb-dictionaries
	// /agencies.txt)
	// file.
	EntityId int64 `json:"entityId,omitempty,string"`

	// EntityName: The name of the entity. This field is automatically
	// fetched based on
	// the type and ID.
	// The value of this field is ignored in create and update operations.
	EntityName string `json:"entityName,omitempty"`

	// EntityType: The type of the client entity: `ADVERTISER`, `BRAND`, or
	// `AGENCY`.
	//
	// Possible values:
	//   "ENTITY_TYPE_UNSPECIFIED" - A placeholder for an undefined client
	// entity type.
	//   "ADVERTISER" - An advertiser.
	//   "BRAND" - A brand.
	//   "AGENCY" - An advertising agency.
	EntityType string `json:"entityType,omitempty"`

	// Role: The role which is assigned to the client buyer. Each role
	// implies a set of
	// permissions granted to the client. Must be one of
	// `CLIENT_DEAL_VIEWER`,
	// `CLIENT_DEAL_NEGOTIATOR`, or `CLIENT_DEAL_APPROVER`.
	//
	// Possible values:
	//   "CLIENT_ROLE_UNSPECIFIED" - A placeholder for an undefined client
	// role.
	//   "CLIENT_DEAL_VIEWER" - Users associated with this client can see
	// publisher deal offers
	// in the Marketplace.
	// They can neither negotiate proposals nor approve deals.
	// If this client is visible to publishers, they can send deal
	// proposals
	// to this client.
	//   "CLIENT_DEAL_NEGOTIATOR" - Users associated with this client can
	// respond to deal proposals
	// sent to them by publishers. They can also initiate deal proposals
	// of their own.
	//   "CLIENT_DEAL_APPROVER" - Users associated with this client can
	// approve eligible deals
	// on your behalf. Some deals may still explicitly require
	// publisher
	// finalization. If this role is not selected, the sponsor buyer
	// will need to manually approve each of their deals.
	Role string `json:"role,omitempty"`

	// Status: The status of the client buyer.
	//
	// Possible values:
	//   "CLIENT_STATUS_UNSPECIFIED" - A placeholder for an undefined client
	// status.
	//   "DISABLED" - A client that is currently disabled.
	//   "ACTIVE" - A client that is currently active.
	Status string `json:"status,omitempty"`

	// VisibleToSeller: Whether the client buyer will be visible to sellers.
	VisibleToSeller bool `json:"visibleToSeller,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ClientAccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Client) MarshalJSON() ([]byte, error) {
	type noMethod Client
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClientUser: A client user is created under a client buyer and has
// restricted access to
// the Ad Exchange Marketplace and certain other sections
// of the Ad Exchange Buyer UI based on the role
// granted to the associated client buyer.
//
// The only way a new client user can be created is via accepting
// an
// email invitation
// (see the
// accounts.clients.invitations.create
// method).
//
// All fields are required unless otherwise specified.
type ClientUser struct {
	// ClientAccountId: Numerical account ID of the client buyer
	// with which the user is associated; the
	// buyer must be a client of the current sponsor buyer.
	// The value of this field is ignored in an update operation.
	ClientAccountId int64 `json:"clientAccountId,omitempty,string"`

	// Email: User's email address. The value of this field
	// is ignored in an update operation.
	Email string `json:"email,omitempty"`

	// Status: The status of the client user.
	//
	// Possible values:
	//   "USER_STATUS_UNSPECIFIED" - A placeholder for an undefined user
	// status.
	//   "PENDING" - A user who was already created but hasn't accepted the
	// invitation yet.
	//   "ACTIVE" - A user that is currently active.
	//   "DISABLED" - A user that is currently disabled.
	Status string `json:"status,omitempty"`

	// UserId: The unique numerical ID of the client user
	// that has accepted an invitation.
	// The value of this field is ignored in an update operation.
	UserId int64 `json:"userId,omitempty,string"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ClientAccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClientUser) MarshalJSON() ([]byte, error) {
	type noMethod ClientUser
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClientUserInvitation: An invitation for a new client user to get
// access to the AdExchange Buyer UI.
//
// All fields are required unless otherwise specified.
type ClientUserInvitation struct {
	// ClientAccountId: Numerical account ID of the client buyer
	// that the invited user is associated with.
	// The value of this field is ignored in create operations.
	ClientAccountId int64 `json:"clientAccountId,omitempty,string"`

	// Email: The email address to which the invitation is sent.
	// Email
	// addresses should be unique among all client users under each
	// sponsor
	// buyer.
	Email string `json:"email,omitempty"`

	// InvitationId: The unique numerical ID of the invitation that is sent
	// to the user.
	// The value of this field is ignored in create operations.
	InvitationId int64 `json:"invitationId,omitempty,string"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ClientAccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClientUserInvitation) MarshalJSON() ([]byte, error) {
	type noMethod ClientUserInvitation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ListClientUserInvitationsResponse struct {
	// Invitations: The returned list of client users.
	Invitations []*ClientUserInvitation `json:"invitations,omitempty"`

	// NextPageToken: A token to retrieve the next page of results.
	// Pass this value in
	// the
	// ListClientUserInvitationsRequest.pageToken
	// field in the subsequent call to the
	// clients.invitations.list
	// method to retrieve the next
	// page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Invitations") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClientUserInvitationsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListClientUserInvitationsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ListClientUsersResponse struct {
	// NextPageToken: A token to retrieve the next page of results.
	// Pass this value in the
	// ListClientUsersRequest.pageToken
	// field in the subsequent call to the
	// clients.invitations.list
	// method to retrieve the next
	// page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Users: The returned list of client users.
	Users []*ClientUser `json:"users,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "NextPageToken") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClientUsersResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListClientUsersResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ListClientsResponse struct {
	// Clients: The returned list of clients.
	Clients []*Client `json:"clients,omitempty"`

	// NextPageToken: A token to retrieve the next page of results.
	// Pass this value in the
	// ListClientsRequest.pageToken
	// field in the subsequent call to the
	// accounts.clients.list method
	// to retrieve the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Clients") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClientsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListClientsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "adexchangebuyer2.accounts.clients.create":

type AccountsClientsCreateCall struct {
	s          *Service
	accountId  int64
	client     *Client
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create: Creates a new client buyer.
func (r *AccountsClientsService) Create(accountId int64, client *Client) *AccountsClientsCreateCall {
	c := &AccountsClientsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.client = client
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsCreateCall) Fields(s ...googleapi.Field) *AccountsClientsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsCreateCall) Context(ctx context.Context) *AccountsClientsCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.client)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.create" call.
// Exactly one of *Client or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Client.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsClientsCreateCall) Do(opts ...googleapi.CallOption) (*Client, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Client{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates a new client buyer.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer2.accounts.clients.create",
	//   "parameterOrder": [
	//     "accountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Unique numerical account ID for the buyer of which the client buyer\nis a customer; the sponsor buyer to create a client for. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients",
	//   "request": {
	//     "$ref": "Client"
	//   },
	//   "response": {
	//     "$ref": "Client"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.get":

type AccountsClientsGetCall struct {
	s               *Service
	accountId       int64
	clientAccountId int64
	urlParams_      gensupport.URLParams
	ifNoneMatch_    string
	ctx_            context.Context
}

// Get: Gets a client buyer with a given client account ID.
func (r *AccountsClientsService) Get(accountId int64, clientAccountId int64) *AccountsClientsGetCall {
	c := &AccountsClientsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsGetCall) Fields(s ...googleapi.Field) *AccountsClientsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsGetCall) IfNoneMatch(entityTag string) *AccountsClientsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsGetCall) Context(ctx context.Context) *AccountsClientsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.get" call.
// Exactly one of *Client or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Client.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsClientsGetCall) Do(opts ...googleapi.CallOption) (*Client, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Client{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Gets a client buyer with a given client account ID.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer to retrieve. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}",
	//   "response": {
	//     "$ref": "Client"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.list":

type AccountsClientsListCall struct {
	s            *Service
	accountId    int64
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists all the clients for the current sponsor buyer.
func (r *AccountsClientsService) List(accountId int64) *AccountsClientsListCall {
	c := &AccountsClientsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	return c
}

// PageSize sets the optional parameter "pageSize": Requested page size.
// The server may return fewer clients than requested.
// If unspecified, the server will pick an appropriate default.
func (c *AccountsClientsListCall) PageSize(pageSize int64) *AccountsClientsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": A token
// identifying a page of results the server should return.
// Typically, this is the value
// of
// ListClientsResponse.nextPageToken
// returned from the previous call to the
// accounts.clients.list method.
func (c *AccountsClientsListCall) PageToken(pageToken string) *AccountsClientsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsListCall) Fields(s ...googleapi.Field) *AccountsClientsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsListCall) IfNoneMatch(entityTag string) *AccountsClientsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsListCall) Context(ctx context.Context) *AccountsClientsListCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId": strconv.FormatInt(c.accountId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.list" call.
// Exactly one of *ListClientsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListClientsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AccountsClientsListCall) Do(opts ...googleapi.CallOption) (*ListClientsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListClientsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all the clients for the current sponsor buyer.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.list",
	//   "parameterOrder": [
	//     "accountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Unique numerical account ID of the sponsor buyer to list the clients for.",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Requested page size. The server may return fewer clients than requested.\nIf unspecified, the server will pick an appropriate default.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "A token identifying a page of results the server should return.\nTypically, this is the value of\nListClientsResponse.nextPageToken\nreturned from the previous call to the\naccounts.clients.list method.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients",
	//   "response": {
	//     "$ref": "ListClientsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *AccountsClientsListCall) Pages(ctx context.Context, f func(*ListClientsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "adexchangebuyer2.accounts.clients.update":

type AccountsClientsUpdateCall struct {
	s               *Service
	accountId       int64
	clientAccountId int64
	client          *Client
	urlParams_      gensupport.URLParams
	ctx_            context.Context
}

// Update: Updates an existing client buyer.
func (r *AccountsClientsService) Update(accountId int64, clientAccountId int64, client *Client) *AccountsClientsUpdateCall {
	c := &AccountsClientsUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	c.client = client
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsUpdateCall) Fields(s ...googleapi.Field) *AccountsClientsUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsUpdateCall) Context(ctx context.Context) *AccountsClientsUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.client)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.update" call.
// Exactly one of *Client or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Client.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *AccountsClientsUpdateCall) Do(opts ...googleapi.CallOption) (*Client, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &Client{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing client buyer.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer2.accounts.clients.update",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Unique numerical account ID for the buyer of which the client buyer\nis a customer; the sponsor buyer to update a client for. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Unique numerical account ID of the client to update. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}",
	//   "request": {
	//     "$ref": "Client"
	//   },
	//   "response": {
	//     "$ref": "Client"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.invitations.create":

type AccountsClientsInvitationsCreateCall struct {
	s                    *Service
	accountId            int64
	clientAccountId      int64
	clientuserinvitation *ClientUserInvitation
	urlParams_           gensupport.URLParams
	ctx_                 context.Context
}

// Create: Creates and sends out an email invitation to access
// an Ad Exchange client buyer account.
func (r *AccountsClientsInvitationsService) Create(accountId int64, clientAccountId int64, clientuserinvitation *ClientUserInvitation) *AccountsClientsInvitationsCreateCall {
	c := &AccountsClientsInvitationsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	c.clientuserinvitation = clientuserinvitation
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsInvitationsCreateCall) Fields(s ...googleapi.Field) *AccountsClientsInvitationsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsInvitationsCreateCall) Context(ctx context.Context) *AccountsClientsInvitationsCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsInvitationsCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.clientuserinvitation)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.invitations.create" call.
// Exactly one of *ClientUserInvitation or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ClientUserInvitation.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AccountsClientsInvitationsCreateCall) Do(opts ...googleapi.CallOption) (*ClientUserInvitation, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientUserInvitation{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Creates and sends out an email invitation to access\nan Ad Exchange client buyer account.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations",
	//   "httpMethod": "POST",
	//   "id": "adexchangebuyer2.accounts.clients.invitations.create",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer that the user\nshould be associated with. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations",
	//   "request": {
	//     "$ref": "ClientUserInvitation"
	//   },
	//   "response": {
	//     "$ref": "ClientUserInvitation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.invitations.get":

type AccountsClientsInvitationsGetCall struct {
	s               *Service
	accountId       int64
	clientAccountId int64
	invitationId    int64
	urlParams_      gensupport.URLParams
	ifNoneMatch_    string
	ctx_            context.Context
}

// Get: Retrieves an existing client user invitation.
func (r *AccountsClientsInvitationsService) Get(accountId int64, clientAccountId int64, invitationId int64) *AccountsClientsInvitationsGetCall {
	c := &AccountsClientsInvitationsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	c.invitationId = invitationId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsInvitationsGetCall) Fields(s ...googleapi.Field) *AccountsClientsInvitationsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsInvitationsGetCall) IfNoneMatch(entityTag string) *AccountsClientsInvitationsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsInvitationsGetCall) Context(ctx context.Context) *AccountsClientsInvitationsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsInvitationsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations/{invitationId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
		"invitationId":    strconv.FormatInt(c.invitationId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.invitations.get" call.
// Exactly one of *ClientUserInvitation or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ClientUserInvitation.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AccountsClientsInvitationsGetCall) Do(opts ...googleapi.CallOption) (*ClientUserInvitation, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientUserInvitation{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves an existing client user invitation.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations/{invitationId}",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.invitations.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId",
	//     "invitationId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer that the user invitation\nto be retrieved is associated with. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "invitationId": {
	//       "description": "Numerical identifier of the user invitation to retrieve. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations/{invitationId}",
	//   "response": {
	//     "$ref": "ClientUserInvitation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.invitations.list":

type AccountsClientsInvitationsListCall struct {
	s               *Service
	accountId       int64
	clientAccountId string
	urlParams_      gensupport.URLParams
	ifNoneMatch_    string
	ctx_            context.Context
}

// List: Lists all the client users invitations for a client
// with a given account ID.
func (r *AccountsClientsInvitationsService) List(accountId int64, clientAccountId string) *AccountsClientsInvitationsListCall {
	c := &AccountsClientsInvitationsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	return c
}

// PageSize sets the optional parameter "pageSize": Requested page size.
// Server may return fewer clients than requested.
// If unspecified, server will pick an appropriate default.
func (c *AccountsClientsInvitationsListCall) PageSize(pageSize int64) *AccountsClientsInvitationsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": A token
// identifying a page of results the server should return.
// Typically, this is the value
// of
// ListClientUserInvitationsResponse.nextPageToken
// returned from the previous call to
// the
// clients.invitations.list
// method.
func (c *AccountsClientsInvitationsListCall) PageToken(pageToken string) *AccountsClientsInvitationsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsInvitationsListCall) Fields(s ...googleapi.Field) *AccountsClientsInvitationsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsInvitationsListCall) IfNoneMatch(entityTag string) *AccountsClientsInvitationsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsInvitationsListCall) Context(ctx context.Context) *AccountsClientsInvitationsListCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsInvitationsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": c.clientAccountId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.invitations.list" call.
// Exactly one of *ListClientUserInvitationsResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *ListClientUserInvitationsResponse.ServerResponse.Header or
// (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *AccountsClientsInvitationsListCall) Do(opts ...googleapi.CallOption) (*ListClientUserInvitationsResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListClientUserInvitationsResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all the client users invitations for a client\nwith a given account ID.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.invitations.list",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer to list invitations for.\n(required)\nYou must either specify a string representation of a\nnumerical account identifier or the `-` character\nto list all the invitations for all the clients\nof a given sponsor buyer.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Requested page size. Server may return fewer clients than requested.\nIf unspecified, server will pick an appropriate default.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "A token identifying a page of results the server should return.\nTypically, this is the value of\nListClientUserInvitationsResponse.nextPageToken\nreturned from the previous call to the\nclients.invitations.list\nmethod.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/invitations",
	//   "response": {
	//     "$ref": "ListClientUserInvitationsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *AccountsClientsInvitationsListCall) Pages(ctx context.Context, f func(*ListClientUserInvitationsResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "adexchangebuyer2.accounts.clients.users.get":

type AccountsClientsUsersGetCall struct {
	s               *Service
	accountId       int64
	clientAccountId int64
	userId          int64
	urlParams_      gensupport.URLParams
	ifNoneMatch_    string
	ctx_            context.Context
}

// Get: Retrieves an existing client user.
func (r *AccountsClientsUsersService) Get(accountId int64, clientAccountId int64, userId int64) *AccountsClientsUsersGetCall {
	c := &AccountsClientsUsersGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	c.userId = userId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsUsersGetCall) Fields(s ...googleapi.Field) *AccountsClientsUsersGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsUsersGetCall) IfNoneMatch(entityTag string) *AccountsClientsUsersGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsUsersGetCall) Context(ctx context.Context) *AccountsClientsUsersGetCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsUsersGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
		"userId":          strconv.FormatInt(c.userId, 10),
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.users.get" call.
// Exactly one of *ClientUser or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ClientUser.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *AccountsClientsUsersGetCall) Do(opts ...googleapi.CallOption) (*ClientUser, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientUser{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Retrieves an existing client user.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.users.get",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId",
	//     "userId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer\nthat the user to be retrieved is associated with. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "userId": {
	//       "description": "Numerical identifier of the user to retrieve. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}",
	//   "response": {
	//     "$ref": "ClientUser"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// method id "adexchangebuyer2.accounts.clients.users.list":

type AccountsClientsUsersListCall struct {
	s               *Service
	accountId       int64
	clientAccountId string
	urlParams_      gensupport.URLParams
	ifNoneMatch_    string
	ctx_            context.Context
}

// List: Lists all the known client users for a specified
// sponsor buyer account ID.
func (r *AccountsClientsUsersService) List(accountId int64, clientAccountId string) *AccountsClientsUsersListCall {
	c := &AccountsClientsUsersListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	return c
}

// PageSize sets the optional parameter "pageSize": Requested page size.
// The server may return fewer clients than requested.
// If unspecified, the server will pick an appropriate default.
func (c *AccountsClientsUsersListCall) PageSize(pageSize int64) *AccountsClientsUsersListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": A token
// identifying a page of results the server should return.
// Typically, this is the value
// of
// ListClientUsersResponse.nextPageToken
// returned from the previous call to the
// accounts.clients.users.list method.
func (c *AccountsClientsUsersListCall) PageToken(pageToken string) *AccountsClientsUsersListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsUsersListCall) Fields(s ...googleapi.Field) *AccountsClientsUsersListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *AccountsClientsUsersListCall) IfNoneMatch(entityTag string) *AccountsClientsUsersListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsUsersListCall) Context(ctx context.Context) *AccountsClientsUsersListCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsUsersListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": c.clientAccountId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.users.list" call.
// Exactly one of *ListClientUsersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListClientUsersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *AccountsClientsUsersListCall) Do(opts ...googleapi.CallOption) (*ListClientUsersResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ListClientUsersResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Lists all the known client users for a specified\nsponsor buyer account ID.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users",
	//   "httpMethod": "GET",
	//   "id": "adexchangebuyer2.accounts.clients.users.list",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the sponsor buyer of the client to list users for.\n(required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "The account ID of the client buyer to list users for. (required)\nYou must specify either a string representation of a\nnumerical account identifier or the `-` character\nto list all the client users for all the clients\nof a given sponsor buyer.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Requested page size. The server may return fewer clients than requested.\nIf unspecified, the server will pick an appropriate default.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "A token identifying a page of results the server should return.\nTypically, this is the value of\nListClientUsersResponse.nextPageToken\nreturned from the previous call to the\naccounts.clients.users.list method.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users",
	//   "response": {
	//     "$ref": "ListClientUsersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *AccountsClientsUsersListCall) Pages(ctx context.Context, f func(*ListClientUsersResponse) error) error {
	c.ctx_ = ctx
	defer c.PageToken(c.urlParams_.Get("pageToken")) // reset paging to original point
	for {
		x, err := c.Do()
		if err != nil {
			return err
		}
		if err := f(x); err != nil {
			return err
		}
		if x.NextPageToken == "" {
			return nil
		}
		c.PageToken(x.NextPageToken)
	}
}

// method id "adexchangebuyer2.accounts.clients.users.update":

type AccountsClientsUsersUpdateCall struct {
	s               *Service
	accountId       int64
	clientAccountId int64
	userId          int64
	clientuser      *ClientUser
	urlParams_      gensupport.URLParams
	ctx_            context.Context
}

// Update: Updates an existing client user.
// Only the user status can be changed on update.
func (r *AccountsClientsUsersService) Update(accountId int64, clientAccountId int64, userId int64, clientuser *ClientUser) *AccountsClientsUsersUpdateCall {
	c := &AccountsClientsUsersUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.accountId = accountId
	c.clientAccountId = clientAccountId
	c.userId = userId
	c.clientuser = clientuser
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *AccountsClientsUsersUpdateCall) Fields(s ...googleapi.Field) *AccountsClientsUsersUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *AccountsClientsUsersUpdateCall) Context(ctx context.Context) *AccountsClientsUsersUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *AccountsClientsUsersUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.clientuser)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"accountId":       strconv.FormatInt(c.accountId, 10),
		"clientAccountId": strconv.FormatInt(c.clientAccountId, 10),
		"userId":          strconv.FormatInt(c.userId, 10),
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "adexchangebuyer2.accounts.clients.users.update" call.
// Exactly one of *ClientUser or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ClientUser.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *AccountsClientsUsersUpdateCall) Do(opts ...googleapi.CallOption) (*ClientUser, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &ClientUser{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Updates an existing client user.\nOnly the user status can be changed on update.",
	//   "flatPath": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}",
	//   "httpMethod": "PUT",
	//   "id": "adexchangebuyer2.accounts.clients.users.update",
	//   "parameterOrder": [
	//     "accountId",
	//     "clientAccountId",
	//     "userId"
	//   ],
	//   "parameters": {
	//     "accountId": {
	//       "description": "Numerical account ID of the client's sponsor buyer. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "clientAccountId": {
	//       "description": "Numerical account ID of the client buyer that the user to be retrieved\nis associated with. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "userId": {
	//       "description": "Numerical identifier of the user to retrieve. (required)",
	//       "format": "int64",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v2beta1/accounts/{accountId}/clients/{clientAccountId}/users/{userId}",
	//   "request": {
	//     "$ref": "ClientUser"
	//   },
	//   "response": {
	//     "$ref": "ClientUser"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/adexchange.buyer"
	//   ]
	// }

}
