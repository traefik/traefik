package client

const (
	STORAGE_DRIVER_TYPE = "storageDriver"
)

type StorageDriver struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	BlockDevicePath string `json:"blockDevicePath,omitempty" yaml:"block_device_path,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	Scope string `json:"scope,omitempty" yaml:"scope,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	VolumeAccessMode string `json:"volumeAccessMode,omitempty" yaml:"volume_access_mode,omitempty"`

	VolumeCapabilities []string `json:"volumeCapabilities,omitempty" yaml:"volume_capabilities,omitempty"`
}

type StorageDriverCollection struct {
	Collection
	Data   []StorageDriver `json:"data,omitempty"`
	client *StorageDriverClient
}

type StorageDriverClient struct {
	rancherClient *RancherClient
}

type StorageDriverOperations interface {
	List(opts *ListOpts) (*StorageDriverCollection, error)
	Create(opts *StorageDriver) (*StorageDriver, error)
	Update(existing *StorageDriver, updates interface{}) (*StorageDriver, error)
	ById(id string) (*StorageDriver, error)
	Delete(container *StorageDriver) error

	ActionActivate(*StorageDriver) (*StorageDriver, error)

	ActionCreate(*StorageDriver) (*StorageDriver, error)

	ActionDeactivate(*StorageDriver) (*StorageDriver, error)

	ActionRemove(*StorageDriver) (*StorageDriver, error)

	ActionUpdate(*StorageDriver) (*StorageDriver, error)
}

func newStorageDriverClient(rancherClient *RancherClient) *StorageDriverClient {
	return &StorageDriverClient{
		rancherClient: rancherClient,
	}
}

func (c *StorageDriverClient) Create(container *StorageDriver) (*StorageDriver, error) {
	resp := &StorageDriver{}
	err := c.rancherClient.doCreate(STORAGE_DRIVER_TYPE, container, resp)
	return resp, err
}

func (c *StorageDriverClient) Update(existing *StorageDriver, updates interface{}) (*StorageDriver, error) {
	resp := &StorageDriver{}
	err := c.rancherClient.doUpdate(STORAGE_DRIVER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StorageDriverClient) List(opts *ListOpts) (*StorageDriverCollection, error) {
	resp := &StorageDriverCollection{}
	err := c.rancherClient.doList(STORAGE_DRIVER_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *StorageDriverCollection) Next() (*StorageDriverCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &StorageDriverCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *StorageDriverClient) ById(id string) (*StorageDriver, error) {
	resp := &StorageDriver{}
	err := c.rancherClient.doById(STORAGE_DRIVER_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *StorageDriverClient) Delete(container *StorageDriver) error {
	return c.rancherClient.doResourceDelete(STORAGE_DRIVER_TYPE, &container.Resource)
}

func (c *StorageDriverClient) ActionActivate(resource *StorageDriver) (*StorageDriver, error) {

	resp := &StorageDriver{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverClient) ActionCreate(resource *StorageDriver) (*StorageDriver, error) {

	resp := &StorageDriver{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverClient) ActionDeactivate(resource *StorageDriver) (*StorageDriver, error) {

	resp := &StorageDriver{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverClient) ActionRemove(resource *StorageDriver) (*StorageDriver, error) {

	resp := &StorageDriver{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *StorageDriverClient) ActionUpdate(resource *StorageDriver) (*StorageDriver, error) {

	resp := &StorageDriver{}

	err := c.rancherClient.doAction(STORAGE_DRIVER_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}
