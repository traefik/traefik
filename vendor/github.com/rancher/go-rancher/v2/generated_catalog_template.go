package client

const (
	CATALOG_TEMPLATE_TYPE = "catalogTemplate"
)

type CatalogTemplate struct {
	Resource

	Answers map[string]interface{} `json:"answers,omitempty" yaml:"answers,omitempty"`

	Binding Binding `json:"binding,omitempty" yaml:"binding,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	DockerCompose string `json:"dockerCompose,omitempty" yaml:"docker_compose,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RancherCompose string `json:"rancherCompose,omitempty" yaml:"rancher_compose,omitempty"`

	TemplateId string `json:"templateId,omitempty" yaml:"template_id,omitempty"`

	TemplateVersionId string `json:"templateVersionId,omitempty" yaml:"template_version_id,omitempty"`
}

type CatalogTemplateCollection struct {
	Collection
	Data   []CatalogTemplate `json:"data,omitempty"`
	client *CatalogTemplateClient
}

type CatalogTemplateClient struct {
	rancherClient *RancherClient
}

type CatalogTemplateOperations interface {
	List(opts *ListOpts) (*CatalogTemplateCollection, error)
	Create(opts *CatalogTemplate) (*CatalogTemplate, error)
	Update(existing *CatalogTemplate, updates interface{}) (*CatalogTemplate, error)
	ById(id string) (*CatalogTemplate, error)
	Delete(container *CatalogTemplate) error
}

func newCatalogTemplateClient(rancherClient *RancherClient) *CatalogTemplateClient {
	return &CatalogTemplateClient{
		rancherClient: rancherClient,
	}
}

func (c *CatalogTemplateClient) Create(container *CatalogTemplate) (*CatalogTemplate, error) {
	resp := &CatalogTemplate{}
	err := c.rancherClient.doCreate(CATALOG_TEMPLATE_TYPE, container, resp)
	return resp, err
}

func (c *CatalogTemplateClient) Update(existing *CatalogTemplate, updates interface{}) (*CatalogTemplate, error) {
	resp := &CatalogTemplate{}
	err := c.rancherClient.doUpdate(CATALOG_TEMPLATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CatalogTemplateClient) List(opts *ListOpts) (*CatalogTemplateCollection, error) {
	resp := &CatalogTemplateCollection{}
	err := c.rancherClient.doList(CATALOG_TEMPLATE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *CatalogTemplateCollection) Next() (*CatalogTemplateCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &CatalogTemplateCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *CatalogTemplateClient) ById(id string) (*CatalogTemplate, error) {
	resp := &CatalogTemplate{}
	err := c.rancherClient.doById(CATALOG_TEMPLATE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *CatalogTemplateClient) Delete(container *CatalogTemplate) error {
	return c.rancherClient.doResourceDelete(CATALOG_TEMPLATE_TYPE, &container.Resource)
}
