package vault

// StoredData represents the data managed by the Store
type StoredData struct {
	Certificates []*Certificate
}

// Store is a generic interface to represents a storage
type Store interface {
	GetCertificates() ([]*Certificate, error)
	SaveCertificates([]*Certificate) error
}
