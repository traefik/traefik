package client

const (
	HOST_TEMPLATE_TYPE = "hostTemplate"
)

type HostTemplate struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Driver string `json:"driver,omitempty" yaml:"driver,omitempty"`

	FlavorPrefix string `json:"flavorPrefix,omitempty" yaml:"flavor_prefix,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PublicValues map[string]interface{} `json:"publicValues,omitempty" yaml:"public_values,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	SecretValues map[string]interface{} `json:"secretValues,omitempty" yaml:"secret_values,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type HostTemplateCollection struct {
	Collection
	Data   []HostTemplate `json:"data,omitempty"`
	client *HostTemplateClient
}

type HostTemplateClient struct {
	rancherClient *RancherClient
}

type HostTemplateOperations interface {
	List(opts *ListOpts) (*HostTemplateCollection, error)
	Create(opts *HostTemplate) (*HostTemplate, error)
	Update(existing *HostTemplate, updates interface{}) (*HostTemplate, error)
	ById(id string) (*HostTemplate, error)
	Delete(container *HostTemplate) error

	ActionCreate(*HostTemplate) (*HostTemplate, error)

	ActionRemove(*HostTemplate) (*HostTemplate, error)
}

func newHostTemplateClient(rancherClient *RancherClient) *HostTemplateClient {
	return &HostTemplateClient{
		rancherClient: rancherClient,
	}
}

func (c *HostTemplateClient) Create(container *HostTemplate) (*HostTemplate, error) {
	resp := &HostTemplate{}
	err := c.rancherClient.doCreate(HOST_TEMPLATE_TYPE, container, resp)
	return resp, err
}

func (c *HostTemplateClient) Update(existing *HostTemplate, updates interface{}) (*HostTemplate, error) {
	resp := &HostTemplate{}
	err := c.rancherClient.doUpdate(HOST_TEMPLATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HostTemplateClient) List(opts *ListOpts) (*HostTemplateCollection, error) {
	resp := &HostTemplateCollection{}
	err := c.rancherClient.doList(HOST_TEMPLATE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HostTemplateCollection) Next() (*HostTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HostTemplateCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HostTemplateClient) ById(id string) (*HostTemplate, error) {
	resp := &HostTemplate{}
	err := c.rancherClient.doById(HOST_TEMPLATE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *HostTemplateClient) Delete(container *HostTemplate) error {
	return c.rancherClient.doResourceDelete(HOST_TEMPLATE_TYPE, &container.Resource)
}

func (c *HostTemplateClient) ActionCreate(resource *HostTemplate) (*HostTemplate, error) {

	resp := &HostTemplate{}

	err := c.rancherClient.doAction(HOST_TEMPLATE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *HostTemplateClient) ActionRemove(resource *HostTemplate) (*HostTemplate, error) {

	resp := &HostTemplate{}

	err := c.rancherClient.doAction(HOST_TEMPLATE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}
