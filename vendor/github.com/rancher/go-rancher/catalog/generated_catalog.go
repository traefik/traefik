package catalog

const (
	CATALOG_TYPE = "catalog"
)

type Catalog struct {
	Resource

	Actions map[string]interface{} `json:"actions,omitempty" yaml:"actions,omitempty"`

	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`

	Commit string `json:"commit,omitempty" yaml:"commit,omitempty"`

	EnvironmentId string `json:"environmentId,omitempty" yaml:"environment_id,omitempty"`

	Links map[string]interface{} `json:"links,omitempty" yaml:"links,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	Url string `json:"url,omitempty" yaml:"url,omitempty"`
}

type CatalogCollection struct {
	Collection
	Data   []Catalog `json:"data,omitempty"`
	client *CatalogClient
}

type CatalogClient struct {
	rancherClient *RancherClient
}

type CatalogOperations interface {
	List(opts *ListOpts) (*CatalogCollection, error)
	Create(opts *Catalog) (*Catalog, error)
	Update(existing *Catalog, updates interface{}) (*Catalog, error)
	ById(id string) (*Catalog, error)
	Delete(container *Catalog) error
}

func newCatalogClient(rancherClient *RancherClient) *CatalogClient {
	return &CatalogClient{
		rancherClient: rancherClient,
	}
}

func (c *CatalogClient) Create(container *Catalog) (*Catalog, error) {
	resp := &Catalog{}
	err := c.rancherClient.doCreate(CATALOG_TYPE, container, resp)
	return resp, err
}

func (c *CatalogClient) Update(existing *Catalog, updates interface{}) (*Catalog, error) {
	resp := &Catalog{}
	err := c.rancherClient.doUpdate(CATALOG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CatalogClient) List(opts *ListOpts) (*CatalogCollection, error) {
	resp := &CatalogCollection{}
	err := c.rancherClient.doList(CATALOG_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *CatalogCollection) Next() (*CatalogCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &CatalogCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *CatalogClient) ById(id string) (*Catalog, error) {
	resp := &Catalog{}
	err := c.rancherClient.doById(CATALOG_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *CatalogClient) Delete(container *Catalog) error {
	return c.rancherClient.doResourceDelete(CATALOG_TYPE, &container.Resource)
}
