package client

const (
	PROCESS_SUMMARY_TYPE = "processSummary"
)

type ProcessSummary struct {
	Resource

	Delay int64 `json:"delay,omitempty" yaml:"delay,omitempty"`

	ProcessName string `json:"processName,omitempty" yaml:"process_name,omitempty"`

	Ready int64 `json:"ready,omitempty" yaml:"ready,omitempty"`

	Running int64 `json:"running,omitempty" yaml:"running,omitempty"`
}

type ProcessSummaryCollection struct {
	Collection
	Data   []ProcessSummary `json:"data,omitempty"`
	client *ProcessSummaryClient
}

type ProcessSummaryClient struct {
	rancherClient *RancherClient
}

type ProcessSummaryOperations interface {
	List(opts *ListOpts) (*ProcessSummaryCollection, error)
	Create(opts *ProcessSummary) (*ProcessSummary, error)
	Update(existing *ProcessSummary, updates interface{}) (*ProcessSummary, error)
	ById(id string) (*ProcessSummary, error)
	Delete(container *ProcessSummary) error
}

func newProcessSummaryClient(rancherClient *RancherClient) *ProcessSummaryClient {
	return &ProcessSummaryClient{
		rancherClient: rancherClient,
	}
}

func (c *ProcessSummaryClient) Create(container *ProcessSummary) (*ProcessSummary, error) {
	resp := &ProcessSummary{}
	err := c.rancherClient.doCreate(PROCESS_SUMMARY_TYPE, container, resp)
	return resp, err
}

func (c *ProcessSummaryClient) Update(existing *ProcessSummary, updates interface{}) (*ProcessSummary, error) {
	resp := &ProcessSummary{}
	err := c.rancherClient.doUpdate(PROCESS_SUMMARY_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProcessSummaryClient) List(opts *ListOpts) (*ProcessSummaryCollection, error) {
	resp := &ProcessSummaryCollection{}
	err := c.rancherClient.doList(PROCESS_SUMMARY_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProcessSummaryCollection) Next() (*ProcessSummaryCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProcessSummaryCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProcessSummaryClient) ById(id string) (*ProcessSummary, error) {
	resp := &ProcessSummary{}
	err := c.rancherClient.doById(PROCESS_SUMMARY_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ProcessSummaryClient) Delete(container *ProcessSummary) error {
	return c.rancherClient.doResourceDelete(PROCESS_SUMMARY_TYPE, &container.Resource)
}
