/*
Copyright
*/
package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/containous/traefik/middlewares"
	"github.com/gorilla/mux"
	"github.com/xenolf/lego/acme"
	"io/ioutil"
	fmtlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// ACMEAccount is used to store lets encrypt registration info
type ACMEAccount struct {
	Email           string
	Registration    *acme.RegistrationResource
	PrivateKey      []byte
	CertificatesMap DomainsCertificates
}

// DomainsCertificates stores a certificate for multiple domains
type DomainsCertificates []DomainsCertificate

func (dc DomainsCertificates) getCertificateForDomain(domainToFind string) (*AcmeCertificate, bool) {
	for _, domainsCertificate := range dc {
		for _, domain := range domainsCertificate.Domains {
			if domain == domainToFind {
				return domainsCertificate.Certificate, true
			}
		}
	}
	return nil, false
}

// DomainsCertificate contains a certificate for multiple domains
type DomainsCertificate struct {
	Domains     []string
	Certificate *AcmeCertificate
}

// GetEmail returns email
func (a ACMEAccount) GetEmail() string {
	return a.Email
}

// GetRegistration returns lets encrypt registration resource
func (a ACMEAccount) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

// GetPrivateKey returns private key
func (a ACMEAccount) GetPrivateKey() crypto.PrivateKey {
	if privateKey, err := x509.ParsePKCS1PrivateKey(a.PrivateKey); err == nil {
		return privateKey
	}
	log.Errorf("Cannot unmarshall private key %+v", a.PrivateKey)
	return nil
}

// AcmeCertificate is used to store certificate info
type AcmeCertificate struct {
	Domain        string
	CertURL       string
	CertStableURL string
	PrivateKey    []byte
	Certificate   []byte
}

func (a *ACME) createACMEConfig(router *middlewares.HandlerSwitcher, proxyRouter *middlewares.HandlerSwitcher) (*tls.Config, error) {
	acme.Logger = fmtlog.New(ioutil.Discard, "", 0)

	if len(a.StorageFile) == 0 {
		return nil, errors.New("Empty StorageFile, please provide a filenmae for certs storage")
	}

	// if certificates in storage, load them
	if fileInfo, err := os.Stat(a.StorageFile); err == nil && fileInfo.Size() != 0 {
		// load account
		acmeAccount, err := a.loadACMEAccount(a)
		if err != nil {
			return nil, err
		}

		// build client
		client, err := a.buildACMEClient(acmeAccount)
		if err != nil {
			return nil, err
		}
		config := &tls.Config{}
		config.Certificates = []tls.Certificate{}
		for _, certificateResource := range acmeAccount.CertificatesMap {
			cert, err := tls.X509KeyPair(certificateResource.Certificate.Certificate, certificateResource.Certificate.PrivateKey)
			if err != nil {
				return nil, err
			}
			leaf, err := x509.ParseCertificate(cert.Certificate[0])
			if err != nil {
				return nil, err
			}
			// <= 30 days left, renew certificate
			if leaf.NotAfter.Before(time.Now().Add(time.Duration(24 * 30 * time.Hour))) {
				renewedCert, err := client.RenewCertificate(acme.CertificateResource{
					Domain:        certificateResource.Certificate.Domain,
					CertURL:       certificateResource.Certificate.CertURL,
					CertStableURL: certificateResource.Certificate.CertStableURL,
					PrivateKey:    certificateResource.Certificate.PrivateKey,
					Certificate:   certificateResource.Certificate.Certificate,
				}, false)
				if err != nil {
					return nil, err
				}
				log.Debugf("Renewed certificate %s", renewedCert.Domain)
				certificateResource.Certificate = &AcmeCertificate{
					Domain:        renewedCert.Domain,
					CertURL:       renewedCert.CertURL,
					CertStableURL: renewedCert.CertStableURL,
					PrivateKey:    renewedCert.PrivateKey,
					Certificate:   renewedCert.Certificate,
				}
				if err = a.saveACMEAccount(acmeAccount); err != nil {
					return nil, err
				}
				cert, err = tls.X509KeyPair(renewedCert.Certificate, renewedCert.PrivateKey)
				if err != nil {
					return nil, err
				}
			}
			config.Certificates = append(config.Certificates, cert)
		}
		config.BuildNameToCertificate()
		if a.OnDemand {
			config.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if !router.GetHandler().Match(&http.Request{URL: &url.URL{}, Host: clientHello.ServerName}, &mux.RouteMatch{}) {
					return nil, nil
				}
				return a.loadCertificateOnDemand(client, acmeAccount, clientHello, proxyRouter)
			}
		}
		return config, nil
	}
	log.Infof("Loading ACME certificates...")

	// Create a user. New accounts need an email and private key to start
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	acmeAccount := &ACMEAccount{
		Email:      a.Email,
		PrivateKey: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	client, err := a.buildACMEClient(acmeAccount)
	if err != nil {
		return nil, err
	}

	//client.SetTLSAddress(acmeConfig.TLSAddress)
	// New users will need to register; be sure to save it
	reg, err := client.Register()
	if err != nil {
		return nil, err
	}
	acmeAccount.Registration = reg

	// The client has a URL to the current Let's Encrypt Subscriber
	// Agreement. The user will need to agree to it.
	err = client.AgreeToTOS()
	if err != nil {
		return nil, err
	}

	config := &tls.Config{}
	config.Certificates = []tls.Certificate{}
	acmeAccount.CertificatesMap = []DomainsCertificate{}

	for _, domain := range a.Domains {
		domains := append([]string{domain.Main}, domain.SANs...)
		certificateResource, err := a.getDomainsCertificates(client, domains, proxyRouter)
		if err != nil {
			return nil, err
		}
		cert, err := tls.X509KeyPair(certificateResource.Certificate, certificateResource.PrivateKey)
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, cert)
		acmeAccount.CertificatesMap = append(acmeAccount.CertificatesMap, DomainsCertificate{Domains: domains, Certificate: certificateResource})
	}
	// BuildNameToCertificate parses the CommonName and SubjectAlternateName fields
	// in each certificate and populates the config.NameToCertificate map.
	config.BuildNameToCertificate()
	if a.OnDemand {
		config.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if !router.GetHandler().Match(&http.Request{URL: &url.URL{}, Host: clientHello.ServerName}, &mux.RouteMatch{}) {
				return nil, nil
			}
			return a.loadCertificateOnDemand(client, acmeAccount, clientHello, proxyRouter)
		}
	}
	if err = a.saveACMEAccount(acmeAccount); err != nil {
		return nil, err
	}
	return config, nil
}

func (a *ACME) buildACMEClient(acmeAccount *ACMEAccount) (*acme.Client, error) {

	// A client facilitates communication with the CA server. This CA URL is
	// configured for a local dev instance of Boulder running in Docker in a VM.
	caServer := "https://acme-v01.api.letsencrypt.org/directory"
	if len(a.CAServer) > 0 {
		caServer = a.CAServer
	}
	client, err := acme.NewClient(caServer, acmeAccount, acme.RSA4096)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Ask the kernel for a free open port that is ready to use
func (a *ACME) getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

func (a *ACME) loadCertificateOnDemand(client *acme.Client, acmeAccount *ACMEAccount, clientHello *tls.ClientHelloInfo, proxyRouter *middlewares.HandlerSwitcher) (*tls.Certificate, error) {
	if certificateResource, ok := acmeAccount.CertificatesMap.getCertificateForDomain(clientHello.ServerName); ok {
		cert, err := tls.X509KeyPair(certificateResource.Certificate, certificateResource.PrivateKey)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}
	certificateResource, err := a.getDomainsCertificates(client, []string{clientHello.ServerName}, proxyRouter)
	if err != nil {
		return nil, err
	}
	log.Debugf("Got certificate on demand for domain %s", clientHello.ServerName)
	acmeAccount.CertificatesMap = append(acmeAccount.CertificatesMap, DomainsCertificate{Domains: []string{clientHello.ServerName}, Certificate: certificateResource})
	if err = a.saveACMEAccount(acmeAccount); err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certificateResource.Certificate, certificateResource.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (a *ACME) loadACMEAccount(acmeConfig *ACME) (*ACMEAccount, error) {
	a.storageLock.Lock()
	defer a.storageLock.Unlock()
	acmeAccount := ACMEAccount{
		CertificatesMap: DomainsCertificates{},
	}
	file, err := ioutil.ReadFile(acmeConfig.StorageFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &acmeAccount); err != nil {
		return nil, err
	}
	log.Infof("Loaded ACME config from storage %s", acmeConfig.StorageFile)
	return &acmeAccount, nil
}

func (a *ACME) saveACMEAccount(acmeAccount *ACMEAccount) error {
	a.storageLock.Lock()
	defer a.storageLock.Unlock()
	// write account to file
	data, err := json.MarshalIndent(acmeAccount, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(a.StorageFile, data, 0644)
}

func (a *ACME) getDomainsCertificates(client *acme.Client, domains []string, proxyRouter *middlewares.HandlerSwitcher) (*AcmeCertificate, error) {
	var proxyRoute *mux.Route
	proxyRoute = proxyRouter.GetHandler().Get("9141156b44763db2a504b8c63cf6f81c")
	if proxyRoute == nil {
		proxyRoute = proxyRouter.GetHandler().NewRoute().PathPrefix("/.well-known/acme-challenge/").Name("9141156b44763db2a504b8c63cf6f81c")
	}
	url, err := url.Parse("http://127.0.0.1:5002")
	if err != nil {
		return nil, err
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	proxyRoute.Handler(reverseProxy)
	defer proxyRoute.Handler(http.NotFoundHandler())
	// The acme library takes care of completing the challenges to obtain the certificate(s).
	// Of course, the hostnames must resolve to this machine or it will fail.
	log.Debugf("Loading ACME certificates %s", domains)
	bundle := false
	client.ExcludeChallenges([]acme.Challenge{acme.TLSSNI01, acme.DNS01})
	client.SetHTTPAddress("127.0.0.1:5002")
	certificate, failures := client.ObtainCertificate(domains, bundle, nil)
	if len(failures) > 0 {
		log.Error(failures)
		return nil, fmt.Errorf("Cannot obtain certificates %s+v", failures)
	}
	return &AcmeCertificate{
		Domain:        certificate.Domain,
		CertURL:       certificate.CertURL,
		CertStableURL: certificate.CertStableURL,
		PrivateKey:    certificate.PrivateKey,
		Certificate:   certificate.Certificate,
	}, nil
}
