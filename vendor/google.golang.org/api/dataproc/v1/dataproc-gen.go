// Package dataproc provides access to the Google Cloud Dataproc API.
//
// See https://cloud.google.com/dataproc/
//
// Usage example:
//
//   import "google.golang.org/api/dataproc/v1"
//   ...
//   dataprocService, err := dataproc.New(oauthHttpClient)
package dataproc // import "google.golang.org/api/dataproc/v1"

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

const apiId = "dataproc:v1"
const apiName = "dataproc"
const apiVersion = "v1"
const basePath = "https://dataproc.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	// Administrate log data for your projects
	LoggingAdminScope = "https://www.googleapis.com/auth/logging.admin"

	// View log data for your projects
	LoggingReadScope = "https://www.googleapis.com/auth/logging.read"

	// Submit log data for your projects
	LoggingWriteScope = "https://www.googleapis.com/auth/logging.write"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Media = NewMediaService(s)
	s.Projects = NewProjectsService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Media *MediaService

	Projects *ProjectsService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewMediaService(s *Service) *MediaService {
	rs := &MediaService{s: s}
	return rs
}

type MediaService struct {
	s *Service
}

func NewProjectsService(s *Service) *ProjectsService {
	rs := &ProjectsService{s: s}
	rs.Regions = NewProjectsRegionsService(s)
	return rs
}

type ProjectsService struct {
	s *Service

	Regions *ProjectsRegionsService
}

func NewProjectsRegionsService(s *Service) *ProjectsRegionsService {
	rs := &ProjectsRegionsService{s: s}
	rs.Clusters = NewProjectsRegionsClustersService(s)
	rs.Jobs = NewProjectsRegionsJobsService(s)
	rs.Operations = NewProjectsRegionsOperationsService(s)
	return rs
}

type ProjectsRegionsService struct {
	s *Service

	Clusters *ProjectsRegionsClustersService

	Jobs *ProjectsRegionsJobsService

	Operations *ProjectsRegionsOperationsService
}

func NewProjectsRegionsClustersService(s *Service) *ProjectsRegionsClustersService {
	rs := &ProjectsRegionsClustersService{s: s}
	return rs
}

type ProjectsRegionsClustersService struct {
	s *Service
}

func NewProjectsRegionsJobsService(s *Service) *ProjectsRegionsJobsService {
	rs := &ProjectsRegionsJobsService{s: s}
	return rs
}

type ProjectsRegionsJobsService struct {
	s *Service
}

func NewProjectsRegionsOperationsService(s *Service) *ProjectsRegionsOperationsService {
	rs := &ProjectsRegionsOperationsService{s: s}
	return rs
}

type ProjectsRegionsOperationsService struct {
	s *Service
}

// CancelJobRequest: A request to cancel a job.
type CancelJobRequest struct {
}

// Cluster: Describes the identifying information, config, and status of
// a cluster of Google Compute Engine instances.
type Cluster struct {
	// ClusterName: [Required] The cluster name. Cluster names within a
	// project must be unique. Names of deleted clusters can be reused.
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterUuid: [Output-only] A cluster UUID (Unique Universal
	// Identifier). Cloud Dataproc generates this value when it creates the
	// cluster.
	ClusterUuid string `json:"clusterUuid,omitempty"`

	// Config: [Required] The cluster config. Note that Cloud Dataproc may
	// set default values, and values may change when clusters are updated.
	Config *ClusterConfig `json:"config,omitempty"`

	// ProjectId: [Required] The Google Cloud Platform project ID that the
	// cluster belongs to.
	ProjectId string `json:"projectId,omitempty"`

	// Status: [Output-only] Cluster status.
	Status *ClusterStatus `json:"status,omitempty"`

	// StatusHistory: [Output-only] The previous cluster status.
	StatusHistory []*ClusterStatus `json:"statusHistory,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ClusterName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Cluster) MarshalJSON() ([]byte, error) {
	type noMethod Cluster
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClusterConfig: The cluster config.
type ClusterConfig struct {
	// ConfigBucket: [Optional] A Google Cloud Storage staging bucket used
	// for sharing generated SSH keys and config. If you do not specify a
	// staging bucket, Cloud Dataproc will determine an appropriate Cloud
	// Storage location (US, ASIA, or EU) for your cluster's staging bucket
	// according to the Google Compute Engine zone where your cluster is
	// deployed, and then it will create and manage this project-level,
	// per-location bucket for you.
	ConfigBucket string `json:"configBucket,omitempty"`

	// GceClusterConfig: [Optional] The shared Google Compute Engine config
	// settings for all instances in a cluster.
	GceClusterConfig *GceClusterConfig `json:"gceClusterConfig,omitempty"`

	// InitializationActions: [Optional] Commands to execute on each node
	// after config is completed. By default, executables are run on master
	// and all worker nodes. You can test a node's role metadata to run an
	// executable on a master or worker node, as shown below:
	// ROLE=$(/usr/share/google/get_metadata_value attributes/role) if [[
	// "${ROLE}" == 'Master' ]]; then ... master specific actions ... else
	// ... worker specific actions ... fi
	InitializationActions []*NodeInitializationAction `json:"initializationActions,omitempty"`

	// MasterConfig: [Optional] The Google Compute Engine config settings
	// for the master instance in a cluster.
	MasterConfig *InstanceGroupConfig `json:"masterConfig,omitempty"`

	// SecondaryWorkerConfig: [Optional] The Google Compute Engine config
	// settings for additional worker instances in a cluster.
	SecondaryWorkerConfig *InstanceGroupConfig `json:"secondaryWorkerConfig,omitempty"`

	// SoftwareConfig: [Optional] The config settings for software inside
	// the cluster.
	SoftwareConfig *SoftwareConfig `json:"softwareConfig,omitempty"`

	// WorkerConfig: [Optional] The Google Compute Engine config settings
	// for worker instances in a cluster.
	WorkerConfig *InstanceGroupConfig `json:"workerConfig,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ConfigBucket") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClusterConfig) MarshalJSON() ([]byte, error) {
	type noMethod ClusterConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClusterOperationMetadata: Metadata describing the operation.
type ClusterOperationMetadata struct {
	// ClusterName: Name of the cluster for the operation.
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterUuid: Cluster UUId for the operation.
	ClusterUuid string `json:"clusterUuid,omitempty"`

	// Description: [Output-only] Short description of operation.
	Description string `json:"description,omitempty"`

	// OperationType: [Output-only] The operation type.
	OperationType string `json:"operationType,omitempty"`

	// Status: [Output-only] Current operation status.
	Status *ClusterOperationStatus `json:"status,omitempty"`

	// StatusHistory: [Output-only] The previous operation status.
	StatusHistory []*ClusterOperationStatus `json:"statusHistory,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ClusterName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClusterOperationMetadata) MarshalJSON() ([]byte, error) {
	type noMethod ClusterOperationMetadata
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClusterOperationStatus: The status of the operation.
type ClusterOperationStatus struct {
	// Details: A message containing any operation metadata details.
	Details string `json:"details,omitempty"`

	// InnerState: A message containing the detailed operation state.
	InnerState string `json:"innerState,omitempty"`

	// State: A message containing the operation state.
	//
	// Possible values:
	//   "UNKNOWN"
	//   "PENDING"
	//   "RUNNING"
	//   "DONE"
	State string `json:"state,omitempty"`

	// StateStartTime: The time this state was entered.
	StateStartTime string `json:"stateStartTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Details") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClusterOperationStatus) MarshalJSON() ([]byte, error) {
	type noMethod ClusterOperationStatus
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ClusterStatus: The status of a cluster and its instances.
type ClusterStatus struct {
	// Detail: Optional details of cluster's state.
	Detail string `json:"detail,omitempty"`

	// State: The cluster's state.
	//
	// Possible values:
	//   "UNKNOWN"
	//   "CREATING"
	//   "RUNNING"
	//   "ERROR"
	//   "DELETING"
	//   "UPDATING"
	State string `json:"state,omitempty"`

	// StateStartTime: Time when this state was entered.
	StateStartTime string `json:"stateStartTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Detail") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ClusterStatus) MarshalJSON() ([]byte, error) {
	type noMethod ClusterStatus
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DiagnoseClusterOutputLocation: The location where output from
// diagnostic command can be found.
type DiagnoseClusterOutputLocation struct {
	// OutputUri: [Output-only] The Google Cloud Storage URI of the
	// diagnostic output. This will be a plain text file with summary of
	// collected diagnostics.
	OutputUri string `json:"outputUri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "OutputUri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DiagnoseClusterOutputLocation) MarshalJSON() ([]byte, error) {
	type noMethod DiagnoseClusterOutputLocation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DiagnoseClusterRequest: A request to collect cluster diagnostic
// information.
type DiagnoseClusterRequest struct {
}

// DiagnoseClusterResults: The location of diagnostic output.
type DiagnoseClusterResults struct {
	// OutputUri: [Output-only] The Google Cloud Storage URI of the
	// diagnostic output. This is a plain text file with a summary of
	// collected diagnostics.
	OutputUri string `json:"outputUri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "OutputUri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DiagnoseClusterResults) MarshalJSON() ([]byte, error) {
	type noMethod DiagnoseClusterResults
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DiskConfig: Specifies the config of disk options for a group of VM
// instances.
type DiskConfig struct {
	// BootDiskSizeGb: [Optional] Size in GB of the boot disk (default is
	// 500GB).
	BootDiskSizeGb int64 `json:"bootDiskSizeGb,omitempty"`

	// NumLocalSsds: [Optional] Number of attached SSDs, from 0 to 4
	// (default is 0). If SSDs are not attached, the boot disk is used to
	// store runtime logs and HDFS data. If one or more SSDs are attached,
	// this runtime bulk data is spread across them, and the boot disk
	// contains only basic config and installed binaries.
	NumLocalSsds int64 `json:"numLocalSsds,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BootDiskSizeGb") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DiskConfig) MarshalJSON() ([]byte, error) {
	type noMethod DiskConfig
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

// GceClusterConfig: Common config settings for resources of Google
// Compute Engine cluster instances, applicable to all instances in the
// cluster.
type GceClusterConfig struct {
	// Metadata: The Google Compute Engine metadata entries to add to all
	// instances.
	Metadata map[string]string `json:"metadata,omitempty"`

	// NetworkUri: The Google Compute Engine network to be used for machine
	// communications. Cannot be specified with subnetwork_uri. If neither
	// network_uri nor subnetwork_uri is specified, the "default" network of
	// the project is used, if it exists. Cannot be a "Custom Subnet
	// Network" (see https://cloud.google.com/compute/docs/subnetworks for
	// more information). Example:
	// `https://www.googleapis.com/compute/v1/projects/[project_id]/regions/g
	// lobal/default`.
	NetworkUri string `json:"networkUri,omitempty"`

	// ServiceAccountScopes: The URIs of service account scopes to be
	// included in Google Compute Engine instances. The following base set
	// of scopes is always included: *
	// https://www.googleapis.com/auth/cloud.useraccounts.readonly *
	// https://www.googleapis.com/auth/devstorage.read_write *
	// https://www.googleapis.com/auth/logging.write If no scopes are
	// specfied, the following defaults are also provided: *
	// https://www.googleapis.com/auth/bigquery *
	// https://www.googleapis.com/auth/bigtable.admin.table *
	// https://www.googleapis.com/auth/bigtable.data *
	// https://www.googleapis.com/auth/devstorage.full_control
	ServiceAccountScopes []string `json:"serviceAccountScopes,omitempty"`

	// SubnetworkUri: The Google Compute Engine subnetwork to be used for
	// machine communications. Cannot be specified with network_uri.
	// Example:
	// `https://www.googleapis.com/compute/v1/projects/[project_id]/regions/u
	// s-east1/sub0`.
	SubnetworkUri string `json:"subnetworkUri,omitempty"`

	// Tags: The Google Compute Engine tags to add to all instances.
	Tags []string `json:"tags,omitempty"`

	// ZoneUri: [Required] The zone where the Google Compute Engine cluster
	// will be located. Example:
	// `https://www.googleapis.com/compute/v1/projects/[project_id]/zones/[zo
	// ne]`.
	ZoneUri string `json:"zoneUri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Metadata") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *GceClusterConfig) MarshalJSON() ([]byte, error) {
	type noMethod GceClusterConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// HadoopJob: A Cloud Dataproc job for running Hadoop MapReduce jobs on
// YARN.
type HadoopJob struct {
	// ArchiveUris: [Optional] HCFS URIs of archives to be extracted in the
	// working directory of Hadoop drivers and tasks. Supported file types:
	// .jar, .tar, .tar.gz, .tgz, or .zip.
	ArchiveUris []string `json:"archiveUris,omitempty"`

	// Args: [Optional] The arguments to pass to the driver. Do not include
	// arguments, such as `-libjars` or `-Dfoo=bar`, that can be set as job
	// properties, since a collision may occur that causes an incorrect job
	// submission.
	Args []string `json:"args,omitempty"`

	// FileUris: [Optional] HCFS (Hadoop Compatible Filesystem) URIs of
	// files to be copied to the working directory of Hadoop drivers and
	// distributed tasks. Useful for naively parallel tasks.
	FileUris []string `json:"fileUris,omitempty"`

	// JarFileUris: [Optional] Jar file URIs to add to the CLASSPATHs of the
	// Hadoop driver and tasks.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// LoggingConfig: [Optional] The runtime log config for job execution.
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`

	// MainClass: The name of the driver's main class. The jar file
	// containing the class must be in the default CLASSPATH or specified in
	// `jar_file_uris`.
	MainClass string `json:"mainClass,omitempty"`

	// MainJarFileUri: The HCFS URI of the jar file containing the main
	// class. Examples:
	// 'gs://foo-bucket/analytics-binaries/extract-useful-metrics-mr.jar'
	// 'hdfs:/tmp/test-samples/custom-wordcount.jar'
	// 'file:///home/usr/lib/hadoop-mapreduce/hadoop-mapreduce-examples.jar'
	MainJarFileUri string `json:"mainJarFileUri,omitempty"`

	// Properties: [Optional] A mapping of property names to values, used to
	// configure Hadoop. Properties that conflict with values set by the
	// Cloud Dataproc API may be overwritten. Can include properties set in
	// /etc/hadoop/conf/*-site and classes in user code.
	Properties map[string]string `json:"properties,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ArchiveUris") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *HadoopJob) MarshalJSON() ([]byte, error) {
	type noMethod HadoopJob
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// HiveJob: A Cloud Dataproc job for running Hive queries on YARN.
type HiveJob struct {
	// ContinueOnFailure: [Optional] Whether to continue executing queries
	// if a query fails. The default value is `false`. Setting to `true` can
	// be useful when executing independent parallel queries.
	ContinueOnFailure bool `json:"continueOnFailure,omitempty"`

	// JarFileUris: [Optional] HCFS URIs of jar files to add to the
	// CLASSPATH of the Hive server and Hadoop MapReduce (MR) tasks. Can
	// contain Hive SerDes and UDFs.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// Properties: [Optional] A mapping of property names and values, used
	// to configure Hive. Properties that conflict with values set by the
	// Cloud Dataproc API may be overwritten. Can include properties set in
	// /etc/hadoop/conf/*-site.xml, /etc/hive/conf/hive-site.xml, and
	// classes in user code.
	Properties map[string]string `json:"properties,omitempty"`

	// QueryFileUri: The HCFS URI of the script that contains Hive queries.
	QueryFileUri string `json:"queryFileUri,omitempty"`

	// QueryList: A list of queries.
	QueryList *QueryList `json:"queryList,omitempty"`

	// ScriptVariables: [Optional] Mapping of query variable names to values
	// (equivalent to the Hive command: `SET name="value";`).
	ScriptVariables map[string]string `json:"scriptVariables,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ContinueOnFailure")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *HiveJob) MarshalJSON() ([]byte, error) {
	type noMethod HiveJob
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// InstanceGroupConfig: The config settings for Google Compute Engine
// resources in an instance group, such as a master or worker group.
type InstanceGroupConfig struct {
	// DiskConfig: Disk option config settings.
	DiskConfig *DiskConfig `json:"diskConfig,omitempty"`

	// ImageUri: [Output-only] The Google Compute Engine image resource used
	// for cluster instances. Inferred from `SoftwareConfig.image_version`.
	ImageUri string `json:"imageUri,omitempty"`

	// InstanceNames: The list of instance names. Cloud Dataproc derives the
	// names from `cluster_name`, `num_instances`, and the instance group if
	// not set by user (recommended practice is to let Cloud Dataproc derive
	// the name).
	InstanceNames []string `json:"instanceNames,omitempty"`

	// IsPreemptible: Specifies that this instance group contains
	// Preemptible Instances.
	IsPreemptible bool `json:"isPreemptible,omitempty"`

	// MachineTypeUri: The Google Compute Engine machine type used for
	// cluster instances. Example:
	// `https://www.googleapis.com/compute/v1/projects/[project_id]/zones/us-
	// east1-a/machineTypes/n1-standard-2`.
	MachineTypeUri string `json:"machineTypeUri,omitempty"`

	// ManagedGroupConfig: [Output-only] The config for Google Compute
	// Engine Instance Group Manager that manages this group. This is only
	// used for preemptible instance groups.
	ManagedGroupConfig *ManagedGroupConfig `json:"managedGroupConfig,omitempty"`

	// NumInstances: The number of VM instances in the instance group. For
	// master instance groups, must be set to 1.
	NumInstances int64 `json:"numInstances,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DiskConfig") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *InstanceGroupConfig) MarshalJSON() ([]byte, error) {
	type noMethod InstanceGroupConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Job: A Cloud Dataproc job resource.
type Job struct {
	// DriverControlFilesUri: [Output-only] If present, the location of
	// miscellaneous control files which may be used as part of job setup
	// and handling. If not present, control files may be placed in the same
	// location as `driver_output_uri`.
	DriverControlFilesUri string `json:"driverControlFilesUri,omitempty"`

	// DriverOutputResourceUri: [Output-only] A URI pointing to the location
	// of the stdout of the job's driver program.
	DriverOutputResourceUri string `json:"driverOutputResourceUri,omitempty"`

	// HadoopJob: Job is a Hadoop job.
	HadoopJob *HadoopJob `json:"hadoopJob,omitempty"`

	// HiveJob: Job is a Hive job.
	HiveJob *HiveJob `json:"hiveJob,omitempty"`

	// PigJob: Job is a Pig job.
	PigJob *PigJob `json:"pigJob,omitempty"`

	// Placement: [Required] Job information, including how, when, and where
	// to run the job.
	Placement *JobPlacement `json:"placement,omitempty"`

	// PysparkJob: Job is a Pyspark job.
	PysparkJob *PySparkJob `json:"pysparkJob,omitempty"`

	// Reference: [Optional] The fully qualified reference to the job, which
	// can be used to obtain the equivalent REST path of the job resource.
	// If this property is not specified when a job is created, the server
	// generates a job_id.
	Reference *JobReference `json:"reference,omitempty"`

	// SparkJob: Job is a Spark job.
	SparkJob *SparkJob `json:"sparkJob,omitempty"`

	// SparkSqlJob: Job is a SparkSql job.
	SparkSqlJob *SparkSqlJob `json:"sparkSqlJob,omitempty"`

	// Status: [Output-only] The job status. Additional application-specific
	// status information may be contained in the type_job and
	// yarn_applications fields.
	Status *JobStatus `json:"status,omitempty"`

	// StatusHistory: [Output-only] The previous job status.
	StatusHistory []*JobStatus `json:"statusHistory,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g.
	// "DriverControlFilesUri") to unconditionally include in API requests.
	// By default, fields with empty values are omitted from API requests.
	// However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Job) MarshalJSON() ([]byte, error) {
	type noMethod Job
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// JobPlacement: Cloud Dataproc job config.
type JobPlacement struct {
	// ClusterName: [Required] The name of the cluster where the job will be
	// submitted.
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterUuid: [Output-only] A cluster UUID generated by the Cloud
	// Dataproc service when the job is submitted.
	ClusterUuid string `json:"clusterUuid,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ClusterName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *JobPlacement) MarshalJSON() ([]byte, error) {
	type noMethod JobPlacement
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// JobReference: Encapsulates the full scoping used to reference a job.
type JobReference struct {
	// JobId: [Required] The job ID, which must be unique within the
	// project. The job ID is generated by the server upon job submission or
	// provided by the user as a means to perform retries without creating
	// duplicate jobs. The ID must contain only letters (a-z, A-Z), numbers
	// (0-9), underscores (_), or hyphens (-). The maximum length is 512
	// characters.
	JobId string `json:"jobId,omitempty"`

	// ProjectId: [Required] The ID of the Google Cloud Platform project
	// that the job belongs to.
	ProjectId string `json:"projectId,omitempty"`

	// ForceSendFields is a list of field names (e.g. "JobId") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *JobReference) MarshalJSON() ([]byte, error) {
	type noMethod JobReference
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// JobStatus: Cloud Dataproc job status.
type JobStatus struct {
	// Details: [Optional] Job state details, such as an error description
	// if the state is ERROR.
	Details string `json:"details,omitempty"`

	// State: [Required] A state message specifying the overall job state.
	//
	// Possible values:
	//   "STATE_UNSPECIFIED"
	//   "PENDING"
	//   "SETUP_DONE"
	//   "RUNNING"
	//   "CANCEL_PENDING"
	//   "CANCEL_STARTED"
	//   "CANCELLED"
	//   "DONE"
	//   "ERROR"
	State string `json:"state,omitempty"`

	// StateStartTime: [Output-only] The time when this state was entered.
	StateStartTime string `json:"stateStartTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Details") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *JobStatus) MarshalJSON() ([]byte, error) {
	type noMethod JobStatus
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ListClustersResponse: The list of all clusters in a project.
type ListClustersResponse struct {
	// Clusters: [Output-only] The clusters in the project.
	Clusters []*Cluster `json:"clusters,omitempty"`

	// NextPageToken: [Optional] This token is included in the response if
	// there are more results to fetch. To fetch additional results, provide
	// this value as the `page_token` in a subsequent ListClustersRequest.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Clusters") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListClustersResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListClustersResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ListJobsResponse: A list of jobs in a project.
type ListJobsResponse struct {
	// Jobs: [Output-only] Jobs list.
	Jobs []*Job `json:"jobs,omitempty"`

	// NextPageToken: [Optional] This token is included in the response if
	// there are more results to fetch. To fetch additional results, provide
	// this value as the `page_token` in a subsequent ListJobsRequest.
	NextPageToken string `json:"nextPageToken,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Jobs") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ListJobsResponse) MarshalJSON() ([]byte, error) {
	type noMethod ListJobsResponse
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

// LoggingConfig: The runtime logging config of the job.
type LoggingConfig struct {
	// DriverLogLevels: The per-package log levels for the driver. This may
	// include "root" package name to configure rootLogger. Examples:
	// 'com.google = FATAL', 'root = INFO', 'org.apache = DEBUG'
	DriverLogLevels map[string]string `json:"driverLogLevels,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DriverLogLevels") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LoggingConfig) MarshalJSON() ([]byte, error) {
	type noMethod LoggingConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ManagedGroupConfig: Specifies the resources used to actively manage
// an instance group.
type ManagedGroupConfig struct {
	// InstanceGroupManagerName: [Output-only] The name of the Instance
	// Group Manager for this group.
	InstanceGroupManagerName string `json:"instanceGroupManagerName,omitempty"`

	// InstanceTemplateName: [Output-only] The name of the Instance Template
	// used for the Managed Instance Group.
	InstanceTemplateName string `json:"instanceTemplateName,omitempty"`

	// ForceSendFields is a list of field names (e.g.
	// "InstanceGroupManagerName") to unconditionally include in API
	// requests. By default, fields with empty values are omitted from API
	// requests. However, any non-pointer, non-interface field appearing in
	// ForceSendFields will be sent to the server regardless of whether the
	// field is empty or not. This may be used to include empty fields in
	// Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ManagedGroupConfig) MarshalJSON() ([]byte, error) {
	type noMethod ManagedGroupConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Media: Media resource.
type Media struct {
	// ResourceName: Name of the media resource.
	ResourceName string `json:"resourceName,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "ResourceName") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Media) MarshalJSON() ([]byte, error) {
	type noMethod Media
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// NodeInitializationAction: Specifies an executable to run on a fully
// configured node and a timeout period for executable completion.
type NodeInitializationAction struct {
	// ExecutableFile: [Required] Google Cloud Storage URI of executable
	// file.
	ExecutableFile string `json:"executableFile,omitempty"`

	// ExecutionTimeout: [Optional] Amount of time executable has to
	// complete. Default is 10 minutes. Cluster creation fails with an
	// explanatory error message (the name of the executable that caused the
	// error and the exceeded timeout period) if the executable is not
	// completed at end of the timeout period.
	ExecutionTimeout string `json:"executionTimeout,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ExecutableFile") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *NodeInitializationAction) MarshalJSON() ([]byte, error) {
	type noMethod NodeInitializationAction
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

	// Metadata: Service-specific metadata associated with the operation. It
	// typically contains progress information and common metadata such as
	// create time. Some services might not provide such metadata. Any
	// method that returns a long-running operation should document the
	// metadata type, if any.
	Metadata OperationMetadata `json:"metadata,omitempty"`

	// Name: The server-assigned name, which is only unique within the same
	// service that originally returns it. If you use the default HTTP
	// mapping, the `name` should have the format of
	// `operations/some/unique/name`.
	Name string `json:"name,omitempty"`

	// Response: The normal response of the operation in case of success. If
	// the original method returns no data on success, such as `Delete`, the
	// response is `google.protobuf.Empty`. If the original method is
	// standard `Get`/`Create`/`Update`, the response should be the
	// resource. For other methods, the response should have the type
	// `XxxResponse`, where `Xxx` is the original method name. For example,
	// if the original method name is `TakeSnapshot()`, the inferred
	// response type is `TakeSnapshotResponse`.
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

// OperationMetadata1: Metadata describing the operation.
type OperationMetadata1 struct {
	// ClusterName: Name of the cluster for the operation.
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterUuid: Cluster UUId for the operation.
	ClusterUuid string `json:"clusterUuid,omitempty"`

	// Description: [Output-only] Short description of operation.
	Description string `json:"description,omitempty"`

	// Details: A message containing any operation metadata details.
	Details string `json:"details,omitempty"`

	// EndTime: The time that the operation completed.
	EndTime string `json:"endTime,omitempty"`

	// InnerState: A message containing the detailed operation state.
	InnerState string `json:"innerState,omitempty"`

	// InsertTime: The time that the operation was requested.
	InsertTime string `json:"insertTime,omitempty"`

	// OperationType: [Output-only] The operation type.
	OperationType string `json:"operationType,omitempty"`

	// StartTime: The time that the operation was started by the server.
	StartTime string `json:"startTime,omitempty"`

	// State: A message containing the operation state.
	//
	// Possible values:
	//   "UNKNOWN"
	//   "PENDING"
	//   "RUNNING"
	//   "DONE"
	State string `json:"state,omitempty"`

	// Status: [Output-only] Current operation status.
	Status *OperationStatus `json:"status,omitempty"`

	// StatusHistory: [Output-only] Previous operation status.
	StatusHistory []*OperationStatus `json:"statusHistory,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ClusterName") to
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

// OperationStatus: The status of the operation.
type OperationStatus struct {
	// Details: A message containing any operation metadata details.
	Details string `json:"details,omitempty"`

	// InnerState: A message containing the detailed operation state.
	InnerState string `json:"innerState,omitempty"`

	// State: A message containing the operation state.
	//
	// Possible values:
	//   "UNKNOWN"
	//   "PENDING"
	//   "RUNNING"
	//   "DONE"
	State string `json:"state,omitempty"`

	// StateStartTime: The time this state was entered.
	StateStartTime string `json:"stateStartTime,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Details") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *OperationStatus) MarshalJSON() ([]byte, error) {
	type noMethod OperationStatus
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PigJob: A Cloud Dataproc job for running Pig queries on YARN.
type PigJob struct {
	// ContinueOnFailure: [Optional] Whether to continue executing queries
	// if a query fails. The default value is `false`. Setting to `true` can
	// be useful when executing independent parallel queries.
	ContinueOnFailure bool `json:"continueOnFailure,omitempty"`

	// JarFileUris: [Optional] HCFS URIs of jar files to add to the
	// CLASSPATH of the Pig Client and Hadoop MapReduce (MR) tasks. Can
	// contain Pig UDFs.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// LoggingConfig: [Optional] The runtime log config for job execution.
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`

	// Properties: [Optional] A mapping of property names to values, used to
	// configure Pig. Properties that conflict with values set by the Cloud
	// Dataproc API may be overwritten. Can include properties set in
	// /etc/hadoop/conf/*-site.xml, /etc/pig/conf/pig.properties, and
	// classes in user code.
	Properties map[string]string `json:"properties,omitempty"`

	// QueryFileUri: The HCFS URI of the script that contains the Pig
	// queries.
	QueryFileUri string `json:"queryFileUri,omitempty"`

	// QueryList: A list of queries.
	QueryList *QueryList `json:"queryList,omitempty"`

	// ScriptVariables: [Optional] Mapping of query variable names to values
	// (equivalent to the Pig command: `name=[value]`).
	ScriptVariables map[string]string `json:"scriptVariables,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ContinueOnFailure")
	// to unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PigJob) MarshalJSON() ([]byte, error) {
	type noMethod PigJob
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// PySparkJob: A Cloud Dataproc job for running PySpark applications on
// YARN.
type PySparkJob struct {
	// ArchiveUris: [Optional] HCFS URIs of archives to be extracted in the
	// working directory of .jar, .tar, .tar.gz, .tgz, and .zip.
	ArchiveUris []string `json:"archiveUris,omitempty"`

	// Args: [Optional] The arguments to pass to the driver. Do not include
	// arguments, such as `--conf`, that can be set as job properties, since
	// a collision may occur that causes an incorrect job submission.
	Args []string `json:"args,omitempty"`

	// FileUris: [Optional] HCFS URIs of files to be copied to the working
	// directory of Python drivers and distributed tasks. Useful for naively
	// parallel tasks.
	FileUris []string `json:"fileUris,omitempty"`

	// JarFileUris: [Optional] HCFS URIs of jar files to add to the
	// CLASSPATHs of the Python driver and tasks.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// LoggingConfig: [Optional] The runtime log config for job execution.
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`

	// MainPythonFileUri: [Required] The HCFS URI of the main Python file to
	// use as the driver. Must be a .py file.
	MainPythonFileUri string `json:"mainPythonFileUri,omitempty"`

	// Properties: [Optional] A mapping of property names to values, used to
	// configure PySpark. Properties that conflict with values set by the
	// Cloud Dataproc API may be overwritten. Can include properties set in
	// /etc/spark/conf/spark-defaults.conf and classes in user code.
	Properties map[string]string `json:"properties,omitempty"`

	// PythonFileUris: [Optional] HCFS file URIs of Python files to pass to
	// the PySpark framework. Supported file types: .py, .egg, and .zip.
	PythonFileUris []string `json:"pythonFileUris,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ArchiveUris") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *PySparkJob) MarshalJSON() ([]byte, error) {
	type noMethod PySparkJob
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// QueryList: A list of queries to run on a cluster.
type QueryList struct {
	// Queries: [Required] The queries to execute. You do not need to
	// terminate a query with a semicolon. Multiple queries can be specified
	// in one string by separating each with a semicolon. Here is an example
	// of an Cloud Dataproc API snippet that uses a QueryList to specify a
	// HiveJob: "hiveJob": { "queryList": { "queries": [ "query1", "query2",
	// "query3;query4", ] } }
	Queries []string `json:"queries,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Queries") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *QueryList) MarshalJSON() ([]byte, error) {
	type noMethod QueryList
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SoftwareConfig: Specifies the selection and config of software inside
// the cluster.
type SoftwareConfig struct {
	// ImageVersion: [Optional] The version of software inside the cluster.
	// It must match the regular expression `[0-9]+\.[0-9]+`. If
	// unspecified, it defaults to the latest version (see [Cloud Dataproc
	// Versioning](/dataproc/versioning)).
	ImageVersion string `json:"imageVersion,omitempty"`

	// Properties: [Optional] The properties to set on daemon config files.
	// Property keys are specified in `prefix:property` format, such as
	// `core:fs.defaultFS`. The following are supported prefixes and their
	// mappings: * core: `core-site.xml` * hdfs: `hdfs-site.xml` * mapred:
	// `mapred-site.xml` * yarn: `yarn-site.xml` * hive: `hive-site.xml` *
	// pig: `pig.properties` * spark: `spark-defaults.conf`
	Properties map[string]string `json:"properties,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ImageVersion") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SoftwareConfig) MarshalJSON() ([]byte, error) {
	type noMethod SoftwareConfig
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SparkJob: A Cloud Dataproc job for running Spark applications on
// YARN.
type SparkJob struct {
	// ArchiveUris: [Optional] HCFS URIs of archives to be extracted in the
	// working directory of Spark drivers and tasks. Supported file types:
	// .jar, .tar, .tar.gz, .tgz, and .zip.
	ArchiveUris []string `json:"archiveUris,omitempty"`

	// Args: [Optional] The arguments to pass to the driver. Do not include
	// arguments, such as `--conf`, that can be set as job properties, since
	// a collision may occur that causes an incorrect job submission.
	Args []string `json:"args,omitempty"`

	// FileUris: [Optional] HCFS URIs of files to be copied to the working
	// directory of Spark drivers and distributed tasks. Useful for naively
	// parallel tasks.
	FileUris []string `json:"fileUris,omitempty"`

	// JarFileUris: [Optional] HCFS URIs of jar files to add to the
	// CLASSPATHs of the Spark driver and tasks.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// LoggingConfig: [Optional] The runtime log config for job execution.
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`

	// MainClass: The name of the driver's main class. The jar file that
	// contains the class must be in the default CLASSPATH or specified in
	// `jar_file_uris`.
	MainClass string `json:"mainClass,omitempty"`

	// MainJarFileUri: The HCFS URI of the jar file that contains the main
	// class.
	MainJarFileUri string `json:"mainJarFileUri,omitempty"`

	// Properties: [Optional] A mapping of property names to values, used to
	// configure Spark. Properties that conflict with values set by the
	// Cloud Dataproc API may be overwritten. Can include properties set in
	// /etc/spark/conf/spark-defaults.conf and classes in user code.
	Properties map[string]string `json:"properties,omitempty"`

	// ForceSendFields is a list of field names (e.g. "ArchiveUris") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SparkJob) MarshalJSON() ([]byte, error) {
	type noMethod SparkJob
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SparkSqlJob: A Cloud Dataproc job for running Spark SQL queries.
type SparkSqlJob struct {
	// JarFileUris: [Optional] HCFS URIs of jar files to be added to the
	// Spark CLASSPATH.
	JarFileUris []string `json:"jarFileUris,omitempty"`

	// LoggingConfig: [Optional] The runtime log config for job execution.
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`

	// Properties: [Optional] A mapping of property names to values, used to
	// configure Spark SQL's SparkConf. Properties that conflict with values
	// set by the Cloud Dataproc API may be overwritten.
	Properties map[string]string `json:"properties,omitempty"`

	// QueryFileUri: The HCFS URI of the script that contains SQL queries.
	QueryFileUri string `json:"queryFileUri,omitempty"`

	// QueryList: A list of queries.
	QueryList *QueryList `json:"queryList,omitempty"`

	// ScriptVariables: [Optional] Mapping of query variable names to values
	// (equivalent to the Spark SQL command: SET `name="value";`).
	ScriptVariables map[string]string `json:"scriptVariables,omitempty"`

	// ForceSendFields is a list of field names (e.g. "JarFileUris") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SparkSqlJob) MarshalJSON() ([]byte, error) {
	type noMethod SparkSqlJob
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

// SubmitJobRequest: A request to submit a job.
type SubmitJobRequest struct {
	// Job: [Required] The job resource.
	Job *Job `json:"job,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Job") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SubmitJobRequest) MarshalJSON() ([]byte, error) {
	type noMethod SubmitJobRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "dataproc.media.download":

type MediaDownloadCall struct {
	s            *Service
	resourceName string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Download: Method for media download. Download is supported on the URI
// `/v1/media/{+name}?alt=media`.
func (r *MediaService) Download(resourceName string) *MediaDownloadCall {
	c := &MediaDownloadCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resourceName = resourceName
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MediaDownloadCall) Fields(s ...googleapi.Field) *MediaDownloadCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *MediaDownloadCall) IfNoneMatch(entityTag string) *MediaDownloadCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do and Download
// methods. Any pending HTTP request will be aborted if the provided
// context is canceled.
func (c *MediaDownloadCall) Context(ctx context.Context) *MediaDownloadCall {
	c.ctx_ = ctx
	return c
}

func (c *MediaDownloadCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/media/{+resourceName}")
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

// Download fetches the API endpoint's "media" value, instead of the normal
// API response value. If the returned error is nil, the Response is guaranteed to
// have a 2xx status code. Callers must close the Response.Body as usual.
func (c *MediaDownloadCall) Download(opts ...googleapi.CallOption) (*http.Response, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("media")
	if err != nil {
		return nil, err
	}
	if err := googleapi.CheckMediaResponse(res); err != nil {
		res.Body.Close()
		return nil, err
	}
	return res, nil
}

// Do executes the "dataproc.media.download" call.
// Exactly one of *Media or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Media.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *MediaDownloadCall) Do(opts ...googleapi.CallOption) (*Media, error) {
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
	ret := &Media{
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
	//   "description": "Method for media download. Download is supported on the URI `/v1/media/{+name}?alt=media`.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.media.download",
	//   "parameterOrder": [
	//     "resourceName"
	//   ],
	//   "parameters": {
	//     "resourceName": {
	//       "description": "Name of the media that is being downloaded. See ByteStream.ReadRequest.resource_name.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/media/{+resourceName}",
	//   "response": {
	//     "$ref": "Media"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/logging.admin",
	//     "https://www.googleapis.com/auth/logging.read",
	//     "https://www.googleapis.com/auth/logging.write"
	//   ],
	//   "supportsMediaDownload": true
	// }

}

// method id "dataproc.media.upload":

type MediaUploadCall struct {
	s                *Service
	resourceName     string
	media            *Media
	urlParams_       gensupport.URLParams
	media_           io.Reader
	resumableBuffer_ *gensupport.ResumableBuffer
	mediaType_       string
	mediaSize_       int64 // mediaSize, if known.  Used only for calls to progressUpdater_.
	progressUpdater_ googleapi.ProgressUpdater
	ctx_             context.Context
}

// Upload: Method for media upload. Upload is supported on the URI
// `/upload/v1/media/{+name}`.
func (r *MediaService) Upload(resourceName string, media *Media) *MediaUploadCall {
	c := &MediaUploadCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.resourceName = resourceName
	c.media = media
	return c
}

// Media specifies the media to upload in one or more chunks. The chunk
// size may be controlled by supplying a MediaOption generated by
// googleapi.ChunkSize. The chunk size defaults to
// googleapi.DefaultUploadChunkSize.The Content-Type header used in the
// upload request will be determined by sniffing the contents of r,
// unless a MediaOption generated by googleapi.ContentType is
// supplied.
// At most one of Media and ResumableMedia may be set.
func (c *MediaUploadCall) Media(r io.Reader, options ...googleapi.MediaOption) *MediaUploadCall {
	opts := googleapi.ProcessMediaOptions(options)
	chunkSize := opts.ChunkSize
	if !opts.ForceEmptyContentType {
		r, c.mediaType_ = gensupport.DetermineContentType(r, opts.ContentType)
	}
	c.media_, c.resumableBuffer_ = gensupport.PrepareUpload(r, chunkSize)
	return c
}

// ResumableMedia specifies the media to upload in chunks and can be
// canceled with ctx.
//
// Deprecated: use Media instead.
//
// At most one of Media and ResumableMedia may be set. mediaType
// identifies the MIME media type of the upload, such as "image/png". If
// mediaType is "", it will be auto-detected. The provided ctx will
// supersede any context previously provided to the Context method.
func (c *MediaUploadCall) ResumableMedia(ctx context.Context, r io.ReaderAt, size int64, mediaType string) *MediaUploadCall {
	c.ctx_ = ctx
	rdr := gensupport.ReaderAtToReader(r, size)
	rdr, c.mediaType_ = gensupport.DetermineContentType(rdr, mediaType)
	c.resumableBuffer_ = gensupport.NewResumableBuffer(rdr, googleapi.DefaultUploadChunkSize)
	c.media_ = nil
	c.mediaSize_ = size
	return c
}

// ProgressUpdater provides a callback function that will be called
// after every chunk. It should be a low-latency function in order to
// not slow down the upload operation. This should only be called when
// using ResumableMedia (as opposed to Media).
func (c *MediaUploadCall) ProgressUpdater(pu googleapi.ProgressUpdater) *MediaUploadCall {
	c.progressUpdater_ = pu
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *MediaUploadCall) Fields(s ...googleapi.Field) *MediaUploadCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
// This context will supersede any context previously provided to the
// ResumableMedia method.
func (c *MediaUploadCall) Context(ctx context.Context) *MediaUploadCall {
	c.ctx_ = ctx
	return c
}

func (c *MediaUploadCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.media)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/media/{+resourceName}")
	if c.media_ != nil || c.resumableBuffer_ != nil {
		urls = strings.Replace(urls, "https://www.googleapis.com/", "https://www.googleapis.com/upload/", 1)
		protocol := "multipart"
		if c.resumableBuffer_ != nil {
			protocol = "resumable"
		}
		c.urlParams_.Set("uploadType", protocol)
	}
	urls += "?" + c.urlParams_.Encode()
	if c.media_ != nil {
		var combined io.ReadCloser
		combined, ctype = gensupport.CombineBodyMedia(body, ctype, c.media_, c.mediaType_)
		defer combined.Close()
		body = combined
	}
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"resourceName": c.resourceName,
	})
	if c.resumableBuffer_ != nil && c.mediaType_ != "" {
		req.Header.Set("X-Upload-Content-Type", c.mediaType_)
	}
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.media.upload" call.
// Exactly one of *Media or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Media.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *MediaUploadCall) Do(opts ...googleapi.CallOption) (*Media, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := gensupport.Retry(c.ctx_, func() (*http.Response, error) {
		return c.doRequest("json")
	}, gensupport.DefaultBackoffStrategy())
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
	if c.resumableBuffer_ != nil {
		loc := res.Header.Get("Location")
		rx := &gensupport.ResumableUpload{
			Client:    c.s.client,
			UserAgent: c.s.userAgent(),
			URI:       loc,
			Media:     c.resumableBuffer_,
			MediaType: c.mediaType_,
			Callback: func(curr int64) {
				if c.progressUpdater_ != nil {
					c.progressUpdater_(curr, c.mediaSize_)
				}
			},
		}
		ctx := c.ctx_
		if ctx == nil {
			ctx = context.TODO()
		}
		res, err = rx.Upload(ctx)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if err := googleapi.CheckResponse(res); err != nil {
			return nil, err
		}
	}
	ret := &Media{
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
	//   "description": "Method for media upload. Upload is supported on the URI `/upload/v1/media/{+name}`.",
	//   "httpMethod": "POST",
	//   "id": "dataproc.media.upload",
	//   "mediaUpload": {
	//     "accept": [
	//       "*/*"
	//     ],
	//     "protocols": {
	//       "resumable": {
	//         "multipart": true,
	//         "path": "/resumable/upload/v1/media/{+resourceName}"
	//       },
	//       "simple": {
	//         "multipart": true,
	//         "path": "/upload/v1/media/{+resourceName}"
	//       }
	//     }
	//   },
	//   "parameterOrder": [
	//     "resourceName"
	//   ],
	//   "parameters": {
	//     "resourceName": {
	//       "description": "Name of the media that is being downloaded. See ByteStream.ReadRequest.resource_name.",
	//       "location": "path",
	//       "pattern": "^.*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/media/{+resourceName}",
	//   "request": {
	//     "$ref": "Media"
	//   },
	//   "response": {
	//     "$ref": "Media"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform",
	//     "https://www.googleapis.com/auth/logging.admin",
	//     "https://www.googleapis.com/auth/logging.read",
	//     "https://www.googleapis.com/auth/logging.write"
	//   ],
	//   "supportsMediaUpload": true
	// }

}

// method id "dataproc.projects.regions.clusters.create":

type ProjectsRegionsClustersCreateCall struct {
	s          *Service
	projectId  string
	region     string
	cluster    *Cluster
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Create: Creates a cluster in a project.
func (r *ProjectsRegionsClustersService) Create(projectId string, region string, cluster *Cluster) *ProjectsRegionsClustersCreateCall {
	c := &ProjectsRegionsClustersCreateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.cluster = cluster
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersCreateCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersCreateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersCreateCall) Context(ctx context.Context) *ProjectsRegionsClustersCreateCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersCreateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.cluster)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.clusters.create" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsRegionsClustersCreateCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	//   "description": "Creates a cluster in a project.",
	//   "httpMethod": "POST",
	//   "id": "dataproc.projects.regions.clusters.create",
	//   "parameterOrder": [
	//     "projectId",
	//     "region"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters",
	//   "request": {
	//     "$ref": "Cluster"
	//   },
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.clusters.delete":

type ProjectsRegionsClustersDeleteCall struct {
	s           *Service
	projectId   string
	region      string
	clusterName string
	urlParams_  gensupport.URLParams
	ctx_        context.Context
}

// Delete: Deletes a cluster in a project.
func (r *ProjectsRegionsClustersService) Delete(projectId string, region string, clusterName string) *ProjectsRegionsClustersDeleteCall {
	c := &ProjectsRegionsClustersDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.clusterName = clusterName
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersDeleteCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersDeleteCall) Context(ctx context.Context) *ProjectsRegionsClustersDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId":   c.projectId,
		"region":      c.region,
		"clusterName": c.clusterName,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.clusters.delete" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsRegionsClustersDeleteCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	//   "description": "Deletes a cluster in a project.",
	//   "httpMethod": "DELETE",
	//   "id": "dataproc.projects.regions.clusters.delete",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "clusterName"
	//   ],
	//   "parameters": {
	//     "clusterName": {
	//       "description": "[Required] The cluster name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}",
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.clusters.diagnose":

type ProjectsRegionsClustersDiagnoseCall struct {
	s                      *Service
	projectId              string
	region                 string
	clusterName            string
	diagnoseclusterrequest *DiagnoseClusterRequest
	urlParams_             gensupport.URLParams
	ctx_                   context.Context
}

// Diagnose: Gets cluster diagnostic information. After the operation
// completes, the Operation.response field contains
// `DiagnoseClusterOutputLocation`.
func (r *ProjectsRegionsClustersService) Diagnose(projectId string, region string, clusterName string, diagnoseclusterrequest *DiagnoseClusterRequest) *ProjectsRegionsClustersDiagnoseCall {
	c := &ProjectsRegionsClustersDiagnoseCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.clusterName = clusterName
	c.diagnoseclusterrequest = diagnoseclusterrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersDiagnoseCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersDiagnoseCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersDiagnoseCall) Context(ctx context.Context) *ProjectsRegionsClustersDiagnoseCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersDiagnoseCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.diagnoseclusterrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}:diagnose")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId":   c.projectId,
		"region":      c.region,
		"clusterName": c.clusterName,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.clusters.diagnose" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsRegionsClustersDiagnoseCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	//   "description": "Gets cluster diagnostic information. After the operation completes, the Operation.response field contains `DiagnoseClusterOutputLocation`.",
	//   "httpMethod": "POST",
	//   "id": "dataproc.projects.regions.clusters.diagnose",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "clusterName"
	//   ],
	//   "parameters": {
	//     "clusterName": {
	//       "description": "[Required] The cluster name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}:diagnose",
	//   "request": {
	//     "$ref": "DiagnoseClusterRequest"
	//   },
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.clusters.get":

type ProjectsRegionsClustersGetCall struct {
	s            *Service
	projectId    string
	region       string
	clusterName  string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets the resource representation for a cluster in a project.
func (r *ProjectsRegionsClustersService) Get(projectId string, region string, clusterName string) *ProjectsRegionsClustersGetCall {
	c := &ProjectsRegionsClustersGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.clusterName = clusterName
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersGetCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsClustersGetCall) IfNoneMatch(entityTag string) *ProjectsRegionsClustersGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersGetCall) Context(ctx context.Context) *ProjectsRegionsClustersGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId":   c.projectId,
		"region":      c.region,
		"clusterName": c.clusterName,
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

// Do executes the "dataproc.projects.regions.clusters.get" call.
// Exactly one of *Cluster or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Cluster.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsRegionsClustersGetCall) Do(opts ...googleapi.CallOption) (*Cluster, error) {
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
	ret := &Cluster{
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
	//   "description": "Gets the resource representation for a cluster in a project.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.projects.regions.clusters.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "clusterName"
	//   ],
	//   "parameters": {
	//     "clusterName": {
	//       "description": "[Required] The cluster name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}",
	//   "response": {
	//     "$ref": "Cluster"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.clusters.list":

type ProjectsRegionsClustersListCall struct {
	s            *Service
	projectId    string
	region       string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists all regions/{region}/clusters in a project.
func (r *ProjectsRegionsClustersService) List(projectId string, region string) *ProjectsRegionsClustersListCall {
	c := &ProjectsRegionsClustersListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	return c
}

// PageSize sets the optional parameter "pageSize": The standard List
// page size.
func (c *ProjectsRegionsClustersListCall) PageSize(pageSize int64) *ProjectsRegionsClustersListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The standard List
// page token.
func (c *ProjectsRegionsClustersListCall) PageToken(pageToken string) *ProjectsRegionsClustersListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersListCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsClustersListCall) IfNoneMatch(entityTag string) *ProjectsRegionsClustersListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersListCall) Context(ctx context.Context) *ProjectsRegionsClustersListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
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

// Do executes the "dataproc.projects.regions.clusters.list" call.
// Exactly one of *ListClustersResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListClustersResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsRegionsClustersListCall) Do(opts ...googleapi.CallOption) (*ListClustersResponse, error) {
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
	ret := &ListClustersResponse{
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
	//   "description": "Lists all regions/{region}/clusters in a project.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.projects.regions.clusters.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "region"
	//   ],
	//   "parameters": {
	//     "pageSize": {
	//       "description": "The standard List page size.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "The standard List page token.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters",
	//   "response": {
	//     "$ref": "ListClustersResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsRegionsClustersListCall) Pages(ctx context.Context, f func(*ListClustersResponse) error) error {
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

// method id "dataproc.projects.regions.clusters.patch":

type ProjectsRegionsClustersPatchCall struct {
	s           *Service
	projectId   string
	region      string
	clusterName string
	cluster     *Cluster
	urlParams_  gensupport.URLParams
	ctx_        context.Context
}

// Patch: Updates a cluster in a project.
func (r *ProjectsRegionsClustersService) Patch(projectId string, region string, clusterName string, cluster *Cluster) *ProjectsRegionsClustersPatchCall {
	c := &ProjectsRegionsClustersPatchCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.clusterName = clusterName
	c.cluster = cluster
	return c
}

// UpdateMask sets the optional parameter "updateMask": [Required]
// Specifies the path, relative to Cluster, of the field to update. For
// example, to change the number of workers in a cluster to 5, the
// update_mask parameter would be specified as
// config.worker_config.num_instances, and the `PATCH` request body
// would specify the new value, as follows: { "config":{
// "workerConfig":{ "numInstances":"5" } } } Note: Currently,
// config.worker_config.num_instances is the only field that can be
// updated.
func (c *ProjectsRegionsClustersPatchCall) UpdateMask(updateMask string) *ProjectsRegionsClustersPatchCall {
	c.urlParams_.Set("updateMask", updateMask)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsClustersPatchCall) Fields(s ...googleapi.Field) *ProjectsRegionsClustersPatchCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsClustersPatchCall) Context(ctx context.Context) *ProjectsRegionsClustersPatchCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsClustersPatchCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.cluster)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("PATCH", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId":   c.projectId,
		"region":      c.region,
		"clusterName": c.clusterName,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.clusters.patch" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsRegionsClustersPatchCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	//   "description": "Updates a cluster in a project.",
	//   "httpMethod": "PATCH",
	//   "id": "dataproc.projects.regions.clusters.patch",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "clusterName"
	//   ],
	//   "parameters": {
	//     "clusterName": {
	//       "description": "[Required] The cluster name.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project the cluster belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "updateMask": {
	//       "description": "[Required] Specifies the path, relative to Cluster, of the field to update. For example, to change the number of workers in a cluster to 5, the update_mask parameter would be specified as config.worker_config.num_instances, and the `PATCH` request body would specify the new value, as follows: { \"config\":{ \"workerConfig\":{ \"numInstances\":\"5\" } } } Note: Currently, config.worker_config.num_instances is the only field that can be updated.",
	//       "location": "query",
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/clusters/{clusterName}",
	//   "request": {
	//     "$ref": "Cluster"
	//   },
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.jobs.cancel":

type ProjectsRegionsJobsCancelCall struct {
	s                *Service
	projectId        string
	region           string
	jobId            string
	canceljobrequest *CancelJobRequest
	urlParams_       gensupport.URLParams
	ctx_             context.Context
}

// Cancel: Starts a job cancellation request. To access the job resource
// after cancellation, call
// [regions/{region}/jobs.list](/dataproc/reference/rest/v1/projects.regi
// ons.jobs/list) or
// [regions/{region}/jobs.get](/dataproc/reference/rest/v1/projects.regio
// ns.jobs/get).
func (r *ProjectsRegionsJobsService) Cancel(projectId string, region string, jobId string, canceljobrequest *CancelJobRequest) *ProjectsRegionsJobsCancelCall {
	c := &ProjectsRegionsJobsCancelCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.jobId = jobId
	c.canceljobrequest = canceljobrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsJobsCancelCall) Fields(s ...googleapi.Field) *ProjectsRegionsJobsCancelCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsJobsCancelCall) Context(ctx context.Context) *ProjectsRegionsJobsCancelCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsJobsCancelCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.canceljobrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/jobs/{jobId}:cancel")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
		"jobId":     c.jobId,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.jobs.cancel" call.
// Exactly one of *Job or error will be non-nil. Any non-2xx status code
// is an error. Response headers are in either
// *Job.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsRegionsJobsCancelCall) Do(opts ...googleapi.CallOption) (*Job, error) {
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
	ret := &Job{
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
	//   "description": "Starts a job cancellation request. To access the job resource after cancellation, call [regions/{region}/jobs.list](/dataproc/reference/rest/v1/projects.regions.jobs/list) or [regions/{region}/jobs.get](/dataproc/reference/rest/v1/projects.regions.jobs/get).",
	//   "httpMethod": "POST",
	//   "id": "dataproc.projects.regions.jobs.cancel",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "jobId"
	//   ],
	//   "parameters": {
	//     "jobId": {
	//       "description": "[Required] The job ID.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the job belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/jobs/{jobId}:cancel",
	//   "request": {
	//     "$ref": "CancelJobRequest"
	//   },
	//   "response": {
	//     "$ref": "Job"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.jobs.delete":

type ProjectsRegionsJobsDeleteCall struct {
	s          *Service
	projectId  string
	region     string
	jobId      string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes the job from the project. If the job is active, the
// delete fails, and the response returns `FAILED_PRECONDITION`.
func (r *ProjectsRegionsJobsService) Delete(projectId string, region string, jobId string) *ProjectsRegionsJobsDeleteCall {
	c := &ProjectsRegionsJobsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.jobId = jobId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsJobsDeleteCall) Fields(s ...googleapi.Field) *ProjectsRegionsJobsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsJobsDeleteCall) Context(ctx context.Context) *ProjectsRegionsJobsDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsJobsDeleteCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/jobs/{jobId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("DELETE", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
		"jobId":     c.jobId,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.jobs.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsRegionsJobsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Deletes the job from the project. If the job is active, the delete fails, and the response returns `FAILED_PRECONDITION`.",
	//   "httpMethod": "DELETE",
	//   "id": "dataproc.projects.regions.jobs.delete",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "jobId"
	//   ],
	//   "parameters": {
	//     "jobId": {
	//       "description": "[Required] The job ID.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the job belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/jobs/{jobId}",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.jobs.get":

type ProjectsRegionsJobsGetCall struct {
	s            *Service
	projectId    string
	region       string
	jobId        string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets the resource representation for a job in a project.
func (r *ProjectsRegionsJobsService) Get(projectId string, region string, jobId string) *ProjectsRegionsJobsGetCall {
	c := &ProjectsRegionsJobsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.jobId = jobId
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsJobsGetCall) Fields(s ...googleapi.Field) *ProjectsRegionsJobsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsJobsGetCall) IfNoneMatch(entityTag string) *ProjectsRegionsJobsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsJobsGetCall) Context(ctx context.Context) *ProjectsRegionsJobsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsJobsGetCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/jobs/{jobId}")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
		"jobId":     c.jobId,
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

// Do executes the "dataproc.projects.regions.jobs.get" call.
// Exactly one of *Job or error will be non-nil. Any non-2xx status code
// is an error. Response headers are in either
// *Job.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsRegionsJobsGetCall) Do(opts ...googleapi.CallOption) (*Job, error) {
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
	ret := &Job{
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
	//   "description": "Gets the resource representation for a job in a project.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.projects.regions.jobs.get",
	//   "parameterOrder": [
	//     "projectId",
	//     "region",
	//     "jobId"
	//   ],
	//   "parameters": {
	//     "jobId": {
	//       "description": "[Required] The job ID.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the job belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/jobs/{jobId}",
	//   "response": {
	//     "$ref": "Job"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.jobs.list":

type ProjectsRegionsJobsListCall struct {
	s            *Service
	projectId    string
	region       string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists regions/{region}/jobs in a project.
func (r *ProjectsRegionsJobsService) List(projectId string, region string) *ProjectsRegionsJobsListCall {
	c := &ProjectsRegionsJobsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	return c
}

// ClusterName sets the optional parameter "clusterName": [Optional] If
// set, the returned jobs list includes only jobs that were submitted to
// the named cluster.
func (c *ProjectsRegionsJobsListCall) ClusterName(clusterName string) *ProjectsRegionsJobsListCall {
	c.urlParams_.Set("clusterName", clusterName)
	return c
}

// JobStateMatcher sets the optional parameter "jobStateMatcher":
// [Optional] Specifies enumerated categories of jobs to list.
//
// Possible values:
//   "ALL"
//   "ACTIVE"
//   "NON_ACTIVE"
func (c *ProjectsRegionsJobsListCall) JobStateMatcher(jobStateMatcher string) *ProjectsRegionsJobsListCall {
	c.urlParams_.Set("jobStateMatcher", jobStateMatcher)
	return c
}

// PageSize sets the optional parameter "pageSize": [Optional] The
// number of results to return in each response.
func (c *ProjectsRegionsJobsListCall) PageSize(pageSize int64) *ProjectsRegionsJobsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": [Optional] The
// page token, returned by a previous call, to request the next page of
// results.
func (c *ProjectsRegionsJobsListCall) PageToken(pageToken string) *ProjectsRegionsJobsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsJobsListCall) Fields(s ...googleapi.Field) *ProjectsRegionsJobsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsJobsListCall) IfNoneMatch(entityTag string) *ProjectsRegionsJobsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsJobsListCall) Context(ctx context.Context) *ProjectsRegionsJobsListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsJobsListCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/jobs")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("GET", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
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

// Do executes the "dataproc.projects.regions.jobs.list" call.
// Exactly one of *ListJobsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListJobsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsRegionsJobsListCall) Do(opts ...googleapi.CallOption) (*ListJobsResponse, error) {
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
	ret := &ListJobsResponse{
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
	//   "description": "Lists regions/{region}/jobs in a project.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.projects.regions.jobs.list",
	//   "parameterOrder": [
	//     "projectId",
	//     "region"
	//   ],
	//   "parameters": {
	//     "clusterName": {
	//       "description": "[Optional] If set, the returned jobs list includes only jobs that were submitted to the named cluster.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "jobStateMatcher": {
	//       "description": "[Optional] Specifies enumerated categories of jobs to list.",
	//       "enum": [
	//         "ALL",
	//         "ACTIVE",
	//         "NON_ACTIVE"
	//       ],
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "[Optional] The number of results to return in each response.",
	//       "format": "int32",
	//       "location": "query",
	//       "type": "integer"
	//     },
	//     "pageToken": {
	//       "description": "[Optional] The page token, returned by a previous call, to request the next page of results.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the job belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/jobs",
	//   "response": {
	//     "$ref": "ListJobsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsRegionsJobsListCall) Pages(ctx context.Context, f func(*ListJobsResponse) error) error {
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

// method id "dataproc.projects.regions.jobs.submit":

type ProjectsRegionsJobsSubmitCall struct {
	s                *Service
	projectId        string
	region           string
	submitjobrequest *SubmitJobRequest
	urlParams_       gensupport.URLParams
	ctx_             context.Context
}

// Submit: Submits a job to a cluster.
func (r *ProjectsRegionsJobsService) Submit(projectId string, region string, submitjobrequest *SubmitJobRequest) *ProjectsRegionsJobsSubmitCall {
	c := &ProjectsRegionsJobsSubmitCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.projectId = projectId
	c.region = region
	c.submitjobrequest = submitjobrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsJobsSubmitCall) Fields(s ...googleapi.Field) *ProjectsRegionsJobsSubmitCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsJobsSubmitCall) Context(ctx context.Context) *ProjectsRegionsJobsSubmitCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsJobsSubmitCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.submitjobrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/projects/{projectId}/regions/{region}/jobs:submit")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"projectId": c.projectId,
		"region":    c.region,
	})
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.jobs.submit" call.
// Exactly one of *Job or error will be non-nil. Any non-2xx status code
// is an error. Response headers are in either
// *Job.ServerResponse.Header or (if a response was returned at all) in
// error.(*googleapi.Error).Header. Use googleapi.IsNotModified to check
// whether the returned error was because http.StatusNotModified was
// returned.
func (c *ProjectsRegionsJobsSubmitCall) Do(opts ...googleapi.CallOption) (*Job, error) {
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
	ret := &Job{
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
	//   "description": "Submits a job to a cluster.",
	//   "httpMethod": "POST",
	//   "id": "dataproc.projects.regions.jobs.submit",
	//   "parameterOrder": [
	//     "projectId",
	//     "region"
	//   ],
	//   "parameters": {
	//     "projectId": {
	//       "description": "[Required] The ID of the Google Cloud Platform project that the job belongs to.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "region": {
	//       "description": "[Required] The Cloud Dataproc region in which to handle the request.",
	//       "location": "path",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/projects/{projectId}/regions/{region}/jobs:submit",
	//   "request": {
	//     "$ref": "SubmitJobRequest"
	//   },
	//   "response": {
	//     "$ref": "Job"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.operations.cancel":

type ProjectsRegionsOperationsCancelCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Cancel: Starts asynchronous cancellation on a long-running operation.
// The server makes a best effort to cancel the operation, but success
// is not guaranteed. If the server doesn't support this method, it
// returns `google.rpc.Code.UNIMPLEMENTED`. Clients can use
// Operations.GetOperation or other methods to check whether the
// cancellation succeeded or whether the operation completed despite
// cancellation.
func (r *ProjectsRegionsOperationsService) Cancel(name string) *ProjectsRegionsOperationsCancelCall {
	c := &ProjectsRegionsOperationsCancelCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsOperationsCancelCall) Fields(s ...googleapi.Field) *ProjectsRegionsOperationsCancelCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsOperationsCancelCall) Context(ctx context.Context) *ProjectsRegionsOperationsCancelCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsOperationsCancelCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/{+name}:cancel")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.Expand(req.URL, map[string]string{
		"name": c.name,
	})
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "dataproc.projects.regions.operations.cancel" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsRegionsOperationsCancelCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Starts asynchronous cancellation on a long-running operation. The server makes a best effort to cancel the operation, but success is not guaranteed. If the server doesn't support this method, it returns `google.rpc.Code.UNIMPLEMENTED`. Clients can use Operations.GetOperation or other methods to check whether the cancellation succeeded or whether the operation completed despite cancellation.",
	//   "httpMethod": "POST",
	//   "id": "dataproc.projects.regions.operations.cancel",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The name of the operation resource to be cancelled.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/regions/[^/]*/operations/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}:cancel",
	//   "response": {
	//     "$ref": "Empty"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.operations.delete":

type ProjectsRegionsOperationsDeleteCall struct {
	s          *Service
	name       string
	urlParams_ gensupport.URLParams
	ctx_       context.Context
}

// Delete: Deletes a long-running operation. This method indicates that
// the client is no longer interested in the operation result. It does
// not cancel the operation. If the server doesn't support this method,
// it returns `google.rpc.Code.UNIMPLEMENTED`.
func (r *ProjectsRegionsOperationsService) Delete(name string) *ProjectsRegionsOperationsDeleteCall {
	c := &ProjectsRegionsOperationsDeleteCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsOperationsDeleteCall) Fields(s ...googleapi.Field) *ProjectsRegionsOperationsDeleteCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsOperationsDeleteCall) Context(ctx context.Context) *ProjectsRegionsOperationsDeleteCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsOperationsDeleteCall) doRequest(alt string) (*http.Response, error) {
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

// Do executes the "dataproc.projects.regions.operations.delete" call.
// Exactly one of *Empty or error will be non-nil. Any non-2xx status
// code is an error. Response headers are in either
// *Empty.ServerResponse.Header or (if a response was returned at all)
// in error.(*googleapi.Error).Header. Use googleapi.IsNotModified to
// check whether the returned error was because http.StatusNotModified
// was returned.
func (c *ProjectsRegionsOperationsDeleteCall) Do(opts ...googleapi.CallOption) (*Empty, error) {
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
	//   "description": "Deletes a long-running operation. This method indicates that the client is no longer interested in the operation result. It does not cancel the operation. If the server doesn't support this method, it returns `google.rpc.Code.UNIMPLEMENTED`.",
	//   "httpMethod": "DELETE",
	//   "id": "dataproc.projects.regions.operations.delete",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The name of the operation resource to be deleted.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/regions/[^/]*/operations/[^/]*$",
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

// method id "dataproc.projects.regions.operations.get":

type ProjectsRegionsOperationsGetCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// Get: Gets the latest state of a long-running operation. Clients can
// use this method to poll the operation result at intervals as
// recommended by the API service.
func (r *ProjectsRegionsOperationsService) Get(name string) *ProjectsRegionsOperationsGetCall {
	c := &ProjectsRegionsOperationsGetCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsOperationsGetCall) Fields(s ...googleapi.Field) *ProjectsRegionsOperationsGetCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsOperationsGetCall) IfNoneMatch(entityTag string) *ProjectsRegionsOperationsGetCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsOperationsGetCall) Context(ctx context.Context) *ProjectsRegionsOperationsGetCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsOperationsGetCall) doRequest(alt string) (*http.Response, error) {
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

// Do executes the "dataproc.projects.regions.operations.get" call.
// Exactly one of *Operation or error will be non-nil. Any non-2xx
// status code is an error. Response headers are in either
// *Operation.ServerResponse.Header or (if a response was returned at
// all) in error.(*googleapi.Error).Header. Use googleapi.IsNotModified
// to check whether the returned error was because
// http.StatusNotModified was returned.
func (c *ProjectsRegionsOperationsGetCall) Do(opts ...googleapi.CallOption) (*Operation, error) {
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
	//   "id": "dataproc.projects.regions.operations.get",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "name": {
	//       "description": "The name of the operation resource.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/regions/[^/]*/operations/[^/]*$",
	//       "required": true,
	//       "type": "string"
	//     }
	//   },
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "Operation"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// method id "dataproc.projects.regions.operations.list":

type ProjectsRegionsOperationsListCall struct {
	s            *Service
	name         string
	urlParams_   gensupport.URLParams
	ifNoneMatch_ string
	ctx_         context.Context
}

// List: Lists operations that match the specified filter in the
// request. If the server doesn't support this method, it returns
// `UNIMPLEMENTED`. NOTE: the `name` binding below allows API services
// to override the binding to use different resource name schemes, such
// as `users/*/operations`.
func (r *ProjectsRegionsOperationsService) List(name string) *ProjectsRegionsOperationsListCall {
	c := &ProjectsRegionsOperationsListCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.name = name
	return c
}

// Filter sets the optional parameter "filter": The standard list
// filter.
func (c *ProjectsRegionsOperationsListCall) Filter(filter string) *ProjectsRegionsOperationsListCall {
	c.urlParams_.Set("filter", filter)
	return c
}

// PageSize sets the optional parameter "pageSize": The standard list
// page size.
func (c *ProjectsRegionsOperationsListCall) PageSize(pageSize int64) *ProjectsRegionsOperationsListCall {
	c.urlParams_.Set("pageSize", fmt.Sprint(pageSize))
	return c
}

// PageToken sets the optional parameter "pageToken": The standard list
// page token.
func (c *ProjectsRegionsOperationsListCall) PageToken(pageToken string) *ProjectsRegionsOperationsListCall {
	c.urlParams_.Set("pageToken", pageToken)
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ProjectsRegionsOperationsListCall) Fields(s ...googleapi.Field) *ProjectsRegionsOperationsListCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// IfNoneMatch sets the optional parameter which makes the operation
// fail if the object's ETag matches the given value. This is useful for
// getting updates only after the object has changed since the last
// request. Use googleapi.IsNotModified to check whether the response
// error from Do is the result of In-None-Match.
func (c *ProjectsRegionsOperationsListCall) IfNoneMatch(entityTag string) *ProjectsRegionsOperationsListCall {
	c.ifNoneMatch_ = entityTag
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ProjectsRegionsOperationsListCall) Context(ctx context.Context) *ProjectsRegionsOperationsListCall {
	c.ctx_ = ctx
	return c
}

func (c *ProjectsRegionsOperationsListCall) doRequest(alt string) (*http.Response, error) {
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

// Do executes the "dataproc.projects.regions.operations.list" call.
// Exactly one of *ListOperationsResponse or error will be non-nil. Any
// non-2xx status code is an error. Response headers are in either
// *ListOperationsResponse.ServerResponse.Header or (if a response was
// returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ProjectsRegionsOperationsListCall) Do(opts ...googleapi.CallOption) (*ListOperationsResponse, error) {
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
	//   "description": "Lists operations that match the specified filter in the request. If the server doesn't support this method, it returns `UNIMPLEMENTED`. NOTE: the `name` binding below allows API services to override the binding to use different resource name schemes, such as `users/*/operations`.",
	//   "httpMethod": "GET",
	//   "id": "dataproc.projects.regions.operations.list",
	//   "parameterOrder": [
	//     "name"
	//   ],
	//   "parameters": {
	//     "filter": {
	//       "description": "The standard list filter.",
	//       "location": "query",
	//       "type": "string"
	//     },
	//     "name": {
	//       "description": "The name of the operation collection.",
	//       "location": "path",
	//       "pattern": "^projects/[^/]*/regions/[^/]*/operations$",
	//       "required": true,
	//       "type": "string"
	//     },
	//     "pageSize": {
	//       "description": "The standard list page size.",
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
	//   "path": "v1/{+name}",
	//   "response": {
	//     "$ref": "ListOperationsResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}

// Pages invokes f for each page of results.
// A non-nil error returned from f will halt the iteration.
// The provided context supersedes any context provided to the Context method.
func (c *ProjectsRegionsOperationsListCall) Pages(ctx context.Context, f func(*ListOperationsResponse) error) error {
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
