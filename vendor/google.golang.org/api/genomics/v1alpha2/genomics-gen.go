// Package genomics provides access to the Genomics API.
//
// See https://cloud.google.com/genomics/
//
// Usage example:
//
//   import "google.golang.org/api/genomics/v1alpha2"
//   ...
//   genomicsService, err := genomics.New(oauthHttpClient)
package genomics // import "google.golang.org/api/genomics/v1alpha2"

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

const apiId = "genomics:v1alpha2"
const apiName = "genomics"
const apiVersion = "v1alpha2"
const basePath = "https://genomics.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	// View and manage Genomics data
	GenomicsScope = "https://www.googleapis.com/auth/genomics"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Operations = NewOperationsService(s)
	s.Pipelines = NewPipelinesService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Operations *OperationsService

	Pipelines *PipelinesService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewOperationsService(s *Service) *OperationsService {
	rs := &OperationsService{s: s}
	return rs
}

type OperationsService struct {
	s *Service
}

func NewPipelinesService(s *Service) *PipelinesService {
	rs := &PipelinesService{s: s}
	return rs
}

type PipelinesService struct {
	s *Service
}

// CancelOperationRequest: The request message for
// Operations.CancelOperation.
type CancelOperationRequest struct {
}

// ControllerConfig: Stores the information that the controller will
// fetch from the server in order to run. Should only be used by VMs
// created by the Pipelines Service and not by end users.
type ControllerConfig struct {
	Cmd string `json:"cmd,omitempty"`

	Disks map[string]string `json:"disks,omitempty"`

	GcsLogPath string `json:"gcsLogPath,omitempty"`

	GcsSinks map[string]RepeatedString `json:"gcsSinks,omitempty"`

	GcsSources map[string]RepeatedString `json:"gcsSources,omitempty"`

	Image string `json:"image,omitempty"`

	MachineType string `json:"machineType,omitempty"`

	Vars map[string]string `json:"vars,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Cmd") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ControllerConfig) MarshalJSON() ([]byte, error) {
	type noMethod ControllerConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Disk: A Google Compute Engine disk resource specification.
type Disk struct {
	// AutoDelete: Specifies whether or not to delete the disk when the
	// pipeline completes. This field is applicable only for newly created
	// disks. See
	// https://cloud.google.com/compute/docs/reference/latest/instances#resource for more details. Optional. At create time means that an auto delete disk may be used. At run time, means it should be used. Cannot be true at run time if false at create
	// time.
	AutoDelete bool `json:"autoDelete,omitempty"`

	// MountPoint: Required at create time and cannot be overridden at run
	// time. Specifies the path in the docker container where files on this
	// disk should be located. For example, if `mountPoint` is `/mnt/disk`,
	// and the parameter has `localPath` `inputs/file.txt`, the docker
	// container can access the data at `/mnt/disk/inputs/file.txt`.
	MountPoint string `json:"mountPoint,omitempty"`

	// Name: Required. The name of the disk that can be used in the pipeline
	// parameters. Must be 1 - 63 characters. The name "boot" is reserved
	// for system use.
	Name string `json:"name,omitempty"`

	// ReadOnly: Specifies how a sourced-base persistent disk will be
	// mounted. See
	// https://cloud.google.com/compute/docs/disks/persistent-disks#use_multi_instances for more details. Can only be set at create
	// time.
	ReadOnly bool `json:"readOnly,omitempty"`

	// SizeGb: The size of the disk. This field is not applicable for local
	// SSD.
	SizeGb int64 `json:"sizeGb,omitempty"`

	// Source: The full or partial URL of the persistent disk to attach. See
	// https://cloud.google.com/compute/docs/reference/latest/instances#resource and https://cloud.google.com/compute/docs/disks/persistent-disks#snapshots for more
	// details.
	Source string `json:"source,omitempty"`

	// Type: Required. The type of the disk to create.
	//
	// Possible values:
	//   "TYPE_UNSPECIFIED"
	//   "PERSISTENT_HDD"
	//   "PERSISTENT_SSD"
	//   "LOCAL_SSD"
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AutoDelete") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Disk) MarshalJSON() ([]byte, error) {
	type noMethod Disk
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DockerExecutor: The Docker execuctor specification.
type DockerExecutor struct {
	// Cmd: Required. The command string to run. Parameters that do not have
	// `localCopy` specified should be used as environment variables, while
	// those that do can be accessed at the defined paths.
	Cmd string `json:"cmd,omitempty"`

	// ImageName: Required. Image name from either Docker Hub or Google
	// Container Repository. Users that run pipelines must have READ access
	// to the image.
	ImageName string `json:"imageName,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Cmd") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DockerExecutor) MarshalJSON() ([]byte, error) {
	type noMethod DockerExecutor
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
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

// ImportReadGroupSetsResponse: The read group set import response.
type ImportReadGroupSetsResponse struct {
	// ReadGroupSetIds: IDs of the read group sets that were created.
	ReadGroupSetIds []string `json:"readGroupSetIds,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ReadGroupSetIds") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImportReadGroupSetsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ImportReadGroupSetsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ImportVariantsResponse: The variant data import response.
type ImportVariantsResponse struct {
	// CallSetIds: IDs of the call sets created during the import.
	CallSetIds []string `json:"callSetIds,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CallSetIds") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImportVariantsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ImportVariantsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ListOperationsResponse: The response message for
// Operations.ListOperations.
type ListOperationsResponse struct {
	// NextPageToken: The standard List next-page token.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Operations: A list of operations that matches the specified filter in
	// the request.
	Operations []*Operation `json:"operations,omitempty"`

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

func (s *ListOperationsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListOperationsResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ListPipelinesResponse: The response of ListPipelines. Contains at
// most `pageSize` pipelines. If it contains `pageSize` pipelines, and
// more pipelines exist, then `nextPageToken` will be populated and
// should be used as the `pageToken` argument to a subsequent
// ListPipelines request.
type ListPipelinesResponse struct {
	// NextPageToken: The token to use to get the next page of results.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// Pipelines: The matched pipelines.
	Pipelines []*Pipeline `json:"pipelines,omitempty"`

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

func (s *ListPipelinesResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListPipelinesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LocalCopy: LocalCopy defines how a remote file should be copied to
// and from the VM.
type LocalCopy struct {
	// Disk: Required. The name of the disk where this parameter is located.
	// Can be the name of one of the disks specified in the Resources field,
	// or "boot", which represents the Docker instance's boot disk and has a
	// mount point of `/`.
	Disk string `json:"disk,omitempty"`

	// Path: Required. The path within the user's docker container where
	// this input should be localized to and from, relative to the specified
	// disk's mount point. For example: file.txt,
	Path string `json:"path,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Disk") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LocalCopy) MarshalJSON() ([]byte, error) {
	type noMethod LocalCopy
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LoggingOptions: The logging options for the pipeline run.
type LoggingOptions struct {
	// GcsPath: The location in Google Cloud Storage to which the pipeline
	// logs will be copied. Can be specified as a fully qualified directory
	// path, in which case logs will be output with a unique identifier as
	// the filename in that directory, or as a fully specified path, which
	// must end in `.log`, in which case that path will be used, and the
	// user must ensure that logs are not overwritten. Stdout and stderr
	// logs from the run are also generated and output as `-stdout.log` and
	// `-stderr.log`.
	GcsPath string `json:"gcsPath,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GcsPath") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LoggingOptions) MarshalJSON() ([]byte, error) {
	type noMethod LoggingOptions
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Operation: This resource represents a long-running operation that is
// the result of a network API call.
type Operation struct {
	// Done: If the value is `false`, it means the operation is still in
	// progress. If true, the operation is completed, and either `error` or
	// `response` is available.
	Done bool `json:"done,omitempty"`

	// Error: The error result of the operation in case of failure.
	Error *Status `json:"error,omitempty"`

	// Metadata: An OperationMetadata object. This will always be returned
	// with the Operation.
	Metadata OperationMetadata `json:"metadata,omitempty"`

	// Name: The server-assigned name, which is only unique within the same
	// service that originally returns it. For example:
	// `operations/CJHU7Oi_ChDrveSpBRjfuL-qzoWAgEw`
	Name string `json:"name,omitempty"`

	// Response: If importing ReadGroupSets, an ImportReadGroupSetsResponse
	// is returned. If importing Variants, an ImportVariantsResponse is
	// returned. For exports, an empty response is returned.
	Response OperationResponse `json:"response,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Done") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Operation) MarshalJSON() ([]byte, error) {
	type noMethod Operation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type OperationMetadata interface{}

type OperationResponse interface{}

// OperationEvent: An event that occurred during an Operation.
type OperationEvent struct {
	// Description: Required description of event.
	Description string `json:"description,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *OperationEvent) MarshalJSON() ([]byte, error) {
	type noMethod OperationEvent
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// OperationMetadata1: Metadata describing an Operation.
type OperationMetadata1 struct {
	// CreateTime: The time at which the job was submitted to the Genomics
	// service.
	CreateTime string `json:"createTime,omitempty"`

	// Events: Optional event messages that were generated during the job's
	// execution. This also contains any warnings that were generated during
	// import or export.
	Events []*OperationEvent `json:"events,omitempty"`

	// ProjectId: The Google Cloud Project in which the job is scoped.
	ProjectId string `json:"projectId,omitempty"`

	// Request: The original request that started the operation. Note that
	// this will be in current version of the API. If the operation was
	// started with v1beta2 API and a GetOperation is performed on v1 API, a
	// v1 request will be returned.
	Request OperationMetadataRequest `json:"request,omitempty"`

	// ForceSendFields is a list of field names (e.g. "CreateTime") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *OperationMetadata1) MarshalJSON() ([]byte, error) {
	type noMethod OperationMetadata1
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type OperationMetadataRequest interface{}

// Pipeline: The pipeline object. Represents a transformation from a set
// of input parameters to a set of output parameters. The transformation
// is defined as a docker image and command to run within that image.
// Each pipeline is run on a Google Compute Engine VM. A pipeline can be
// created with the `create` method and then later run with the `run`
// method, or a pipeline can be defined and run all at once with the
// `run` method.
type Pipeline struct {
	// Description: Optional. User-specified description.
	Description string `json:"description,omitempty"`

	// Docker: Specifies the docker run information.
	Docker *DockerExecutor `json:"docker,omitempty"`

	// InputParameters: Input parameters of the pipeline.
	InputParameters []*PipelineParameter `json:"inputParameters,omitempty"`

	// Name: Required. A user specified pipeline name that does not have to
	// be unique. This name can be used for filtering Pipelines in
	// ListPipelines.
	Name string `json:"name,omitempty"`

	// OutputParameters: Output parameters of the pipeline.
	OutputParameters []*PipelineParameter `json:"outputParameters,omitempty"`

	// PipelineId: Unique pipeline id that is generated by the service when
	// CreatePipeline is called. Cannot be specified in the Pipeline used in
	// the CreatePipelineRequest, and will be populated in the response to
	// CreatePipeline and all subsequent Get and List calls. Indicates that
	// the service has registered this pipeline.
	PipelineId string `json:"pipelineId,omitempty"`

	// ProjectId: Required. The project in which to create the pipeline. The
	// caller must have WRITE access.
	ProjectId string `json:"projectId,omitempty"`

	// Resources: Required. Specifies resource requirements for the pipeline
	// run. Required fields: * minimumCpuCores * minimumRamGb
	Resources *PipelineResources `json:"resources,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Pipeline) MarshalJSON() ([]byte, error) {
	type noMethod Pipeline
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PipelineParameter: Parameters facilitate setting and delivering data
// into the pipeline's execution environment. They are defined at create
// time, with optional defaults, and can be overridden at run time. If
// `localCopy` is unset, then the parameter specifies a string that is
// passed as-is into the pipeline, as the value of the environment
// variable with the given name. A default value can be optionally
// specified at create time. The default can be overridden at run time
// using the inputs map. If no default is given, a value must be
// supplied at runtime. If `localCopy` is defined, then the parameter
// specifies a data source or sink, both in Google Cloud Storage and on
// the Docker container where the pipeline computation is run. At run
// time, the Google Cloud Storage paths can be overridden if a default
// was provided at create time, or must be set otherwise. The pipeline
// runner should add a key/value pair to either the inputs or outputs
// map. The indicated data copies will be carried out before/after
// pipeline execution, just as if the corresponding arguments were
// provided to `gsutil cp`. For example: Given the following
// `PipelineParameter`, specified in the `inputParameters` list: ```
// {name: "input_file", localCopy: {path: "file.txt", disk: "pd1"}} ```
// where `disk` is defined in the `PipelineResources` object as: ```
// {name: "pd1", mountPoint: "/mnt/disk/"} ``` We create a disk named
// `pd1`, mount it on the host VM, and map `/mnt/pd1` to `/mnt/disk` in
// the docker container. At runtime, an entry for `input_file` would be
// required in the inputs map, such as: ``` inputs["input_file"] =
// "gs://my-bucket/bar.txt" ``` This would generate the following gsutil
// call: ``` gsutil cp gs://my-bucket/bar.txt /mnt/pd1/file.txt ``` The
// file `/mnt/pd1/file.txt` maps to `/mnt/disk/file.txt` in the Docker
// container. Note that we do not use `cp -r`, so for inputs, the Google
// Cloud Storage path (runtime value) must be a file or a glob, while
// the local path must be a file or a directory, respectively. For
// outputs, the direction of the copy is reversed: ``` gsutil cp
// /mnt/disk/file.txt gs://my-bucket/bar.txt ``` Again note that there
// is no `-r`, so the Google Cloud Storage path (runtime value) must be
// a file or a directory, while the local path can be a file or a glob,
// respectively. One restriction, due to docker limitations, is that for
// outputs that are found on the boot disk, the local path cannot be a
// glob and must be a file.
type PipelineParameter struct {
	// DefaultValue: The default value for this parameter. Can be overridden
	// at runtime. If `localCopy` is present, then this must be a Google
	// Cloud Storage path beginning with `gs://`.
	DefaultValue string `json:"defaultValue,omitempty"`

	// Description: Optional. Human-readable description.
	Description string `json:"description,omitempty"`

	// LocalCopy: If present, this parameter is marked for copying to and
	// from the VM. `LocalCopy` indicates where on the VM the file should
	// be. The value given to this parameter (either at runtime or using
	// `defaultValue`) must be the remote path where the file should be.
	LocalCopy *LocalCopy `json:"localCopy,omitempty"`

	// Name: Required. Name of the parameter - the pipeline runner uses this
	// string as the key to the input and output maps in RunPipeline.
	Name string `json:"name,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DefaultValue") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PipelineParameter) MarshalJSON() ([]byte, error) {
	type noMethod PipelineParameter
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PipelineResources: The system resources for the pipeline run.
type PipelineResources struct {
	// Disks: Disks to attach.
	Disks []*Disk `json:"disks,omitempty"`

	// MinimumCpuCores: Required at create time; optional at run time. The
	// minimum number of cores to use.
	MinimumCpuCores int64 `json:"minimumCpuCores,omitempty"`

	// MinimumRamGb: Required at create time; optional at run time. The
	// minimum amount of RAM to use.
	MinimumRamGb float64 `json:"minimumRamGb,omitempty"`

	// Preemptible: Optional. At create time means that preemptible machines
	// may be used for the run. At run time, means they should be used.
	// Cannot be true at run time if false at create time.
	Preemptible bool `json:"preemptible,omitempty"`

	// Zones: List of Google Compute Engine availability zones to which
	// resource creation will restricted. If empty, any zone may be chosen.
	Zones []string `json:"zones,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Disks") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PipelineResources) MarshalJSON() ([]byte, error) {
	type noMethod PipelineResources
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type RepeatedString struct {
	Values []string `json:"values,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Values") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RepeatedString) MarshalJSON() ([]byte, error) {
	type noMethod RepeatedString
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// RunPipelineArgs: The pipeline run arguments.
type RunPipelineArgs struct {
	// ClientId: Optional. For callers to use in filtering operations
	// returned by this request.
	ClientId string `json:"clientId,omitempty"`

	// Inputs: Pipeline input arguments; keys are defined in the pipeline
	// documentation. All input parameters that do not have default values
	// must be specified. If parameters with defaults are specified here,
	// the defaults will be overridden.
	Inputs map[string]string `json:"inputs,omitempty"`

	// Logging: Required. Logging options. Used by the service to
	// communicate results to the user.
	Logging *LoggingOptions `json:"logging,omitempty"`

	// Outputs: Pipeline output arguments; keys are defined in the pipeline
	// documentation. All output parameters of without default values must
	// be specified. If parameters with defaults are specified here, the
	// defaults will be overridden.
	Outputs map[string]string `json:"outputs,omitempty"`

	// ProjectId: Required. The project in which to run the pipeline. The
	// caller must have WRITER access to all Google Cloud services and
	// resources (e.g. Google Compute Engine) will be used.
	ProjectId string `json:"projectId,omitempty"`

	// Resources: Specifies resource requirements/overrides for the pipeline
	// run.
	Resources *PipelineResources `json:"resources,omitempty"`

	// ServiceAccount: Required. The Google Cloud Service Account that will
	// be used to access data and services.
	ServiceAccount *ServiceAccount `json:"serviceAccount,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ClientId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RunPipelineArgs) MarshalJSON() ([]byte, error) {
	type noMethod RunPipelineArgs
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// RunPipelineRequest: The request to run a pipeline. If `pipelineId` is
// specified, it refers to a saved pipeline created with CreatePipeline
// and set as the `pipelineId` of the returned Pipeline object. If
// `ephemeralPipeline` is specified, that pipeline is run once with the
// given args and not saved. It is an error to specify both `pipelineId`
// and `ephemeralPipeline`. `pipelineArgs` must be specified.
type RunPipelineRequest struct {
	// EphemeralPipeline: A new pipeline object to run once and then delete.
	EphemeralPipeline *Pipeline `json:"ephemeralPipeline,omitempty"`

	// PipelineArgs: The arguments to use when running this pipeline.
	PipelineArgs *RunPipelineArgs `json:"pipelineArgs,omitempty"`

	// PipelineId: The already created pipeline to run.
	PipelineId string `json:"pipelineId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "EphemeralPipeline")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *RunPipelineRequest) MarshalJSON() ([]byte, error) {
	type noMethod RunPipelineRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ServiceAccount: A Google Cloud Service Account.
type ServiceAccount struct {
	// Email: Required. Email address of the service account. 'default' is a
	// valid option and uses the compute service account associated with the
	// project.
	Email string `json:"email,omitempty"`

	// Scopes: Required. List of scopes to be made available for this
	// service account. Should include *
	// https://www.googleapis.com/auth/genomics *
	// https://www.googleapis.com/auth/compute *
	// https://www.googleapis.com/auth/devstorage.full_control
	Scopes []string `json:"scopes,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Email") to
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

// SetOperationStatusRequest: Request to set operation status. Should
// only be used by VMs created by the Pipelines Service and not by end
// users.
type SetOperationStatusRequest struct {
	// Possible values:
	//   "OK"
	//   "CANCELLED"
	//   "UNKNOWN"
	//   "INVALID_ARGUMENT"
	//   "DEADLINE_EXCEEDED"
	//   "NOT_FOUND"
	//   "ALREADY_EXISTS"
	//   "PERMISSION_DENIED"
	//   "UNAUTHENTICATED"
	//   "RESOURCE_EXHAUSTED"
	//   "FAILED_PRECONDITION"
	//   "ABORTED"
	//   "OUT_OF_RANGE"
	//   "UNIMPLEMENTED"
	//   "INTERNAL"
	//   "UNAVAILABLE"
	//   "DATA_LOSS"
	ErrorCode string `json:"errorCode,omitempty"`

	ErrorMessage string `json:"errorMessage,omitempty"`

	OperationId string `json:"operationId,omitempty"`

	TimestampEvents []*TimestampEvent `json:"timestampEvents,omitempty"`

	ValidationToken uint64 `json:"validationToken,omitempty,string"`

	// ForceSendFields is a list of field names (e.g. "ErrorCode") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SetOperationStatusRequest) MarshalJSON() ([]byte, error) {
	type noMethod SetOperationStatusRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Status: The `Status` type defines a logical error model that is
// suitable for different programming environments, including REST APIs
// and RPC APIs. It is used by [gRPC](https://github.com/grpc). The
// error model is designed to be: - Simple to use and understand for
// most users - Flexible enough to meet unexpected needs # Overview The
// `Status` message contains three pieces of data: error code, error
// message, and error details. The error code should be an enum value of
// google.rpc.Code, but it may accept additional error codes if needed.
// The error message should be a developer-facing English message that
// helps developers *understand* and *resolve* the error. If a localized
// user-facing error message is needed, put the localized message in the
// error details or localize it in the client. The optional error
// details may contain arbitrary information about the error. There is a
// predefined set of error detail types in the package `google.rpc`
// which can be used for common error conditions. # Language mapping The
// `Status` message is the logical representation of the error model,
// but it is not necessarily the actual wire format. When the `Status`
// message is exposed in different client libraries and different wire
// protocols, it can be mapped differently. For example, it will likely
// be mapped to some exceptions in Java, but more likely mapped to some
// error codes in C. # Other uses The error model and the `Status`
// message can be used in a variety of environments, either with or
// without APIs, to provide a consistent developer experience across
// different environments. Example uses of this error model include: -
// Partial errors. If a service needs to return partial errors to the
// client, it may embed the `Status` in the normal response to indicate
// the partial errors. - Workflow errors. A typical workflow has
// multiple steps. Each step may have a `Status` message for error
// reporting purpose. - Batch operations. If a client uses batch request
// and batch response, the `Status` message should be used directly
// inside batch response, one for each error sub-response. -
// Asynchronous operations. If an API call embeds asynchronous operation
// results in its response, the status of those operations should be
// represented directly using the `Status` message. - Logging. If some
// API errors are stored in logs, the message `Status` could be used
// directly after any stripping needed for security/privacy reasons.
type Status struct {
	// Code: The status code, which should be an enum value of
	// google.rpc.Code.
	Code int64 `json:"code,omitempty"`

	// Details: A list of messages that carry the error details. There will
	// be a common set of message types for APIs to use.
	Details []StatusDetails `json:"details,omitempty"`

	// Message: A developer-facing error message, which should be in
	// English. Any user-facing error message should be localized and sent
	// in the google.rpc.Status.details field, or localized by the client.
	Message string `json:"message,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Code") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Status) MarshalJSON() ([]byte, error) {
	type noMethod Status
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type StatusDetails interface{}

// TimestampEvent: Stores the list of events and times they occured for
// major events in job execution.
type TimestampEvent struct {
	// Description: String indicating the type of event
	Description string `json:"description,omitempty"`

	// Timestamp: The time this event occured.
	Timestamp string `json:"timestamp,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Description") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *TimestampEvent) MarshalJSON() ([]byte, error) {
	type noMethod TimestampEvent
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "genomics.operations.cancel":

type OperationsCancelCall struct {
	s                      *Service
	name                   string
	canceloperationrequest *CancelOperationRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// Cancel: Starts asynchronous cancellation on a long-running operation.
// The server makes a best effort to cancel the operation, but success
// is not guaranteed. Clients may use Operations.GetOperation or
// Operations.ListOperations to check whether the cancellation succeeded
// or the operation completed despite cancellation.
func (r *OperationsService) Cancel(name string, canceloperationrequest *CancelOperationRequest) *OperationsCancelCall {
	c := &OperationsCancelCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	c.canceloperationrequest = canceloperationrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OperationsCancelCall) Fields(s ...googleapi.Field) *OperationsCancelCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *OperationsCancelCall) Context(ctx context.Context) *OperationsCancelCall {
	c.ctx_ = ctx
	return c
}

func (c *OperationsCancelCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.canceloperationrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/{+name}:cancel")
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

// Do executes the "genomics.operations.cancel" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *OperationsCancelCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Starts asynchronous cancellation on a long-running operation. The server makes a best effort to cancel the operation, but success is not guaranteed. Clients may use Operations.GetOperation or Operations.ListOperations to check whether the cancellation succeeded or the operation completed despite cancellation.",
	//   "httpMethod": "POST",
	//   "id": "genomics.operations.cancel",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The name of the operation resource to be cancelled.",
	//       "location": "path",
	//       "pattern": "^operations/.*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/{+name}:cancel",
	//   "request": {
	//     "$ref": "CancelOperationRequest"
	//   },
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.operations.get":

type OperationsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets the latest state of a long-running operation. Clients can
// use this method to poll the operation result at intervals as
// recommended by the API service.
func (r *OperationsService) Get(name string) *OperationsGetCall {
	c := &OperationsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OperationsGetCall) Fields(s ...googleapi.Field) *OperationsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *OperationsGetCall) IfNoneMatch(entityTag string) *OperationsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *OperationsGetCall) Context(ctx context.Context) *OperationsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *OperationsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/{+name}")
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

// Do executes the "genomics.operations.get" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *OperationsGetCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	ret := &Operation{
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
	//   "description": "Gets the latest state of a long-running operation. Clients can use this method to poll the operation result at intervals as recommended by the API service.",
	//   "httpMethod": "GET",
	//   "id": "genomics.operations.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The name of the operation resource.",
	//       "location": "path",
	//       "pattern": "^operations/.*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/{+name}",
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.operations.list":

type OperationsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists operations that match the specified filter in the
// request.
func (r *OperationsService) List(name string) *OperationsListCall {
	c := &OperationsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Filter sets the optional parameter "filter": A string for filtering
// Operations. The following filter fields are supported: * projectId:
// Required. Corresponds to OperationMetadata.projectId. * createTime:
// The time this job was created, in seconds from the
// [epoch](http://en.wikipedia.org/wiki/Unix_time). Can use `>=` and/or
// `= 1432140000` * `projectId = my-project AND createTime >= 1432140000
// AND createTime <= 1432150000 AND status = RUNNING`
func (c *OperationsListCall) Filter(filter string) *OperationsListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// PageSize sets the optional parameter "pageSize": The maximum number
// of results to return. If unspecified, defaults to 256. The maximum
// value is 2048.
func (c *OperationsListCall) PageSize(pageSize int64) *OperationsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The standard list
// page token.
func (c *OperationsListCall) PageToken(pageToken string) *OperationsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *OperationsListCall) Fields(s ...googleapi.Field) *OperationsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *OperationsListCall) IfNoneMatch(entityTag string) *OperationsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *OperationsListCall) Context(ctx context.Context) *OperationsListCall {
	c.ctx_ = ctx
	return c
}

func (c *OperationsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/{+name}")
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

// Do executes the "genomics.operations.list" call.
// Exactly one of *ListOperationsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListOperationsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *OperationsListCall) Do(opts ...googleapi.CallOption) (*ListOperationsResponse, error) {
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
	ret := &ListOperationsResponse{
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
	//   "description": "Lists operations that match the specified filter in the request.",
	//   "httpMethod": "GET",
	//   "id": "genomics.operations.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "filter": {
	//       "description": "A string for filtering Operations. The following filter fields are supported: * projectId: Required. Corresponds to OperationMetadata.projectId. * createTime: The time this job was created, in seconds from the [epoch](http://en.wikipedia.org/wiki/Unix_time). Can use `\u003e=` and/or `= 1432140000` * `projectId = my-project AND createTime \u003e= 1432140000 AND createTime \u003c= 1432150000 AND status = RUNNING`",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The name of the operation collection.",
	//       "location": "path",
	//       "pattern": "^operations$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The maximum number of results to return. If unspecified, defaults to 256. The maximum value is 2048.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The standard list page token.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/{+name}",
	//   "response": {
	//     "$ref": "ListOperationsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *OperationsListCall) Pages(ctx context.Context, f func(*ListOperationsResponse) error) error {
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

// method id "genomics.pipelines.create":

type PipelinesCreateCall struct {
	s          *Service
	pipeline   *Pipeline
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create: Creates a pipeline that can be run later. Create takes a
// Pipeline that has all fields other than `pipelineId` populated, and
// then returns the same pipeline with `pipelineId` populated. This id
// can be used to run the pipeline. Caller must have WRITE permission to
// the project.
func (r *PipelinesService) Create(pipeline *Pipeline) *PipelinesCreateCall {
	c := &PipelinesCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.pipeline = pipeline
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesCreateCall) Fields(s ...googleapi.Field) *PipelinesCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesCreateCall) Context(ctx context.Context) *PipelinesCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.pipeline)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "genomics.pipelines.create" call.
// Exactly one of *Pipeline or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Pipeline.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PipelinesCreateCall) Do(opts ...googleapi.CallOption) (*Pipeline, error) {
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
	ret := &Pipeline{
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
	//   "description": "Creates a pipeline that can be run later. Create takes a Pipeline that has all fields other than `pipelineId` populated, and then returns the same pipeline with `pipelineId` populated. This id can be used to run the pipeline. Caller must have WRITE permission to the project.",
	//   "httpMethod": "POST",
	//   "id": "genomics.pipelines.create",
	//   "path": "v1alpha2/pipelines",
	//   "request": {
	//     "$ref": "Pipeline"
	//   },
	//   "response": {
	//     "$ref": "Pipeline"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.pipelines.delete":

type PipelinesDeleteCall struct {
	s          *Service
	pipelineId string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a pipeline based on ID. Caller must have WRITE
// permission to the project.
func (r *PipelinesService) Delete(pipelineId string) *PipelinesDeleteCall {
	c := &PipelinesDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.pipelineId = pipelineId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesDeleteCall) Fields(s ...googleapi.Field) *PipelinesDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesDeleteCall) Context(ctx context.Context) *PipelinesDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines/{pipelineId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"pipelineId": c.pipelineId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "genomics.pipelines.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *PipelinesDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Deletes a pipeline based on ID. Caller must have WRITE permission to the project.",
	//   "httpMethod": "DELETE",
	//   "id": "genomics.pipelines.delete",
	//   "parameterOrder": [
	//     "pipelineId"
	//   ],
	//   "parameters": {
	//     "pipelineId": {
	//       "description": "Caller must have WRITE access to the project in which this pipeline is defined.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/pipelines/{pipelineId}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.pipelines.get":

type PipelinesGetCall struct {
	s            *Service
	pipelineId   string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Retrieves a pipeline based on ID. Caller must have READ
// permission to the project.
func (r *PipelinesService) Get(pipelineId string) *PipelinesGetCall {
	c := &PipelinesGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.pipelineId = pipelineId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesGetCall) Fields(s ...googleapi.Field) *PipelinesGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PipelinesGetCall) IfNoneMatch(entityTag string) *PipelinesGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesGetCall) Context(ctx context.Context) *PipelinesGetCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines/{pipelineId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"pipelineId": c.pipelineId,
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

// Do executes the "genomics.pipelines.get" call.
// Exactly one of *Pipeline or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Pipeline.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PipelinesGetCall) Do(opts ...googleapi.CallOption) (*Pipeline, error) {
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
	ret := &Pipeline{
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
	//   "description": "Retrieves a pipeline based on ID. Caller must have READ permission to the project.",
	//   "httpMethod": "GET",
	//   "id": "genomics.pipelines.get",
	//   "parameterOrder": [
	//     "pipelineId"
	//   ],
	//   "parameters": {
	//     "pipelineId": {
	//       "description": "Caller must have READ access to the project in which this pipeline is defined.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/pipelines/{pipelineId}",
	//   "response": {
	//     "$ref": "Pipeline"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.pipelines.getControllerConfig":

type PipelinesGetControllerConfigCall struct {
	s            *Service
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// GetControllerConfig: Gets controller configuration information.
// Should only be called by VMs created by the Pipelines Service and not
// by end users.
func (r *PipelinesService) GetControllerConfig() *PipelinesGetControllerConfigCall {
	c := &PipelinesGetControllerConfigCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	return c
}

// OperationId sets the optional parameter "operationId": The operation
// to retrieve controller configuration for.
func (c *PipelinesGetControllerConfigCall) OperationId(operationId string) *PipelinesGetControllerConfigCall {
	c.urlParams_.Set("operationId", operationId)
	return c
}

// ValidationToken sets the optional parameter "validationToken":
func (c *PipelinesGetControllerConfigCall) ValidationToken(validationToken uint64) *PipelinesGetControllerConfigCall {
	c.urlParams_.Set("validationToken", fmt.Sprint(validationToken))
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesGetControllerConfigCall) Fields(s ...googleapi.Field) *PipelinesGetControllerConfigCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PipelinesGetControllerConfigCall) IfNoneMatch(entityTag string) *PipelinesGetControllerConfigCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesGetControllerConfigCall) Context(ctx context.Context) *PipelinesGetControllerConfigCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesGetControllerConfigCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines:getControllerConfig")
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

// Do executes the "genomics.pipelines.getControllerConfig" call.
// Exactly one of *ControllerConfig or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ControllerConfig.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PipelinesGetControllerConfigCall) Do(opts ...googleapi.CallOption) (*ControllerConfig, error) {
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
	ret := &ControllerConfig{
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
	//   "description": "Gets controller configuration information. Should only be called by VMs created by the Pipelines Service and not by end users.",
	//   "httpMethod": "GET",
	//   "id": "genomics.pipelines.getControllerConfig",
	//   "parameters": {
	//     "operationId": {
	//       "description": "The operation to retrieve controller configuration for.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "validationToken": {
	//       "format": "uint64",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/pipelines:getControllerConfig",
	//   "response": {
	//     "$ref": "ControllerConfig"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.pipelines.list":

type PipelinesListCall struct {
	s            *Service
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists pipelines. Caller must have READ permission to the
// project.
func (r *PipelinesService) List() *PipelinesListCall {
	c := &PipelinesListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	return c
}

// NamePrefix sets the optional parameter "namePrefix": Pipelines with
// names that match this prefix should be returned. If unspecified, all
// pipelines in the project, up to `pageSize`, will be returned.
func (c *PipelinesListCall) NamePrefix(namePrefix string) *PipelinesListCall {
	c.urlParams_.Set("namePrefix", namePrefix)
	return c
}

// PageSize sets the optional parameter "pageSize": Number of pipelines
// to return at once. Defaults to 256, and max is 2048.
func (c *PipelinesListCall) PageSize(pageSize int64) *PipelinesListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": Token to use to
// indicate where to start getting results. If unspecified, returns the
// first page of results.
func (c *PipelinesListCall) PageToken(pageToken string) *PipelinesListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// ProjectId sets the optional parameter "projectId": Required. The name
// of the project to search for pipelines. Caller must have READ access
// to this project.
func (c *PipelinesListCall) ProjectId(projectId string) *PipelinesListCall {
	c.urlParams_.Set("projectId", projectId)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesListCall) Fields(s ...googleapi.Field) *PipelinesListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *PipelinesListCall) IfNoneMatch(entityTag string) *PipelinesListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesListCall) Context(ctx context.Context) *PipelinesListCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines")
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

// Do executes the "genomics.pipelines.list" call.
// Exactly one of *ListPipelinesResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListPipelinesResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *PipelinesListCall) Do(opts ...googleapi.CallOption) (*ListPipelinesResponse, error) {
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
	ret := &ListPipelinesResponse{
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
	//   "description": "Lists pipelines. Caller must have READ permission to the project.",
	//   "httpMethod": "GET",
	//   "id": "genomics.pipelines.list",
	//   "parameters": {
	//     "namePrefix": {
	//       "description": "Optional. Pipelines with names that match this prefix should be returned. If unspecified, all pipelines in the project, up to `pageSize`, will be returned.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "Optional. Number of pipelines to return at once. Defaults to 256, and max is 2048.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "Optional. Token to use to indicate where to start getting results. If unspecified, returns the first page of results.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "Required. The name of the project to search for pipelines. Caller must have READ access to this project.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1alpha2/pipelines",
	//   "response": {
	//     "$ref": "ListPipelinesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *PipelinesListCall) Pages(ctx context.Context, f func(*ListPipelinesResponse) error) error {
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

// method id "genomics.pipelines.run":

type PipelinesRunCall struct {
	s                  *Service
	runpipelinerequest *RunPipelineRequest
	urlParams_         gensupport.URLParams
	ctx_               context.Context
}

// Run: Runs a pipeline. If `pipelineId` is specified in the request,
// then run a saved pipeline. If `ephemeralPipeline` is specified, then
// run that pipeline once without saving a copy. The caller must have
// READ permission to the project where the pipeline is stored and WRITE
// permission to the project where the pipeline will be run, as VMs will
// be created and storage will be used.
func (r *PipelinesService) Run(runpipelinerequest *RunPipelineRequest) *PipelinesRunCall {
	c := &PipelinesRunCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.runpipelinerequest = runpipelinerequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesRunCall) Fields(s ...googleapi.Field) *PipelinesRunCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesRunCall) Context(ctx context.Context) *PipelinesRunCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesRunCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.runpipelinerequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines:run")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "genomics.pipelines.run" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *PipelinesRunCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	ret := &Operation{
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
	//   "description": "Runs a pipeline. If `pipelineId` is specified in the request, then run a saved pipeline. If `ephemeralPipeline` is specified, then run that pipeline once without saving a copy. The caller must have READ permission to the project where the pipeline is stored and WRITE permission to the project where the pipeline will be run, as VMs will be created and storage will be used.",
	//   "httpMethod": "POST",
	//   "id": "genomics.pipelines.run",
	//   "path": "v1alpha2/pipelines:run",
	//   "request": {
	//     "$ref": "RunPipelineRequest"
	//   },
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}

// method id "genomics.pipelines.setOperationStatus":

type PipelinesSetOperationStatusCall struct {
	s                         *Service
	setoperationstatusrequest *SetOperationStatusRequest
	urlParams_                gensupport.URLParams
	ctx_                      context.Context
}

// SetOperationStatus: Sets status of a given operation. All timestamps
// are sent on each call, and the whole series of events is replaced, in
// case intermediate calls are lost. Should only be called by VMs
// created by the Pipelines Service and not by end users.
func (r *PipelinesService) SetOperationStatus(setoperationstatusrequest *SetOperationStatusRequest) *PipelinesSetOperationStatusCall {
	c := &PipelinesSetOperationStatusCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.setoperationstatusrequest = setoperationstatusrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *PipelinesSetOperationStatusCall) Fields(s ...googleapi.Field) *PipelinesSetOperationStatusCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *PipelinesSetOperationStatusCall) Context(ctx context.Context) *PipelinesSetOperationStatusCall {
	c.ctx_ = ctx
	return c
}

func (c *PipelinesSetOperationStatusCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.setoperationstatusrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1alpha2/pipelines:setOperationStatus")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PUT", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "genomics.pipelines.setOperationStatus" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *PipelinesSetOperationStatusCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Sets status of a given operation. All timestamps are sent on each call, and the whole series of events is replaced, in case intermediate calls are lost. Should only be called by VMs created by the Pipelines Service and not by end users.",
	//   "httpMethod": "PUT",
	//   "id": "genomics.pipelines.setOperationStatus",
	//   "path": "v1alpha2/pipelines:setOperationStatus",
	//   "request": {
	//     "$ref": "SetOperationStatusRequest"
	//   },
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/genomics"
	//   ]
	// }

}
