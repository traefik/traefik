package egoscale

// SSHKeyPair represents an SSH key pair
type SSHKeyPair struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
	PrivateKey  string `json:"privatekey,omitempty"`
}

// CreateSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createSSHKeyPair.html
type CreateSSHKeyPair struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*CreateSSHKeyPair) name() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(CreateSSHKeyPairResponse)
}

// CreateSSHKeyPairResponse represents the creation of an SSH Key Pair
type CreateSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// DeleteSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSSHKeyPair.html
type DeleteSSHKeyPair struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*DeleteSSHKeyPair) name() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanSyncResponse)
}

// RegisterSSHKeyPair represents a new registration of a public key in a keypair
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerSSHKeyPair.html
type RegisterSSHKeyPair struct {
	Name      string `json:"name"`
	PublicKey string `json:"publickey"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*RegisterSSHKeyPair) name() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(RegisterSSHKeyPairResponse)
}

// RegisterSSHKeyPairResponse represents the creation of an SSH Key Pair
type RegisterSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// ListSSHKeyPairs represents a query for a list of SSH KeyPairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listSSHKeyPairs.html
type ListSSHKeyPairs struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	IsRecursive bool   `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	ListAll     bool   `json:"listall,omitempty"`
	Name        string `json:"name,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
}

func (*ListSSHKeyPairs) name() string {
	return "listSSHKeyPairs"
}

func (*ListSSHKeyPairs) response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

// ListSSHKeyPairsResponse represents a list of SSH key pairs
type ListSSHKeyPairsResponse struct {
	Count      int          `json:"count"`
	SSHKeyPair []SSHKeyPair `json:"sshkeypair"`
}

// ResetSSHKeyForVirtualMachine (Async) represents a change for the key pairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/resetSSHKeyForVirtualMachine.html
type ResetSSHKeyForVirtualMachine struct {
	ID        string `json:"id"`
	KeyPair   string `json:"keypair"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

func (*ResetSSHKeyForVirtualMachine) name() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(ResetSSHKeyForVirtualMachineResponse)
}

// ResetSSHKeyForVirtualMachineResponse represents the modified VirtualMachine
type ResetSSHKeyForVirtualMachineResponse VirtualMachineResponse

// CreateKeypair create a new SSH Key Pair
//
// Deprecated: will go away, use the API directly
func (exo *Client) CreateKeypair(name string) (*SSHKeyPair, error) {
	req := &CreateSSHKeyPair{
		Name: name,
	}
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}

	keypair := resp.(*CreateSSHKeyPairResponse).KeyPair
	return &keypair, nil
}

// DeleteKeypair deletes an SSH key pair
//
// Deprecated: will go away, use the API directly
func (exo *Client) DeleteKeypair(name string) error {
	req := &DeleteSSHKeyPair{
		Name: name,
	}
	return exo.BooleanRequest(req)
}

// RegisterKeypair registers a public key in a keypair
//
// Deprecated: will go away, use the API directly
func (exo *Client) RegisterKeypair(name string, publicKey string) (*SSHKeyPair, error) {
	req := &RegisterSSHKeyPair{
		Name:      name,
		PublicKey: publicKey,
	}
	resp, err := exo.Request(req)
	if err != nil {
		return nil, err
	}

	keypair := resp.(*RegisterSSHKeyPairResponse).KeyPair
	return &keypair, nil
}
