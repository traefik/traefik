package client

const (
	REGISTER_TYPE = "register"
)

type Register struct {
	Resource

	AccessKey string `json:"accessKey,omitempty" yaml:"access_key,omitempty"`

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

	SecretKey string `json:"secretKey,omitempty" yaml:"secret_key,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type RegisterCollection struct {
	Collection
	Data   []Register `json:"data,omitempty"`
	client *RegisterClient
}

type RegisterClient struct {
	rancherClient *RancherClient
}

type RegisterOperations interface {
	List(opts *ListOpts) (*RegisterCollection, error)
	Create(opts *Register) (*Register, error)
	Update(existing *Register, updates interface{}) (*Register, error)
	ById(id string) (*Register, error)
	Delete(container *Register) error

	ActionCreate(*Register) (*GenericObject, error)

	ActionRemove(*Register) (*GenericObject, error)

	ActionStop(*Register, *InstanceStop) (*Instance, error)
}

func newRegisterClient(rancherClient *RancherClient) *RegisterClient {
	return &RegisterClient{
		rancherClient: rancherClient,
	}
}

func (c *RegisterClient) Create(container *Register) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doCreate(REGISTER_TYPE, container, resp)
	return resp, err
}

func (c *RegisterClient) Update(existing *Register, updates interface{}) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doUpdate(REGISTER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RegisterClient) List(opts *ListOpts) (*RegisterCollection, error) {
	resp := &RegisterCollection{}
	err := c.rancherClient.doList(REGISTER_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *RegisterCollection) Next() (*RegisterCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &RegisterCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *RegisterClient) ById(id string) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doById(REGISTER_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *RegisterClient) Delete(container *Register) error {
	return c.rancherClient.doResourceDelete(REGISTER_TYPE, &container.Resource)
}

func (c *RegisterClient) ActionCreate(resource *Register) (*GenericObject, error) {

	resp := &GenericObject{}

	err := c.rancherClient.doAction(REGISTER_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *RegisterClient) ActionRemove(resource *Register) (*GenericObject, error) {

	resp := &GenericObject{}

	err := c.rancherClient.doAction(REGISTER_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *RegisterClient) ActionStop(resource *Register, input *InstanceStop) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(REGISTER_TYPE, "stop", &resource.Resource, input, resp)

	return resp, err
}
