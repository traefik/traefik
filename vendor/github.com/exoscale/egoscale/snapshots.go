package egoscale

// SnapshotState represents the Snapshot.State enum
//
// See: https://github.com/apache/cloudstack/blob/master/api/src/main/java/com/cloud/storage/Snapshot.java
type SnapshotState string

const (
	// Allocated ... (TODO)
	Allocated SnapshotState = "Allocated"
	// Creating ... (TODO)
	Creating SnapshotState = "Creating"
	// CreatedOnPrimary ... (TODO)
	CreatedOnPrimary SnapshotState = "CreatedOnPrimary"
	// BackingUp ... (TODO)
	BackingUp SnapshotState = "BackingUp"
	// BackedUp ... (TODO)
	BackedUp SnapshotState = "BackedUp"
	// Copying ... (TODO)
	Copying SnapshotState = "Copying"
	// Destroying ... (TODO)
	Destroying SnapshotState = "Destroying"
	// Destroyed ... (TODO)
	Destroyed SnapshotState = "Destroyed"
	// Error is a state where the user can't see the snapshot while the snapshot may still exist on the storage
	Error SnapshotState = "Error"
)

// Snapshot represents a volume snapshot
type Snapshot struct {
	Account      string        `json:"account,omitempty" doc:"the account associated with the snapshot"`
	AccountID    *UUID         `json:"accountid,omitempty" doc:"the account ID associated with the snapshot"`
	Created      string        `json:"created,omitempty" doc:"the date the snapshot was created"`
	ID           *UUID         `json:"id,omitempty" doc:"ID of the snapshot"`
	IntervalType string        `json:"intervaltype,omitempty" doc:"valid types are hourly, daily, weekly, monthy, template, and none."`
	Name         string        `json:"name,omitempty" doc:"name of the snapshot"`
	PhysicalSize int64         `json:"physicalsize,omitempty" doc:"physical size of the snapshot on image store"`
	Revertable   *bool         `json:"revertable,omitempty" doc:"indicates whether the underlying storage supports reverting the volume to this snapshot"`
	Size         int64         `json:"size,omitempty" doc:"the size of original volume"`
	SnapshotType string        `json:"snapshottype,omitempty" doc:"the type of the snapshot"`
	State        string        `json:"state,omitempty" doc:"the state of the snapshot. BackedUp means that snapshot is ready to be used; Creating - the snapshot is being allocated on the primary storage; BackingUp - the snapshot is being backed up on secondary storage"`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with snapshot"`
	VolumeID     *UUID         `json:"volumeid,omitempty" doc:"ID of the disk volume"`
	VolumeName   string        `json:"volumename,omitempty" doc:"name of the disk volume"`
	VolumeType   string        `json:"volumetype,omitempty" doc:"type of the disk volume"`
	ZoneID       *UUID         `json:"zoneid,omitempty" doc:"id of the availability zone"`
}

// ResourceType returns the type of the resource
func (Snapshot) ResourceType() string {
	return "Snapshot"
}

// CreateSnapshot (Async) creates an instant snapshot of a volume
type CreateSnapshot struct {
	VolumeID  *UUID `json:"volumeid" doc:"The ID of the disk volume"`
	QuiesceVM *bool `json:"quiescevm,omitempty" doc:"quiesce vm if true"`
	_         bool  `name:"createSnapshot" description:"Creates an instant snapshot of a volume."`
}

// Response returns the struct to unmarshal
func (CreateSnapshot) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (CreateSnapshot) AsyncResponse() interface{} {
	return new(Snapshot)
}

// ListRequest builds the ListSnapshot request
func (ss Snapshot) ListRequest() (ListCommand, error) {
	// Restricted cannot be applied here because it really has three states
	req := &ListSnapshots{
		ID:           ss.ID,
		Name:         ss.Name,
		VolumeID:     ss.VolumeID,
		SnapshotType: ss.SnapshotType,
		ZoneID:       ss.ZoneID,
		// TODO: tags
	}

	return req, nil
}

//go:generate go run generate/main.go -interface=Listable ListSnapshots

// ListSnapshots lists the volume snapshots
type ListSnapshots struct {
	ID           *UUID         `json:"id,omitempty" doc:"lists snapshot by snapshot ID"`
	IntervalType string        `json:"intervaltype,omitempty" doc:"valid values are HOURLY, DAILY, WEEKLY, and MONTHLY."`
	Keyword      string        `json:"keyword,omitempty" doc:"List by keyword"`
	Name         string        `json:"name,omitempty" doc:"lists snapshot by snapshot name"`
	Page         int           `json:"page,omitempty"`
	PageSize     int           `json:"pagesize,omitempty"`
	SnapshotType string        `json:"snapshottype,omitempty" doc:"valid values are MANUAL or RECURRING."`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	VolumeID     *UUID         `json:"volumeid,omitempty" doc:"the ID of the disk volume"`
	ZoneID       *UUID         `json:"zoneid,omitempty" doc:"list snapshots by zone id"`
	_            bool          `name:"listSnapshots" description:"Lists all available snapshots for the account."`
}

// ListSnapshotsResponse represents a list of volume snapshots
type ListSnapshotsResponse struct {
	Count    int        `json:"count"`
	Snapshot []Snapshot `json:"snapshot"`
}

// DeleteSnapshot (Async) deletes a snapshot of a disk volume
type DeleteSnapshot struct {
	ID *UUID `json:"id" doc:"The ID of the snapshot"`
	_  bool  `name:"deleteSnapshot" description:"Deletes a snapshot of a disk volume."`
}

// Response returns the struct to unmarshal
func (DeleteSnapshot) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (DeleteSnapshot) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

// RevertSnapshot (Async) reverts a volume snapshot
type RevertSnapshot struct {
	ID *UUID `json:"id" doc:"The ID of the snapshot"`
	_  bool  `name:"revertSnapshot" description:"revert a volume snapshot."`
}

// Response returns the struct to unmarshal
func (RevertSnapshot) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (RevertSnapshot) AsyncResponse() interface{} {
	return new(BooleanResponse)
}
