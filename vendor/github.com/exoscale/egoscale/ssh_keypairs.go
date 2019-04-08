package egoscale

import (
	"context"
	"fmt"
)

// SSHKeyPair represents an SSH key pair
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#creating-the-ssh-keypair
type SSHKeyPair struct {
	Fingerprint string `json:"fingerprint,omitempty" doc:"Fingerprint of the public key"`
	Name        string `json:"name,omitempty" doc:"Name of the keypair"`
	PrivateKey  string `json:"privatekey,omitempty" doc:"Private key"`
}

// Delete removes the given SSH key, by Name
func (ssh SSHKeyPair) Delete(ctx context.Context, client *Client) error {
	if ssh.Name == "" {
		return fmt.Errorf("an SSH Key Pair may only be deleted using Name")
	}

	return client.BooleanRequestWithContext(ctx, &DeleteSSHKeyPair{
		Name: ssh.Name,
	})
}

// ListRequest builds the ListSSHKeyPairs request
func (ssh SSHKeyPair) ListRequest() (ListCommand, error) {
	req := &ListSSHKeyPairs{
		Fingerprint: ssh.Fingerprint,
		Name:        ssh.Name,
	}

	return req, nil
}

// CreateSSHKeyPair represents a new keypair to be created
type CreateSSHKeyPair struct {
	Name string `json:"name" doc:"Name of the keypair"`
	_    bool   `name:"createSSHKeyPair" description:"Create a new keypair and returns the private key"`
}

// Response returns the struct to unmarshal
func (CreateSSHKeyPair) Response() interface{} {
	return new(SSHKeyPair)
}

// DeleteSSHKeyPair represents a new keypair to be created
type DeleteSSHKeyPair struct {
	Name string `json:"name" doc:"Name of the keypair"`
	_    bool   `name:"deleteSSHKeyPair" description:"Deletes a keypair by name"`
}

// Response returns the struct to unmarshal
func (DeleteSSHKeyPair) Response() interface{} {
	return new(BooleanResponse)
}

// RegisterSSHKeyPair represents a new registration of a public key in a keypair
type RegisterSSHKeyPair struct {
	Name      string `json:"name" doc:"Name of the keypair"`
	PublicKey string `json:"publickey" doc:"Public key material of the keypair"`
	_         bool   `name:"registerSSHKeyPair" description:"Register a public key in a keypair under a certain name"`
}

// Response returns the struct to unmarshal
func (RegisterSSHKeyPair) Response() interface{} {
	return new(SSHKeyPair)
}

//go:generate go run generate/main.go -interface=Listable ListSSHKeyPairs

// ListSSHKeyPairs represents a query for a list of SSH KeyPairs
type ListSSHKeyPairs struct {
	Fingerprint string `json:"fingerprint,omitempty" doc:"A public key fingerprint to look for"`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	Name        string `json:"name,omitempty" doc:"A key pair name to look for"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	_           bool   `name:"listSSHKeyPairs" description:"List registered keypairs"`
}

// ListSSHKeyPairsResponse represents a list of SSH key pairs
type ListSSHKeyPairsResponse struct {
	Count      int          `json:"count"`
	SSHKeyPair []SSHKeyPair `json:"sshkeypair"`
}

// ResetSSHKeyForVirtualMachine (Async) represents a change for the key pairs
type ResetSSHKeyForVirtualMachine struct {
	ID      *UUID  `json:"id" doc:"The ID of the virtual machine"`
	KeyPair string `json:"keypair" doc:"Name of the ssh key pair used to login to the virtual machine"`
	_       bool   `name:"resetSSHKeyForVirtualMachine" description:"Resets the SSH Key for virtual machine. The virtual machine must be in a \"Stopped\" state."`
}

// Response returns the struct to unmarshal
func (ResetSSHKeyForVirtualMachine) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job
func (ResetSSHKeyForVirtualMachine) AsyncResponse() interface{} {
	return new(VirtualMachine)
}
