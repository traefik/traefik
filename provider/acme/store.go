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

	GetHTTPChallenges() (map[string]map[string][]byte, error)
	SaveHTTPChallenges(map[string]map[string][]byte) error

	AddTLSChallenge(domain string, cert *Certificate) error
	GetTLSChallenge(domain string) (*Certificate, error)
	RemoveTLSChallenge(domain string) error
}
