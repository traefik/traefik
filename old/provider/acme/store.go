package acme

// StoredData represents the data managed by the Store
type StoredData struct {
	Account        *Account
	Certificates   []*Certificate
	HTTPChallenges map[string]map[string][]byte
	TLSChallenges  map[string]*Certificate
}

// Store is a generic interface to represents a storage
type Store interface {
	GetAccount() (*Account, error)
	SaveAccount(*Account) error
	GetCertificates() ([]*Certificate, error)
	SaveCertificates([]*Certificate) error

	GetHTTPChallengeToken(token, domain string) ([]byte, error)
	SetHTTPChallengeToken(token, domain string, keyAuth []byte) error
	RemoveHTTPChallengeToken(token, domain string) error

	AddTLSChallenge(domain string, cert *Certificate) error
	GetTLSChallenge(domain string) (*Certificate, error)
	RemoveTLSChallenge(domain string) error
}
