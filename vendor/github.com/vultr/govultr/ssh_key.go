package govultr

import (
	"context"
	"net/http"
	"net/url"
)

// SSHKeyService is the interface to interact with the SSH Key endpoints on the Vultr API
// Link: https://www.vultr.com/api/#sshkey
type SSHKeyService interface {
	Create(ctx context.Context, name, sshKey string) (*SSHKey, error)
	Delete(ctx context.Context, sshKeyID string) error
	List(ctx context.Context) ([]SSHKey, error)
	Update(ctx context.Context, sshKey *SSHKey) error
}

// SSHKeyServiceHandler handles interaction with the SSH Key methods for the Vultr API
type SSHKeyServiceHandler struct {
	client *Client
}

// SSHKey represents an SSH Key on Vultr
type SSHKey struct {
	SSHKeyID    string `json:"SSHKEYID"`
	Name        string `json:"name"`
	Key         string `json:"ssh_key"`
	DateCreated string `json:"date_created"`
}

// Create will add the specified SSH Key to your Vultr account
func (s *SSHKeyServiceHandler) Create(ctx context.Context, name, sshKey string) (*SSHKey, error) {

	uri := "/v1/sshkey/create"

	values := url.Values{
		"name":    {name},
		"ssh_key": {sshKey},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	key := new(SSHKey)

	err = s.client.DoWithContext(ctx, req, key)

	if err != nil {
		return nil, err
	}

	key.Name = name
	key.Key = sshKey

	return key, nil
}

// Delete will delete the specified SHH Key from your Vultr account
func (s *SSHKeyServiceHandler) Delete(ctx context.Context, sshKeyID string) error {

	uri := "/v1/sshkey/destroy"

	values := url.Values{
		"SSHKEYID": {sshKeyID},
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will list all the SSH Keys associated with your Vultr account
func (s *SSHKeyServiceHandler) List(ctx context.Context) ([]SSHKey, error) {

	uri := "/v1/sshkey/list"

	req, err := s.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	sshKeysMap := make(map[string]SSHKey)
	err = s.client.DoWithContext(ctx, req, &sshKeysMap)
	if err != nil {
		return nil, err
	}

	var sshKeys []SSHKey
	for _, key := range sshKeysMap {
		sshKeys = append(sshKeys, key)
	}

	return sshKeys, nil
}

// Update will update the given SSH Key. Empty strings will be ignored.
func (s *SSHKeyServiceHandler) Update(ctx context.Context, sshKey *SSHKey) error {

	uri := "/v1/sshkey/update"

	values := url.Values{
		"SSHKEYID": {sshKey.SSHKeyID},
	}

	// Optional
	if sshKey.Name != "" {
		values.Add("name", sshKey.Name)
	}
	if sshKey.Key != "" {
		values.Add("ssh_key", sshKey.Key)
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = s.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
