// Package iam provides access to the Google Identity and Access Management API.
//
// See https://cloud.google.com/iam/
//
// Usage example:
//
//   import "google.golang.org/api/iam/v1"
//   ...
//   iamService, err := iam.New(oauthHttpClient)
package iam // import "google.golang.org/api/iam/v1"

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

const apiId = "iam:v1"
const apiName = "iam"
const apiVersion = "v1"
const basePath = "https://iam.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Projects = NewProjectsService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Projects *ProjectsService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewProjectsService(s *Service) *ProjectsService {
	rs := &ProjectsService{s: s}
	rs.ServiceAccounts = NewProjectsServiceAccountsService(s)
	return rs
}

type ProjectsService struct {
	s *Service

	ServiceAccounts *ProjectsServiceAccountsService
}

func NewProjectsServiceAccountsService(s *Service) *ProjectsServiceAccountsService {
	rs := &ProjectsServiceAccountsService{s: s}
	rs.Keys = NewProjectsServiceAccountsKeysService(s)
	return rs
}

type ProjectsServiceAccountsService struct {
	s *Service

	Keys *ProjectsServiceAccountsKeysService
}

func NewProjectsServiceAccountsKeysService(s *Service) *ProjectsServiceAccountsKeysService {
	rs := &ProjectsServiceAccountsKeysService{s: s}
	return rs
}

type ProjectsServiceAccountsKeysService struct {
	s *Service
}

// Binding: Associates `members` with a `role`.
type Binding struct {
	// Members: Specifies the identities requesting access for a Cloud
	// Platform resource. `members` can have the following values: *
	// `allUsers`: A special identifier that represents anyone who is on the
	// internet; with or without a Google account. *
	// `allAuthenticatedUsers`: A special identifier that represents anyone
	// who is authenticated with a Google account or a service account. *
	// `user:{emailid}`: An email address that represents a specific Google
	// account. For example, `alice@gmail.com` or `joe@example.com`. *
	// `serviceAccount:{emailid}`: An email address that represents a
	// service account. For example,
	// `my-other-app@appspot.gserviceaccount.com`. * `group:{emailid}`: An
	// email address that represents a Google group. For example,
	// `admins@example.com`. * `domain:{domain}`: A Google Apps domain name
	// that represents all the users of that domain. For example,
	// `google.com` or `example.com`.
	Members []string `json:"members,omitempty"`

	// Role: Role that is assigned to `members`. For example,
	// `roles/viewer`, `roles/editor`, or `roles/owner`. Required
	Role string `json:"role,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Members") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Binding) MarshalJSON() ([]byte, error) {
	type noMethod Binding
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// CloudAuditOptions: Write a Cloud Audit log
type CloudAuditOptions struct {
}

// Condition: A condition to be met.
type Condition struct {
	// Iam: Trusted attributes supplied by the IAM system.
	//
	// Possible values:
	//   "NO_ATTR"
	//   "AUTHORITY"
	//   "ATTRIBUTION"
	Iam string `json:"iam,omitempty"`

	// Op: An operator to apply the subject with.
	//
	// Possible values:
	//   "NO_OP"
	//   "EQUALS"
	//   "NOT_EQUALS"
	//   "IN"
	//   "NOT_IN"
	//   "DISCHARGED"
	Op string `json:"op,omitempty"`

	// Svc: Trusted attributes discharged by the service.
	Svc string `json:"svc,omitempty"`

	// Sys: Trusted attributes supplied by any service that owns resources
	// and uses the IAM system for access control.
	//
	// Possible values:
	//   "NO_ATTR"
	//   "REGION"
	//   "SERVICE"
	//   "NAME"
	//   "IP"
	Sys string `json:"sys,omitempty"`

	// Value: The object of the condition. Exactly one of these must be set.
	Value string `json:"value,omitempty"`

	// Values: The objects of the condition. This is mutually exclusive with
	// 'value'.
	Values []string `json:"values,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Iam") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Condition) MarshalJSON() ([]byte, error) {
	type noMethod Condition
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// CounterOptions: Options for counters
type CounterOptions struct {
	// Field: The field value to attribute.
	Field string `json:"field,omitempty"`

	// Metric: The metric to update.
	Metric string `json:"metric,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Field") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CounterOptions) MarshalJSON() ([]byte, error) {
	type noMethod CounterOptions
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// CreateServiceAccountKeyRequest: The service account key create
// request.
type CreateServiceAccountKeyRequest struct {
	// PrivateKeyType: The type of the key requested. GOOGLE_CREDENTIALS is
	// the default key type.
	//
	// Possible values:
	//   "TYPE_UNSPECIFIED"
	//   "TYPE_PKCS12_FILE"
	//   "TYPE_GOOGLE_CREDENTIALS_FILE"
	PrivateKeyType string `json:"privateKeyType,omitempty"`

	// ForceSendFields is a list of field names (e.g. "PrivateKeyType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreateServiceAccountKeyRequest) MarshalJSON() ([]byte, error) {
	type noMethod CreateServiceAccountKeyRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// CreateServiceAccountRequest: The service account create request.
type CreateServiceAccountRequest struct {
	// AccountId: Required. The account id that is used to generate the
	// service account email address and a stable unique id. It is unique
	// within a project, must be 1-63 characters long, and match the regular
	// expression [a-z]([-a-z0-9]*[a-z0-9]) to comply with RFC1035.
	AccountId string `json:"accountId,omitempty"`

	// ServiceAccount: The ServiceAccount resource to create. Currently,
	// only the following values are user assignable: display_name .
	ServiceAccount *ServiceAccount `json:"serviceAccount,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AccountId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CreateServiceAccountRequest) MarshalJSON() ([]byte, error) {
	type noMethod CreateServiceAccountRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DataAccessOptions: Write a Data Access (Gin) log
type DataAccessOptions struct {
}

// Empty: A generic empty message that you can re-use to avoid defining
// duplicated empty messages in your APIs. A typical example is to use
// it as the request or the response type of an API method. For
// instance: service Foo { rpc Bar(google.protobuf.Empty) returns
// (google.protobuf.Empty); } The JSON representation for `Empty` is
// empty JSON object `{}`.
type Empty struct {
	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`
}

// ListServiceAccountKeysResponse: The service account keys list
// response.
type ListServiceAccountKeysResponse struct {
	// Keys: The public keys for the service account.
	Keys []*ServiceAccountKey `json:"keys,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListServiceAccountKeysResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListServiceAccountKeysResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ListServiceAccountsResponse: The service account list response.
type ListServiceAccountsResponse struct {
	// Accounts: The list of matching service accounts.
	Accounts []*ServiceAccount `json:"accounts,omitempty"`

	// NextPageToken: To retrieve the next page of results, set
	// [ListServiceAccountsRequest.page_token] to this value.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Accounts") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListServiceAccountsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListServiceAccountsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LogConfig: Specifies what kind of log the caller must write Increment
// a streamz counter with the specified metric and field names. Metric
// names should start with a '/', generally be lowercase-only, and end
// in "_count". Field names should not contain an initial slash. The
// actual exported metric names will have "/iam/policy" prepended. Field
// names correspond to IAM request parameters and field values are their
// respective values. At present only "iam_principal", corresponding to
// IAMContext.principal, is supported. Examples: counter { metric:
// "/debug_access_count" field: "iam_principal" } ==> increment counter
// /iam/policy/backend_debug_access_count {iam_principal=[value of
// IAMContext.principal]} At this time we do not support: * multiple
// field names (though this may be supported in the future) *
// decrementing the counter * incrementing it by anything other than 1
type LogConfig struct {
	// CloudAudit: Cloud audit options.
	CloudAudit *CloudAuditOptions `json:"cloudAudit,omitempty"`

	// Counter: Counter options.
	Counter *CounterOptions `json:"counter,omitempty"`

	// DataAccess: Data access options.
	DataAccess *DataAccessOptions `json:"dataAccess,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CloudAudit") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LogConfig) MarshalJSON() ([]byte, error) {
	type noMethod LogConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Policy: Defines an Identity and Access Management (IAM) policy. It is
// used to specify access control policies for Cloud Platform resources.
// A `Policy` consists of a list of `bindings`. A `Binding` binds a list
// of `members` to a `role`, where the members can be user accounts,
// Google groups, Google domains, and service accounts. A `role` is a
// named list of permissions defined by IAM. **Example** { "bindings": [
// { "role": "roles/owner", "members": [ "user:mike@example.com",
// "group:admins@example.com", "domain:google.com",
// "serviceAccount:my-other-app@appspot.gserviceaccount.com"] }, {
// "role": "roles/viewer", "members": ["user:sean@example.com"] } ] }
// For a description of IAM and its features, see the [IAM developer's
// guide](https://cloud.google.com/iam).
type Policy struct {
	// Bindings: Associates a list of `members` to a `role`. Multiple
	// `bindings` must not be specified for the same `role`. `bindings` with
	// no members will result in an error.
	Bindings []*Binding `json:"bindings,omitempty"`

	// Etag: `etag` is used for optimistic concurrency control as a way to
	// help prevent simultaneous updates of a policy from overwriting each
	// other. It is strongly suggested that systems make use of the `etag`
	// in the read-modify-write cycle to perform policy updates in order to
	// avoid race conditions: An `etag` is returned in the response to
	// `getIamPolicy`, and systems are expected to put that etag in the
	// request to `setIamPolicy` to ensure that their change will be applied
	// to the same version of the policy. If no `etag` is provided in the
	// call to `setIamPolicy`, then the existing policy is overwritten
	// blindly.
	Etag string `json:"etag,omitempty"`

	Rules []*Rule `json:"rules,omitempty"`

	// Version: Version of the `Policy`. The default version is 0.
	Version int64 `json:"version,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Bindings") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Policy) MarshalJSON() ([]byte, error) {
	type noMethod Policy
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Rule: A rule to be applied in a Policy.
type Rule struct {
	// Action: Required
	//
	// Possible values:
	//   "NO_ACTION"
	//   "ALLOW"
	//   "ALLOW_WITH_LOG"
	//   "DENY"
	//   "DENY_WITH_LOG"
	//   "LOG"
	Action string `json:"action,omitempty"`

	// Conditions: Additional restrictions that must be met
	Conditions []*Condition `json:"conditions,omitempty"`

	// Description: Human-readable description of the rule.
	Description string `json:"description,omitempty"`

	// In: The rule matches if the PRINCIPAL/AUTHORITY_SELECTOR is in this
	// set of entries.
	In []string `json:"in,omitempty"`

	// LogConfig: The config returned to callers of tech.iam.IAM.CheckPolicy
	// for any entries that match the LOG action.
	LogConfig []*LogConfig `json:"logConfig,omitempty"`

	// NotIn: The rule matches if the PRINCIPAL/AUTHORITY_SELECTOR is not in
	// this set of entries. The format for in and not_in entries is the same
	// as for members in a Binding (see google/iam/v1/policy.proto).
	NotIn []string `json:"notIn,omitempty"`

	// Permissions: A permission is a string of form '..' (e.g.,
	// 'storage.buckets.list'). A value of '*' matches all permissions, and
	// a verb part of '*' (e.g., 'storage.buckets.*') matches all verbs.
	Permissions []string `json:"permissions,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Action") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Rule) MarshalJSON() ([]byte, error) {
	type noMethod Rule
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ServiceAccount: A service account in the Identity and Access
// Management API. To create a service account, you specify the
// project_id and account_id for the account. The account_id is unique
// within the project, and used to generate the service account email
// address and a stable unique id. All other methods can identify
// accounts using the format
// "projects/{project}/serviceAccounts/{account}". Using '-' as a
// wildcard for the project, will infer the project from the account.
// The account value can be the email address or the unique_id of the
// service account.
type ServiceAccount struct {
	// DisplayName: Optional. A user-specified description of the service
	// account. Must be fewer than 100 UTF-8 bytes.
	DisplayName string `json:"displayName,omitempty"`

	// Email: @OutputOnly Email address of the service account.
	Email string `json:"email,omitempty"`

	// Etag: Used to perform a consistent read-modify-write.
	Etag string `json:"etag,omitempty"`

	// Name: The resource name of the service account in the format
	// "projects/{project}/serviceAccounts/{account}". In requests using '-'
	// as a wildcard for the project, will infer the project from the
	// account and the account value can be the email address or the
	// unique_id of the service account. In responses the resource name will
	// always be in the format "projects/{project}/serviceAccounts/{email}".
	Name string `json:"name,omitempty"`

	// Oauth2ClientId: @OutputOnly. The OAuth2 client id for the service
	// account. This is used in conjunction with the OAuth2 clientconfig API
	// to make three legged OAuth2 (3LO) flows to access the data of Google
	// users.
	Oauth2ClientId string `json:"oauth2ClientId,omitempty"`

	// ProjectId: @OutputOnly The id of the project that owns the service
	// account.
	ProjectId string `json:"projectId,omitempty"`

	// UniqueId: @OutputOnly unique and stable id of the service account.
	UniqueId string `json:"uniqueId,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ServiceAccount) MarshalJSON() ([]byte, error) {
	type noMethod ServiceAccount
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ServiceAccountKey: Represents a service account key. A service
// account can have 0 or more key pairs. The private keys for these are
// not stored by Google. ServiceAccountKeys are immutable.
type ServiceAccountKey struct {
	// Name: The resource name of the service account key in the format
	// "projects/{project}/serviceAccounts/{email}/keys/{key}".
	Name string `json:"name,omitempty"`

	// PrivateKeyData: The key data.
	PrivateKeyData string `json:"privateKeyData,omitempty"`

	// PrivateKeyType: The type of the private key.
	//
	// Possible values:
	//   "TYPE_UNSPECIFIED"
	//   "TYPE_PKCS12_FILE"
	//   "TYPE_GOOGLE_CREDENTIALS_FILE"
	PrivateKeyType string `json:"privateKeyType,omitempty"`

	// ValidAfterTime: The key can be used after this timestamp.
	ValidAfterTime string `json:"validAfterTime,omitempty"`

	// ValidBeforeTime: The key can be used before this timestamp.
	ValidBeforeTime string `json:"validBeforeTime,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ServiceAccountKey) MarshalJSON() ([]byte, error) {
	type noMethod ServiceAccountKey
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SetIamPolicyRequest: Request message for `SetIamPolicy` method.
type SetIamPolicyRequest struct {
	// Policy: REQUIRED: The complete policy to be applied to the
	// `resource`. The size of the policy is limited to a few 10s of KB. An
	// empty policy is a valid policy but certain Cloud Platform services
	// (such as Projects) might reject them.
	Policy *Policy `json:"policy,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Policy") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SetIamPolicyRequest) MarshalJSON() ([]byte, error) {
	type noMethod SetIamPolicyRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SignBlobRequest: The service account sign blob request.
type SignBlobRequest struct {
	// BytesToSign: The bytes to sign
	BytesToSign string `json:"bytesToSign,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BytesToSign") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignBlobRequest) MarshalJSON() ([]byte, error) {
	type noMethod SignBlobRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SignBlobResponse: The service account sign blob response.
type SignBlobResponse struct {
	// KeyId: The id of the key used to sign the blob.
	KeyId string `json:"keyId,omitempty"`

	// Signature: The signed blob.
	Signature string `json:"signature,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "KeyId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SignBlobResponse) MarshalJSON() ([]byte, error) {
	type noMethod SignBlobResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// TestIamPermissionsRequest: Request message for `TestIamPermissions`
// method.
type TestIamPermissionsRequest struct {
	// Permissions: The set of permissions to check for the `resource`.
	// Permissions with wildcards (such as '*' or 'storage.*') are not
	// allowed. For more information see IAM Overview.
	Permissions []string `json:"permissions,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Permissions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TestIamPermissionsRequest) MarshalJSON() ([]byte, error) {
	type noMethod TestIamPermissionsRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// TestIamPermissionsResponse: Response message for `TestIamPermissions`
// method.
type TestIamPermissionsResponse struct {
	// Permissions: A subset of `TestPermissionsRequest.permissions` that
	// the caller is allowed.
	Permissions []string `json:"permissions,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Permissions") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TestIamPermissionsResponse) MarshalJSON() ([]byte, error) {
	type noMethod TestIamPermissionsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "iam.projects.serviceAccounts.create":

type ProjectsServiceAccountsCreateCall struct {
	s                           *Service
	name                        string
	createserviceaccountrequest *CreateServiceAccountRequest
	urlParams_                  gensupport.URLParams
	ctx_                        context.Context
}

// Create: Creates a service account and returns it.
func (r *ProjectsServiceAccountsService) Create(name string, createserviceaccountrequest *CreateServiceAccountRequest) *ProjectsServiceAccountsCreateCall {
	c := &ProjectsServiceAccountsCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.createserviceaccountrequest = createserviceaccountrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsCreateCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsCreateCall) Context(ctx context.Context) *ProjectsServiceAccountsCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createserviceaccountrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}/serviceAccounts")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.create" call.
// Exactly one of *ServiceAccount or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ServiceAccount.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsCreateCall) Do(opts ...googleapi.CallOption) (*ServiceAccount, error) {
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
	ret := &ServiceAccount{
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
	//   "description": "Creates a service account and returns it.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "Required. The resource name of the project associated with the service accounts, such as \"projects/123\"",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}/serviceAccounts",
	//   "request": {
	//     "$ref": "CreateServiceAccountRequest"
	//   },
	//   "response": {
	//     "$ref": "ServiceAccount"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.delete":

type ProjectsServiceAccountsDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a service acount.
func (r *ProjectsServiceAccountsService) Delete(name string) *ProjectsServiceAccountsDeleteCall {
	c := &ProjectsServiceAccountsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsDeleteCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsDeleteCall) Context(ctx context.Context) *ProjectsServiceAccountsDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsServiceAccountsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	ret := &Empty{
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
	//   "description": "Deletes a service acount.",
	//   "httpMethod": "DELETE",
	//   "id": "iam.projects.serviceAccounts.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". Using '-' as a wildcard for the project, will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.get":

type ProjectsServiceAccountsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets a ServiceAccount
func (r *ProjectsServiceAccountsService) Get(name string) *ProjectsServiceAccountsGetCall {
	c := &ProjectsServiceAccountsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsGetCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsServiceAccountsGetCall) IfNoneMatch(entityTag string) *ProjectsServiceAccountsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsGetCall) Context(ctx context.Context) *ProjectsServiceAccountsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
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

// Do executes the "iam.projects.serviceAccounts.get" call.
// Exactly one of *ServiceAccount or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ServiceAccount.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsGetCall) Do(opts ...googleapi.CallOption) (*ServiceAccount, error) {
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
	ret := &ServiceAccount{
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
	//   "description": "Gets a ServiceAccount",
	//   "httpMethod": "GET",
	//   "id": "iam.projects.serviceAccounts.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". Using '-' as a wildcard for the project, will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "ServiceAccount"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.getIamPolicy":

type ProjectsServiceAccountsGetIamPolicyCall struct {
	s          *Service
	resource   string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// GetIamPolicy: Returns the IAM access control policy for specified IAM
// resource.
func (r *ProjectsServiceAccountsService) GetIamPolicy(resource string) *ProjectsServiceAccountsGetIamPolicyCall {
	c := &ProjectsServiceAccountsGetIamPolicyCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resource = resource
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsGetIamPolicyCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsGetIamPolicyCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsGetIamPolicyCall) Context(ctx context.Context) *ProjectsServiceAccountsGetIamPolicyCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsGetIamPolicyCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+resource}:getIamPolicy")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resource": c.resource,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.getIamPolicy" call.
// Exactly one of *Policy or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Policy.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsServiceAccountsGetIamPolicyCall) Do(opts ...googleapi.CallOption) (*Policy, error) {
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
	ret := &Policy{
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
	//   "description": "Returns the IAM access control policy for specified IAM resource.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.getIamPolicy",
	//   "parameterOrder": [
	//     "resource"
	//   ],
	//   "parameters": {
	//     "resource": {
	//       "description": "REQUIRED: The resource for which the policy is being requested. `resource` is usually specified as a path, such as `projects/*project*/zones/*zone*/disks/*disk*`. The format for the path specified in this value is resource specific and is specified in the `getIamPolicy` documentation.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+resource}:getIamPolicy",
	//   "response": {
	//     "$ref": "Policy"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.list":

type ProjectsServiceAccountsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists service accounts for a project.
func (r *ProjectsServiceAccountsService) List(name string) *ProjectsServiceAccountsListCall {
	c := &ProjectsServiceAccountsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// PageSize sets the optional parameter "pageSize": Optional limit on
// the number of service accounts to include in the response. Further
// accounts can subsequently be obtained by including the
// [ListServiceAccountsResponse.next_page_token] in a subsequent
// request.
func (c *ProjectsServiceAccountsListCall) PageSize(pageSize int64) *ProjectsServiceAccountsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": Optional
// pagination token returned in an earlier
// [ListServiceAccountsResponse.next_page_token].
func (c *ProjectsServiceAccountsListCall) PageToken(pageToken string) *ProjectsServiceAccountsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsListCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsServiceAccountsListCall) IfNoneMatch(entityTag string) *ProjectsServiceAccountsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsListCall) Context(ctx context.Context) *ProjectsServiceAccountsListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}/serviceAccounts")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
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

// Do executes the "iam.projects.serviceAccounts.list" call.
// Exactly one of *ListServiceAccountsResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *ListServiceAccountsResponse.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsListCall) Do(opts ...googleapi.CallOption) (*ListServiceAccountsResponse, error) {
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
	ret := &ListServiceAccountsResponse{
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
	//   "description": "Lists service accounts for a project.",
	//   "httpMethod": "GET",
	//   "id": "iam.projects.serviceAccounts.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "Required. The resource name of the project associated with the service accounts, such as \"projects/123\"",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Optional limit on the number of service accounts to include in the response. Further accounts can subsequently be obtained by including the [ListServiceAccountsResponse.next_page_token] in a subsequent request.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "Optional pagination token returned in an earlier [ListServiceAccountsResponse.next_page_token].",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}/serviceAccounts",
	//   "response": {
	//     "$ref": "ListServiceAccountsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsServiceAccountsListCall) Pages(ctx context.Context, f func(*ListServiceAccountsResponse) error) error {
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

// method id "iam.projects.serviceAccounts.setIamPolicy":

type ProjectsServiceAccountsSetIamPolicyCall struct {
	s                   *Service
	resource            string
	setiampolicyrequest *SetIamPolicyRequest
	urlParams_          gensupport.URLParams
	ctx_                context.Context
}

// SetIamPolicy: Sets the IAM access control policy for the specified
// IAM resource.
func (r *ProjectsServiceAccountsService) SetIamPolicy(resource string, setiampolicyrequest *SetIamPolicyRequest) *ProjectsServiceAccountsSetIamPolicyCall {
	c := &ProjectsServiceAccountsSetIamPolicyCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resource = resource
	c.setiampolicyrequest = setiampolicyrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsSetIamPolicyCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsSetIamPolicyCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsSetIamPolicyCall) Context(ctx context.Context) *ProjectsServiceAccountsSetIamPolicyCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsSetIamPolicyCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.setiampolicyrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+resource}:setIamPolicy")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resource": c.resource,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.setIamPolicy" call.
// Exactly one of *Policy or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Policy.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsServiceAccountsSetIamPolicyCall) Do(opts ...googleapi.CallOption) (*Policy, error) {
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
	ret := &Policy{
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
	//   "description": "Sets the IAM access control policy for the specified IAM resource.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.setIamPolicy",
	//   "parameterOrder": [
	//     "resource"
	//   ],
	//   "parameters": {
	//     "resource": {
	//       "description": "REQUIRED: The resource for which the policy is being specified. `resource` is usually specified as a path, such as `projects/*project*/zones/*zone*/disks/*disk*`. The format for the path specified in this value is resource specific and is specified in the `setIamPolicy` documentation.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+resource}:setIamPolicy",
	//   "request": {
	//     "$ref": "SetIamPolicyRequest"
	//   },
	//   "response": {
	//     "$ref": "Policy"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.signBlob":

type ProjectsServiceAccountsSignBlobCall struct {
	s               *Service
	name            string
	signblobrequest *SignBlobRequest
	urlParams_      gensupport.URLParams
	ctx_            context.Context
}

// SignBlob: Signs a blob using a service account.
func (r *ProjectsServiceAccountsService) SignBlob(name string, signblobrequest *SignBlobRequest) *ProjectsServiceAccountsSignBlobCall {
	c := &ProjectsServiceAccountsSignBlobCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.signblobrequest = signblobrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsSignBlobCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsSignBlobCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsSignBlobCall) Context(ctx context.Context) *ProjectsServiceAccountsSignBlobCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsSignBlobCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.signblobrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}:signBlob")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.signBlob" call.
// Exactly one of *SignBlobResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *SignBlobResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsSignBlobCall) Do(opts ...googleapi.CallOption) (*SignBlobResponse, error) {
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
	ret := &SignBlobResponse{
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
	//   "description": "Signs a blob using a service account.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.signBlob",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". Using '-' as a wildcard for the project, will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}:signBlob",
	//   "request": {
	//     "$ref": "SignBlobRequest"
	//   },
	//   "response": {
	//     "$ref": "SignBlobResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.testIamPermissions":

type ProjectsServiceAccountsTestIamPermissionsCall struct {
	s                         *Service
	resource                  string
	testiampermissionsrequest *TestIamPermissionsRequest
	urlParams_                gensupport.URLParams
	ctx_                      context.Context
}

// TestIamPermissions: Tests the specified permissions against the IAM
// access control policy for the specified IAM resource.
func (r *ProjectsServiceAccountsService) TestIamPermissions(resource string, testiampermissionsrequest *TestIamPermissionsRequest) *ProjectsServiceAccountsTestIamPermissionsCall {
	c := &ProjectsServiceAccountsTestIamPermissionsCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resource = resource
	c.testiampermissionsrequest = testiampermissionsrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsTestIamPermissionsCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsTestIamPermissionsCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsTestIamPermissionsCall) Context(ctx context.Context) *ProjectsServiceAccountsTestIamPermissionsCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsTestIamPermissionsCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.testiampermissionsrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+resource}:testIamPermissions")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resource": c.resource,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.testIamPermissions" call.
// Exactly one of *TestIamPermissionsResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *TestIamPermissionsResponse.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsTestIamPermissionsCall) Do(opts ...googleapi.CallOption) (*TestIamPermissionsResponse, error) {
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
	ret := &TestIamPermissionsResponse{
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
	//   "description": "Tests the specified permissions against the IAM access control policy for the specified IAM resource.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.testIamPermissions",
	//   "parameterOrder": [
	//     "resource"
	//   ],
	//   "parameters": {
	//     "resource": {
	//       "description": "REQUIRED: The resource for which the policy detail is being requested. `resource` is usually specified as a path, such as `projects/*project*/zones/*zone*/disks/*disk*`. The format for the path specified in this value is resource specific and is specified in the `testIamPermissions` documentation.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+resource}:testIamPermissions",
	//   "request": {
	//     "$ref": "TestIamPermissionsRequest"
	//   },
	//   "response": {
	//     "$ref": "TestIamPermissionsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.update":

type ProjectsServiceAccountsUpdateCall struct {
	s              *Service
	name           string
	serviceaccount *ServiceAccount
	urlParams_     gensupport.URLParams
	ctx_           context.Context
}

// Update: Updates a service account. Currently, only the following
// fields are updatable: 'display_name' . The 'etag' is mandatory.
func (r *ProjectsServiceAccountsService) Update(name string, serviceaccount *ServiceAccount) *ProjectsServiceAccountsUpdateCall {
	c := &ProjectsServiceAccountsUpdateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.serviceaccount = serviceaccount
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsUpdateCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsUpdateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsUpdateCall) Context(ctx context.Context) *ProjectsServiceAccountsUpdateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsUpdateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.serviceaccount)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.update" call.
// Exactly one of *ServiceAccount or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *ServiceAccount.ServerResponse.Header or (if a response was returned
// at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsUpdateCall) Do(opts ...googleapi.CallOption) (*ServiceAccount, error) {
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
	ret := &ServiceAccount{
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
	//   "description": "Updates a service account. Currently, only the following fields are updatable: 'display_name' . The 'etag' is mandatory.",
	//   "httpMethod": "PUT",
	//   "id": "iam.projects.serviceAccounts.update",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". In requests using '-' as a wildcard for the project, will infer the project from the account and the account value can be the email address or the unique_id of the service account. In responses the resource name will always be in the format \"projects/{project}/serviceAccounts/{email}\".",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "request": {
	//     "$ref": "ServiceAccount"
	//   },
	//   "response": {
	//     "$ref": "ServiceAccount"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.keys.create":

type ProjectsServiceAccountsKeysCreateCall struct {
	s                              *Service
	name                           string
	createserviceaccountkeyrequest *CreateServiceAccountKeyRequest
	urlParams_                     gensupport.URLParams
	ctx_                           context.Context
}

// Create: Creates a service account key and returns it.
func (r *ProjectsServiceAccountsKeysService) Create(name string, createserviceaccountkeyrequest *CreateServiceAccountKeyRequest) *ProjectsServiceAccountsKeysCreateCall {
	c := &ProjectsServiceAccountsKeysCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.createserviceaccountkeyrequest = createserviceaccountkeyrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsKeysCreateCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsKeysCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsKeysCreateCall) Context(ctx context.Context) *ProjectsServiceAccountsKeysCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsKeysCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.createserviceaccountkeyrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}/keys")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.keys.create" call.
// Exactly one of *ServiceAccountKey or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ServiceAccountKey.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsKeysCreateCall) Do(opts ...googleapi.CallOption) (*ServiceAccountKey, error) {
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
	ret := &ServiceAccountKey{
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
	//   "description": "Creates a service account key and returns it.",
	//   "httpMethod": "POST",
	//   "id": "iam.projects.serviceAccounts.keys.create",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". Using '-' as a wildcard for the project, will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}/keys",
	//   "request": {
	//     "$ref": "CreateServiceAccountKeyRequest"
	//   },
	//   "response": {
	//     "$ref": "ServiceAccountKey"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.keys.delete":

type ProjectsServiceAccountsKeysDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a service account key.
func (r *ProjectsServiceAccountsKeysService) Delete(name string) *ProjectsServiceAccountsKeysDeleteCall {
	c := &ProjectsServiceAccountsKeysDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsKeysDeleteCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsKeysDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsKeysDeleteCall) Context(ctx context.Context) *ProjectsServiceAccountsKeysDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsKeysDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "iam.projects.serviceAccounts.keys.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsServiceAccountsKeysDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	ret := &Empty{
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
	//   "description": "Deletes a service account key.",
	//   "httpMethod": "DELETE",
	//   "id": "iam.projects.serviceAccounts.keys.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account key in the format \"projects/{project}/serviceAccounts/{account}/keys/{key}\". Using '-' as a wildcard for the project will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*/keys/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.keys.get":

type ProjectsServiceAccountsKeysGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets the ServiceAccountKey by key id.
func (r *ProjectsServiceAccountsKeysService) Get(name string) *ProjectsServiceAccountsKeysGetCall {
	c := &ProjectsServiceAccountsKeysGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsKeysGetCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsKeysGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsServiceAccountsKeysGetCall) IfNoneMatch(entityTag string) *ProjectsServiceAccountsKeysGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsKeysGetCall) Context(ctx context.Context) *ProjectsServiceAccountsKeysGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsKeysGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
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

// Do executes the "iam.projects.serviceAccounts.keys.get" call.
// Exactly one of *ServiceAccountKey or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ServiceAccountKey.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsKeysGetCall) Do(opts ...googleapi.CallOption) (*ServiceAccountKey, error) {
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
	ret := &ServiceAccountKey{
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
	//   "description": "Gets the ServiceAccountKey by key id.",
	//   "httpMethod": "GET",
	//   "id": "iam.projects.serviceAccounts.keys.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The resource name of the service account key in the format \"projects/{project}/serviceAccounts/{account}/keys/{key}\". Using '-' as a wildcard for the project will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*/keys/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "ServiceAccountKey"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "iam.projects.serviceAccounts.keys.list":

type ProjectsServiceAccountsKeysListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists service account keys
func (r *ProjectsServiceAccountsKeysService) List(name string) *ProjectsServiceAccountsKeysListCall {
	c := &ProjectsServiceAccountsKeysListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// KeyTypes sets the optional parameter "keyTypes": The type of keys the
// user wants to list. If empty, all key types are included in the
// response. Duplicate key types are not allowed.
//
// Possible values:
//   "KEY_TYPE_UNSPECIFIED"
//   "USER_MANAGED"
//   "SYSTEM_MANAGED"
func (c *ProjectsServiceAccountsKeysListCall) KeyTypes(keyTypes ...string) *ProjectsServiceAccountsKeysListCall {
	c.urlParams_.SetMulti("keyTypes", append([]string{}, keyTypes...))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsServiceAccountsKeysListCall) Fields(s ...googleapi.Field) *ProjectsServiceAccountsKeysListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsServiceAccountsKeysListCall) IfNoneMatch(entityTag string) *ProjectsServiceAccountsKeysListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsServiceAccountsKeysListCall) Context(ctx context.Context) *ProjectsServiceAccountsKeysListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsServiceAccountsKeysListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}/keys")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
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

// Do executes the "iam.projects.serviceAccounts.keys.list" call.
// Exactly one of *ListServiceAccountKeysResponse or error will be
// non-nil. Any non-2xx status code is an error. Response headers are in
// either *ListServiceAccountKeysResponse.ServerResponse.Header or (if a
// response was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsServiceAccountsKeysListCall) Do(opts ...googleapi.CallOption) (*ListServiceAccountKeysResponse, error) {
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
	ret := &ListServiceAccountKeysResponse{
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
	//   "description": "Lists service account keys",
	//   "httpMethod": "GET",
	//   "id": "iam.projects.serviceAccounts.keys.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "keyTypes": {
	//       "description": "The type of keys the user wants to list. If empty, all key types are included in the response. Duplicate key types are not allowed.",
	//       "enum": [
	//         "KEY_TYPE_UNSPECIFIED",
	//         "USER_MANAGED",
	//         "SYSTEM_MANAGED"
	//       ],
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The resource name of the service account in the format \"projects/{project}/serviceAccounts/{account}\". Using '-' as a wildcard for the project, will infer the project from the account. The account value can be the email address or the unique_id of the service account.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/serviceAccounts/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}/keys",
	//   "response": {
	//     "$ref": "ListServiceAccountKeysResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}
