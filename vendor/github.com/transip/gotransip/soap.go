package gotransip

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// format for SOAP envelopes
	soapEnvelopeFixture string = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://www.transip.nl/soap" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
	<SOAP-ENV:Body>%s</SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
)

// getSOAPArgs returns XML representing given name and argument as SOAP body
func getSOAPArgs(name string, input ...string) []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("<ns1:%s>", name))
	for _, x := range input {
		buf.WriteString(x)
	}
	buf.WriteString(fmt.Sprintf("</ns1:%s>", name))

	return buf.Bytes()
}

// getSOAPArg returns XML representing given input argument as SOAP parameters
// in combination with getSOAPArgs you can build SOAP body
func getSOAPArg(name string, input interface{}) (output string) {
	switch input.(type) {
	case []string:
		i := input.([]string)
		output = fmt.Sprintf(`<%s SOAP-ENC:arrayType="xsd:string[%d]" xsi:type="ns1:ArrayOfString">`, name, len(i))
		for _, x := range i {
			output = output + fmt.Sprintf(`<item xsi:type="xsd:string">%s</item>`, x)
		}
		output = output + fmt.Sprintf("</%s>", name)
	case string:
		output = fmt.Sprintf(`<%s xsi:type="xsd:string">%s</%s>`, name, input, name)
	case int, int32, int64:
		output = fmt.Sprintf(`<%s xsi:type="xsd:integer">%d</%s>`, name, input, name)
	}

	return
}

type soapFault struct {
	Code        string `xml:"faultcode,omitempty"`
	Description string `xml:"faultstring,omitempty"`
}

func (s soapFault) String() string {
	return fmt.Sprintf("SOAP Fault %s: %s", s.Code, s.Description)
}

// paramsEncoder allows SoapParams to hook into encoding theirselves, useful when
// fields consist of complex structs
type paramsEncoder interface {
	EncodeParams(ParamsContainer, string)
	EncodeArgs(string) string
}

// ParamsContainer is the interface a type should implement to be able to hold
// SOAP parameters
type ParamsContainer interface {
	Len() int
	Add(string, interface{})
}

// soapParams is a utility to make sure parameter data is encoded into a query
// in the same order as we set them. The TransIP API requires this order for
// verifying the signature
type soapParams struct {
	keys   []string
	values []interface{}
}

// Add adds parameter data to the end of this SoapParams
func (s *soapParams) Add(k string, v interface{}) {
	if s.keys == nil {
		s.keys = make([]string, 0)
	}

	if s.values == nil {
		s.values = make([]interface{}, 0)
	}

	s.keys = append(s.keys, k)
	s.values = append(s.values, v)
}

// Len returns amount of parameters set in this SoapParams
func (s soapParams) Len() int {
	return len(s.keys)
}

// Encode returns a URL-like query string that can be used to generate a request's
// signature. It's similar to url.Values.Encode() but without sorting of the keys
// and based on the value's type it tries to encode accordingly.
func (s soapParams) Encode() string {
	var buf bytes.Buffer
	var key string

	for i, v := range s.values {
		// if this is not the first parameter, prefix with &
		if i > 0 {
			buf.WriteString("&")
		}

		// for empty data fields, don't encode anything
		if v == nil {
			continue
		}

		// get key for this value
		key = s.keys[i]

		switch v.(type) {
		case []string:
			c := v.([]string)
			for j, cc := range c {
				if j > 0 {
					buf.WriteString("&")
				}
				buf.WriteString(fmt.Sprintf("%s[%d]=", key, j))
				buf.WriteString(strings.Replace(url.QueryEscape(cc), "+", "%20", -1))
			}
		case string:
			c := v.(string)
			buf.WriteString(fmt.Sprintf("%s=", key))
			buf.WriteString(strings.Replace(url.QueryEscape(c), "+", "%20", -1))
		case int, int8, int16, int32, int64:
			buf.WriteString(fmt.Sprintf("%s=", key))
			buf.WriteString(fmt.Sprintf("%d", v))
		case bool:
			c := v.(bool)
			buf.WriteString(fmt.Sprintf("%s=", key))
			if c {
				buf.WriteString("1")
			}
		default:
			continue
		}
	}

	return buf.String()
}

type soapHeader struct {
	XMLName  struct{} `xml:"Header"`
	Contents []byte   `xml:",innerxml"`
}

type soapBody struct {
	XMLName  struct{} `xml:"Body"`
	Contents []byte   `xml:",innerxml"`
}

type soapResponse struct {
	Response struct {
		InnerXML []byte `xml:",innerxml"`
	} `xml:"return"`
}

type soapEnvelope struct {
	XMLName struct{} `xml:"Envelope"`
	Header  soapHeader
	Body    soapBody
}

// SoapRequest holds all information for perfoming a SOAP request
// Arguments to the request can be specified with AddArgument
// If padding is defined, the SOAP response will be parsed after it being padded
// with items in Padding in reverse order
type SoapRequest struct {
	Service string
	Method  string
	params  *soapParams // params used for creating signature
	args    []string    // XML body arguments
	Padding []string
}

// AddArgument adds an argument to the SoapRequest; the arguments ared used to
// fill the XML request body as well as to create a valid signature for the
// request
func (sr *SoapRequest) AddArgument(key string, value interface{}) {
	if sr.params == nil {
		sr.params = &soapParams{}
	}

	// check if value implements paramsEncoder
	if pe, ok := value.(paramsEncoder); ok {
		sr.args = append(sr.args, pe.EncodeArgs(key))
		pe.EncodeParams(sr.params, "")
		return
	}

	switch value.(type) {
	case []string:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(key, value))
	case string:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(key, value.(string)))
	case int, int8, int16, int32, int64:
		sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), value)
		sr.args = append(sr.args, getSOAPArg(key, fmt.Sprintf("%d", value)))
	default:
		// check if value implements the String interface
		if str, ok := value.(fmt.Stringer); ok {
			sr.params.Add(fmt.Sprintf("%d", sr.params.Len()), str.String())
			sr.args = append(sr.args, getSOAPArg(key, str.String()))
		}
	}
}

func (sr SoapRequest) getEnvelope() string {
	return fmt.Sprintf(soapEnvelopeFixture, getSOAPArgs(sr.Method, sr.args...))
}

type soapClient struct {
	Login      string
	Mode       APIMode
	PrivateKey []byte
}

// httpReqForSoapRequest creates the HTTP request for a specific SoapRequest
// this includes setting the URL, POST body and cookies
func (s soapClient) httpReqForSoapRequest(req SoapRequest) (*http.Request, error) {
	// format URL
	url := fmt.Sprintf("https://%s/soap/?service=%s", transipAPIHost, req.Service)

	// create HTTP request
	// TransIP API SOAP requests are always POST requests
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(req.getEnvelope()))
	if err != nil {
		return nil, err
	}

	// generate a number-used-once, a.k.a. nonce
	// seeding the RNG is important if we want to do prevent using the same nonce
	// in 2 sequential requests
	rand.Seed(time.Now().UnixNano())
	nonce := fmt.Sprintf("%d", rand.Int())
	// set time of request, used later for signature as well
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// set cookies required for the request
	// most of these cookies are used for the signature as well so they should
	// obviously match
	httpReq.AddCookie(&http.Cookie{
		Name:  "login",
		Value: s.Login,
	})
	httpReq.AddCookie(&http.Cookie{
		Name:  "mode",
		Value: string(s.Mode),
	})
	httpReq.AddCookie(&http.Cookie{
		Name:  "timestamp",
		Value: timestamp,
	})
	httpReq.AddCookie(&http.Cookie{
		Name:  "nonce",
		Value: nonce,
	})

	// add params required for signature to the request parameters
	if req.params == nil {
		req.params = &soapParams{}
	}
	// TransIP API is quite picky on the order of the parameters
	// so don't change anything in the order below
	req.params.Add("__method", req.Method)
	req.params.Add("__service", req.Service)
	req.params.Add("__hostname", transipAPIHost)
	req.params.Add("__timestamp", timestamp)
	req.params.Add("__nonce", nonce)

	signature, err := signWithKey(req.params, s.PrivateKey)
	if err != nil {
		return nil, err
	}

	// add signature of the request to the cookies as well
	httpReq.AddCookie(&http.Cookie{
		Name:  "signature",
		Value: signature,
	})

	return httpReq, nil
}

func parseSoapResponse(data []byte, padding []string, statusCode int, result interface{}) error {
	// try to decode the resulting XML
	var env soapEnvelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return err
	}

	// try to decode the body to a soapFault
	var fault soapFault
	if err := xml.Unmarshal(env.Body.Contents, &fault); err != nil {
		return err
	}

	// by checking fault's Code, we can determine if the response body in fact
	// was a SOAP fault and if it was: return it as an error
	if len(fault.Code) > 0 {
		return errors.New(fault.String())
	}

	// try to decode into soapResponse
	sr := soapResponse{}
	if err := xml.Unmarshal(env.Body.Contents, &sr); err != nil {
		return err
	}

	// if the response was empty and HTTP status was 200, consider it a success
	if len(sr.Response.InnerXML) == 0 && statusCode == 200 {
		return nil
	}

	// it seems like xml.Unmarshal won't work well on the most outer element
	// so even when no Padding is defined, always pad with "transip" element
	p := append([]string{"transip"}, padding...)
	innerXML := padXMLData(sr.Response.InnerXML, p)

	// try to decode to given result interface
	return xml.Unmarshal([]byte(innerXML), &result)
}

func (s *soapClient) call(req SoapRequest, result interface{}) error {
	// get http request for soap request
	httpReq, err := s.httpReqForSoapRequest(req)
	if err != nil {
		return err
	}

	// create HTTP client and do the actual request
	client := &http.Client{Timeout: time.Second * 10}
	// make sure to verify the validity of remote certificate
	// this is the default, but adding this flag here makes it easier to toggle
	// it for testing/debugging
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request error:\n%s", err.Error())
	}
	defer resp.Body.Close()

	// read entire response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse SOAP response into given result interface
	return parseSoapResponse(b, req.Padding, resp.StatusCode, result)
}

// apply given padding around the XML data fed into this function
// padding is applied in reverse order, so last element of padding is the
// innermost element in the resulting XML
func padXMLData(data []byte, padding []string) []byte {
	// get right information from padding elements by matching to regex
	re, _ := regexp.Compile("^<?(?:([^ ]+) )?([^>]+)>?$")

	var prefix, suffix []byte
	var tag, attr string
	// go over each padding element
	for i := len(padding); i > 0; i-- {
		res := re.FindStringSubmatch(padding[i-1])
		// no attribute was given
		if len(res[1]) == 0 {
			tag = res[2]
			attr = ""
		} else {
			tag = res[1]
			attr = " " + res[2]
		}

		prefix = []byte(fmt.Sprintf("<%s%s>", tag, attr))
		suffix = []byte(fmt.Sprintf("</%s>", tag))
		data = append(append(prefix, data...), suffix...)
	}

	return data
}

// TestParamsContainer is only useful for unit testing the ParamsContainer
// implementation of other type
type TestParamsContainer struct {
	Prm string
}

// Add just makes sure we use Len(), key and value in the result so it can be
// tested
func (t *TestParamsContainer) Add(key string, value interface{}) {
	var suffix string
	if t.Len() > 0 {
		suffix = "&"
	}

	if value != nil {
		var prm string
		switch value.(type) {
		case bool:
			c := value.(bool)
			if c {
				prm = "1"
			}
		default:
			prm = fmt.Sprintf("%s", value)
		}

		suffix = suffix + fmt.Sprintf("%d%s=%s", t.Len(), key, prm)
	}

	t.Prm = t.Prm + suffix
}

// Len returns current length of test data in TestParamsContainer
func (t TestParamsContainer) Len() int {
	return len(t.Prm)
}
