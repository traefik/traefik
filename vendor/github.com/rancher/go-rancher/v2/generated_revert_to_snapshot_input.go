package client

const (
	REVERT_TO_SNAPSHOT_INPUT_TYPE = "revertToSnapshotInput"
)

type RevertToSnapshotInput struct {
	Resource

	SnapshotId string `json:"snapshotId,omitempty" yaml:"snapshot_id,omitempty"`
}

type RevertToSnapshotInputCollection struct {
	Collection
	Data   []RevertToSnapshotInput `json:"data,omitempty"`
	client *RevertToSnapshotInputClient
}

type RevertToSnapshotInputClient struct {
	rancherClient *RancherClient
}

type RevertToSnapshotInputOperations interface {
	List(opts *ListOpts) (*RevertToSnapshotInputCollection, error)
	Create(opts *RevertToSnapshotInput) (*RevertToSnapshotInput, error)
	Update(existing *RevertToSnapshotInput, updates interface{}) (*RevertToSnapshotInput, error)
	ById(id string) (*RevertToSnapshotInput, error)
	Delete(container *RevertToSnapshotInput) error
}

func newRevertToSnapshotInputClient(rancherClient *RancherClient) *RevertToSnapshotInputClient {
	return &RevertToSnapshotInputClient{
		rancherClient: rancherClient,
	}
}

func (c *RevertToSnapshotInputClient) Create(container *RevertToSnapshotInput) (*RevertToSnapshotInput, error) {
	resp := &RevertToSnapshotInput{}
	err := c.rancherClient.doCreate(REVERT_TO_SNAPSHOT_INPUT_TYPE, container, resp)
	return resp, err
}

func (c *RevertToSnapshotInputClient) Update(existing *RevertToSnapshotInput, updates interface{}) (*RevertToSnapshotInput, error) {
	resp := &RevertToSnapshotInput{}
	err := c.rancherClient.doUpdate(REVERT_TO_SNAPSHOT_INPUT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RevertToSnapshotInputClient) List(opts *ListOpts) (*RevertToSnapshotInputCollection, error) {
	resp := &RevertToSnapshotInputCollection{}
	err := c.rancherClient.doList(REVERT_TO_SNAPSHOT_INPUT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *RevertToSnapshotInputCollection) Next() (*RevertToSnapshotInputCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &RevertToSnapshotInputCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *RevertToSnapshotInputClient) ById(id string) (*RevertToSnapshotInput, error) {
	resp := &RevertToSnapshotInput{}
	err := c.rancherClient.doById(REVERT_TO_SNAPSHOT_INPUT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *RevertToSnapshotInputClient) Delete(container *RevertToSnapshotInput) error {
	return c.rancherClient.doResourceDelete(REVERT_TO_SNAPSHOT_INPUT_TYPE, &container.Resource)
}
