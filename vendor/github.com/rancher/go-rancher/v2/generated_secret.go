package client

const (
	SECRET_TYPE = "secret"
)

type Secret struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

type SecretCollection struct {
	Collection
	Data   []Secret `json:"data,omitempty"`
	client *SecretClient
}

type SecretClient struct {
	rancherClient *RancherClient
}

type SecretOperations interface {
	List(opts *ListOpts) (*SecretCollection, error)
	Create(opts *Secret) (*Secret, error)
	Update(existing *Secret, updates interface{}) (*Secret, error)
	ById(id string) (*Secret, error)
	Delete(container *Secret) error

	ActionCreate(*Secret) (*Secret, error)

	ActionRemove(*Secret) (*Secret, error)
}

func newSecretClient(rancherClient *RancherClient) *SecretClient {
	return &SecretClient{
		rancherClient: rancherClient,
	}
}

func (c *SecretClient) Create(container *Secret) (*Secret, error) {
	resp := &Secret{}
	err := c.rancherClient.doCreate(SECRET_TYPE, container, resp)
	return resp, err
}

func (c *SecretClient) Update(existing *Secret, updates interface{}) (*Secret, error) {
	resp := &Secret{}
	err := c.rancherClient.doUpdate(SECRET_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SecretClient) List(opts *ListOpts) (*SecretCollection, error) {
	resp := &SecretCollection{}
	err := c.rancherClient.doList(SECRET_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SecretCollection) Next() (*SecretCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SecretCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SecretClient) ById(id string) (*Secret, error) {
	resp := &Secret{}
	err := c.rancherClient.doById(SECRET_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *SecretClient) Delete(container *Secret) error {
	return c.rancherClient.doResourceDelete(SECRET_TYPE, &container.Resource)
}

func (c *SecretClient) ActionCreate(resource *Secret) (*Secret, error) {

	resp := &Secret{}

	err := c.rancherClient.doAction(SECRET_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *SecretClient) ActionRemove(resource *Secret) (*Secret, error) {

	resp := &Secret{}

	err := c.rancherClient.doAction(SECRET_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}
