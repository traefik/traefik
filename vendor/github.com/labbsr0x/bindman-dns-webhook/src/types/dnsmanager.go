package types

// DNSManager defines the operations a DNS Manager provider should implement
type DNSManager interface {

	// GetDNSRecords retrieves all the dns records being managed
	GetDNSRecords() ([]DNSRecord, error)

	// GetDNSRecord retrieves the dns record identified by name
	GetDNSRecord(name, recordType string) (*DNSRecord, error)

	// RemoveDNSRecord removes a DNS record
	RemoveDNSRecord(name, recordType string) error

	// AddDNSRecord adds a new DNS record
	AddDNSRecord(record DNSRecord) error

	// UpdateDNSRecord updates an existing DNS record
	UpdateDNSRecord(record DNSRecord) error
}
