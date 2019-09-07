package govultr

import (
	"context"
	"net/http"
)

// BackupService is the interface to interact with the backup endpoint on the Vultr API
// Link: https://www.vultr.com/api/#backup
type BackupService interface {
	List(ctx context.Context) ([]Backup, error)
	Get(ctx context.Context, backupID string) (*Backup, error)
	ListBySub(ctx context.Context, subID string) ([]Backup, error)
}

// BackupServiceHandler handles interaction with the backup methods for the Vultr API
type BackupServiceHandler struct {
	client *Client
}

// Backup represents a Vultr backup
type Backup struct {
	BackupID    string `json:"BACKUPID"`
	DateCreated string `json:"date_created"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Status      string `json:"status"`
}

// List retrieves a list of all backups on the current account
func (b *BackupServiceHandler) List(ctx context.Context) ([]Backup, error) {
	uri := "/v1/backup/list"
	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	backupsMap := make(map[string]Backup)
	err = b.client.DoWithContext(ctx, req, &backupsMap)
	if err != nil {
		return nil, err
	}

	var backups []Backup
	for _, backup := range backupsMap {
		backups = append(backups, backup)
	}

	return backups, nil
}

// Get retrieves a backup that matches the given backupID
func (b *BackupServiceHandler) Get(ctx context.Context, backupID string) (*Backup, error) {
	uri := "/v1/backup/list"
	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("BACKUPID", backupID)
	req.URL.RawQuery = q.Encode()

	backupsMap := make(map[string]Backup)
	err = b.client.DoWithContext(ctx, req, &backupsMap)
	if err != nil {
		return nil, err
	}

	backup := new(Backup)
	for _, bk := range backupsMap {
		backup = &bk
	}

	return backup, nil
}

// ListBySub retrieves a list of all backups on the current account that match the given subID
func (b *BackupServiceHandler) ListBySub(ctx context.Context, subID string) ([]Backup, error) {
	uri := "/v1/backup/list"
	req, err := b.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("SUBID", subID)
	req.URL.RawQuery = q.Encode()

	backupsMap := make(map[string]Backup)
	err = b.client.DoWithContext(ctx, req, &backupsMap)
	if err != nil {
		return nil, err
	}

	var backups []Backup
	for _, backup := range backupsMap {
		backups = append(backups, backup)
	}

	return backups, nil
}
