package acme

// StoredData represents the data managed by Store.
type StoredData struct {
	Account      *Account
	Certificates []*CertAndStore
}

// Store is a generic interface that represents a storage.
type Store interface {
	GetAccount(string) (*Account, error)
	SaveAccount(string, *Account) error
	GetCertificates(string) ([]*CertAndStore, error)
	SaveCertificates(string, []*CertAndStore) error
}
