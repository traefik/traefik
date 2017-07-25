package restjson_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/awstesting"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/private/protocol/restjson"
	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
	"github.com/aws/aws-sdk-go/private/util"
	"github.com/stretchr/testify/assert"
)

var _ bytes.Buffer // always import bytes
var _ http.Request
var _ json.Marshaler
var _ time.Time
var _ xmlutil.XMLNode
var _ xml.Attr
var _ = ioutil.Discard
var _ = util.Trim("")
var _ = url.Values{}
var _ = io.EOF
var _ = aws.String
var _ = fmt.Println

func init() {
	protocol.RandReader = &awstesting.ZeroReader{}
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService1ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService1ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService1ProtocolTest client from just a session.
//     svc := outputservice1protocoltest.New(mySession)
//
//     // Create a OutputService1ProtocolTest client with additional configuration
//     svc := outputservice1protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService1ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService1ProtocolTest {
	c := p.ClientConfig("outputservice1protocoltest", cfgs...)
	return newOutputService1ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService1ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService1ProtocolTest {
	svc := &OutputService1ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice1protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService1ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService1ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService1TestCaseOperation1 = "OperationName"

// OutputService1TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService1TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService1TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService1TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService1TestCaseOperation1Request method.
//    req, resp := client.OutputService1TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService1ProtocolTest) OutputService1TestCaseOperation1Request(input *OutputService1TestShapeOutputService1TestCaseOperation1Input) (req *request.Request, output *OutputService1TestShapeOutputService1TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService1TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService1TestShapeOutputService1TestCaseOperation1Input{}
	}

	output = &OutputService1TestShapeOutputService1TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService1TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService1TestCaseOperation1 for usage and error information.
func (c *OutputService1ProtocolTest) OutputService1TestCaseOperation1(input *OutputService1TestShapeOutputService1TestCaseOperation1Input) (*OutputService1TestShapeOutputService1TestCaseOperation1Output, error) {
	req, out := c.OutputService1TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService1TestShapeOutputService1TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService1TestShapeOutputService1TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	Char *string `type:"character"`

	Double *float64 `type:"double"`

	FalseBool *bool `type:"boolean"`

	Float *float64 `type:"float"`

	ImaHeader *string `location:"header" type:"string"`

	ImaHeaderLocation *string `location:"header" locationName:"X-Foo" type:"string"`

	Long *int64 `type:"long"`

	Num *int64 `type:"integer"`

	Status *int64 `location:"statusCode" type:"integer"`

	Str *string `type:"string"`

	TrueBool *bool `type:"boolean"`
}

// SetChar sets the Char field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetChar(v string) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Char = &v
	return s
}

// SetDouble sets the Double field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetDouble(v float64) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Double = &v
	return s
}

// SetFalseBool sets the FalseBool field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetFalseBool(v bool) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.FalseBool = &v
	return s
}

// SetFloat sets the Float field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetFloat(v float64) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Float = &v
	return s
}

// SetImaHeader sets the ImaHeader field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetImaHeader(v string) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.ImaHeader = &v
	return s
}

// SetImaHeaderLocation sets the ImaHeaderLocation field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetImaHeaderLocation(v string) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.ImaHeaderLocation = &v
	return s
}

// SetLong sets the Long field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetLong(v int64) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Long = &v
	return s
}

// SetNum sets the Num field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetNum(v int64) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Num = &v
	return s
}

// SetStatus sets the Status field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetStatus(v int64) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Status = &v
	return s
}

// SetStr sets the Str field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetStr(v string) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.Str = &v
	return s
}

// SetTrueBool sets the TrueBool field's value.
func (s *OutputService1TestShapeOutputService1TestCaseOperation1Output) SetTrueBool(v bool) *OutputService1TestShapeOutputService1TestCaseOperation1Output {
	s.TrueBool = &v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService2ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService2ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService2ProtocolTest client from just a session.
//     svc := outputservice2protocoltest.New(mySession)
//
//     // Create a OutputService2ProtocolTest client with additional configuration
//     svc := outputservice2protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService2ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService2ProtocolTest {
	c := p.ClientConfig("outputservice2protocoltest", cfgs...)
	return newOutputService2ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService2ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService2ProtocolTest {
	svc := &OutputService2ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice2protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService2ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService2ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService2TestCaseOperation1 = "OperationName"

// OutputService2TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService2TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService2TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService2TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService2TestCaseOperation1Request method.
//    req, resp := client.OutputService2TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService2ProtocolTest) OutputService2TestCaseOperation1Request(input *OutputService2TestShapeOutputService2TestCaseOperation1Input) (req *request.Request, output *OutputService2TestShapeOutputService2TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService2TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService2TestShapeOutputService2TestCaseOperation1Input{}
	}

	output = &OutputService2TestShapeOutputService2TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService2TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService2TestCaseOperation1 for usage and error information.
func (c *OutputService2ProtocolTest) OutputService2TestCaseOperation1(input *OutputService2TestShapeOutputService2TestCaseOperation1Input) (*OutputService2TestShapeOutputService2TestCaseOperation1Output, error) {
	req, out := c.OutputService2TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService2TestShapeBlobContainer struct {
	_ struct{} `type:"structure"`

	// Foo is automatically base64 encoded/decoded by the SDK.
	Foo []byte `locationName:"foo" type:"blob"`
}

// SetFoo sets the Foo field's value.
func (s *OutputService2TestShapeBlobContainer) SetFoo(v []byte) *OutputService2TestShapeBlobContainer {
	s.Foo = v
	return s
}

type OutputService2TestShapeOutputService2TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService2TestShapeOutputService2TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	// BlobMember is automatically base64 encoded/decoded by the SDK.
	BlobMember []byte `type:"blob"`

	StructMember *OutputService2TestShapeBlobContainer `type:"structure"`
}

// SetBlobMember sets the BlobMember field's value.
func (s *OutputService2TestShapeOutputService2TestCaseOperation1Output) SetBlobMember(v []byte) *OutputService2TestShapeOutputService2TestCaseOperation1Output {
	s.BlobMember = v
	return s
}

// SetStructMember sets the StructMember field's value.
func (s *OutputService2TestShapeOutputService2TestCaseOperation1Output) SetStructMember(v *OutputService2TestShapeBlobContainer) *OutputService2TestShapeOutputService2TestCaseOperation1Output {
	s.StructMember = v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService3ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService3ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService3ProtocolTest client from just a session.
//     svc := outputservice3protocoltest.New(mySession)
//
//     // Create a OutputService3ProtocolTest client with additional configuration
//     svc := outputservice3protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService3ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService3ProtocolTest {
	c := p.ClientConfig("outputservice3protocoltest", cfgs...)
	return newOutputService3ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService3ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService3ProtocolTest {
	svc := &OutputService3ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice3protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService3ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService3ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService3TestCaseOperation1 = "OperationName"

// OutputService3TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService3TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService3TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService3TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService3TestCaseOperation1Request method.
//    req, resp := client.OutputService3TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService3ProtocolTest) OutputService3TestCaseOperation1Request(input *OutputService3TestShapeOutputService3TestCaseOperation1Input) (req *request.Request, output *OutputService3TestShapeOutputService3TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService3TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService3TestShapeOutputService3TestCaseOperation1Input{}
	}

	output = &OutputService3TestShapeOutputService3TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService3TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService3TestCaseOperation1 for usage and error information.
func (c *OutputService3ProtocolTest) OutputService3TestCaseOperation1(input *OutputService3TestShapeOutputService3TestCaseOperation1Input) (*OutputService3TestShapeOutputService3TestCaseOperation1Output, error) {
	req, out := c.OutputService3TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService3TestShapeOutputService3TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService3TestShapeOutputService3TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	StructMember *OutputService3TestShapeTimeContainer `type:"structure"`

	TimeMember *time.Time `type:"timestamp" timestampFormat:"unix"`
}

// SetStructMember sets the StructMember field's value.
func (s *OutputService3TestShapeOutputService3TestCaseOperation1Output) SetStructMember(v *OutputService3TestShapeTimeContainer) *OutputService3TestShapeOutputService3TestCaseOperation1Output {
	s.StructMember = v
	return s
}

// SetTimeMember sets the TimeMember field's value.
func (s *OutputService3TestShapeOutputService3TestCaseOperation1Output) SetTimeMember(v time.Time) *OutputService3TestShapeOutputService3TestCaseOperation1Output {
	s.TimeMember = &v
	return s
}

type OutputService3TestShapeTimeContainer struct {
	_ struct{} `type:"structure"`

	Foo *time.Time `locationName:"foo" type:"timestamp" timestampFormat:"unix"`
}

// SetFoo sets the Foo field's value.
func (s *OutputService3TestShapeTimeContainer) SetFoo(v time.Time) *OutputService3TestShapeTimeContainer {
	s.Foo = &v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService4ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService4ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService4ProtocolTest client from just a session.
//     svc := outputservice4protocoltest.New(mySession)
//
//     // Create a OutputService4ProtocolTest client with additional configuration
//     svc := outputservice4protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService4ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService4ProtocolTest {
	c := p.ClientConfig("outputservice4protocoltest", cfgs...)
	return newOutputService4ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService4ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService4ProtocolTest {
	svc := &OutputService4ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice4protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService4ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService4ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService4TestCaseOperation1 = "OperationName"

// OutputService4TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService4TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService4TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService4TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService4TestCaseOperation1Request method.
//    req, resp := client.OutputService4TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService4ProtocolTest) OutputService4TestCaseOperation1Request(input *OutputService4TestShapeOutputService4TestCaseOperation1Input) (req *request.Request, output *OutputService4TestShapeOutputService4TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService4TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService4TestShapeOutputService4TestCaseOperation1Input{}
	}

	output = &OutputService4TestShapeOutputService4TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService4TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService4TestCaseOperation1 for usage and error information.
func (c *OutputService4ProtocolTest) OutputService4TestCaseOperation1(input *OutputService4TestShapeOutputService4TestCaseOperation1Input) (*OutputService4TestShapeOutputService4TestCaseOperation1Output, error) {
	req, out := c.OutputService4TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService4TestShapeOutputService4TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService4TestShapeOutputService4TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	ListMember []*string `type:"list"`
}

// SetListMember sets the ListMember field's value.
func (s *OutputService4TestShapeOutputService4TestCaseOperation1Output) SetListMember(v []*string) *OutputService4TestShapeOutputService4TestCaseOperation1Output {
	s.ListMember = v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService5ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService5ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService5ProtocolTest client from just a session.
//     svc := outputservice5protocoltest.New(mySession)
//
//     // Create a OutputService5ProtocolTest client with additional configuration
//     svc := outputservice5protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService5ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService5ProtocolTest {
	c := p.ClientConfig("outputservice5protocoltest", cfgs...)
	return newOutputService5ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService5ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService5ProtocolTest {
	svc := &OutputService5ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice5protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService5ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService5ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService5TestCaseOperation1 = "OperationName"

// OutputService5TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService5TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService5TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService5TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService5TestCaseOperation1Request method.
//    req, resp := client.OutputService5TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService5ProtocolTest) OutputService5TestCaseOperation1Request(input *OutputService5TestShapeOutputService5TestCaseOperation1Input) (req *request.Request, output *OutputService5TestShapeOutputService5TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService5TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService5TestShapeOutputService5TestCaseOperation1Input{}
	}

	output = &OutputService5TestShapeOutputService5TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService5TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService5TestCaseOperation1 for usage and error information.
func (c *OutputService5ProtocolTest) OutputService5TestCaseOperation1(input *OutputService5TestShapeOutputService5TestCaseOperation1Input) (*OutputService5TestShapeOutputService5TestCaseOperation1Output, error) {
	req, out := c.OutputService5TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService5TestShapeOutputService5TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService5TestShapeOutputService5TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	ListMember []*OutputService5TestShapeSingleStruct `type:"list"`
}

// SetListMember sets the ListMember field's value.
func (s *OutputService5TestShapeOutputService5TestCaseOperation1Output) SetListMember(v []*OutputService5TestShapeSingleStruct) *OutputService5TestShapeOutputService5TestCaseOperation1Output {
	s.ListMember = v
	return s
}

type OutputService5TestShapeSingleStruct struct {
	_ struct{} `type:"structure"`

	Foo *string `type:"string"`
}

// SetFoo sets the Foo field's value.
func (s *OutputService5TestShapeSingleStruct) SetFoo(v string) *OutputService5TestShapeSingleStruct {
	s.Foo = &v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService6ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService6ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService6ProtocolTest client from just a session.
//     svc := outputservice6protocoltest.New(mySession)
//
//     // Create a OutputService6ProtocolTest client with additional configuration
//     svc := outputservice6protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService6ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService6ProtocolTest {
	c := p.ClientConfig("outputservice6protocoltest", cfgs...)
	return newOutputService6ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService6ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService6ProtocolTest {
	svc := &OutputService6ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice6protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService6ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService6ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService6TestCaseOperation1 = "OperationName"

// OutputService6TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService6TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService6TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService6TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService6TestCaseOperation1Request method.
//    req, resp := client.OutputService6TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService6ProtocolTest) OutputService6TestCaseOperation1Request(input *OutputService6TestShapeOutputService6TestCaseOperation1Input) (req *request.Request, output *OutputService6TestShapeOutputService6TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService6TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService6TestShapeOutputService6TestCaseOperation1Input{}
	}

	output = &OutputService6TestShapeOutputService6TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService6TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService6TestCaseOperation1 for usage and error information.
func (c *OutputService6ProtocolTest) OutputService6TestCaseOperation1(input *OutputService6TestShapeOutputService6TestCaseOperation1Input) (*OutputService6TestShapeOutputService6TestCaseOperation1Output, error) {
	req, out := c.OutputService6TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService6TestShapeOutputService6TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService6TestShapeOutputService6TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	MapMember map[string][]*int64 `type:"map"`
}

// SetMapMember sets the MapMember field's value.
func (s *OutputService6TestShapeOutputService6TestCaseOperation1Output) SetMapMember(v map[string][]*int64) *OutputService6TestShapeOutputService6TestCaseOperation1Output {
	s.MapMember = v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService7ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService7ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService7ProtocolTest client from just a session.
//     svc := outputservice7protocoltest.New(mySession)
//
//     // Create a OutputService7ProtocolTest client with additional configuration
//     svc := outputservice7protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService7ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService7ProtocolTest {
	c := p.ClientConfig("outputservice7protocoltest", cfgs...)
	return newOutputService7ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService7ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService7ProtocolTest {
	svc := &OutputService7ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice7protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService7ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService7ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService7TestCaseOperation1 = "OperationName"

// OutputService7TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService7TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService7TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService7TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService7TestCaseOperation1Request method.
//    req, resp := client.OutputService7TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService7ProtocolTest) OutputService7TestCaseOperation1Request(input *OutputService7TestShapeOutputService7TestCaseOperation1Input) (req *request.Request, output *OutputService7TestShapeOutputService7TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService7TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService7TestShapeOutputService7TestCaseOperation1Input{}
	}

	output = &OutputService7TestShapeOutputService7TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService7TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService7TestCaseOperation1 for usage and error information.
func (c *OutputService7ProtocolTest) OutputService7TestCaseOperation1(input *OutputService7TestShapeOutputService7TestCaseOperation1Input) (*OutputService7TestShapeOutputService7TestCaseOperation1Output, error) {
	req, out := c.OutputService7TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService7TestShapeOutputService7TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService7TestShapeOutputService7TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	MapMember map[string]*time.Time `type:"map"`
}

// SetMapMember sets the MapMember field's value.
func (s *OutputService7TestShapeOutputService7TestCaseOperation1Output) SetMapMember(v map[string]*time.Time) *OutputService7TestShapeOutputService7TestCaseOperation1Output {
	s.MapMember = v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService8ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService8ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService8ProtocolTest client from just a session.
//     svc := outputservice8protocoltest.New(mySession)
//
//     // Create a OutputService8ProtocolTest client with additional configuration
//     svc := outputservice8protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService8ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService8ProtocolTest {
	c := p.ClientConfig("outputservice8protocoltest", cfgs...)
	return newOutputService8ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService8ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService8ProtocolTest {
	svc := &OutputService8ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice8protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService8ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService8ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService8TestCaseOperation1 = "OperationName"

// OutputService8TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService8TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService8TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService8TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService8TestCaseOperation1Request method.
//    req, resp := client.OutputService8TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService8ProtocolTest) OutputService8TestCaseOperation1Request(input *OutputService8TestShapeOutputService8TestCaseOperation1Input) (req *request.Request, output *OutputService8TestShapeOutputService8TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService8TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService8TestShapeOutputService8TestCaseOperation1Input{}
	}

	output = &OutputService8TestShapeOutputService8TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService8TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService8TestCaseOperation1 for usage and error information.
func (c *OutputService8ProtocolTest) OutputService8TestCaseOperation1(input *OutputService8TestShapeOutputService8TestCaseOperation1Input) (*OutputService8TestShapeOutputService8TestCaseOperation1Output, error) {
	req, out := c.OutputService8TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService8TestShapeOutputService8TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService8TestShapeOutputService8TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	StrType *string `type:"string"`
}

// SetStrType sets the StrType field's value.
func (s *OutputService8TestShapeOutputService8TestCaseOperation1Output) SetStrType(v string) *OutputService8TestShapeOutputService8TestCaseOperation1Output {
	s.StrType = &v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService9ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService9ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService9ProtocolTest client from just a session.
//     svc := outputservice9protocoltest.New(mySession)
//
//     // Create a OutputService9ProtocolTest client with additional configuration
//     svc := outputservice9protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService9ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService9ProtocolTest {
	c := p.ClientConfig("outputservice9protocoltest", cfgs...)
	return newOutputService9ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService9ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService9ProtocolTest {
	svc := &OutputService9ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice9protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService9ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService9ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService9TestCaseOperation1 = "OperationName"

// OutputService9TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService9TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService9TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService9TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService9TestCaseOperation1Request method.
//    req, resp := client.OutputService9TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService9ProtocolTest) OutputService9TestCaseOperation1Request(input *OutputService9TestShapeOutputService9TestCaseOperation1Input) (req *request.Request, output *OutputService9TestShapeOutputService9TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService9TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService9TestShapeOutputService9TestCaseOperation1Input{}
	}

	output = &OutputService9TestShapeOutputService9TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService9TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService9TestCaseOperation1 for usage and error information.
func (c *OutputService9ProtocolTest) OutputService9TestCaseOperation1(input *OutputService9TestShapeOutputService9TestCaseOperation1Input) (*OutputService9TestShapeOutputService9TestCaseOperation1Output, error) {
	req, out := c.OutputService9TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService9TestShapeOutputService9TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService9TestShapeOutputService9TestCaseOperation1Output struct {
	_ struct{} `type:"structure"`

	AllHeaders map[string]*string `location:"headers" type:"map"`

	PrefixedHeaders map[string]*string `location:"headers" locationName:"X-" type:"map"`
}

// SetAllHeaders sets the AllHeaders field's value.
func (s *OutputService9TestShapeOutputService9TestCaseOperation1Output) SetAllHeaders(v map[string]*string) *OutputService9TestShapeOutputService9TestCaseOperation1Output {
	s.AllHeaders = v
	return s
}

// SetPrefixedHeaders sets the PrefixedHeaders field's value.
func (s *OutputService9TestShapeOutputService9TestCaseOperation1Output) SetPrefixedHeaders(v map[string]*string) *OutputService9TestShapeOutputService9TestCaseOperation1Output {
	s.PrefixedHeaders = v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService10ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService10ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService10ProtocolTest client from just a session.
//     svc := outputservice10protocoltest.New(mySession)
//
//     // Create a OutputService10ProtocolTest client with additional configuration
//     svc := outputservice10protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService10ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService10ProtocolTest {
	c := p.ClientConfig("outputservice10protocoltest", cfgs...)
	return newOutputService10ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService10ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService10ProtocolTest {
	svc := &OutputService10ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice10protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService10ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService10ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService10TestCaseOperation1 = "OperationName"

// OutputService10TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService10TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService10TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService10TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService10TestCaseOperation1Request method.
//    req, resp := client.OutputService10TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService10ProtocolTest) OutputService10TestCaseOperation1Request(input *OutputService10TestShapeOutputService10TestCaseOperation1Input) (req *request.Request, output *OutputService10TestShapeOutputService10TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService10TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService10TestShapeOutputService10TestCaseOperation1Input{}
	}

	output = &OutputService10TestShapeOutputService10TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService10TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService10TestCaseOperation1 for usage and error information.
func (c *OutputService10ProtocolTest) OutputService10TestCaseOperation1(input *OutputService10TestShapeOutputService10TestCaseOperation1Input) (*OutputService10TestShapeOutputService10TestCaseOperation1Output, error) {
	req, out := c.OutputService10TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService10TestShapeBodyStructure struct {
	_ struct{} `type:"structure"`

	Foo *string `type:"string"`
}

// SetFoo sets the Foo field's value.
func (s *OutputService10TestShapeBodyStructure) SetFoo(v string) *OutputService10TestShapeBodyStructure {
	s.Foo = &v
	return s
}

type OutputService10TestShapeOutputService10TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService10TestShapeOutputService10TestCaseOperation1Output struct {
	_ struct{} `type:"structure" payload:"Data"`

	Data *OutputService10TestShapeBodyStructure `type:"structure"`

	Header *string `location:"header" locationName:"X-Foo" type:"string"`
}

// SetData sets the Data field's value.
func (s *OutputService10TestShapeOutputService10TestCaseOperation1Output) SetData(v *OutputService10TestShapeBodyStructure) *OutputService10TestShapeOutputService10TestCaseOperation1Output {
	s.Data = v
	return s
}

// SetHeader sets the Header field's value.
func (s *OutputService10TestShapeOutputService10TestCaseOperation1Output) SetHeader(v string) *OutputService10TestShapeOutputService10TestCaseOperation1Output {
	s.Header = &v
	return s
}

// The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type OutputService11ProtocolTest struct {
	*client.Client
}

// New creates a new instance of the OutputService11ProtocolTest client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a OutputService11ProtocolTest client from just a session.
//     svc := outputservice11protocoltest.New(mySession)
//
//     // Create a OutputService11ProtocolTest client with additional configuration
//     svc := outputservice11protocoltest.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func NewOutputService11ProtocolTest(p client.ConfigProvider, cfgs ...*aws.Config) *OutputService11ProtocolTest {
	c := p.ClientConfig("outputservice11protocoltest", cfgs...)
	return newOutputService11ProtocolTestClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion, c.SigningName)
}

// newClient creates, initializes and returns a new service client instance.
func newOutputService11ProtocolTestClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion, signingName string) *OutputService11ProtocolTest {
	svc := &OutputService11ProtocolTest{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   "outputservice11protocoltest",
				SigningName:   signingName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBackNamed(v4.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restjson.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restjson.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restjson.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restjson.UnmarshalErrorHandler)

	return svc
}

// newRequest creates a new request for a OutputService11ProtocolTest operation and runs any
// custom request initialization.
func (c *OutputService11ProtocolTest) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	return req
}

const opOutputService11TestCaseOperation1 = "OperationName"

// OutputService11TestCaseOperation1Request generates a "aws/request.Request" representing the
// client's request for the OutputService11TestCaseOperation1 operation. The "output" return
// value can be used to capture response data after the request's "Send" method
// is called.
//
// See OutputService11TestCaseOperation1 for usage and error information.
//
// Creating a request object using this method should be used when you want to inject
// custom logic into the request's lifecycle using a custom handler, or if you want to
// access properties on the request object before or after sending the request. If
// you just want the service response, call the OutputService11TestCaseOperation1 method directly
// instead.
//
// Note: You must call the "Send" method on the returned request object in order
// to execute the request.
//
//    // Example sending a request using the OutputService11TestCaseOperation1Request method.
//    req, resp := client.OutputService11TestCaseOperation1Request(params)
//
//    err := req.Send()
//    if err == nil { // resp is now filled
//        fmt.Println(resp)
//    }
func (c *OutputService11ProtocolTest) OutputService11TestCaseOperation1Request(input *OutputService11TestShapeOutputService11TestCaseOperation1Input) (req *request.Request, output *OutputService11TestShapeOutputService11TestCaseOperation1Output) {
	op := &request.Operation{
		Name:     opOutputService11TestCaseOperation1,
		HTTPPath: "/",
	}

	if input == nil {
		input = &OutputService11TestShapeOutputService11TestCaseOperation1Input{}
	}

	output = &OutputService11TestShapeOutputService11TestCaseOperation1Output{}
	req = c.newRequest(op, input, output)
	return
}

// OutputService11TestCaseOperation1 API operation for .
//
// Returns awserr.Error for service API and SDK errors. Use runtime type assertions
// with awserr.Error's Code and Message methods to get detailed information about
// the error.
//
// See the AWS API reference guide for 's
// API operation OutputService11TestCaseOperation1 for usage and error information.
func (c *OutputService11ProtocolTest) OutputService11TestCaseOperation1(input *OutputService11TestShapeOutputService11TestCaseOperation1Input) (*OutputService11TestShapeOutputService11TestCaseOperation1Output, error) {
	req, out := c.OutputService11TestCaseOperation1Request(input)
	err := req.Send()
	return out, err
}

type OutputService11TestShapeOutputService11TestCaseOperation1Input struct {
	_ struct{} `type:"structure"`
}

type OutputService11TestShapeOutputService11TestCaseOperation1Output struct {
	_ struct{} `type:"structure" payload:"Stream"`

	Stream []byte `type:"blob"`
}

// SetStream sets the Stream field's value.
func (s *OutputService11TestShapeOutputService11TestCaseOperation1Output) SetStream(v []byte) *OutputService11TestShapeOutputService11TestCaseOperation1Output {
	s.Stream = v
	return s
}

//
// Tests begin here
//

func TestOutputService1ProtocolTestScalarMembersCase1(t *testing.T) {
	svc := NewOutputService1ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"Str\": \"myname\", \"Num\": 123, \"FalseBool\": false, \"TrueBool\": true, \"Float\": 1.2, \"Double\": 1.3, \"Long\": 200, \"Char\": \"a\"}"))
	req, out := svc.OutputService1TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers
	req.HTTPResponse.Header.Set("ImaHeader", "test")
	req.HTTPResponse.Header.Set("X-Foo", "abc")

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "a", *out.Char)
	assert.Equal(t, 1.3, *out.Double)
	assert.Equal(t, false, *out.FalseBool)
	assert.Equal(t, 1.2, *out.Float)
	assert.Equal(t, "test", *out.ImaHeader)
	assert.Equal(t, "abc", *out.ImaHeaderLocation)
	assert.Equal(t, int64(200), *out.Long)
	assert.Equal(t, int64(123), *out.Num)
	assert.Equal(t, int64(200), *out.Status)
	assert.Equal(t, "myname", *out.Str)
	assert.Equal(t, true, *out.TrueBool)

}

func TestOutputService2ProtocolTestBlobMembersCase1(t *testing.T) {
	svc := NewOutputService2ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"BlobMember\": \"aGkh\", \"StructMember\": {\"foo\": \"dGhlcmUh\"}}"))
	req, out := svc.OutputService2TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "hi!", string(out.BlobMember))
	assert.Equal(t, "there!", string(out.StructMember.Foo))

}

func TestOutputService3ProtocolTestTimestampMembersCase1(t *testing.T) {
	svc := NewOutputService3ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"TimeMember\": 1398796238, \"StructMember\": {\"foo\": 1398796238}}"))
	req, out := svc.OutputService3TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, time.Unix(1.398796238e+09, 0).UTC().String(), out.StructMember.Foo.String())
	assert.Equal(t, time.Unix(1.398796238e+09, 0).UTC().String(), out.TimeMember.String())

}

func TestOutputService4ProtocolTestListsCase1(t *testing.T) {
	svc := NewOutputService4ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"ListMember\": [\"a\", \"b\"]}"))
	req, out := svc.OutputService4TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "a", *out.ListMember[0])
	assert.Equal(t, "b", *out.ListMember[1])

}

func TestOutputService5ProtocolTestListsWithStructureMemberCase1(t *testing.T) {
	svc := NewOutputService5ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"ListMember\": [{\"Foo\": \"a\"}, {\"Foo\": \"b\"}]}"))
	req, out := svc.OutputService5TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "a", *out.ListMember[0].Foo)
	assert.Equal(t, "b", *out.ListMember[1].Foo)

}

func TestOutputService6ProtocolTestMapsCase1(t *testing.T) {
	svc := NewOutputService6ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"MapMember\": {\"a\": [1, 2], \"b\": [3, 4]}}"))
	req, out := svc.OutputService6TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, int64(1), *out.MapMember["a"][0])
	assert.Equal(t, int64(2), *out.MapMember["a"][1])
	assert.Equal(t, int64(3), *out.MapMember["b"][0])
	assert.Equal(t, int64(4), *out.MapMember["b"][1])

}

func TestOutputService7ProtocolTestComplexMapValuesCase1(t *testing.T) {
	svc := NewOutputService7ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"MapMember\": {\"a\": 1398796238, \"b\": 1398796238}}"))
	req, out := svc.OutputService7TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, time.Unix(1.398796238e+09, 0).UTC().String(), out.MapMember["a"].String())
	assert.Equal(t, time.Unix(1.398796238e+09, 0).UTC().String(), out.MapMember["b"].String())

}

func TestOutputService8ProtocolTestIgnoresExtraDataCase1(t *testing.T) {
	svc := NewOutputService8ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"foo\": \"bar\"}"))
	req, out := svc.OutputService8TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used

}

func TestOutputService9ProtocolTestSupportsHeaderMapsCase1(t *testing.T) {
	svc := NewOutputService9ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{}"))
	req, out := svc.OutputService9TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers
	req.HTTPResponse.Header.Set("Content-Length", "10")
	req.HTTPResponse.Header.Set("X-Bam", "boo")
	req.HTTPResponse.Header.Set("X-Foo", "bar")

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "10", *out.AllHeaders["Content-Length"])
	assert.Equal(t, "boo", *out.AllHeaders["X-Bam"])
	assert.Equal(t, "bar", *out.AllHeaders["X-Foo"])
	assert.Equal(t, "boo", *out.PrefixedHeaders["Bam"])
	assert.Equal(t, "bar", *out.PrefixedHeaders["Foo"])

}

func TestOutputService10ProtocolTestJSONPayloadCase1(t *testing.T) {
	svc := NewOutputService10ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("{\"Foo\": \"abc\"}"))
	req, out := svc.OutputService10TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers
	req.HTTPResponse.Header.Set("X-Foo", "baz")

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "abc", *out.Data.Foo)
	assert.Equal(t, "baz", *out.Header)

}

func TestOutputService11ProtocolTestStreamingPayloadCase1(t *testing.T) {
	svc := NewOutputService11ProtocolTest(unit.Session, &aws.Config{Endpoint: aws.String("https://test")})

	buf := bytes.NewReader([]byte("abc"))
	req, out := svc.OutputService11TestCaseOperation1Request(nil)
	req.HTTPResponse = &http.Response{StatusCode: 200, Body: ioutil.NopCloser(buf), Header: http.Header{}}

	// set headers

	// unmarshal response
	restjson.UnmarshalMeta(req)
	restjson.Unmarshal(req)
	assert.NoError(t, req.Error)

	// assert response
	assert.NotNil(t, out) // ensure out variable is used
	assert.Equal(t, "abc", string(out.Stream))

}
