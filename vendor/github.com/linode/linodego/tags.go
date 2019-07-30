package linodego

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Tag represents a Tag object
type Tag struct {
	Label string `json:"label"`
}

// TaggedObject represents a Tagged Object object
type TaggedObject struct {
	Type    string          `json:"type"`
	RawData json.RawMessage `json:"data"`
	Data    interface{}     `json:"-"`
}

// SortedObjects currently only includes Instances
type SortedObjects struct {
	Instances     []Instance
	Domains       []Domain
	Volumes       []Volume
	NodeBalancers []NodeBalancer
	/*
		StackScripts  []Stackscript
	*/
}

// TaggedObjectList are a list of TaggedObjects, as returning by ListTaggedObjects
type TaggedObjectList []TaggedObject

// TagCreateOptions fields are those accepted by CreateTag
type TagCreateOptions struct {
	Label         string `json:"label"`
	Linodes       []int  `json:"linodes,omitempty"`
	Domains       []int  `json:"domains,omitempty"`
	Volumes       []int  `json:"volumes,omitempty"`
	NodeBalancers []int  `json:"nodebalancers,omitempty"`
}

// GetCreateOptions converts a Tag to TagCreateOptions for use in CreateTag
func (i Tag) GetCreateOptions() (o TagCreateOptions) {
	o.Label = i.Label
	return
}

// TaggedObjectsPagedResponse represents a paginated Tag API response
type TaggedObjectsPagedResponse struct {
	*PageOptions
	Data []TaggedObject `json:"data"`
}

// TagsPagedResponse represents a paginated Tag API response
type TagsPagedResponse struct {
	*PageOptions
	Data []Tag `json:"data"`
}

// endpoint gets the endpoint URL for Tag
func (TagsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Tags.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// endpoint gets the endpoint URL for Tag
func (TaggedObjectsPagedResponse) endpointWithID(c *Client, id string) string {
	endpoint, err := c.Tags.Endpoint()
	if err != nil {
		panic(err)
	}
	endpoint = fmt.Sprintf("%s/%s", endpoint, id)
	return endpoint
}

// appendData appends Tags when processing paginated Tag responses
func (resp *TagsPagedResponse) appendData(r *TagsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// appendData appends TaggedObjects when processing paginated TaggedObjects responses
func (resp *TaggedObjectsPagedResponse) appendData(r *TaggedObjectsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListTags lists Tags
func (c *Client) ListTags(ctx context.Context, opts *ListOptions) ([]Tag, error) {
	response := TagsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixData stores an object of the type defined by Type in Data using RawData
func (i *TaggedObject) fixData() (*TaggedObject, error) {
	switch i.Type {
	case "linode":
		obj := Instance{}
		if err := json.Unmarshal(i.RawData, &obj); err != nil {
			return nil, err
		}
		i.Data = obj
	case "nodebalancer":
		obj := NodeBalancer{}
		if err := json.Unmarshal(i.RawData, &obj); err != nil {
			return nil, err
		}
		i.Data = obj
	case "domain":
		obj := Domain{}
		if err := json.Unmarshal(i.RawData, &obj); err != nil {
			return nil, err
		}
		i.Data = obj
	case "volume":
		obj := Volume{}
		if err := json.Unmarshal(i.RawData, &obj); err != nil {
			return nil, err
		}
		i.Data = obj
	}

	return i, nil
}

// ListTaggedObjects lists Tagged Objects
func (c *Client) ListTaggedObjects(ctx context.Context, label string, opts *ListOptions) (TaggedObjectList, error) {
	response := TaggedObjectsPagedResponse{}
	err := c.listHelperWithID(ctx, &response, label, opts)
	if err != nil {
		return nil, err
	}
	for i := range response.Data {
		if _, err := response.Data[i].fixData(); err != nil {
			return nil, err
		}
	}
	return response.Data, nil
}

// SortedObjects converts a list of TaggedObjects into a Sorted Objects struct, for easier access
func (t TaggedObjectList) SortedObjects() (SortedObjects, error) {
	so := SortedObjects{}
	for _, o := range t {
		switch o.Type {
		case "linode":
			if instance, ok := o.Data.(Instance); ok {
				so.Instances = append(so.Instances, instance)
			} else {
				return so, errors.New("expected an Instance when Type was \"linode\"")
			}
		case "domain":
			if domain, ok := o.Data.(Domain); ok {
				so.Domains = append(so.Domains, domain)
			} else {
				return so, errors.New("expected a Domain when Type was \"domain\"")
			}
		case "volume":
			if volume, ok := o.Data.(Volume); ok {
				so.Volumes = append(so.Volumes, volume)
			} else {
				return so, errors.New("expected an Volume when Type was \"volume\"")
			}
		case "nodebalancer":
			if nodebalancer, ok := o.Data.(NodeBalancer); ok {
				so.NodeBalancers = append(so.NodeBalancers, nodebalancer)
			} else {
				return so, errors.New("expected an NodeBalancer when Type was \"nodebalancer\"")
			}
		}
	}
	return so, nil
}

// CreateTag creates a Tag
func (c *Client) CreateTag(ctx context.Context, createOpts TagCreateOptions) (*Tag, error) {
	var body string
	e, err := c.Tags.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Tag{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Tag), nil
}

// DeleteTag deletes the Tag with the specified id
func (c *Client) DeleteTag(ctx context.Context, label string) error {
	e, err := c.Tags.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%s", e, label)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
