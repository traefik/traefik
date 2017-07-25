package storage

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Container represents an Azure container.
type Container struct {
	bsc        *BlobStorageClient
	Name       string              `xml:"Name"`
	Properties ContainerProperties `xml:"Properties"`
}

func (c *Container) buildPath() string {
	return fmt.Sprintf("/%s", c.Name)
}

// ContainerProperties contains various properties of a container returned from
// various endpoints like ListContainers.
type ContainerProperties struct {
	LastModified  string `xml:"Last-Modified"`
	Etag          string `xml:"Etag"`
	LeaseStatus   string `xml:"LeaseStatus"`
	LeaseState    string `xml:"LeaseState"`
	LeaseDuration string `xml:"LeaseDuration"`
}

// ContainerListResponse contains the response fields from
// ListContainers call.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd179352.aspx
type ContainerListResponse struct {
	XMLName    xml.Name    `xml:"EnumerationResults"`
	Xmlns      string      `xml:"xmlns,attr"`
	Prefix     string      `xml:"Prefix"`
	Marker     string      `xml:"Marker"`
	NextMarker string      `xml:"NextMarker"`
	MaxResults int64       `xml:"MaxResults"`
	Containers []Container `xml:"Containers>Container"`
}

// BlobListResponse contains the response fields from ListBlobs call.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd135734.aspx
type BlobListResponse struct {
	XMLName    xml.Name `xml:"EnumerationResults"`
	Xmlns      string   `xml:"xmlns,attr"`
	Prefix     string   `xml:"Prefix"`
	Marker     string   `xml:"Marker"`
	NextMarker string   `xml:"NextMarker"`
	MaxResults int64    `xml:"MaxResults"`
	Blobs      []Blob   `xml:"Blobs>Blob"`

	// BlobPrefix is used to traverse blobs as if it were a file system.
	// It is returned if ListBlobsParameters.Delimiter is specified.
	// The list here can be thought of as "folders" that may contain
	// other folders or blobs.
	BlobPrefixes []string `xml:"Blobs>BlobPrefix>Name"`

	// Delimiter is used to traverse blobs as if it were a file system.
	// It is returned if ListBlobsParameters.Delimiter is specified.
	Delimiter string `xml:"Delimiter"`
}

// ListBlobsParameters defines the set of customizable
// parameters to make a List Blobs call.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd135734.aspx
type ListBlobsParameters struct {
	Prefix     string
	Delimiter  string
	Marker     string
	Include    string
	MaxResults uint
	Timeout    uint
}

func (p ListBlobsParameters) getParameters() url.Values {
	out := url.Values{}

	if p.Prefix != "" {
		out.Set("prefix", p.Prefix)
	}
	if p.Delimiter != "" {
		out.Set("delimiter", p.Delimiter)
	}
	if p.Marker != "" {
		out.Set("marker", p.Marker)
	}
	if p.Include != "" {
		out.Set("include", p.Include)
	}
	if p.MaxResults != 0 {
		out.Set("maxresults", fmt.Sprintf("%v", p.MaxResults))
	}
	if p.Timeout != 0 {
		out.Set("timeout", fmt.Sprintf("%v", p.Timeout))
	}

	return out
}

// ContainerAccessType defines the access level to the container from a public
// request.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd179468.aspx and "x-ms-
// blob-public-access" header.
type ContainerAccessType string

// Access options for containers
const (
	ContainerAccessTypePrivate   ContainerAccessType = ""
	ContainerAccessTypeBlob      ContainerAccessType = "blob"
	ContainerAccessTypeContainer ContainerAccessType = "container"
)

// ContainerAccessPolicy represents each access policy in the container ACL.
type ContainerAccessPolicy struct {
	ID         string
	StartTime  time.Time
	ExpiryTime time.Time
	CanRead    bool
	CanWrite   bool
	CanDelete  bool
}

// ContainerPermissions represents the container ACLs.
type ContainerPermissions struct {
	AccessType     ContainerAccessType
	AccessPolicies []ContainerAccessPolicy
}

// ContainerAccessHeader references header used when setting/getting container ACL
const (
	ContainerAccessHeader string = "x-ms-blob-public-access"
)

// Create creates a blob container within the storage account
// with given name and access level. Returns error if container already exists.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd179468.aspx
func (c *Container) Create() error {
	resp, err := c.create()
	if err != nil {
		return err
	}
	defer readAndCloseBody(resp.body)
	return checkRespCode(resp.statusCode, []int{http.StatusCreated})
}

// CreateIfNotExists creates a blob container if it does not exist. Returns
// true if container is newly created or false if container already exists.
func (c *Container) CreateIfNotExists() (bool, error) {
	resp, err := c.create()
	if resp != nil {
		defer readAndCloseBody(resp.body)
		if resp.statusCode == http.StatusCreated || resp.statusCode == http.StatusConflict {
			return resp.statusCode == http.StatusCreated, nil
		}
	}
	return false, err
}

func (c *Container) create() (*storageResponse, error) {
	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), url.Values{"restype": {"container"}})
	headers := c.bsc.client.getStandardHeaders()
	return c.bsc.client.exec(http.MethodPut, uri, headers, nil, c.bsc.auth)
}

// Exists returns true if a container with given name exists
// on the storage account, otherwise returns false.
func (c *Container) Exists() (bool, error) {
	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), url.Values{"restype": {"container"}})
	headers := c.bsc.client.getStandardHeaders()

	resp, err := c.bsc.client.exec(http.MethodHead, uri, headers, nil, c.bsc.auth)
	if resp != nil {
		defer readAndCloseBody(resp.body)
		if resp.statusCode == http.StatusOK || resp.statusCode == http.StatusNotFound {
			return resp.statusCode == http.StatusOK, nil
		}
	}
	return false, err
}

// SetPermissions sets up container permissions as per https://msdn.microsoft.com/en-us/library/azure/dd179391.aspx
func (c *Container) SetPermissions(permissions ContainerPermissions, timeout int, leaseID string) error {
	params := url.Values{
		"restype": {"container"},
		"comp":    {"acl"},
	}

	if timeout > 0 {
		params.Add("timeout", strconv.Itoa(timeout))
	}

	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), params)
	headers := c.bsc.client.getStandardHeaders()
	if permissions.AccessType != "" {
		headers[ContainerAccessHeader] = string(permissions.AccessType)
	}

	if leaseID != "" {
		headers[headerLeaseID] = leaseID
	}

	body, length, err := generateContainerACLpayload(permissions.AccessPolicies)
	headers["Content-Length"] = strconv.Itoa(length)

	resp, err := c.bsc.client.exec(http.MethodPut, uri, headers, body, c.bsc.auth)
	if err != nil {
		return err
	}
	defer readAndCloseBody(resp.body)

	if err := checkRespCode(resp.statusCode, []int{http.StatusOK}); err != nil {
		return errors.New("Unable to set permissions")
	}

	return nil
}

// GetPermissions gets the container permissions as per https://msdn.microsoft.com/en-us/library/azure/dd179469.aspx
// If timeout is 0 then it will not be passed to Azure
// leaseID will only be passed to Azure if populated
func (c *Container) GetPermissions(timeout int, leaseID string) (*ContainerPermissions, error) {
	params := url.Values{
		"restype": {"container"},
		"comp":    {"acl"},
	}

	if timeout > 0 {
		params.Add("timeout", strconv.Itoa(timeout))
	}

	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), params)
	headers := c.bsc.client.getStandardHeaders()

	if leaseID != "" {
		headers[headerLeaseID] = leaseID
	}

	resp, err := c.bsc.client.exec(http.MethodGet, uri, headers, nil, c.bsc.auth)
	if err != nil {
		return nil, err
	}
	defer resp.body.Close()

	var ap AccessPolicy
	err = xmlUnmarshal(resp.body, &ap.SignedIdentifiersList)
	if err != nil {
		return nil, err
	}
	return buildAccessPolicy(ap, &resp.headers), nil
}

func buildAccessPolicy(ap AccessPolicy, headers *http.Header) *ContainerPermissions {
	// containerAccess. Blob, Container, empty
	containerAccess := headers.Get(http.CanonicalHeaderKey(ContainerAccessHeader))
	permissions := ContainerPermissions{
		AccessType:     ContainerAccessType(containerAccess),
		AccessPolicies: []ContainerAccessPolicy{},
	}

	for _, policy := range ap.SignedIdentifiersList.SignedIdentifiers {
		capd := ContainerAccessPolicy{
			ID:         policy.ID,
			StartTime:  policy.AccessPolicy.StartTime,
			ExpiryTime: policy.AccessPolicy.ExpiryTime,
		}
		capd.CanRead = updatePermissions(policy.AccessPolicy.Permission, "r")
		capd.CanWrite = updatePermissions(policy.AccessPolicy.Permission, "w")
		capd.CanDelete = updatePermissions(policy.AccessPolicy.Permission, "d")

		permissions.AccessPolicies = append(permissions.AccessPolicies, capd)
	}
	return &permissions
}

// Delete deletes the container with given name on the storage
// account. If the container does not exist returns error.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd179408.aspx
func (c *Container) Delete() error {
	resp, err := c.delete()
	if err != nil {
		return err
	}
	defer readAndCloseBody(resp.body)
	return checkRespCode(resp.statusCode, []int{http.StatusAccepted})
}

// DeleteIfExists deletes the container with given name on the storage
// account if it exists. Returns true if container is deleted with this call, or
// false if the container did not exist at the time of the Delete Container
// operation.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd179408.aspx
func (c *Container) DeleteIfExists() (bool, error) {
	resp, err := c.delete()
	if resp != nil {
		defer readAndCloseBody(resp.body)
		if resp.statusCode == http.StatusAccepted || resp.statusCode == http.StatusNotFound {
			return resp.statusCode == http.StatusAccepted, nil
		}
	}
	return false, err
}

func (c *Container) delete() (*storageResponse, error) {
	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), url.Values{"restype": {"container"}})
	headers := c.bsc.client.getStandardHeaders()
	return c.bsc.client.exec(http.MethodDelete, uri, headers, nil, c.bsc.auth)
}

// ListBlobs returns an object that contains list of blobs in the container,
// pagination token and other information in the response of List Blobs call.
//
// See https://msdn.microsoft.com/en-us/library/azure/dd135734.aspx
func (c *Container) ListBlobs(params ListBlobsParameters) (BlobListResponse, error) {
	q := mergeParams(params.getParameters(), url.Values{
		"restype": {"container"},
		"comp":    {"list"}},
	)
	uri := c.bsc.client.getEndpoint(blobServiceName, c.buildPath(), q)
	headers := c.bsc.client.getStandardHeaders()

	var out BlobListResponse
	resp, err := c.bsc.client.exec(http.MethodGet, uri, headers, nil, c.bsc.auth)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	err = xmlUnmarshal(resp.body, &out)
	return out, err
}

func generateContainerACLpayload(policies []ContainerAccessPolicy) (io.Reader, int, error) {
	sil := SignedIdentifiers{
		SignedIdentifiers: []SignedIdentifier{},
	}
	for _, capd := range policies {
		permission := capd.generateContainerPermissions()
		signedIdentifier := convertAccessPolicyToXMLStructs(capd.ID, capd.StartTime, capd.ExpiryTime, permission)
		sil.SignedIdentifiers = append(sil.SignedIdentifiers, signedIdentifier)
	}
	return xmlMarshal(sil)
}

func (capd *ContainerAccessPolicy) generateContainerPermissions() (permissions string) {
	// generate the permissions string (rwd).
	// still want the end user API to have bool flags.
	permissions = ""

	if capd.CanRead {
		permissions += "r"
	}

	if capd.CanWrite {
		permissions += "w"
	}

	if capd.CanDelete {
		permissions += "d"
	}

	return permissions
}
