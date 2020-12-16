package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

// https://shaneutt.com/blog/golang-ca-and-signed-cert-go/

const (
	caPrivPath = "ca.key"
	caCertPath = "ca.pem"
	caCRLPath  = "ca.crl"

	httpPort = "8081"
)

// SimpleCA is a simple CA.
type SimpleCA struct {
	now      time.Time
	validity time.Time

	priv    *rsa.PrivateKey
	privPEM *bytes.Buffer
	cert    *x509.Certificate
	certPEM *bytes.Buffer
	crl     []byte

	clients        []*SimpleCAClient
	revokedClients []*SimpleCAClient

	rsaKeyLength int
}

// SimpleCAClient is a client as created by a SimpleCA.
type SimpleCAClient struct {
	name string

	priv    *rsa.PrivateKey
	privPEM *bytes.Buffer
	cert    *x509.Certificate

	signedPEM *bytes.Buffer
}

func (sca *SimpleCA) init() error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Country:            []string{"CA Country"},
			Organization:       []string{"CA Organization"},
			OrganizationalUnit: []string{"CA OrganizationalUnit"},

			Locality: []string{"CA Locality"},
			Province: []string{"CA Province"},

			StreetAddress: []string{"CA StreetAddress"},
			PostalCode:    []string{"CA PostalCode"},

			// SerialNumber: "CA SerialNumber",
			// CommonName:   "CA CommonName",
		},
		NotBefore:             sca.now,
		NotAfter:              sca.validity,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	log.Print("generating CA private key… ")
	priv, err := rsa.GenerateKey(rand.Reader, sca.rsaKeyLength)
	if err != nil {
		return err
	}
	log.Println("done.")

	privPEM := new(bytes.Buffer)
	if err := pem.Encode(privPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return err
	}

	sca.priv = priv
	sca.privPEM = privPEM
	sca.cert = cert
	sca.certPEM = certPEM

	return nil
}

// NewClient creates a new client.
func (sca *SimpleCA) NewClient(name string, serial int64) (*SimpleCAClient, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Country:            []string{name + " Country"},
			Organization:       []string{name + " Organization"},
			OrganizationalUnit: []string{name + " OrganizationalUnit"},

			Locality: []string{name + " Locality"},
			Province: []string{name + " Province"},

			StreetAddress: []string{name + " StreetAddress"},
			PostalCode:    []string{name + " PostalCode"},

			// SerialNumber: name + " SerialNumber",
			// CommonName:   name + " CommonName",
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    sca.now,
		NotAfter:     sca.validity,
		SubjectKeyId: []byte{1, 2, 3, 4, 5},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,

		CRLDistributionPoints: []string{"http://localhost:" + httpPort + "/" + caCRLPath},
	}

	log.Printf("generating %s's private key… ", name)
	priv, err := rsa.GenerateKey(rand.Reader, sca.rsaKeyLength)
	if err != nil {
		return nil, err
	}
	log.Println("done.")

	privPEM := new(bytes.Buffer)
	if err := pem.Encode(privPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		return nil, err
	}

	client := &SimpleCAClient{
		name: name,

		priv:    priv,
		privPEM: privPEM,
		cert:    cert,
	}
	if err := sca.signClient(client); err != nil {
		return nil, err
	}
	sca.clients = append(sca.clients, client)

	return client, nil
}

func (sca *SimpleCA) signClient(client *SimpleCAClient) error {
	signedBytes, err := x509.CreateCertificate(rand.Reader, client.cert, sca.cert, &client.priv.PublicKey, sca.priv)
	if err != nil {
		return err
	}

	signedPEM := new(bytes.Buffer)
	if err := pem.Encode(signedPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: signedBytes,
	}); err != nil {
		return err
	}

	client.signedPEM = signedPEM

	return nil
}

// RevokeClient revokes a client certificate.
func (sca *SimpleCA) RevokeClient(client *SimpleCAClient) error {
	sca.revokedClients = append(sca.revokedClients, client)

	var rev []pkix.RevokedCertificate
	for _, c := range sca.revokedClients {
		log.Println("revoking", c.name)
		rev = append(rev, pkix.RevokedCertificate{
			SerialNumber:   c.cert.SerialNumber,
			RevocationTime: sca.now.UTC(),
		})
	}

	crlBytes, err := sca.cert.CreateCRL(rand.Reader, sca.priv, rev, sca.now, sca.validity)
	if err != nil {
		return err
	}

	sca.crl = crlBytes

	return nil
}

// WriteFiles writes out all files.
// This includes certificates, private keys and a CRL.
func (sca *SimpleCA) WriteFiles() error {
	const (
		uRW    = 0600
		uRWgoR = 0644
	)

	if err := ioutil.WriteFile(caPrivPath, sca.privPEM.Bytes(), uRW); err != nil {
		return err
	}
	if err := ioutil.WriteFile(caCertPath, sca.certPEM.Bytes(), uRWgoR); err != nil {
		return err
	}
	if err := ioutil.WriteFile(caCRLPath, sca.crl, uRWgoR); err != nil {
		return err
	}
	for _, c := range sca.clients {
		if err := ioutil.WriteFile(c.name+".key", c.privPEM.Bytes(), uRW); err != nil {
			return err
		}
		if err := ioutil.WriteFile(c.name+".pem", c.signedPEM.Bytes(), uRWgoR); err != nil {
			return err
		}
	}

	return nil
}

// NewCA creates a new CA.
func NewCA() (*SimpleCA, error) {
	now := time.Now()
	ca := &SimpleCA{
		now:      now,
		validity: now.AddDate(10 /* years */, 0 /* months */, 0 /* days */),

		rsaKeyLength: 4096,
	}
	if err := ca.init(); err != nil {
		return nil, err
	}
	return ca, nil
}

func run() error {
	ca, err := NewCA()
	if err != nil {
		return err
	}

	client1, err := ca.NewClient("client1", 41)
	if err != nil {
		return err
	}
	_ = client1
	client2, err := ca.NewClient("client2", 42)
	if err != nil {
		return err
	}

	if err := ca.RevokeClient(client2); err != nil {
		return err
	}

	return ca.WriteFiles()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

	log.Println("starting HTTP server to serve CRL on port…", httpPort)
	http.HandleFunc("/"+caCRLPath, func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, caCRLPath)
	})
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
