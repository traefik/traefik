package egoscale

// Taggable represents a resource which can have tags attached
//
// This is a helper to fill the resourcetype of a CreateTags call
type Taggable interface {
	// CloudStack resource type of the Taggable type
	ResourceType() string
}

// ResourceTag is a tag associated with a resource
//
// http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.9/management.html
type ResourceTag struct {
	Account      string `json:"account,omitempty"`
	Customer     string `json:"customer,omitempty"`
	Domain       string `json:"domain,omitempty"`
	DomainID     string `json:"domainid,omitempty"`
	Key          string `json:"key"`
	Project      string `json:"project,omitempty"`
	ProjectID    string `json:"projectid,omitempty"`
	ResourceID   string `json:"resourceid,omitempty"`
	ResourceType string `json:"resourcetype,omitempty"`
	Value        string `json:"value"`
}

// CreateTags (Async) creates resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createTags.html
type CreateTags struct {
	ResourceIDs  []string      `json:"resourceids"`
	ResourceType string        `json:"resourcetype"`
	Tags         []ResourceTag `json:"tags"`
	Customer     string        `json:"customer,omitempty"`
}

func (*CreateTags) name() string {
	return "createTags"
}

func (*CreateTags) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// DeleteTags (Async) deletes the resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteTags.html
type DeleteTags struct {
	ResourceIDs  []string      `json:"resourceids"`
	ResourceType string        `json:"resourcetype"`
	Tags         []ResourceTag `json:"tags,omitempty"`
}

func (*DeleteTags) name() string {
	return "deleteTags"
}

func (*DeleteTags) asyncResponse() interface{} {
	return new(booleanAsyncResponse)
}

// ListTags list resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listTags.html
type ListTags struct {
	Account      string `json:"account,omitempty"`
	Customer     string `json:"customer,omitempty"`
	DomainID     string `json:"domainid,omitempty"`
	IsRecursive  bool   `json:"isrecursive,omitempty"`
	Key          string `json:"key,omitempty"`
	Keyword      string `json:"keyword,omitempty"`
	ListAll      bool   `json:"listall,omitempty"`
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"pagesize,omitempty"`
	ProjectID    string `json:"projectid,omitempty"`
	ResourceID   string `json:"resourceid,omitempty"`
	ResourceType string `json:"resourcetype,omitempty"`
	Value        string `json:"value,omitempty"`
}

func (*ListTags) name() string {
	return "listTags"
}

func (*ListTags) response() interface{} {
	return new(ListTagsResponse)
}

// ListTagsResponse represents a list of resource tags
type ListTagsResponse struct {
	Count int           `json:"count"`
	Tag   []ResourceTag `json:"tag"`
}
