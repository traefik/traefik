package acme

// StoredData represents the data managed by Store.
type StoredData struct {
	Account      *Account
	Certificates []*CertAndStore
}

// StoredChallengeData represents the data managed by ChallengeStore.
type StoredChallengeData struct {
	HTTPChallenges map[string]map[string][]byte
	TLSChallenges  map[string]*Certificate
}

// Store is a generic interface that represents a storage.
type Store interface {
	GetAccount(string) (*Account, error)
	SaveAccount(string, *Account) error
	GetCertificates(string) ([]*CertAndStore, error)
	SaveCertificates(string, []*CertAndStore) error
}

// ChallengeStore is a generic interface that represents a store for challenge data.
type ChallengeStore interface {
	GetHTTPChallengeToken(token, domain string) ([]byte, error)
	SetHTTPChallengeToken(token, domain string, keyAuth []byte) error
	RemoveHTTPChallengeToken(token, domain string) error

	AddTLSChallenge(domain string, cert *Certificate) error
	GetTLSChallenge(domain string) (*Certificate, error)
	RemoveTLSChallenge(domain string) error
}
