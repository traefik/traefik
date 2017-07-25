package client

const (
	GENERIC_OBJECT_TYPE = "genericObject"
)

type GenericObject struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Key string `json:"key,omitempty" yaml:"key,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	ResourceData map[string]interface{} `json:"resourceData,omitempty" yaml:"resource_data,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type GenericObjectCollection struct {
	Collection
	Data   []GenericObject `json:"data,omitempty"`
	client *GenericObjectClient
}

type GenericObjectClient struct {
	rancherClient *RancherClient
}

type GenericObjectOperations interface {
	List(opts *ListOpts) (*GenericObjectCollection, error)
	Create(opts *GenericObject) (*GenericObject, error)
	Update(existing *GenericObject, updates interface{}) (*GenericObject, error)
	ById(id string) (*GenericObject, error)
	Delete(container *GenericObject) error

	ActionCreate(*GenericObject) (*GenericObject, error)

	ActionRemove(*GenericObject) (*GenericObject, error)
}

func newGenericObjectClient(rancherClient *RancherClient) *GenericObjectClient {
	return &GenericObjectClient{
		rancherClient: rancherClient,
	}
}

func (c *GenericObjectClient) Create(container *GenericObject) (*GenericObject, error) {
	resp := &GenericObject{}
	err := c.rancherClient.doCreate(GENERIC_OBJECT_TYPE, container, resp)
	return resp, err
}

func (c *GenericObjectClient) Update(existing *GenericObject, updates interface{}) (*GenericObject, error) {
	resp := &GenericObject{}
	err := c.rancherClient.doUpdate(GENERIC_OBJECT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *GenericObjectClient) List(opts *ListOpts) (*GenericObjectCollection, error) {
	resp := &GenericObjectCollection{}
	err := c.rancherClient.doList(GENERIC_OBJECT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *GenericObjectCollection) Next() (*GenericObjectCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &GenericObjectCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *GenericObjectClient) ById(id string) (*GenericObject, error) {
	resp := &GenericObject{}
	err := c.rancherClient.doById(GENERIC_OBJECT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *GenericObjectClient) Delete(container *GenericObject) error {
	return c.rancherClient.doResourceDelete(GENERIC_OBJECT_TYPE, &container.Resource)
}

func (c *GenericObjectClient) ActionCreate(resource *GenericObject) (*GenericObject, error) {

	resp := &GenericObject{}

	err := c.rancherClient.doAction(GENERIC_OBJECT_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *GenericObjectClient) ActionRemove(resource *GenericObject) (*GenericObject, error) {

	resp := &GenericObject{}

	err := c.rancherClient.doAction(GENERIC_OBJECT_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}
