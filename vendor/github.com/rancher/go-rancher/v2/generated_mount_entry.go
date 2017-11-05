package client

const (
	MOUNT_ENTRY_TYPE = "mountEntry"
)

type MountEntry struct {
	Resource

	InstanceId string `json:"instanceId,omitempty" yaml:"instance_id,omitempty"`

	InstanceName string `json:"instanceName,omitempty" yaml:"instance_name,omitempty"`

	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	VolumeId string `json:"volumeId,omitempty" yaml:"volume_id,omitempty"`

	VolumeName string `json:"volumeName,omitempty" yaml:"volume_name,omitempty"`
}

type MountEntryCollection struct {
	Collection
	Data   []MountEntry `json:"data,omitempty"`
	client *MountEntryClient
}

type MountEntryClient struct {
	rancherClient *RancherClient
}

type MountEntryOperations interface {
	List(opts *ListOpts) (*MountEntryCollection, error)
	Create(opts *MountEntry) (*MountEntry, error)
	Update(existing *MountEntry, updates interface{}) (*MountEntry, error)
	ById(id string) (*MountEntry, error)
	Delete(container *MountEntry) error
}

func newMountEntryClient(rancherClient *RancherClient) *MountEntryClient {
	return &MountEntryClient{
		rancherClient: rancherClient,
	}
}

func (c *MountEntryClient) Create(container *MountEntry) (*MountEntry, error) {
	resp := &MountEntry{}
	err := c.rancherClient.doCreate(MOUNT_ENTRY_TYPE, container, resp)
	return resp, err
}

func (c *MountEntryClient) Update(existing *MountEntry, updates interface{}) (*MountEntry, error) {
	resp := &MountEntry{}
	err := c.rancherClient.doUpdate(MOUNT_ENTRY_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *MountEntryClient) List(opts *ListOpts) (*MountEntryCollection, error) {
	resp := &MountEntryCollection{}
	err := c.rancherClient.doList(MOUNT_ENTRY_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *MountEntryCollection) Next() (*MountEntryCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &MountEntryCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *MountEntryClient) ById(id string) (*MountEntry, error) {
	resp := &MountEntry{}
	err := c.rancherClient.doById(MOUNT_ENTRY_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *MountEntryClient) Delete(container *MountEntry) error {
	return c.rancherClient.doResourceDelete(MOUNT_ENTRY_TYPE, &container.Resource)
}
