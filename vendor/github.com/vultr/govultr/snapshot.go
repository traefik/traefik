package govultr

import (
	"context"
	"net/http"
	"net/url"
)

// SnapshotService is the interface to interact with Snapshot endpoints on the Vultr API
// Link: https://www.vultr.com/api/#snapshot
type SnapshotService interface {
	Create(ctx context.Context, InstanceID, description string) (*Snapshot, error)
	CreateFromURL(ctx context.Context, snapshotURL string) (*Snapshot, error)
	Delete(ctx context.Context, snapshotID string) error
	List(ctx context.Context) ([]Snapshot, error)
	Get(ctx context.Context, snapshotID string) (*Snapshot, error)
}

// SnapshotServiceHandler handles interaction with the snapshot methods for the Vultr API
type SnapshotServiceHandler struct {
	Client *Client
}

// Snapshot represents a Vultr snapshot
type Snapshot struct {
	SnapshotID  string `json:"SNAPSHOTID"`
	DateCreated string `json:"date_created"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Status      string `json:"status"`
	OsID        string `json:"OSID"`
	AppID       string `json:"APPID"`
}

// Snapshots represent a collection of snapshots
type Snapshots []Snapshot

// Create makes a snapshot of a provided server
func (s *SnapshotServiceHandler) Create(ctx context.Context, InstanceID, description string) (*Snapshot, error) {

	uri := "/v1/snapshot/create"

	values := url.Values{
		"SUBID":       {InstanceID},
		"description": {description},
	}

	req, err := s.Client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	snapshot := new(Snapshot)
	err = s.Client.DoWithContext(ctx, req, snapshot)

	if err != nil {
		return nil, err
	}

	snapshot.Description = description
	return snapshot, nil
}

// CreateFromURL will create a snapshot based on an image iso from a URL you provide
func (s *SnapshotServiceHandler) CreateFromURL(ctx context.Context, snapshotURL string) (*Snapshot, error) {
	uri := "/v1/snapshot/create_from_url"

	values := url.Values{
		"url": {snapshotURL},
	}

	req, err := s.Client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	snapshot := new(Snapshot)
	err = s.Client.DoWithContext(ctx, req, snapshot)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// Delete a snapshot based on snapshotID
func (s *SnapshotServiceHandler) Delete(ctx context.Context, snapshotID string) error {
	uri := "/v1/snapshot/destroy"

	values := url.Values{
		"SNAPSHOTID": {snapshotID},
	}

	req, err := s.Client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.Client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List of snapshots details
func (s *SnapshotServiceHandler) List(ctx context.Context) ([]Snapshot, error) {
	uri := "/v1/snapshot/list"

	req, err := s.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	snapshotMap := make(map[string]Snapshot)
	err = s.Client.DoWithContext(ctx, req, &snapshotMap)

	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot

	for _, s := range snapshotMap {
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

// Get individual details of a snapshot based on snapshotID
func (s *SnapshotServiceHandler) Get(ctx context.Context, snapshotID string) (*Snapshot, error) {
	uri := "/v1/snapshot/list"

	req, err := s.Client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	if snapshotID != "" {
		q := req.URL.Query()
		q.Add("SNAPSHOTID", snapshotID)
		req.URL.RawQuery = q.Encode()
	}

	snapshotMap := make(map[string]Snapshot)
	err = s.Client.DoWithContext(ctx, req, &snapshotMap)

	if err != nil {
		return nil, err
	}

	snapshot := new(Snapshot)

	for _, s := range snapshotMap {
		snapshot = &s
	}

	return snapshot, nil
}
