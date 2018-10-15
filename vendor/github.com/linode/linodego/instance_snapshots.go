package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// InstanceBackupsResponse response struct for backup snapshot
type InstanceBackupsResponse struct {
	Automatic []*InstanceSnapshot             `json:"automatic"`
	Snapshot  *InstanceBackupSnapshotResponse `json:"snapshot"`
}

// InstanceBackupSnapshotResponse fields are those representing Instance Backup Snapshots
type InstanceBackupSnapshotResponse struct {
	Current    *InstanceSnapshot `json:"current"`
	InProgress *InstanceSnapshot `json:"in_progress"`
}

// RestoreInstanceOptions fields are those accepted by InstanceRestore
type RestoreInstanceOptions struct {
	LinodeID  int  `json:"linode_id"`
	Overwrite bool `json:"overwrite"`
}

// InstanceSnapshot represents a linode backup snapshot
type InstanceSnapshot struct {
	CreatedStr  string `json:"created"`
	UpdatedStr  string `json:"updated"`
	FinishedStr string `json:"finished"`

	ID       int                     `json:"id"`
	Label    string                  `json:"label"`
	Status   InstanceSnapshotStatus  `json:"status"`
	Type     string                  `json:"type"`
	Created  *time.Time              `json:"-"`
	Updated  *time.Time              `json:"-"`
	Finished *time.Time              `json:"-"`
	Configs  []string                `json:"configs"`
	Disks    []*InstanceSnapshotDisk `json:"disks"`
}

// InstanceSnapshotDisk fields represent the source disk of a Snapshot
type InstanceSnapshotDisk struct {
	Label      string `json:"label"`
	Size       int    `json:"size"`
	Filesystem string `json:"filesystem"`
}

// InstanceSnapshotStatus constants start with Snapshot and include Linode API Instance Backup Snapshot status values
type InstanceSnapshotStatus string

// InstanceSnapshotStatus constants reflect the current status of an Instance Snapshot
var (
	SnapshotPaused              InstanceSnapshotStatus = "paused"
	SnapshotPending             InstanceSnapshotStatus = "pending"
	SnapshotRunning             InstanceSnapshotStatus = "running"
	SnapshotNeedsPostProcessing InstanceSnapshotStatus = "needsPostProcessing"
	SnapshotSuccessful          InstanceSnapshotStatus = "successful"
	SnapshotFailed              InstanceSnapshotStatus = "failed"
	SnapshotUserAborted         InstanceSnapshotStatus = "userAborted"
)

func (l *InstanceSnapshot) fixDates() *InstanceSnapshot {
	l.Created, _ = parseDates(l.CreatedStr)
	l.Updated, _ = parseDates(l.UpdatedStr)
	l.Finished, _ = parseDates(l.FinishedStr)
	return l
}

// GetInstanceSnapshot gets the snapshot with the provided ID
func (c *Client) GetInstanceSnapshot(ctx context.Context, linodeID int, snapshotID int) (*InstanceSnapshot, error) {
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, snapshotID)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceSnapshot{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceSnapshot).fixDates(), nil
}

// CreateInstanceSnapshot Creates or Replaces the snapshot Backup of a Linode. If a previous snapshot exists for this Linode, it will be deleted.
func (c *Client) CreateInstanceSnapshot(ctx context.Context, linodeID int, label string) (*InstanceSnapshot, error) {
	o, err := json.Marshal(map[string]string{"label": label})
	if err != nil {
		return nil, err
	}
	body := string(o)
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}

	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(body).
		SetResult(&InstanceSnapshot{}).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*InstanceSnapshot).fixDates(), nil
}

// GetInstanceBackups gets the Instance's available Backups.
// This is not called ListInstanceBackups because a single object is returned, matching the API response.
func (c *Client) GetInstanceBackups(ctx context.Context, linodeID int) (*InstanceBackupsResponse, error) {
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(&InstanceBackupsResponse{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceBackupsResponse).fixDates(), nil
}

// EnableInstanceBackups Enables backups for the specified Linode.
func (c *Client) EnableInstanceBackups(ctx context.Context, linodeID int) error {
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/enable", e)

	_, err = coupleAPIErrors(c.R(ctx).Post(e))
	return err
}

// CancelInstanceBackups Cancels backups for the specified Linode.
func (c *Client) CancelInstanceBackups(ctx context.Context, linodeID int) error {
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/cancel", e)

	_, err = coupleAPIErrors(c.R(ctx).Post(e))
	return err
}

// RestoreInstanceBackup Restores a Linode's Backup to the specified Linode.
func (c *Client) RestoreInstanceBackup(ctx context.Context, linodeID int, backupID int, opts RestoreInstanceOptions) error {
	o, err := json.Marshal(opts)
	if err != nil {
		return NewError(err)
	}
	body := string(o)
	e, err := c.InstanceSnapshots.endpointWithID(linodeID)
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d/restore", e, backupID)

	_, err = coupleAPIErrors(c.R(ctx).SetBody(body).Post(e))

	return err

}

func (l *InstanceBackupSnapshotResponse) fixDates() *InstanceBackupSnapshotResponse {
	if l.Current != nil {
		l.Current.fixDates()
	}
	if l.InProgress != nil {
		l.InProgress.fixDates()
	}
	return l
}

func (l *InstanceBackupsResponse) fixDates() *InstanceBackupsResponse {
	for i := range l.Automatic {
		l.Automatic[i].fixDates()
	}
	if l.Snapshot != nil {
		l.Snapshot.fixDates()
	}
	return l
}
