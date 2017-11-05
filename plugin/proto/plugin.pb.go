// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugin.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	plugin.proto

It has these top-level messages:
	Request
	Response
	HttpRequest
	HttpResponse
	Empty
	ValueList
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type Request struct {
	RequestUuid string       `protobuf:"bytes,1,opt,name=request_uuid,json=requestUuid" json:"request_uuid,omitempty"`
	Request     *HttpRequest `protobuf:"bytes,2,opt,name=request" json:"request,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto1.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Request) GetRequestUuid() string {
	if m != nil {
		return m.RequestUuid
	}
	return ""
}

func (m *Request) GetRequest() *HttpRequest {
	if m != nil {
		return m.Request
	}
	return nil
}

type Response struct {
	Request       *HttpRequest  `protobuf:"bytes,1,opt,name=request" json:"request,omitempty"`
	Response      *HttpResponse `protobuf:"bytes,2,opt,name=response" json:"response,omitempty"`
	StopChain     bool          `protobuf:"varint,3,opt,name=stopChain" json:"stopChain,omitempty"`
	RenderContent bool          `protobuf:"varint,4,opt,name=renderContent" json:"renderContent,omitempty"`
	Redirect      bool          `protobuf:"varint,5,opt,name=redirect" json:"redirect,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto1.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Response) GetRequest() *HttpRequest {
	if m != nil {
		return m.Request
	}
	return nil
}

func (m *Response) GetResponse() *HttpResponse {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *Response) GetStopChain() bool {
	if m != nil {
		return m.StopChain
	}
	return false
}

func (m *Response) GetRenderContent() bool {
	if m != nil {
		return m.RenderContent
	}
	return false
}

func (m *Response) GetRedirect() bool {
	if m != nil {
		return m.Redirect
	}
	return false
}

type HttpRequest struct {
	// Method specifies the HTTP method (GET, POST, PUT, etc.).
	// For client requests an empty string means GET.
	Method string `protobuf:"bytes,1,opt,name=method" json:"method,omitempty"`
	// URL specifies either the URI being requested (for server
	// requests) or the URL to access (for client requests).
	//
	// For server requests the URL is parsed from the URI
	// supplied on the Request-Line as stored in RequestURI.  For
	// most requests, fields other than Path and RawQuery will be
	// empty. (See RFC 2616, Section 5.1.2)
	//
	// For client requests, the URL's Host specifies the server to
	// connect to, while the Request's Host field optionally
	// specifies the Host header value to send in the HTTP
	// request.
	Url string `protobuf:"bytes,2,opt,name=url" json:"url,omitempty"`
	// The protocol version for incoming server requests.
	//
	// For client requests these fields are ignored. The HTTP
	// client code always uses either HTTP/1.1 or HTTP/2.
	// See the docs on Transport for details.
	Proto            string                `protobuf:"bytes,3,opt,name=proto" json:"proto,omitempty"`
	ProtoMajor       int32                 `protobuf:"varint,4,opt,name=protoMajor" json:"protoMajor,omitempty"`
	ProtoMinor       int32                 `protobuf:"varint,5,opt,name=protoMinor" json:"protoMinor,omitempty"`
	Header           map[string]*ValueList `protobuf:"bytes,6,rep,name=header" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Body             []byte                `protobuf:"bytes,7,opt,name=body,proto3" json:"body,omitempty"`
	ContentLength    int64                 `protobuf:"varint,8,opt,name=contentLength" json:"contentLength,omitempty"`
	TransferEncoding []string              `protobuf:"bytes,9,rep,name=transferEncoding" json:"transferEncoding,omitempty"`
	Close            bool                  `protobuf:"varint,10,opt,name=close" json:"close,omitempty"`
	Host             string                `protobuf:"bytes,11,opt,name=host" json:"host,omitempty"`
	FormValues       map[string]*ValueList `protobuf:"bytes,12,rep,name=formValues" json:"formValues,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	PostFormValues   map[string]*ValueList `protobuf:"bytes,13,rep,name=postFormValues" json:"postFormValues,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	// reserved for multipart.Form = 14
	Trailer    map[string]*ValueList `protobuf:"bytes,15,rep,name=trailer" json:"trailer,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	RemoteAddr string                `protobuf:"bytes,16,opt,name=remoteAddr" json:"remoteAddr,omitempty"`
	// RequestURI is the unmodified Request-URI of the
	// Request-Line (RFC 2616, Section 5.1) as sent by the client
	// to a server. Usually the URL field should be used instead.
	// It is an error to set this field in an HTTP client request.
	RequestUri string `protobuf:"bytes,17,opt,name=requestUri" json:"requestUri,omitempty"`
}

func (m *HttpRequest) Reset()                    { *m = HttpRequest{} }
func (m *HttpRequest) String() string            { return proto1.CompactTextString(m) }
func (*HttpRequest) ProtoMessage()               {}
func (*HttpRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *HttpRequest) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *HttpRequest) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

func (m *HttpRequest) GetProto() string {
	if m != nil {
		return m.Proto
	}
	return ""
}

func (m *HttpRequest) GetProtoMajor() int32 {
	if m != nil {
		return m.ProtoMajor
	}
	return 0
}

func (m *HttpRequest) GetProtoMinor() int32 {
	if m != nil {
		return m.ProtoMinor
	}
	return 0
}

func (m *HttpRequest) GetHeader() map[string]*ValueList {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *HttpRequest) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *HttpRequest) GetContentLength() int64 {
	if m != nil {
		return m.ContentLength
	}
	return 0
}

func (m *HttpRequest) GetTransferEncoding() []string {
	if m != nil {
		return m.TransferEncoding
	}
	return nil
}

func (m *HttpRequest) GetClose() bool {
	if m != nil {
		return m.Close
	}
	return false
}

func (m *HttpRequest) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *HttpRequest) GetFormValues() map[string]*ValueList {
	if m != nil {
		return m.FormValues
	}
	return nil
}

func (m *HttpRequest) GetPostFormValues() map[string]*ValueList {
	if m != nil {
		return m.PostFormValues
	}
	return nil
}

func (m *HttpRequest) GetTrailer() map[string]*ValueList {
	if m != nil {
		return m.Trailer
	}
	return nil
}

func (m *HttpRequest) GetRemoteAddr() string {
	if m != nil {
		return m.RemoteAddr
	}
	return ""
}

func (m *HttpRequest) GetRequestUri() string {
	if m != nil {
		return m.RequestUri
	}
	return ""
}

type HttpResponse struct {
	StatusCode int32                 `protobuf:"varint,1,opt,name=status_code,json=statusCode" json:"status_code,omitempty"`
	Header     map[string]*ValueList `protobuf:"bytes,2,rep,name=header" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Body       []byte                `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *HttpResponse) Reset()                    { *m = HttpResponse{} }
func (m *HttpResponse) String() string            { return proto1.CompactTextString(m) }
func (*HttpResponse) ProtoMessage()               {}
func (*HttpResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *HttpResponse) GetStatusCode() int32 {
	if m != nil {
		return m.StatusCode
	}
	return 0
}

func (m *HttpResponse) GetHeader() map[string]*ValueList {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *HttpResponse) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto1.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ValueList struct {
	Value []string `protobuf:"bytes,1,rep,name=value" json:"value,omitempty"`
}

func (m *ValueList) Reset()                    { *m = ValueList{} }
func (m *ValueList) String() string            { return proto1.CompactTextString(m) }
func (*ValueList) ProtoMessage()               {}
func (*ValueList) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ValueList) GetValue() []string {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto1.RegisterType((*Request)(nil), "proto.Request")
	proto1.RegisterType((*Response)(nil), "proto.Response")
	proto1.RegisterType((*HttpRequest)(nil), "proto.HttpRequest")
	proto1.RegisterType((*HttpResponse)(nil), "proto.HttpResponse")
	proto1.RegisterType((*Empty)(nil), "proto.Empty")
	proto1.RegisterType((*ValueList)(nil), "proto.ValueList")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Middleware service

type MiddlewareClient interface {
	ServeHttp(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type middlewareClient struct {
	cc *grpc.ClientConn
}

func NewMiddlewareClient(cc *grpc.ClientConn) MiddlewareClient {
	return &middlewareClient{cc}
}

func (c *middlewareClient) ServeHttp(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/proto.Middleware/ServeHttp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Middleware service

type MiddlewareServer interface {
	ServeHttp(context.Context, *Request) (*Response, error)
}

func RegisterMiddlewareServer(s *grpc.Server, srv MiddlewareServer) {
	s.RegisterService(&_Middleware_serviceDesc, srv)
}

func _Middleware_ServeHttp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MiddlewareServer).ServeHttp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Middleware/ServeHttp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MiddlewareServer).ServeHttp(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Middleware_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Middleware",
	HandlerType: (*MiddlewareServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ServeHttp",
			Handler:    _Middleware_ServeHttp_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin.proto",
}

func init() { proto1.RegisterFile("plugin.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 625 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0x5f, 0x4f, 0xd4, 0x40,
	0x10, 0x4f, 0x39, 0xee, 0x7a, 0x9d, 0x16, 0x38, 0x17, 0x63, 0x36, 0x17, 0x03, 0xa5, 0x31, 0xa4,
	0x31, 0x06, 0x13, 0x4c, 0xfc, 0xc3, 0x9b, 0x12, 0x0c, 0x89, 0xa0, 0x66, 0x51, 0x1f, 0x7c, 0x21,
	0xe5, 0x76, 0xe1, 0xaa, 0xbd, 0x6e, 0xdd, 0xdd, 0x62, 0xee, 0x03, 0xf9, 0x61, 0xfc, 0x56, 0x66,
	0xff, 0x5c, 0xaf, 0x07, 0x17, 0x7d, 0xc0, 0xa7, 0xce, 0xfc, 0xe6, 0x37, 0xbf, 0x9d, 0x99, 0xed,
	0x2c, 0x44, 0x55, 0x51, 0x5f, 0xe5, 0xe5, 0x5e, 0x25, 0xb8, 0xe2, 0xa8, 0x6b, 0x3e, 0xc9, 0x57,
	0xf0, 0x09, 0xfb, 0x51, 0x33, 0xa9, 0xd0, 0x0e, 0x44, 0xc2, 0x9a, 0xe7, 0x75, 0x9d, 0x53, 0xec,
	0xc5, 0x5e, 0x1a, 0x90, 0xd0, 0x61, 0x9f, 0xeb, 0x9c, 0xa2, 0x27, 0xe0, 0x3b, 0x17, 0xaf, 0xc4,
	0x5e, 0x1a, 0xee, 0x23, 0xab, 0xb6, 0x77, 0xac, 0x54, 0xe5, 0x74, 0xc8, 0x8c, 0x92, 0xfc, 0xf6,
	0xa0, 0x4f, 0x98, 0xac, 0x78, 0x29, 0x59, 0x3b, 0xd5, 0xfb, 0x67, 0x2a, 0x7a, 0x0a, 0x7d, 0xe1,
	0x32, 0xdd, 0x49, 0x9b, 0x0b, 0x74, 0x1b, 0x22, 0x0d, 0x09, 0x3d, 0x84, 0x40, 0x2a, 0x5e, 0x1d,
	0x8e, 0xb3, 0xbc, 0xc4, 0x9d, 0xd8, 0x4b, 0xfb, 0x64, 0x0e, 0xa0, 0x47, 0xb0, 0x26, 0x58, 0x49,
	0x99, 0x38, 0xe4, 0xa5, 0x62, 0xa5, 0xc2, 0xab, 0x86, 0xb1, 0x08, 0xa2, 0xa1, 0x3e, 0x94, 0xe6,
	0x82, 0x8d, 0x14, 0xee, 0x1a, 0x42, 0xe3, 0x27, 0xbf, 0x7c, 0x08, 0x5b, 0x95, 0xa2, 0x07, 0xd0,
	0x9b, 0x30, 0x35, 0xe6, 0xb3, 0x31, 0x39, 0x0f, 0x0d, 0xa0, 0x53, 0x8b, 0xc2, 0xd4, 0x1c, 0x10,
	0x6d, 0xa2, 0xfb, 0x60, 0x47, 0x6d, 0xaa, 0x0a, 0x88, 0x75, 0xd0, 0x16, 0x80, 0x31, 0x4e, 0xb3,
	0x6f, 0x5c, 0x98, 0x72, 0xba, 0xa4, 0x85, 0xcc, 0xe3, 0x79, 0xc9, 0x85, 0xa9, 0xa6, 0x89, 0x6b,
	0x04, 0x3d, 0x87, 0xde, 0x98, 0x65, 0x94, 0x09, 0xdc, 0x8b, 0x3b, 0x69, 0xb8, 0xbf, 0x75, 0x7b,
	0x9a, 0x7b, 0xc7, 0x86, 0x70, 0x54, 0x2a, 0x31, 0x25, 0x8e, 0x8d, 0x10, 0xac, 0x5e, 0x70, 0x3a,
	0xc5, 0x7e, 0xec, 0xa5, 0x11, 0x31, 0xb6, 0x9e, 0xce, 0xc8, 0x8e, 0xe0, 0x84, 0x95, 0x57, 0x6a,
	0x8c, 0xfb, 0xb1, 0x97, 0x76, 0xc8, 0x22, 0x88, 0x1e, 0xc3, 0x40, 0x89, 0xac, 0x94, 0x97, 0x5a,
	0x72, 0xc4, 0x69, 0x5e, 0x5e, 0xe1, 0x20, 0xee, 0xa4, 0x01, 0xb9, 0x85, 0xeb, 0x9e, 0x47, 0x05,
	0x97, 0x0c, 0x83, 0x19, 0xa3, 0x75, 0xf4, 0xd9, 0x63, 0x2e, 0x15, 0x0e, 0xcd, 0x20, 0x8c, 0x8d,
	0xde, 0x00, 0x5c, 0x72, 0x31, 0xf9, 0x92, 0x15, 0x35, 0x93, 0x38, 0x32, 0xbd, 0x24, 0x4b, 0x7a,
	0x79, 0xdb, 0x90, 0x6c, 0x3f, 0xad, 0x2c, 0xf4, 0x1e, 0xd6, 0x2b, 0x2e, 0xd5, 0x9c, 0x82, 0xd7,
	0x8c, 0xce, 0xee, 0x12, 0x9d, 0x8f, 0x0b, 0x44, 0xab, 0x75, 0x23, 0x1b, 0xbd, 0x02, 0x5f, 0x89,
	0x2c, 0x2f, 0x98, 0xc0, 0x1b, 0x46, 0x68, 0x7b, 0x89, 0xd0, 0x27, 0xcb, 0xb0, 0x0a, 0x33, 0xbe,
	0xbe, 0x36, 0xc1, 0x26, 0x5c, 0xb1, 0xd7, 0x94, 0x0a, 0x3c, 0x30, 0x8d, 0xb6, 0x10, 0x1b, 0xb7,
	0xfb, 0x24, 0x72, 0x7c, 0x6f, 0x16, 0x9f, 0x21, 0xc3, 0x77, 0x10, 0xb6, 0x6e, 0x4d, 0xff, 0x4d,
	0xdf, 0xd9, 0xd4, 0xfd, 0x62, 0xda, 0x44, 0xbb, 0xd0, 0xbd, 0xd6, 0x55, 0xba, 0xad, 0x18, 0xb8,
	0xca, 0x4c, 0xe5, 0x27, 0xb9, 0x54, 0xc4, 0x86, 0x0f, 0x56, 0x5e, 0x7a, 0xc3, 0x0f, 0xb0, 0x71,
	0xa3, 0xd5, 0x3b, 0x0a, 0x9e, 0xc1, 0xe6, 0x92, 0xf9, 0xdd, 0x51, 0xf4, 0x04, 0xa2, 0xf6, 0x2c,
	0xef, 0xa6, 0xa6, 0xdf, 0x9c, 0xa8, 0xfd, 0x44, 0xa0, 0x6d, 0x08, 0xa5, 0xca, 0x54, 0x2d, 0xcf,
	0x47, 0x9c, 0x32, 0x23, 0xdb, 0x25, 0x60, 0xa1, 0x43, 0x4e, 0x19, 0x7a, 0xd1, 0x6c, 0xd2, 0xca,
	0x92, 0xcb, 0xb6, 0x2a, 0x7f, 0x5d, 0xa5, 0xce, 0x7c, 0x95, 0xfe, 0xeb, 0xfd, 0x25, 0x3e, 0x74,
	0x8f, 0x26, 0x95, 0x9a, 0x26, 0x3b, 0x10, 0x34, 0x04, 0xbd, 0x5b, 0x56, 0xc1, 0x33, 0xcb, 0x67,
	0x9d, 0xfd, 0x03, 0x80, 0xd3, 0x9c, 0xd2, 0x82, 0xfd, 0xcc, 0x84, 0x7e, 0x6c, 0x83, 0x33, 0x26,
	0xae, 0x99, 0xee, 0x01, 0xad, 0xbb, 0x33, 0xdc, 0x9f, 0x3b, 0xdc, 0x68, 0x7c, 0xdb, 0xdc, 0x45,
	0xcf, 0xf8, 0xcf, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0x59, 0x85, 0x83, 0xf2, 0x21, 0x06, 0x00,
	0x00,
}
