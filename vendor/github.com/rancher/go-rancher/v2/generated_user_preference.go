package client

const (
	USER_PREFERENCE_TYPE = "userPreference"
)

type UserPreference struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	All bool `json:"all,omitempty" yaml:"all,omitempty"`

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

type UserPreferenceCollection struct {
	Collection
	Data   []UserPreference `json:"data,omitempty"`
	client *UserPreferenceClient
}

type UserPreferenceClient struct {
	rancherClient *RancherClient
}

type UserPreferenceOperations interface {
	List(opts *ListOpts) (*UserPreferenceCollection, error)
	Create(opts *UserPreference) (*UserPreference, error)
	Update(existing *UserPreference, updates interface{}) (*UserPreference, error)
	ById(id string) (*UserPreference, error)
	Delete(container *UserPreference) error

	ActionActivate(*UserPreference) (*UserPreference, error)

	ActionCreate(*UserPreference) (*UserPreference, error)

	ActionDeactivate(*UserPreference) (*UserPreference, error)

	ActionPurge(*UserPreference) (*UserPreference, error)

	ActionRemove(*UserPreference) (*UserPreference, error)

	ActionUpdate(*UserPreference) (*UserPreference, error)
}

func newUserPreferenceClient(rancherClient *RancherClient) *UserPreferenceClient {
	return &UserPreferenceClient{
		rancherClient: rancherClient,
	}
}

func (c *UserPreferenceClient) Create(container *UserPreference) (*UserPreference, error) {
	resp := &UserPreference{}
	err := c.rancherClient.doCreate(USER_PREFERENCE_TYPE, container, resp)
	return resp, err
}

func (c *UserPreferenceClient) Update(existing *UserPreference, updates interface{}) (*UserPreference, error) {
	resp := &UserPreference{}
	err := c.rancherClient.doUpdate(USER_PREFERENCE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *UserPreferenceClient) List(opts *ListOpts) (*UserPreferenceCollection, error) {
	resp := &UserPreferenceCollection{}
	err := c.rancherClient.doList(USER_PREFERENCE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *UserPreferenceCollection) Next() (*UserPreferenceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &UserPreferenceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *UserPreferenceClient) ById(id string) (*UserPreference, error) {
	resp := &UserPreference{}
	err := c.rancherClient.doById(USER_PREFERENCE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *UserPreferenceClient) Delete(container *UserPreference) error {
	return c.rancherClient.doResourceDelete(USER_PREFERENCE_TYPE, &container.Resource)
}

func (c *UserPreferenceClient) ActionActivate(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *UserPreferenceClient) ActionCreate(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *UserPreferenceClient) ActionDeactivate(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "deactivate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *UserPreferenceClient) ActionPurge(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "purge", &resource.Resource, nil, resp)

	return resp, err
}

func (c *UserPreferenceClient) ActionRemove(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *UserPreferenceClient) ActionUpdate(resource *UserPreference) (*UserPreference, error) {

	resp := &UserPreference{}

	err := c.rancherClient.doAction(USER_PREFERENCE_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}
