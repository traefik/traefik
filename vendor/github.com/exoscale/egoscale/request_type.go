package egoscale

import (
	"net/url"
)

// Command represents a generic request
type Command interface {
	Response() interface{}
}

// AsyncCommand represents a async request
type AsyncCommand interface {
	Command
	AsyncResponse() interface{}
}

// ListCommand represents a listing request
type ListCommand interface {
	Listable
	Command
	// SetPage defines the current pages
	SetPage(int)
	// SetPageSize defines the size of the page
	SetPageSize(int)
	// Each reads the data from the response and feeds channels, and returns true if we are on the last page
	Each(interface{}, IterateItemFunc)
}

// onBeforeHook represents an action to be done on the params before sending them
//
// This little took helps with issue of relying on JSON serialization logic only.
// `omitempty` may make sense in some cases but not all the time.
type onBeforeHook interface {
	onBeforeSend(params url.Values) error
}

// CommandInfo represents the meta data related to a Command
type CommandInfo struct {
	Name        string
	Description string
	RootOnly    bool
}

// JobStatusType represents the status of a Job
type JobStatusType int

//go:generate stringer -type JobStatusType
const (
	// Pending represents a job in progress
	Pending JobStatusType = iota
	// Success represents a successfully completed job
	Success
	// Failure represents a job that has failed to complete
	Failure
)

// ErrorCode represents the CloudStack ApiErrorCode enum
//
// See: https://github.com/apache/cloudstack/blob/master/api/src/main/java/org/apache/cloudstack/api/ApiErrorCode.java
type ErrorCode int

//go:generate stringer -type ErrorCode
const (
	// Unauthorized represents ... (TODO)
	Unauthorized ErrorCode = 401
	// MethodNotAllowed represents ... (TODO)
	MethodNotAllowed ErrorCode = 405
	// UnsupportedActionError represents ... (TODO)
	UnsupportedActionError ErrorCode = 422
	// APILimitExceeded represents ... (TODO)
	APILimitExceeded ErrorCode = 429
	// MalformedParameterError represents ... (TODO)
	MalformedParameterError ErrorCode = 430
	// ParamError represents ... (TODO)
	ParamError ErrorCode = 431

	// InternalError represents a server error
	InternalError ErrorCode = 530
	// AccountError represents ... (TODO)
	AccountError ErrorCode = 531
	// AccountResourceLimitError represents ... (TODO)
	AccountResourceLimitError ErrorCode = 532
	// InsufficientCapacityError represents ... (TODO)
	InsufficientCapacityError ErrorCode = 533
	// ResourceUnavailableError represents ... (TODO)
	ResourceUnavailableError ErrorCode = 534
	// ResourceAllocationError represents ... (TODO)
	ResourceAllocationError ErrorCode = 535
	// ResourceInUseError represents ... (TODO)
	ResourceInUseError ErrorCode = 536
	// NetworkRuleConflictError represents ... (TODO)
	NetworkRuleConflictError ErrorCode = 537
)

// CSErrorCode represents the CloudStack CSExceptionErrorCode enum
//
// See: https://github.com/apache/cloudstack/blob/master/utils/src/main/java/com/cloud/utils/exception/CSExceptionErrorCode.java
type CSErrorCode int

//go:generate stringer -type CSErrorCode
const (
	// CloudRuntimeException ... (TODO)
	CloudRuntimeException CSErrorCode = 4250
	// ExecutionException ... (TODO)
	ExecutionException CSErrorCode = 4260
	// HypervisorVersionChangedException ... (TODO)
	HypervisorVersionChangedException CSErrorCode = 4265
	// CloudException ... (TODO)
	CloudException CSErrorCode = 4275
	// AccountLimitException ... (TODO)
	AccountLimitException CSErrorCode = 4280
	// AgentUnavailableException ... (TODO)
	AgentUnavailableException CSErrorCode = 4285
	// CloudAuthenticationException ... (TODO)
	CloudAuthenticationException CSErrorCode = 4290
	// ConcurrentOperationException ... (TODO)
	ConcurrentOperationException CSErrorCode = 4300
	// ConflictingNetworksException ... (TODO)
	ConflictingNetworkSettingsException CSErrorCode = 4305
	// DiscoveredWithErrorException ... (TODO)
	DiscoveredWithErrorException CSErrorCode = 4310
	// HAStateException ... (TODO)
	HAStateException CSErrorCode = 4315
	// InsufficientAddressCapacityException ... (TODO)
	InsufficientAddressCapacityException CSErrorCode = 4320
	// InsufficientCapacityException ... (TODO)
	InsufficientCapacityException CSErrorCode = 4325
	// InsufficientNetworkCapacityException ... (TODO)
	InsufficientNetworkCapacityException CSErrorCode = 4330
	// InsufficientServerCapaticyException ... (TODO)
	InsufficientServerCapacityException CSErrorCode = 4335
	// InsufficientStorageCapacityException ... (TODO)
	InsufficientStorageCapacityException CSErrorCode = 4340
	// InternalErrorException ... (TODO)
	InternalErrorException CSErrorCode = 4345
	// InvalidParameterValueException ... (TODO)
	InvalidParameterValueException CSErrorCode = 4350
	// ManagementServerException ... (TODO)
	ManagementServerException CSErrorCode = 4355
	// NetworkRuleConflictException  ... (TODO)
	NetworkRuleConflictException CSErrorCode = 4360
	// PermissionDeniedException ... (TODO)
	PermissionDeniedException CSErrorCode = 4365
	// ResourceAllocationException ... (TODO)
	ResourceAllocationException CSErrorCode = 4370
	// ResourceInUseException ... (TODO)
	ResourceInUseException CSErrorCode = 4375
	// ResourceUnavailableException ... (TODO)
	ResourceUnavailableException CSErrorCode = 4380
	// StorageUnavailableException ... (TODO)
	StorageUnavailableException CSErrorCode = 4385
	// UnsupportedServiceException ... (TODO)
	UnsupportedServiceException CSErrorCode = 4390
	// VirtualMachineMigrationException ... (TODO)
	VirtualMachineMigrationException CSErrorCode = 4395
	// AsyncCommandQueued ... (TODO)
	AsyncCommandQueued CSErrorCode = 4540
	// RequestLimitException ... (TODO)
	RequestLimitException CSErrorCode = 4545
	// ServerAPIException ... (TODO)
	ServerAPIException CSErrorCode = 9999
)

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	CSErrorCode CSErrorCode `json:"cserrorcode"`
	ErrorCode   ErrorCode   `json:"errorcode"`
	ErrorText   string      `json:"errortext"`
	UUIDList    []UUIDItem  `json:"uuidList,omitempty"` // uuid*L*ist is not a typo
}

// UUIDItem represents an item of the UUIDList part of an ErrorResponse
type UUIDItem struct {
	Description      string `json:"description,omitempty"`
	SerialVersionUID int64  `json:"serialVersionUID,omitempty"`
	UUID             string `json:"uuid"`
}

// BooleanResponse represents a boolean response (usually after a deletion)
type BooleanResponse struct {
	DisplayText string `json:"displaytext,omitempty"`
	Success     bool   `json:"success"`
}
