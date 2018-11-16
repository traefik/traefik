package gotransip

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	transipAPIHost      = "api.transip.nl"
	transipAPINamespace = "http://www.transip.nl/soap"
)

// APIMode specifies in which mode the API is used. Currently this is only
// supports either readonly or readwrite
type APIMode string

var (
	// APIModeReadOnly specifies that no changes can be made from API calls
	APIModeReadOnly APIMode = "readonly"
	// APIModeReadWrite specifies that changes can be made from API calls
	APIModeReadWrite APIMode = "readwrite"
)

// ClientConfig is a tool to easily create a new Client object
type ClientConfig struct {
	AccountName    string
	PrivateKeyPath string
	PrivateKeyBody []byte
	Mode           APIMode
}

// Client is the interface which all clients should implement
type Client interface {
	Call(SoapRequest, interface{}) error // execute request on client
}

// SOAPClient represents a TransIP API SOAP client, implementing the Client
// interface
type SOAPClient struct {
	soapClient soapClient
}

// Call performs given SOAP request and fills the response into result
func (c SOAPClient) Call(req SoapRequest, result interface{}) error {
	return c.soapClient.call(req, result)
}

// NewSOAPClient returns a new SOAPClient object for given config
// ClientConfig's PrivateKeyPath will override potentially given PrivateKeyBody
func NewSOAPClient(c ClientConfig) (SOAPClient, error) {
	// check account name
	if len(c.AccountName) == 0 {
		return SOAPClient{}, errors.New("AccountName is required")
	}

	// check if private key was given in any form
	if len(c.PrivateKeyPath) == 0 && len(c.PrivateKeyBody) == 0 {
		return SOAPClient{}, errors.New("PrivateKeyPath or PrivateKeyBody is required")
	}

	// if PrivateKeyPath was set, this will override any given PrivateKeyBody
	if len(c.PrivateKeyPath) > 0 {
		// try to open private key and read contents
		if _, err := os.Stat(c.PrivateKeyPath); err != nil {
			return SOAPClient{}, fmt.Errorf("could not open private key: %s", err.Error())
		}

		// read private key so we can pass the body to the soapClient
		var err error
		c.PrivateKeyBody, err = ioutil.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return SOAPClient{}, err
		}
	}

	// default to APIMode read/write
	if len(c.Mode) == 0 {
		c.Mode = APIModeReadWrite
	}

	// create soapClient and pass it to a new Client pointer
	sc := soapClient{
		Login:      c.AccountName,
		Mode:       c.Mode,
		PrivateKey: c.PrivateKeyBody,
	}

	return SOAPClient{
		soapClient: sc,
	}, nil
}

// FakeSOAPClient is a client doing nothing except implementing the gotransip.Client
// interface
// you can however set a fixture XML body which Call will try to Unmarshal into
// result
// useful for testing
type FakeSOAPClient struct {
	fixture []byte // preset this fixture so Call can use it to Unmarshal
}

// FixtureFromFile reads file and sets content as FakeSOAPClient's fixture
func (f *FakeSOAPClient) FixtureFromFile(file string) (err error) {
	// read fixture file
	f.fixture, err = ioutil.ReadFile(file)
	if err != nil {
		err = fmt.Errorf("could not read fixture from file %s: %s", file, err.Error())
	}

	return
}

// Call doesn't do anything except fill the XML unmarshalled result
func (f FakeSOAPClient) Call(req SoapRequest, result interface{}) error {
	// this fake client just parses given fixture as if it was a SOAP response
	return parseSoapResponse(f.fixture, req.Padding, 200, result)
}
