package main

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns"
	"github.com/xenolf/lego/providers/http/memcached"
	"github.com/xenolf/lego/providers/http/webroot"
)

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func setup(c *cli.Context) (*Configuration, *Account, *acme.Client) {

	if c.GlobalIsSet("http-timeout") {
		acme.HTTPClient = http.Client{Timeout: time.Duration(c.GlobalInt("http-timeout")) * time.Second}
	}

	if c.GlobalIsSet("dns-timeout") {
		acme.DNSTimeout = time.Duration(c.GlobalInt("dns-timeout")) * time.Second
	}

	if len(c.GlobalStringSlice("dns-resolvers")) > 0 {
		resolvers := []string{}
		for _, resolver := range c.GlobalStringSlice("dns-resolvers") {
			if !strings.Contains(resolver, ":") {
				resolver += ":53"
			}
			resolvers = append(resolvers, resolver)
		}
		acme.RecursiveNameservers = resolvers
	}

	err := checkFolder(c.GlobalString("path"))
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

	conf := NewConfiguration(c)
	if len(c.GlobalString("email")) == 0 {
		logger().Fatal("You have to pass an account (email address) to the program using --email or -m")
	}

	//TODO: move to account struct? Currently MUST pass email.
	acc := NewAccount(c.GlobalString("email"), conf)

	keyType, err := conf.KeyType()
	if err != nil {
		logger().Fatal(err.Error())
	}

	client, err := acme.NewClient(c.GlobalString("server"), acc, keyType)
	if err != nil {
		logger().Fatalf("Could not create client: %s", err.Error())
	}

	if len(c.GlobalStringSlice("exclude")) > 0 {
		client.ExcludeChallenges(conf.ExcludedSolvers())
	}

	if c.GlobalIsSet("webroot") {
		provider, err := webroot.NewHTTPProvider(c.GlobalString("webroot"))
		if err != nil {
			logger().Fatal(err)
		}

		client.SetChallengeProvider(acme.HTTP01, provider)

		// --webroot=foo indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01, acme.TLSSNI01})
	}
	if c.GlobalIsSet("memcached-host") {
		provider, err := memcached.NewMemcachedProvider(c.GlobalStringSlice("memcached-host"))
		if err != nil {
			logger().Fatal(err)
		}

		client.SetChallengeProvider(acme.HTTP01, provider)

		// --memcached-host=foo:11211 indicates that the user specifically want to do a HTTP challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.DNS01, acme.TLSSNI01})
	}
	if c.GlobalIsSet("http") {
		if strings.Index(c.GlobalString("http"), ":") == -1 {
			logger().Fatalf("The --http switch only accepts interface:port or :port for its argument.")
		}
		client.SetHTTPAddress(c.GlobalString("http"))
	}

	if c.GlobalIsSet("tls") {
		if strings.Index(c.GlobalString("tls"), ":") == -1 {
			logger().Fatalf("The --tls switch only accepts interface:port or :port for its argument.")
		}
		client.SetTLSAddress(c.GlobalString("tls"))
	}

	if c.GlobalIsSet("dns") {
    provider, err := dns.NewDNSChallengeProviderByName(c.GlobalString("dns"))
		if err != nil {
			logger().Fatal(err)
		}

		client.SetChallengeProvider(acme.DNS01, provider)

		// --dns=foo indicates that the user specifically want to do a DNS challenge
		// infer that the user also wants to exclude all other challenges
		client.ExcludeChallenges([]acme.Challenge{acme.HTTP01, acme.TLSSNI01})
	}

	return conf, acc, client
}

func saveCertRes(certRes acme.CertificateResource, conf *Configuration) {
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certOut := path.Join(conf.CertPath(), certRes.Domain+".crt")
	privOut := path.Join(conf.CertPath(), certRes.Domain+".key")
	pemOut := path.Join(conf.CertPath(), certRes.Domain+".pem")
	metaOut := path.Join(conf.CertPath(), certRes.Domain+".json")
	issuerOut := path.Join(conf.CertPath(), certRes.Domain+".issuer.crt")

	err := ioutil.WriteFile(certOut, certRes.Certificate, 0600)
	if err != nil {
		logger().Fatalf("Unable to save Certificate for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	if certRes.IssuerCertificate != nil {
		err = ioutil.WriteFile(issuerOut, certRes.IssuerCertificate, 0600)
		if err != nil {
			logger().Fatalf("Unable to save IssuerCertificate for domain %s\n\t%s", certRes.Domain, err.Error())
		}
	}

	if certRes.PrivateKey != nil {
		// if we were given a CSR, we don't know the private key
		err = ioutil.WriteFile(privOut, certRes.PrivateKey, 0600)
		if err != nil {
			logger().Fatalf("Unable to save PrivateKey for domain %s\n\t%s", certRes.Domain, err.Error())
		}

		if conf.context.GlobalBool("pem") {
			err = ioutil.WriteFile(pemOut, bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil), 0600)
			if err != nil {
				logger().Fatalf("Unable to save Certificate and PrivateKey in .pem for domain %s\n\t%s", certRes.Domain, err.Error())
			}
		}

	} else if conf.context.GlobalBool("pem") {
		// we don't have the private key; can't write the .pem file
		logger().Fatalf("Unable to save pem without private key for domain %s\n\t%s; are you using a CSR?", certRes.Domain, err.Error())
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		logger().Fatalf("Unable to marshal CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		logger().Fatalf("Unable to save CertResource for domain %s\n\t%s", certRes.Domain, err.Error())
	}
}

func handleTOS(c *cli.Context, client *acme.Client, acc *Account) {
	// Check for a global accept override
	if c.GlobalBool("accept-tos") {
		err := client.AgreeToTOS()
		if err != nil {
			logger().Fatalf("Could not agree to TOS: %s", err.Error())
		}

		acc.Save()
		return
	}

	reader := bufio.NewReader(os.Stdin)
	logger().Printf("Please review the TOS at %s", acc.Registration.TosURL)

	for {
		logger().Println("Do you accept the TOS? Y/n")
		text, err := reader.ReadString('\n')
		if err != nil {
			logger().Fatalf("Could not read from console: %s", err.Error())
		}

		text = strings.Trim(text, "\r\n")

		if text == "n" {
			logger().Fatal("You did not accept the TOS. Unable to proceed.")
		}

		if text == "Y" || text == "y" || text == "" {
			err = client.AgreeToTOS()
			if err != nil {
				logger().Fatalf("Could not agree to TOS: %s", err.Error())
			}
			acc.Save()
			break
		}

		logger().Println("Your input was invalid. Please answer with one of Y/y, n or by pressing enter.")
	}
}

func readCSRFile(filename string) (*x509.CertificateRequest, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	raw := bytes

	// see if we can find a PEM-encoded CSR
	var p *pem.Block
	rest := bytes
	for {
		// decode a PEM block
		p, rest = pem.Decode(rest)

		// did we fail?
		if p == nil {
			break
		}

		// did we get a CSR?
		if p.Type == "CERTIFICATE REQUEST" {
			raw = p.Bytes
		}
	}

	// no PEM-encoded CSR
	// assume we were given a DER-encoded ASN.1 CSR
	// (if this assumption is wrong, parsing these bytes will fail)
	return x509.ParseCertificateRequest(raw)
}

func run(c *cli.Context) error {
	conf, acc, client := setup(c)
	if acc.Registration == nil {
		reg, err := client.Register()
		if err != nil {
			logger().Fatalf("Could not complete registration\n\t%s", err.Error())
		}

		acc.Registration = reg
		acc.Save()

		logger().Print("!!!! HEADS UP !!!!")
		logger().Printf(`
		Your account credentials have been saved in your Let's Encrypt
		configuration directory at "%s".
		You should make a secure backup	of this folder now. This
		configuration directory will also contain certificates and
		private keys obtained from Let's Encrypt so making regular
		backups of this folder is ideal.`, conf.AccountPath(c.GlobalString("email")))

	}

	// If the agreement URL is empty, the account still needs to accept the LE TOS.
	if acc.Registration.Body.Agreement == "" {
		handleTOS(c, client, acc)
	}

	// we require either domains or csr, but not both
	hasDomains := len(c.GlobalStringSlice("domains")) > 0
	hasCsr := len(c.GlobalString("csr")) > 0
	if hasDomains && hasCsr {
		logger().Fatal("Please specify either --domains/-d or --csr/-c, but not both")
	}
	if !hasDomains && !hasCsr {
		logger().Fatal("Please specify --domains/-d (or --csr/-c if you already have a CSR)")
	}

	var cert acme.CertificateResource
	var failures map[string]error

	if hasDomains {
		// obtain a certificate, generating a new private key
		cert, failures = client.ObtainCertificate(c.GlobalStringSlice("domains"), !c.Bool("no-bundle"), nil, c.Bool("must-staple"))
	} else {
		// read the CSR
		csr, err := readCSRFile(c.GlobalString("csr"))
		if err != nil {
			// we couldn't read the CSR
			failures = map[string]error{"csr": err}
		} else {
			// obtain a certificate for this CSR
			cert, failures = client.ObtainCertificateForCSR(*csr, !c.Bool("no-bundle"))
		}
	}

	if len(failures) > 0 {
		for k, v := range failures {
			logger().Printf("[%s] Could not obtain certificates\n\t%s", k, v.Error())
		}

		// Make sure to return a non-zero exit code if ObtainSANCertificate
		// returned at least one error. Due to us not returning partial
		// certificate we can just exit here instead of at the end.
		os.Exit(1)
	}

	err := checkFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

	saveCertRes(cert, conf)

	return nil
}

func revoke(c *cli.Context) error {

	conf, _, client := setup(c)

	err := checkFolder(conf.CertPath())
	if err != nil {
		logger().Fatalf("Could not check/create path: %s", err.Error())
	}

	for _, domain := range c.GlobalStringSlice("domains") {
		logger().Printf("Trying to revoke certificate for domain %s", domain)

		certPath := path.Join(conf.CertPath(), domain+".crt")
		certBytes, err := ioutil.ReadFile(certPath)

		err = client.RevokeCertificate(certBytes)
		if err != nil {
			logger().Fatalf("Error while revoking the certificate for domain %s\n\t%s", domain, err.Error())
		} else {
			logger().Print("Certificate was revoked.")
		}
	}

	return nil
}

func renew(c *cli.Context) error {
	conf, _, client := setup(c)

	if len(c.GlobalStringSlice("domains")) <= 0 {
		logger().Fatal("Please specify at least one domain.")
	}

	domain := c.GlobalStringSlice("domains")[0]

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certPath := path.Join(conf.CertPath(), domain+".crt")
	privPath := path.Join(conf.CertPath(), domain+".key")
	metaPath := path.Join(conf.CertPath(), domain+".json")

	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		logger().Fatalf("Error while loading the certificate for domain %s\n\t%s", domain, err.Error())
	}

	if c.IsSet("days") {
		expTime, err := acme.GetPEMCertExpiration(certBytes)
		if err != nil {
			logger().Printf("Could not get Certification expiration for domain %s", domain)
		}

		if int(expTime.Sub(time.Now()).Hours()/24.0) > c.Int("days") {
			return nil
		}
	}

	metaBytes, err := ioutil.ReadFile(metaPath)
	if err != nil {
		logger().Fatalf("Error while loading the meta data for domain %s\n\t%s", domain, err.Error())
	}

	var certRes acme.CertificateResource
	err = json.Unmarshal(metaBytes, &certRes)
	if err != nil {
		logger().Fatalf("Error while marshalling the meta data for domain %s\n\t%s", domain, err.Error())
	}

	if c.Bool("reuse-key") {
		keyBytes, err := ioutil.ReadFile(privPath)
		if err != nil {
			logger().Fatalf("Error while loading the private key for domain %s\n\t%s", domain, err.Error())
		}
		certRes.PrivateKey = keyBytes
	}

	certRes.Certificate = certBytes

	newCert, err := client.RenewCertificate(certRes, !c.Bool("no-bundle"), c.Bool("must-staple"))
	if err != nil {
		logger().Fatalf("%s", err.Error())
	}

	saveCertRes(newCert, conf)

	return nil
}
