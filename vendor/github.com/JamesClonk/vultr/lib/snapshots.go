package lib

import "net/url"

// Snapshot of a virtual machine on Vultr account
type Snapshot struct {
	ID          string `json:"SNAPSHOTID"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Status      string `json:"status"`
	Created     string `json:"date_created"`
}

// GetSnapshots retrieves a list of all snapshots on Vultr account
func (c *Client) GetSnapshots() (snapshots []Snapshot, err error) {
	var snapshotMap map[string]Snapshot
	if err := c.get(`snapshot/list`, &snapshotMap); err != nil {
		return nil, err
	}

	for _, snapshot := range snapshotMap {
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

// CreateSnapshot creates a new virtual machine snapshot
func (c *Client) CreateSnapshot(id, description string) (Snapshot, error) {
	values := url.Values{
		"SUBID":       {id},
		"description": {description},
	}

	var snapshot Snapshot
	if err := c.post(`snapshot/create`, values, &snapshot); err != nil {
		return Snapshot{}, err
	}
	snapshot.Description = description

	return snapshot, nil
}

// DeleteSnapshot deletes an existing virtual machine snapshot
func (c *Client) DeleteSnapshot(id string) error {
	values := url.Values{
		"SNAPSHOTID": {id},
	}

	if err := c.post(`snapshot/destroy`, values, nil); err != nil {
		return err
	}
	return nil
}
