package client

const (
	PROJECT_TEMPLATE_TYPE = "projectTemplate"
)

type ProjectTemplate struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	IsPublic bool `json:"isPublic,omitempty" yaml:"is_public,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	Stacks []CatalogTemplate `json:"stacks,omitempty" yaml:"stacks,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ProjectTemplateCollection struct {
	Collection
	Data   []ProjectTemplate `json:"data,omitempty"`
	client *ProjectTemplateClient
}

type ProjectTemplateClient struct {
	rancherClient *RancherClient
}

type ProjectTemplateOperations interface {
	List(opts *ListOpts) (*ProjectTemplateCollection, error)
	Create(opts *ProjectTemplate) (*ProjectTemplate, error)
	Update(existing *ProjectTemplate, updates interface{}) (*ProjectTemplate, error)
	ById(id string) (*ProjectTemplate, error)
	Delete(container *ProjectTemplate) error

	ActionCreate(*ProjectTemplate) (*ProjectTemplate, error)

	ActionRemove(*ProjectTemplate) (*ProjectTemplate, error)
}

func newProjectTemplateClient(rancherClient *RancherClient) *ProjectTemplateClient {
	return &ProjectTemplateClient{
		rancherClient: rancherClient,
	}
}

func (c *ProjectTemplateClient) Create(container *ProjectTemplate) (*ProjectTemplate, error) {
	resp := &ProjectTemplate{}
	err := c.rancherClient.doCreate(PROJECT_TEMPLATE_TYPE, container, resp)
	return resp, err
}

func (c *ProjectTemplateClient) Update(existing *ProjectTemplate, updates interface{}) (*ProjectTemplate, error) {
	resp := &ProjectTemplate{}
	err := c.rancherClient.doUpdate(PROJECT_TEMPLATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProjectTemplateClient) List(opts *ListOpts) (*ProjectTemplateCollection, error) {
	resp := &ProjectTemplateCollection{}
	err := c.rancherClient.doList(PROJECT_TEMPLATE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProjectTemplateCollection) Next() (*ProjectTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProjectTemplateCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProjectTemplateClient) ById(id string) (*ProjectTemplate, error) {
	resp := &ProjectTemplate{}
	err := c.rancherClient.doById(PROJECT_TEMPLATE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ProjectTemplateClient) Delete(container *ProjectTemplate) error {
	return c.rancherClient.doResourceDelete(PROJECT_TEMPLATE_TYPE, &container.Resource)
}

func (c *ProjectTemplateClient) ActionCreate(resource *ProjectTemplate) (*ProjectTemplate, error) {

	resp := &ProjectTemplate{}

	err := c.rancherClient.doAction(PROJECT_TEMPLATE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ProjectTemplateClient) ActionRemove(resource *ProjectTemplate) (*ProjectTemplate, error) {

	resp := &ProjectTemplate{}

	err := c.rancherClient.doAction(PROJECT_TEMPLATE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}
