package acme

// StoredData represents the data managed by Store.
type StoredData struct {
	Account      *Account
	Certificates []*CertAndStore
}

// Store is a generic interface that represents a storage.
type Store interface {
	GetAccount(resolverName string) (*Account, error)
	SaveAccount(resolverName string, account *Account) error
	GetCertificates(resolverName string) ([]*CertAndStore, error)
	SaveCertificates(resolverName string, certificates []*CertAndStore) error
}
