package dnsimple

import (
	"fmt"
	"strconv"
)

// CertificatesService handles communication with the certificate related
// methods of the DNSimple API.
//
// See https://developer.dnsimple.com/v2/domains/certificates
type CertificatesService struct {
	client *Client
}

// Certificate represents a Certificate in DNSimple.
type Certificate struct {
	ID                  int    `json:"id,omitempty"`
	DomainID            int    `json:"domain_id,omitempty"`
	CommonName          string `json:"common_name,omitempty"`
	Years               int    `json:"years,omitempty"`
	State               string `json:"state,omitempty"`
	AuthorityIdentifier string `json:"authority_identifier,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
	ExpiresOn           string `json:"expires_on,omitempty"`
	CertificateRequest  string `json:"csr,omitempty"`
}

// CertificateBundle represents a container for all the PEM-encoded X509 certificate entities,
// such as the private key, the server certificate and the intermediate chain.
type CertificateBundle struct {
	// CertificateRequest       string   `json:"csr,omitempty"`
	PrivateKey               string   `json:"private_key,omitempty"`
	ServerCertificate        string   `json:"server,omitempty"`
	RootCertificate          string   `json:"root,omitempty"`
	IntermediateCertificates []string `json:"chain,omitempty"`
}

func certificatePath(accountID, domainIdentifier, certificateID string) (path string) {
	path = fmt.Sprintf("%v/certificates", domainPath(accountID, domainIdentifier))
	if certificateID != "" {
		path += fmt.Sprintf("/%v", certificateID)
	}
	return
}

// certificateResponse represents a response from an API method that returns a Certificate struct.
type certificateResponse struct {
	Response
	Data *Certificate `json:"data"`
}

// certificateBundleResponse represents a response from an API method that returns a CertificatBundle struct.
type certificateBundleResponse struct {
	Response
	Data *CertificateBundle `json:"data"`
}

// certificatesResponse represents a response from an API method that returns a collection of Certificate struct.
type certificatesResponse struct {
	Response
	Data []Certificate `json:"data"`
}

// ListCertificates list the certificates for a domain.
//
// See https://developer.dnsimple.com/v2/domains/certificates#list
func (s *CertificatesService) ListCertificates(accountID, domainIdentifier string, options *ListOptions) (*certificatesResponse, error) {
	path := versioned(certificatePath(accountID, domainIdentifier, ""))
	certificatesResponse := &certificatesResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, certificatesResponse)
	if err != nil {
		return certificatesResponse, err
	}

	certificatesResponse.HttpResponse = resp
	return certificatesResponse, nil
}

// GetCertificate fetches the certificate.
//
// See https://developer.dnsimple.com/v2/domains/certificates#get
func (s *CertificatesService) GetCertificate(accountID, domainIdentifier string, certificateID int) (*certificateResponse, error) {
	path := versioned(certificatePath(accountID, domainIdentifier, strconv.Itoa(certificateID)))
	certificateResponse := &certificateResponse{}

	resp, err := s.client.get(path, certificateResponse)
	if err != nil {
		return nil, err
	}

	certificateResponse.HttpResponse = resp
	return certificateResponse, nil
}

// DownloadCertificate download the issued server certificate,
// as well the root certificate and the intermediate chain.
//
// See https://developer.dnsimple.com/v2/domains/certificates#download
func (s *CertificatesService) DownloadCertificate(accountID, domainIdentifier string, certificateID int) (*certificateBundleResponse, error) {
	path := versioned(certificatePath(accountID, domainIdentifier, strconv.Itoa(certificateID)) + "/download")
	certificateBundleResponse := &certificateBundleResponse{}

	resp, err := s.client.get(path, certificateBundleResponse)
	if err != nil {
		return nil, err
	}

	certificateBundleResponse.HttpResponse = resp
	return certificateBundleResponse, nil
}

// GetCertificatePrivateKey fetches the certificate private key.
//
// See https://developer.dnsimple.com/v2/domains/certificates#get-private-key
func (s *CertificatesService) GetCertificatePrivateKey(accountID, domainIdentifier string, certificateID int) (*certificateBundleResponse, error) {
	path := versioned(certificatePath(accountID, domainIdentifier, strconv.Itoa(certificateID)) + "/private_key")
	certificateBundleResponse := &certificateBundleResponse{}

	resp, err := s.client.get(path, certificateBundleResponse)
	if err != nil {
		return nil, err
	}

	certificateBundleResponse.HttpResponse = resp
	return certificateBundleResponse, nil
}
