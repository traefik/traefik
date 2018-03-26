package egoscale

// Event represents an event in the system
type Event struct {
	ID          string `json:"id"`
	Account     string `json:"account"`
	Created     string `json:"created"`
	Description string `json:"description,omitempty"`
	Domain      string `json:"domain,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	Level       string `json:"level"` // INFO, WARN, ERROR
	ParentID    string `json:"parentid,omitempty"`
	Project     string `json:"project,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
	State       string `json:"state,omitempty"`
	Type        string `json:"type"`
	UserName    string `json:"username,omitempty"`
}

// EventType represent a type of event
type EventType struct {
	Name string `json:"name"`
}

// ListEvents list the events
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listEvents.html
type ListEvents struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	EndDate     string `json:"enddate,omitempty"`
	EntryTime   int    `json:"entrytime,omitempty"`
	ID          string `json:"id,omitempty"`
	IsRecursive bool   `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	Level       string `json:"level,omitempty"` // INFO, WARN, ERROR
	ListAll     bool   `json:"listall,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
	StartDate   string `json:"startdate,omitempty"`
	Type        string `json:"type,omitempty"`
}

func (*ListEvents) name() string {
	return "listEvents"
}

func (*ListEvents) response() interface{} {
	return new(ListEventsResponse)
}

// ListEventsResponse represents a response of a list query
type ListEventsResponse struct {
	Count int     `json:"count"`
	Event []Event `json:"event"`
}

// ListEventTypes list the event types
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listEventTypes.html
type ListEventTypes struct{}

func (*ListEventTypes) name() string {
	return "listEventTypes"
}

func (*ListEventTypes) response() interface{} {
	return new(ListEventTypesResponse)
}

// ListEventTypesResponse represents a response of a list query
type ListEventTypesResponse struct {
	Count     int         `json:"count"`
	EventType []EventType `json:"eventtype"`
}
