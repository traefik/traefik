package egoscale

// Snapshot represents a volume snapshot
type Snapshot struct {
	ID           string        `json:"id"`
	Account      string        `json:"account"`
	Created      string        `json:"created,omitempty"`
	Domain       string        `json:"domain"`
	DomainID     string        `json:"domainid"`
	IntervalType string        `json:"intervaltype,omitempty"` // hourly, daily, weekly, monthly, ..., none
	Name         string        `json:"name,omitempty"`
	PhysicalSize int64         `json:"physicalsize"`
	Project      string        `json:"project"`
	ProjectID    string        `json:"projectid"`
	Revertable   bool          `json:"revertable,omitempty"`
	Size         int64         `json:"size,omitempty"`
	SnapshotType string        `json:"snapshottype,omitempty"`
	State        string        `json:"state"` // BackedUp, Creating, BackingUp, ...
	VolumeID     string        `json:"volumeid"`
	VolumeName   string        `json:"volumename,omitempty"`
	VolumeType   string        `json:"volumetype,omitempty"`
	ZoneID       string        `json:"zoneid"`
	Tags         []ResourceTag `json:"tags"`
	JobID        string        `json:"jobid,omitempty"`
	JobStatus    JobStatusType `json:"jobstatus,omitempty"`
}

// ResourceType returns the type of the resource
func (*Snapshot) ResourceType() string {
	return "Snapshot"
}

// CreateSnapshot represents a request to create a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/createSnapshot.html
type CreateSnapshot struct {
	VolumeID  string `json:"volumeid"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	PolicyID  string `json:"policyid,omitempty"`
	QuiesceVM bool   `json:"quiescevm,omitempty"`
}

func (*CreateSnapshot) name() string {
	return "createSnapshot"
}

func (*CreateSnapshot) asyncResponse() interface{} {
	return new(CreateSnapshotResponse)
}

// CreateSnapshotResponse represents a freshly created snapshot
type CreateSnapshotResponse struct {
	Snapshot Snapshot `json:"snapshot"`
}

// ListSnapshots lists the volume snapshots
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/listSnapshots.html
type ListSnapshots struct {
	Account      string        `json:"account,omitempty"`
	DomainID     string        `json:"domainid,omitempty"`
	ID           string        `json:"id,omitempty"`
	IntervalType string        `json:"intervaltype,omitempty"`
	IsRecursive  bool          `json:"isrecursive,omitempty"`
	Keyword      string        `json:"keyword,omitempty"`
	ListAll      bool          `json:"listall,omitempty"`
	Name         string        `json:"name,omitempty"`
	Page         int           `json:"page,omitempty"`
	PageSize     int           `json:"pagesize,omitempty"`
	ProjectID    string        `json:"projectid,omitempty"`
	SnapshotType string        `json:"snapshottype,omitempty"`
	Tags         []ResourceTag `json:"tags,omitempty"`
	VolumeID     string        `json:"volumeid,omitempty"`
	ZoneID       string        `json:"zoneid,omitempty"`
}

func (*ListSnapshots) name() string {
	return "listSnapshots"
}

func (*ListSnapshots) response() interface{} {
	return new(ListSnapshotsResponse)
}

// ListSnapshotsResponse represents a list of volume snapshots
type ListSnapshotsResponse struct {
	Count    int        `json:"count"`
	Snapshot []Snapshot `json:"snapshot"`
}

// DeleteSnapshot represents the deletion of a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSnapshot.html
type DeleteSnapshot struct {
	ID string `json:"id"`
}

func (*DeleteSnapshot) name() string {
	return "deleteSnapshot"
}

func (*DeleteSnapshot) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// RevertSnapshot revert a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/revertSnapshot.html
type RevertSnapshot struct {
	ID string `json:"id"`
}

func (*RevertSnapshot) name() string {
	return "revertSnapshot"
}

func (*RevertSnapshot) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}
