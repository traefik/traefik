// Package people provides access to the Google People API.
//
// See https://developers.google.com/people/
//
// Usage example:
//
//   import "google.golang.org/api/people/v1"
//   ...
//   peopleService, err := people.New(oauthHttpClient)
package people // import "google.golang.org/api/people/v1"

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

const apiId = "people:v1"
const apiName = "people"
const apiVersion = "v1"
const basePath = "https://people.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// Manage your contacts
	ContactsScope = "https://www.googleapis.com/auth/contacts"

	// View your contacts
	ContactsReadonlyScope = "https://www.googleapis.com/auth/contacts.readonly"

	// Know your basic profile info and list of people in your circles.
	PlusLoginScope = "https://www.googleapis.com/auth/plus.login"

	// View your street addresses
	UserAddressesReadScope = "https://www.googleapis.com/auth/user.addresses.read"

	// View your complete date of birth
	UserBirthdayReadScope = "https://www.googleapis.com/auth/user.birthday.read"

	// View your email addresses
	UserEmailsReadScope = "https://www.googleapis.com/auth/user.emails.read"

	// View your phone numbers
	UserPhonenumbersReadScope = "https://www.googleapis.com/auth/user.phonenumbers.read"

	// View your email address
	UserinfoEmailScope = "https://www.googleapis.com/auth/userinfo.email"

	// View your basic profile info
	UserinfoProfileScope = "https://www.googleapis.com/auth/userinfo.profile"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.People = NewPeopleService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	People *PeopleService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewPeopleService(s *Service) *PeopleService {
	rs := &PeopleService{s: s}
	rs.Connections = NewPeopleConnectionsService(s)
	return rs
}

type PeopleService struct {
	s *Service

	Connections *PeopleConnectionsService
}

func NewPeopleConnectionsService(s *Service) *PeopleConnectionsService {
	rs := &PeopleConnectionsService{s: s}
	return rs
}

type PeopleConnectionsService struct {
	s *Service
}

// Address: A person's physical address. May be a P.O. box or street
// address. All fields are optional.
type Address struct {
	// City: The city of the address.
	City string `json:"city,omitempty"`

	// Country: The country of the address.
	Country string `json:"country,omitempty"`

	// CountryCode: The [ISO 3166-1
	// alpha-2](http://www.iso.org/iso/country_codes.htm) country code of
	// the address.
	CountryCode string `json:"countryCode,omitempty"`

	// ExtendedAddress: The extended address of the address; for example,
	// the apartment number.
	ExtendedAddress string `json:"extendedAddress,omitempty"`

	// FormattedType: The read-only type of the address translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// FormattedValue: The read-only value of the address formatted in the
	// viewer's account locale or the `Accept-Language` HTTP header locale.
	FormattedValue string `json:"formattedValue,omitempty"`

	// Metadata: Metadata about the address.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// PoBox: The P.O. box of the address.
	PoBox string `json:"poBox,omitempty"`

	// PostalCode: The postal code of the address.
	PostalCode string `json:"postalCode,omitempty"`

	// Region: The region of the address; for example, the state or
	// province.
	Region string `json:"region,omitempty"`

	// StreetAddress: The street address.
	StreetAddress string `json:"streetAddress,omitempty"`

	// Type: The type of the address. The type can be custom or predefined.
	// Possible values include, but are not limited to, the following: *
	// `home` * `work` * `other`
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "City") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Address) MarshalJSON() ([]byte, error) {
	type noMethod Address
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Biography: A person's short biography.
type Biography struct {
	// Metadata: Metadata about the biography.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The short biography.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Biography) MarshalJSON() ([]byte, error) {
	type noMethod Biography
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Birthday: A person's birthday. At least one of the `date` and `text`
// fields are specified. The `date` and `text` fields typically
// represent the same date, but are not guaranteed to.
type Birthday struct {
	// Date: The date of the birthday.
	Date *Date `json:"date,omitempty"`

	// Metadata: Metadata about the birthday.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Text: A free-form string representing the user's birthday.
	Text string `json:"text,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Date") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Birthday) MarshalJSON() ([]byte, error) {
	type noMethod Birthday
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// BraggingRights: A person's bragging rights.
type BraggingRights struct {
	// Metadata: Metadata about the bragging rights.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The bragging rights; for example, `climbed mount everest`.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BraggingRights) MarshalJSON() ([]byte, error) {
	type noMethod BraggingRights
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ContactGroupMembership: A Google contact group membership.
type ContactGroupMembership struct {
	// ContactGroupId: The contact group ID for the contact group
	// membership. The contact group ID can be custom or predefined.
	// Possible values include, but are not limited to, the following: *
	// `myContacts` * `starred` * A numerical ID for user-created groups.
	ContactGroupId string `json:"contactGroupId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ContactGroupId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ContactGroupMembership) MarshalJSON() ([]byte, error) {
	type noMethod ContactGroupMembership
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// CoverPhoto: A person's cover photo. A large image shown on the
// person's profile page that represents who they are or what they care
// about.
type CoverPhoto struct {
	// Default: True if the cover photo is the default cover photo; false if
	// the cover photo is a user-provided cover photo.
	Default bool `json:"default,omitempty"`

	// Metadata: Metadata about the cover photo.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Url: The URL of the cover photo.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Default") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *CoverPhoto) MarshalJSON() ([]byte, error) {
	type noMethod CoverPhoto
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Date: Represents a whole calendar date, for example a date of birth.
// The time of day and time zone are either specified elsewhere or are
// not significant. The date is relative to the [Proleptic Gregorian
// Calendar](https://en.wikipedia.org/wiki/Proleptic_Gregorian_calendar).
//  The day may be 0 to represent a year and month where the day is not
// significant. The year may be 0 to represent a month and day
// independent of year; for example, anniversary date.
type Date struct {
	// Day: Day of month. Must be from 1 to 31 and valid for the year and
	// month, or 0 if specifying a year/month where the day is not
	// significant.
	Day int64 `json:"day,omitempty"`

	// Month: Month of year. Must be from 1 to 12.
	Month int64 `json:"month,omitempty"`

	// Year: Year of date. Must be from 1 to 9999, or 0 if specifying a date
	// without a year.
	Year int64 `json:"year,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Day") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Date) MarshalJSON() ([]byte, error) {
	type noMethod Date
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DomainMembership: A Google Apps Domain membership.
type DomainMembership struct {
	// InViewerDomain: True if the person is in the viewer's Google Apps
	// domain.
	InViewerDomain bool `json:"inViewerDomain,omitempty"`

	// ForceSendFields is a list of field names (e.g. "InViewerDomain") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DomainMembership) MarshalJSON() ([]byte, error) {
	type noMethod DomainMembership
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// EmailAddress: A person's email address.
type EmailAddress struct {
	// FormattedType: The read-only type of the email address translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the email address.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Type: The type of the email address. The type can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `home` * `work` * `other`
	Type string `json:"type,omitempty"`

	// Value: The email address.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *EmailAddress) MarshalJSON() ([]byte, error) {
	type noMethod EmailAddress
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Event: An event related to the person.
type Event struct {
	// Date: The date of the event.
	Date *Date `json:"date,omitempty"`

	// FormattedType: The read-only type of the event translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the event.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Type: The type of the event. The type can be custom or predefined.
	// Possible values include, but are not limited to, the following: *
	// `anniversary` * `other`
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Date") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Event) MarshalJSON() ([]byte, error) {
	type noMethod Event
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// FieldMetadata: Metadata about a field.
type FieldMetadata struct {
	// Primary: True if the field is the primary field; false if the field
	// is a secondary field.
	Primary bool `json:"primary,omitempty"`

	// Source: The source of the field.
	Source *Source `json:"source,omitempty"`

	// Verified: True if the field is verified; false if the field is
	// unverified. A verified field is typically a name, email address,
	// phone number, or website that has been confirmed to be owned by the
	// person.
	Verified bool `json:"verified,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Primary") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FieldMetadata) MarshalJSON() ([]byte, error) {
	type noMethod FieldMetadata
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Gender: A person's gender.
type Gender struct {
	// FormattedValue: The read-only value of the gender translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedValue string `json:"formattedValue,omitempty"`

	// Metadata: Metadata about the gender.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The gender for the person. The gender can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `male` * `female` * `other` * `unknown`
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Gender) MarshalJSON() ([]byte, error) {
	type noMethod Gender
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type GetPeopleResponse struct {
	// Responses: The response for each requested resource name.
	Responses []*PersonResponse `json:"responses,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Responses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GetPeopleResponse) MarshalJSON() ([]byte, error) {
	type noMethod GetPeopleResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ImClient: A person's instant messaging client.
type ImClient struct {
	// FormattedProtocol: The read-only protocol of the IM client formatted
	// in the viewer's account locale or the `Accept-Language` HTTP header
	// locale.
	FormattedProtocol string `json:"formattedProtocol,omitempty"`

	// FormattedType: The read-only type of the IM client translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the IM client.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Protocol: The protocol of the IM client. The protocol can be custom
	// or predefined. Possible values include, but are not limited to, the
	// following: * `aim` * `msn` * `yahoo` * `skype` * `qq` * `googleTalk`
	// * `icq` * `jabber` * `netMeeting`
	Protocol string `json:"protocol,omitempty"`

	// Type: The type of the IM client. The type can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `home` * `work` * `other`
	Type string `json:"type,omitempty"`

	// Username: The user name used in the IM client.
	Username string `json:"username,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedProtocol")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImClient) MarshalJSON() ([]byte, error) {
	type noMethod ImClient
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Interest: One of the person's interests.
type Interest struct {
	// Metadata: Metadata about the interest.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The interest; for example, `stargazing`.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Interest) MarshalJSON() ([]byte, error) {
	type noMethod Interest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type ListConnectionsResponse struct {
	// Connections: The list of people that the requestor is connected to.
	Connections []*Person `json:"connections,omitempty"`

	// NextPageToken: The token that can be used to retrieve the next page
	// of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// NextSyncToken: The token that can be used to retrieve changes since
	// the last request.
	NextSyncToken string `json:"nextSyncToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Connections") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListConnectionsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListConnectionsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Locale: A person's locale preference.
type Locale struct {
	// Metadata: Metadata about the locale.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The well-formed [IETF BCP
	// 47](https://tools.ietf.org/html/bcp47) language tag representing the
	// locale.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Locale) MarshalJSON() ([]byte, error) {
	type noMethod Locale
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Membership: A person's membership in a group.
type Membership struct {
	// ContactGroupMembership: The contact group membership.
	ContactGroupMembership *ContactGroupMembership `json:"contactGroupMembership,omitempty"`

	// DomainMembership: The domain membership.
	DomainMembership *DomainMembership `json:"domainMembership,omitempty"`

	// Metadata: Metadata about the membership.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "ContactGroupMembership") to unconditionally include in API requests.
	// By default, fields with empty values are omitted from API requests.
	// However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Membership) MarshalJSON() ([]byte, error) {
	type noMethod Membership
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Name: A person's name. If the name is a mononym, the family name is
// empty.
type Name struct {
	// DisplayName: The display name formatted according to the locale
	// specified by the viewer's account or the Accept-Language HTTP header.
	DisplayName string `json:"displayName,omitempty"`

	// FamilyName: The family name.
	FamilyName string `json:"familyName,omitempty"`

	// GivenName: The given name.
	GivenName string `json:"givenName,omitempty"`

	// HonorificPrefix: The honorific prefixes, such as `Mrs.` or `Dr.`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`

	// HonorificSuffix: The honorific suffixes, such as `Jr.`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`

	// Metadata: Metadata about the name.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// MiddleName: The middle name(s).
	MiddleName string `json:"middleName,omitempty"`

	// PhoneticFamilyName: The family name spelled as it sounds.
	PhoneticFamilyName string `json:"phoneticFamilyName,omitempty"`

	// PhoneticGivenName: The given name spelled as it sounds.
	PhoneticGivenName string `json:"phoneticGivenName,omitempty"`

	// PhoneticHonorificPrefix: The honorific prefixes spelled as they
	// sound.
	PhoneticHonorificPrefix string `json:"phoneticHonorificPrefix,omitempty"`

	// PhoneticHonorificSuffix: The honorific suffixes spelled as they
	// sound.
	PhoneticHonorificSuffix string `json:"phoneticHonorificSuffix,omitempty"`

	// PhoneticMiddleName: The middle name(s) spelled as they sound.
	PhoneticMiddleName string `json:"phoneticMiddleName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DisplayName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Name) MarshalJSON() ([]byte, error) {
	type noMethod Name
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Nickname: A person's nickname.
type Nickname struct {
	// Metadata: Metadata about the nickname.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Type: The type of the nickname.
	//
	// Possible values:
	//   "DEFAULT"
	//   "MAIDEN_NAME"
	//   "INITIALS"
	//   "GPLUS"
	//   "OTHER_NAME"
	Type string `json:"type,omitempty"`

	// Value: The nickname.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Nickname) MarshalJSON() ([]byte, error) {
	type noMethod Nickname
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Occupation: A person's occupation.
type Occupation struct {
	// Metadata: Metadata about the occupation.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The occupation; for example, `carpenter`.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Occupation) MarshalJSON() ([]byte, error) {
	type noMethod Occupation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Organization: A person's past or current organization. Overlapping
// date ranges are permitted.
type Organization struct {
	// Current: True if the organization is the person's current
	// organization; false if the organization is a past organization.
	Current bool `json:"current,omitempty"`

	// Department: The person's department at the organization.
	Department string `json:"department,omitempty"`

	// Domain: The domain name associated with the organization; for
	// example, `google.com`.
	Domain string `json:"domain,omitempty"`

	// EndDate: The end date when the person left the organization.
	EndDate *Date `json:"endDate,omitempty"`

	// FormattedType: The read-only type of the organization translated and
	// formatted in the viewer's account locale or the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// JobDescription: The person's job description at the organization.
	JobDescription string `json:"jobDescription,omitempty"`

	// Location: The location of the organization office the person works
	// at.
	Location string `json:"location,omitempty"`

	// Metadata: Metadata about the organization.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Name: The name of the organization.
	Name string `json:"name,omitempty"`

	// PhoneticName: The phonetic name of the organization.
	PhoneticName string `json:"phoneticName,omitempty"`

	// StartDate: The start date when the person joined the organization.
	StartDate *Date `json:"startDate,omitempty"`

	// Symbol: The symbol associated with the organization; for example, a
	// stock ticker symbol, abbreviation, or acronym.
	Symbol string `json:"symbol,omitempty"`

	// Title: The person's job title at the organization.
	Title string `json:"title,omitempty"`

	// Type: The type of the organization. The type can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `work` * `school`
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Current") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Organization) MarshalJSON() ([]byte, error) {
	type noMethod Organization
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Person: Information about a person merged from various data sources
// such as the authenticated user's contacts and profile data. Fields
// other than IDs, metadata, and group memberships are user-edited. Most
// fields can have multiple items. The items in a field have no
// guaranteed order, but each non-empty field is guaranteed to have
// exactly one field with `metadata.primary` set to true.
type Person struct {
	// Addresses: The person's street addresses.
	Addresses []*Address `json:"addresses,omitempty"`

	// AgeRange: The person's age range.
	//
	// Possible values:
	//   "AGE_RANGE_UNSPECIFIED"
	//   "LESS_THAN_EIGHTEEN"
	//   "EIGHTEEN_TO_TWENTY"
	//   "TWENTY_ONE_OR_OLDER"
	AgeRange string `json:"ageRange,omitempty"`

	// Biographies: The person's biographies.
	Biographies []*Biography `json:"biographies,omitempty"`

	// Birthdays: The person's birthdays.
	Birthdays []*Birthday `json:"birthdays,omitempty"`

	// BraggingRights: The person's bragging rights.
	BraggingRights []*BraggingRights `json:"braggingRights,omitempty"`

	// CoverPhotos: The person's cover photos.
	CoverPhotos []*CoverPhoto `json:"coverPhotos,omitempty"`

	// EmailAddresses: The person's email addresses.
	EmailAddresses []*EmailAddress `json:"emailAddresses,omitempty"`

	// Etag: The [HTTP entity tag](https://en.wikipedia.org/wiki/HTTP_ETag)
	// of the resource. Used for web cache validation.
	Etag string `json:"etag,omitempty"`

	// Events: The person's events.
	Events []*Event `json:"events,omitempty"`

	// Genders: The person's genders.
	Genders []*Gender `json:"genders,omitempty"`

	// ImClients: The person's instant messaging clients.
	ImClients []*ImClient `json:"imClients,omitempty"`

	// Interests: The person's interests.
	Interests []*Interest `json:"interests,omitempty"`

	// Locales: The person's locale preferences.
	Locales []*Locale `json:"locales,omitempty"`

	// Memberships: The person's group memberships.
	Memberships []*Membership `json:"memberships,omitempty"`

	// Metadata: Metadata about the person.
	Metadata *PersonMetadata `json:"metadata,omitempty"`

	// Names: The person's names.
	Names []*Name `json:"names,omitempty"`

	// Nicknames: The person's nicknames.
	Nicknames []*Nickname `json:"nicknames,omitempty"`

	// Occupations: The person's occupations.
	Occupations []*Occupation `json:"occupations,omitempty"`

	// Organizations: The person's past or current organizations.
	Organizations []*Organization `json:"organizations,omitempty"`

	// PhoneNumbers: The person's phone numbers.
	PhoneNumbers []*PhoneNumber `json:"phoneNumbers,omitempty"`

	// Photos: The person's photos.
	Photos []*Photo `json:"photos,omitempty"`

	// Relations: The person's relations.
	Relations []*Relation `json:"relations,omitempty"`

	// RelationshipInterests: The kind of relationship the person is looking
	// for.
	RelationshipInterests []*RelationshipInterest `json:"relationshipInterests,omitempty"`

	// RelationshipStatuses: The person's relationship statuses.
	RelationshipStatuses []*RelationshipStatus `json:"relationshipStatuses,omitempty"`

	// Residences: The person's residences.
	Residences []*Residence `json:"residences,omitempty"`

	// ResourceName: The resource name for the person, assigned by the
	// server. An ASCII string with a max length of 27 characters. Always
	// starts with `people/`.
	ResourceName string `json:"resourceName,omitempty"`

	// Skills: The person's skills.
	Skills []*Skill `json:"skills,omitempty"`

	// Taglines: The person's taglines.
	Taglines []*Tagline `json:"taglines,omitempty"`

	// Urls: The person's associated URLs.
	Urls []*Url `json:"urls,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Addresses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Person) MarshalJSON() ([]byte, error) {
	type noMethod Person
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PersonMetadata: Metadata about a person.
type PersonMetadata struct {
	// Deleted: True if the person resource has been deleted. Populated only
	// for [`connections.list`](/people/api/rest/v1/people.connections/list)
	// requests that include a sync token.
	Deleted bool `json:"deleted,omitempty"`

	// ObjectType: The type of the person object.
	//
	// Possible values:
	//   "OBJECT_TYPE_UNSPECIFIED"
	//   "PERSON"
	//   "PAGE"
	ObjectType string `json:"objectType,omitempty"`

	// PreviousResourceNames: Any former resource names this person has had.
	// Populated only for
	// [`connections.list`](/people/api/rest/v1/people.connections/list)
	// requests that include a sync token. The resource name may change when
	// adding or removing fields that link a contact and profile such as a
	// verified email, verified phone number, or profile URL.
	PreviousResourceNames []string `json:"previousResourceNames,omitempty"`

	// Sources: The sources of data for the person.
	Sources []*Source `json:"sources,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Deleted") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonMetadata) MarshalJSON() ([]byte, error) {
	type noMethod PersonMetadata
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PersonResponse: The response for a single person
type PersonResponse struct {
	// HttpStatusCode: [HTTP 1.1 status
	// code](http://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html).
	HttpStatusCode int64 `json:"httpStatusCode,omitempty"`

	// Person: The person.
	Person *Person `json:"person,omitempty"`

	// RequestedResourceName: The original requested resource name. May be
	// different than the resource name on the returned person. The resource
	// name can change when adding or removing fields that link a contact
	// and profile such as a verified email, verified phone number, or a
	// profile URL.
	RequestedResourceName string `json:"requestedResourceName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "HttpStatusCode") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PersonResponse) MarshalJSON() ([]byte, error) {
	type noMethod PersonResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PhoneNumber: A person's phone number.
type PhoneNumber struct {
	// CanonicalForm: The read-only canonicalized [ITU-T
	// E.164](https://law.resource.org/pub/us/cfr/ibr/004/itu-t.E.164.1.2008.
	// pdf) form of the phone number.
	CanonicalForm string `json:"canonicalForm,omitempty"`

	// FormattedType: The read-only type of the phone number translated and
	// formatted in the viewer's account locale or the the `Accept-Language`
	// HTTP header locale.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the phone number.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Type: The type of the phone number. The type can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `home` * `work` * `mobile` * `homeFax` * `workFax` *
	// `otherFax` * `pager` * `workMobile` * `workPager` * `main` *
	// `googleVoice` * `other`
	Type string `json:"type,omitempty"`

	// Value: The phone number.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CanonicalForm") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PhoneNumber) MarshalJSON() ([]byte, error) {
	type noMethod PhoneNumber
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Photo: A person's photo. A picture shown next to the person's name to
// help others recognize the person.
type Photo struct {
	// Metadata: Metadata about the photo.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Url: The URL of the photo.
	Url string `json:"url,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Photo) MarshalJSON() ([]byte, error) {
	type noMethod Photo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Relation: A person's relation to another person.
type Relation struct {
	// FormattedType: The type of the relation translated and formatted in
	// the viewer's account locale or the locale specified in the
	// Accept-Language HTTP header.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the relation.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Person: The name of the other person this relation refers to.
	Person string `json:"person,omitempty"`

	// Type: The person's relation to the other person. The type can be
	// custom or predefined. Possible values include, but are not limited
	// to, the following values: * `spouse` * `child` * `mother` * `father`
	// * `parent` * `brother` * `sister` * `friend` * `relative` *
	// `domesticPartner` * `manager` * `assistant` * `referredBy` *
	// `partner`
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Relation) MarshalJSON() ([]byte, error) {
	type noMethod Relation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// RelationshipInterest: The kind of relationship the person is looking
// for.
type RelationshipInterest struct {
	// FormattedValue: The value of the relationship interest translated and
	// formatted in the viewer's account locale or the locale specified in
	// the Accept-Language HTTP header.
	FormattedValue string `json:"formattedValue,omitempty"`

	// Metadata: Metadata about the relationship interest.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The kind of relationship the person is looking for. The value
	// can be custom or predefined. Possible values include, but are not
	// limited to, the following values: * `friend` * `date` *
	// `relationship` * `networking`
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RelationshipInterest) MarshalJSON() ([]byte, error) {
	type noMethod RelationshipInterest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// RelationshipStatus: A person's relationship status.
type RelationshipStatus struct {
	// FormattedValue: The read-only value of the relationship status
	// translated and formatted in the viewer's account locale or the
	// `Accept-Language` HTTP header locale.
	FormattedValue string `json:"formattedValue,omitempty"`

	// Metadata: Metadata about the relationship status.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The relationship status. The value can be custom or
	// predefined. Possible values include, but are not limited to, the
	// following: * `single` * `inARelationship` * `engaged` * `married` *
	// `itsComplicated` * `openRelationship` * `widowed` *
	// `inDomesticPartnership` * `inCivilUnion`
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RelationshipStatus) MarshalJSON() ([]byte, error) {
	type noMethod RelationshipStatus
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Residence: A person's past or current residence.
type Residence struct {
	// Current: True if the residence is the person's current residence;
	// false if the residence is a past residence.
	Current bool `json:"current,omitempty"`

	// Metadata: Metadata about the residence.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The address of the residence.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Current") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Residence) MarshalJSON() ([]byte, error) {
	type noMethod Residence
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Skill: A skill that the person has.
type Skill struct {
	// Metadata: Metadata about the skill.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The skill; for example, `underwater basket weaving`.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Skill) MarshalJSON() ([]byte, error) {
	type noMethod Skill
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Source: The source of a field.
type Source struct {
	// Id: A unique identifier within the source type generated by the
	// server.
	Id string `json:"id,omitempty"`

	// Type: The source type.
	//
	// Possible values:
	//   "OTHER"
	//   "ACCOUNT"
	//   "PROFILE"
	//   "DOMAIN_PROFILE"
	//   "CONTACT"
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Id") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Source) MarshalJSON() ([]byte, error) {
	type noMethod Source
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Tagline: A brief one-line description of the person.
type Tagline struct {
	// Metadata: Metadata about the tagline.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Value: The tagline.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Tagline) MarshalJSON() ([]byte, error) {
	type noMethod Tagline
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Url: A person's associated URLs.
type Url struct {
	// FormattedType: The read-only type of the URL translated and formatted
	// in the viewer's account locale or the `Accept-Language` HTTP header
	// locale.
	FormattedType string `json:"formattedType,omitempty"`

	// Metadata: Metadata about the URL.
	Metadata *FieldMetadata `json:"metadata,omitempty"`

	// Type: The type of the URL. The type can be custom or predefined.
	// Possible values include, but are not limited to, the following: *
	// `home` * `work` * `blog` * `profile` * `homePage` * `ftp` *
	// `reservations` * `appInstallPage`: website for a Google+ application.
	// * `other`
	Type string `json:"type,omitempty"`

	// Value: The URL.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "FormattedType") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Url) MarshalJSON() ([]byte, error) {
	type noMethod Url
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "people.people.get":

type PeopleGetCall struct {
	s            *Service
	resourceName string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Provides information about a person resource for a resource
// name. Use `people/me` to indicate the authenticated user.
func (r *PeopleService) Get(resourceName string) *PeopleGetCall {
	c := &PeopleGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resourceName = resourceName
	return c
}

// RequestMaskIncludeField sets the optional parameter
// "requestMask.includeField": Comma-separated list of fields to be
// included in the response. Omitting this field will include all
// fields. Each path should start with `person.`: for example,
// `person.names` or `person.photos`.
func (c *PeopleGetCall) RequestMaskIncludeField(requestMaskIncludeField string) *PeopleGetCall {
	c.urlParams_.Set("requestMask.includeField", requestMaskIncludeField)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleGetCall) Fields(s ...googleapi.Field) *PeopleGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleGetCall) IfNoneMatch(entityTag string) *PeopleGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PeopleGetCall) Context(ctx context.Context) *PeopleGetCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+resourceName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resourceName": c.resourceName,
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

// Do executes the "people.people.get" call.
// Exactly one of *Person or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Person.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *PeopleGetCall) Do(opts ...googleapi.CallOption) (*Person, error) {
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
	ret := &Person{
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
	//   "description": "Provides information about a person resource for a resource name. Use `people/me` to indicate the authenticated user.",
	//   "httpMethod": "GET",
	//   "id": "people.people.get",
	//   "parameterOrder": [
	//     "resourceName"
	//   ],
	//   "parameters": {
	//     "requestMask.includeField": {
	//       "description": "Comma-separated list of fields to be included in the response. Omitting this field will include all fields. Each path should start with `person.`: for example, `person.names` or `person.photos`.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "resourceName": {
	//       "description": "The resource name of the person to provide information about. - To get information about the authenticated user, specify `people/me`. - To get information about any user, specify the resource name that identifies the user, such as the resource names returned by [`people.connections.list`](/people/api/rest/v1/people.connections/list).",
	//       "location": "path",
	//       "pattern": "^people/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+resourceName}",
	//   "response": {
	//     "$ref": "Person"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/contacts",
	//     "https://www.googleapis.com/auth/contacts.readonly",
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/user.addresses.read",
	//     "https://www.googleapis.com/auth/user.birthday.read",
	//     "https://www.googleapis.com/auth/user.emails.read",
	//     "https://www.googleapis.com/auth/user.phonenumbers.read",
	//     "https://www.googleapis.com/auth/userinfo.email",
	//     "https://www.googleapis.com/auth/userinfo.profile"
	//   ]
	// }

}

// method id "people.people.getBatchGet":

type PeopleGetBatchGetCall struct {
	s            *Service
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// GetBatchGet: Provides information about a list of specific people by
// specifying a list of requested resource names. Use `people/me` to
// indicate the authenticated user.
func (r *PeopleService) GetBatchGet() *PeopleGetBatchGetCall {
	c := &PeopleGetBatchGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	return c
}

// RequestMaskIncludeField sets the optional parameter
// "requestMask.includeField": Comma-separated list of fields to be
// included in the response. Omitting this field will include all
// fields. Each path should start with `person.`: for example,
// `person.names` or `person.photos`.
func (c *PeopleGetBatchGetCall) RequestMaskIncludeField(requestMaskIncludeField string) *PeopleGetBatchGetCall {
	c.urlParams_.Set("requestMask.includeField", requestMaskIncludeField)
	return c
}

// ResourceNames sets the optional parameter "resourceNames": The
// resource name, such as one returned by
// [`people.connections.list`](/people/api/rest/v1/people.connections/lis
// t), of one of the people to provide information about. You can
// include this parameter up to 50 times in one request.
func (c *PeopleGetBatchGetCall) ResourceNames(resourceNames ...string) *PeopleGetBatchGetCall {
	c.urlParams_.SetMulti("resourceNames", append([]string{}, resourceNames...))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleGetBatchGetCall) Fields(s ...googleapi.Field) *PeopleGetBatchGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleGetBatchGetCall) IfNoneMatch(entityTag string) *PeopleGetBatchGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PeopleGetBatchGetCall) Context(ctx context.Context) *PeopleGetBatchGetCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleGetBatchGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/people:batchGet")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ifNoneMatch_ != "" {
		req.Header.Set("If-None-Match", c.ifNoneMatch_)
	}
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "people.people.getBatchGet" call.
// Exactly one of *GetPeopleResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *GetPeopleResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PeopleGetBatchGetCall) Do(opts ...googleapi.CallOption) (*GetPeopleResponse, error) {
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
	ret := &GetPeopleResponse{
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
	//   "description": "Provides information about a list of specific people by specifying a list of requested resource names. Use `people/me` to indicate the authenticated user.",
	//   "httpMethod": "GET",
	//   "id": "people.people.getBatchGet",
	//   "parameters": {
	//     "requestMask.includeField": {
	//       "description": "Comma-separated list of fields to be included in the response. Omitting this field will include all fields. Each path should start with `person.`: for example, `person.names` or `person.photos`.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "resourceNames": {
	//       "description": "The resource name, such as one returned by [`people.connections.list`](/people/api/rest/v1/people.connections/list), of one of the people to provide information about. You can include this parameter up to 50 times in one request.",
	//       "location": "query",
	//       "repeated": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/people:batchGet",
	//   "response": {
	//     "$ref": "GetPeopleResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/contacts",
	//     "https://www.googleapis.com/auth/contacts.readonly",
	//     "https://www.googleapis.com/auth/plus.login",
	//     "https://www.googleapis.com/auth/user.addresses.read",
	//     "https://www.googleapis.com/auth/user.birthday.read",
	//     "https://www.googleapis.com/auth/user.emails.read",
	//     "https://www.googleapis.com/auth/user.phonenumbers.read",
	//     "https://www.googleapis.com/auth/userinfo.email",
	//     "https://www.googleapis.com/auth/userinfo.profile"
	//   ]
	// }

}

// method id "people.people.connections.list":

type PeopleConnectionsListCall struct {
	s            *Service
	resourceName string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Provides a list of the authenticated user's contacts merged
// with any linked profiles.
func (r *PeopleConnectionsService) List(resourceName string) *PeopleConnectionsListCall {
	c := &PeopleConnectionsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resourceName = resourceName
	return c
}

// PageSize sets the optional parameter "pageSize": The number of
// connections to include in the response. Valid values are between 1
// and 500, inclusive. Defaults to 100.
func (c *PeopleConnectionsListCall) PageSize(pageSize int64) *PeopleConnectionsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The token of the
// page to be returned.
func (c *PeopleConnectionsListCall) PageToken(pageToken string) *PeopleConnectionsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// RequestMaskIncludeField sets the optional parameter
// "requestMask.includeField": Comma-separated list of fields to be
// included in the response. Omitting this field will include all
// fields. Each path should start with `person.`: for example,
// `person.names` or `person.photos`.
func (c *PeopleConnectionsListCall) RequestMaskIncludeField(requestMaskIncludeField string) *PeopleConnectionsListCall {
	c.urlParams_.Set("requestMask.includeField", requestMaskIncludeField)
	return c
}

// SortOrder sets the optional parameter "sortOrder": The order in which
// the connections should be sorted. Defaults to
// `LAST_MODIFIED_ASCENDING`.
//
// Possible values:
//   "LAST_MODIFIED_ASCENDING"
//   "FIRST_NAME_ASCENDING"
//   "LAST_NAME_ASCENDING"
func (c *PeopleConnectionsListCall) SortOrder(sortOrder string) *PeopleConnectionsListCall {
	c.urlParams_.Set("sortOrder", sortOrder)
	return c
}

// SyncToken sets the optional parameter "syncToken": A sync token,
// returned by a previous call to `people.connections.list`. Only
// resources changed since the sync token was created are returned.
func (c *PeopleConnectionsListCall) SyncToken(syncToken string) *PeopleConnectionsListCall {
	c.urlParams_.Set("syncToken", syncToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PeopleConnectionsListCall) Fields(s ...googleapi.Field) *PeopleConnectionsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PeopleConnectionsListCall) IfNoneMatch(entityTag string) *PeopleConnectionsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PeopleConnectionsListCall) Context(ctx context.Context) *PeopleConnectionsListCall {
	c.ctx_ = ctx
	return c
}

func (c *PeopleConnectionsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+resourceName}/connections")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resourceName": c.resourceName,
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

// Do executes the "people.people.connections.list" call.
// Exactly one of *ListConnectionsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListConnectionsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PeopleConnectionsListCall) Do(opts ...googleapi.CallOption) (*ListConnectionsResponse, error) {
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
	ret := &ListConnectionsResponse{
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
	//   "description": "Provides a list of the authenticated user's contacts merged with any linked profiles.",
	//   "httpMethod": "GET",
	//   "id": "people.people.connections.list",
	//   "parameterOrder": [
	//     "resourceName"
	//   ],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The number of connections to include in the response. Valid values are between 1 and 500, inclusive. Defaults to 100.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The token of the page to be returned.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "requestMask.includeField": {
	//       "description": "Comma-separated list of fields to be included in the response. Omitting this field will include all fields. Each path should start with `person.`: for example, `person.names` or `person.photos`.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "resourceName": {
	//       "description": "The resource name to return connections for. Only `people/me` is valid.",
	//       "location": "path",
	//       "pattern": "^people/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "sortOrder": {
	//       "description": "The order in which the connections should be sorted. Defaults to `LAST_MODIFIED_ASCENDING`.",
	//       "enum": [
	//         "LAST_MODIFIED_ASCENDING",
	//         "FIRST_NAME_ASCENDING",
	//         "LAST_NAME_ASCENDING"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "syncToken": {
	//       "description": "A sync token, returned by a previous call to `people.connections.list`. Only resources changed since the sync token was created are returned.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+resourceName}/connections",
	//   "response": {
	//     "$ref": "ListConnectionsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/contacts",
	//     "https://www.googleapis.com/auth/contacts.readonly"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *PeopleConnectionsListCall) Pages(ctx context.Context, f func(*ListConnectionsResponse) error) error {
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
