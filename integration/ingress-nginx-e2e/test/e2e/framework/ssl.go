/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	net_url "net/url"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	rsaBits  = 2048
	validFor = 365 * 24 * time.Hour
)

// CreateIngressTLSSecret creates or updates a Secret containing a TLS
// certificate for the given Ingress and returns a TLS configuration suitable
// for HTTP clients to use against that particular Ingress.
func CreateIngressTLSSecret(client kubernetes.Interface, hosts []string, secretName, namespace string) (*tls.Config, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("require a non-empty host for client hello")
	}

	var serverKey, serverCert bytes.Buffer
	var data map[string][]byte
	host := strings.Join(hosts, ",")

	if err := generateRSACert(host, true, &serverKey, &serverCert); err != nil {
		return nil, err
	}

	data = map[string][]byte{
		v1.TLSCertKey:       serverCert.Bytes(),
		v1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	newSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: data,
	}

	var apierr error
	curSecret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err == nil && curSecret != nil {
		curSecret.Data = newSecret.Data
		_, apierr = client.CoreV1().Secrets(namespace).Update(context.TODO(), curSecret, metav1.UpdateOptions{})
	} else {
		_, apierr = client.CoreV1().Secrets(namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{})
	}
	if apierr != nil {
		return nil, apierr
	}

	serverName := hosts[0]
	return tlsConfig(serverName, serverCert.Bytes())
}

// CreateIngressMASecret creates or updates a Secret containing a Mutual Auth
// certificate-chain for the given Ingress and returns a TLS configuration suitable
// for HTTP clients to use against that particular Ingress.
func CreateIngressMASecret(client kubernetes.Interface, host, secretName, namespace string) (*tls.Config, error) {
	if host == "" {
		return nil, fmt.Errorf("requires a non-empty host")
	}

	var caCert, serverKey, serverCert, clientKey, clientCert bytes.Buffer
	var data map[string][]byte

	if err := generateRSAMutualAuthCerts(host, &caCert, &serverKey, &serverCert, &clientKey, &clientCert); err != nil {
		return nil, err
	}

	data = map[string][]byte{
		v1.TLSCertKey:       serverCert.Bytes(),
		v1.TLSPrivateKeyKey: serverKey.Bytes(),
		"ca.crt":            caCert.Bytes(),
	}

	newSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: data,
	}

	var apierr error
	curSecret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err == nil && curSecret != nil {
		curSecret.Data = newSecret.Data
		_, apierr = client.CoreV1().Secrets(namespace).Update(context.TODO(), curSecret, metav1.UpdateOptions{})
	} else {
		_, apierr = client.CoreV1().Secrets(namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{})
	}
	if apierr != nil {
		return nil, apierr
	}

	clientPair, err := tls.X509KeyPair(clientCert.Bytes(), clientKey.Bytes())
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ServerName:         host,
		Certificates:       []tls.Certificate{clientPair},
		InsecureSkipVerify: true, //nolint:gosec // Ignore the gosec error in testing
	}, nil
}

// WaitForTLS waits until the TLS handshake with a given server completes successfully.
func WaitForTLS(url string, tlsConfig *tls.Config) {
	//nolint:staticcheck // TODO: will replace it since wait.Poll is deprecated
	err := wait.Poll(Poll, DefaultTimeout, matchTLSServerName(url, tlsConfig))
	assert.Nil(ginkgo.GinkgoT(), err, "waiting for TLS configuration in URL %s", url)
}

// generateRSACert generates a basic self signed certificate using a key length
// of rsaBits, valid for validFor time.
func generateRSACert(host string, isCA bool, keyOut, certOut io.Writer) error {
	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "default",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return fmt.Errorf("failed creating key: %v", err)
	}

	return nil
}

// generateRSAMutualAuthCerts generates a complete basic self-signed certificate-chain (ca, server, client) using a
// key-length of rsaBits, valid for validFor time.
func generateRSAMutualAuthCerts(host string, caCertOut, serverKeyOut, serverCertOut, clientKeyOut, clientCertOut io.Writer) error {
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}

	// Generate the CA key and CA cert
	caKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	caTemplate := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   host + "-ca",
			Organization: []string{"Acme Co"},
		},
		SerialNumber: serialNumber,
		NotBefore:    notBefore,
		NotAfter:     notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	caTemplate.IsCA = true
	caTemplate.KeyUsage |= x509.KeyUsageCertSign

	caBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	if err := pem.Encode(caCertOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}

	// Generate the Server Key and CSR for the server
	serverKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	// Create the server cert and sign with the csr
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}

	serverTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	serverBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	if err := pem.Encode(serverCertOut, &pem.Block{Type: "CERTIFICATE", Bytes: serverBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}
	if err := pem.Encode(serverKeyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)}); err != nil {
		return fmt.Errorf("failed creating key: %v", err)
	}

	// Create the client key and certificate
	clientKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return fmt.Errorf("failed to generate key: %v", err)
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}
	clientTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host + "-client",
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	clientBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &caTemplate, &clientKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}
	if err := pem.Encode(clientCertOut, &pem.Block{Type: "CERTIFICATE", Bytes: clientBytes}); err != nil {
		return fmt.Errorf("failed creating cert: %v", err)
	}
	if err := pem.Encode(clientKeyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)}); err != nil {
		return fmt.Errorf("failed creating key: %v", err)
	}

	return nil
}

// tlsConfig returns a client TLS configuration for the given server name and
// CA certificate (PEM).
func tlsConfig(serverName string, pemCA []byte) (*tls.Config, error) {
	rootCAPool := x509.NewCertPool()
	if !rootCAPool.AppendCertsFromPEM(pemCA) {
		return nil, fmt.Errorf("error creating CA certificate pool (%s)", serverName)
	}
	return &tls.Config{ //nolint:gosec // Ignore the gosec error in testing
		ServerName: serverName,
		RootCAs:    rootCAPool,
	}, nil
}

// matchTLSServerName connects to the network address corresponding to the
// given URL using the given TLS configuration and returns whether the TLS
// handshake completed successfully.
func matchTLSServerName(url string, tlsConfig *tls.Config) wait.ConditionFunc {
	return func() (bool, error) {
		u, err := net_url.Parse(url)
		if err != nil {
			return false, err
		}

		port := u.Port()
		if port == "" {
			port = "443"
		}

		conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", u.Host, port), tlsConfig)
		if err != nil {
			Logf("Unexpected TLS error: %v", err)
			return false, nil
		}
		defer conn.Close()

		return true, nil
	}
}
