package plugin

import (
	//"io/ioutil"
	"net/http"
	"net/rpc"
	"os/exec"

	"google.golang.org/grpc"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/plugin/proto"
	"github.com/hashicorp/go-plugin"
	"github.com/satori/go.uuid"
	"github.com/vulcand/oxy/forward"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

// RemoteHandshake is a common handshake that is shared by plugin and host.
var RemoteHandshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TRAEFIK_PLUGIN",
	MagicCookieValue: "traefik",
}

// RemotePluginMap is the map of plugins we can dispense.
var RemotePluginMap = map[string]plugin.Plugin{
	"middleware": &RemotePlugin{},
}

// RemotePluginMiddleware is the interface that we're exposing as a plugin.
type RemotePluginMiddleware interface {
	ServeHttp(req *proto.Request) (*proto.Response, error)
}

var _ plugin.Plugin = (*RemotePlugin)(nil)
var _ plugin.GRPCPlugin = (*RemotePlugin)(nil)

// This is the implementation of plugin.Plugin so we can serve/consume this.
// We also implement GRPCPlugin so that this plugin can be served over
// gRPC.
type RemotePlugin struct {
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl RemotePluginMiddleware
}

func (p *RemotePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (*RemotePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

func (p *RemotePlugin) GRPCServer(s *grpc.Server) error {
	proto.RegisterMiddlewareServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *RemotePlugin) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewMiddlewareClient(c)}, nil
}

// RemotePluginMiddleware
type RemotePluginMiddlewareHandler struct {
	client   *plugin.Client
	remote   RemotePluginMiddleware
	registry metrics.Registry
	plugin   Plugin
}

// NewRemotePluginMiddleware creates a new PluginMiddleware instance.
func NewRemotePluginMiddleware(p Plugin, registry metrics.Registry) *RemotePluginMiddlewareHandler {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  RemoteHandshake,
		Plugins:          RemotePluginMap,
		Cmd:              exec.Command("sh", "-c", p.Path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		// TODO: Add Logger Wrapper Logger:           log,
	})

	rpcClient, err := client.Client()
	if err != nil {
		log.Error("Unable to allocate plugin client")
	}
	raw, err := rpcClient.Dispense("middleware")
	if err != nil {
		log.Error("Unable to invoke plugin")
	}
	remote := raw.(RemotePluginMiddleware)

	return &RemotePluginMiddlewareHandler{
		client:   client,
		remote:   remote,
		registry: registry,
		plugin:   p,
	}
}

func (h *RemotePluginMiddlewareHandler) Stop() {
	log.Debug("Stopping Plugins")
	h.client.Kill()
}

// ServeHTTP delegates to a plugin subprocess, if plugin order is `before` or `around` and then
// invokes the next handler in the middleware chain, if no result rendered, then delegates to a plugin subprocess again, if plugin order is `around` or `after`.
func (h *RemotePluginMiddlewareHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	stopChain := false
	guid := uuid.NewV4().String()
	if h.plugin.Before() {
		stopChain = h.executeRemotePlugin(rw, r, guid, true)
	}
	if !stopChain {
		log.Debug("Executing next handler from plugin middleware")
		next.ServeHTTP(rw, r)
	}
	if h.plugin.After() {
		h.executeRemotePlugin(rw, r, guid, false)
	}
}

// executeRemotePlugin processes the remote plugin response and returns `false` if "next" middleware in the chain should be executed, otherwise returns `true`
func (h *RemotePluginMiddlewareHandler) executeRemotePlugin(rw http.ResponseWriter, r *http.Request, guid string, before bool) bool {
	if h.client != nil {
		start := time.Now()
		//log.Debugf("Calling Remote Plugin %+v for Request %+v", remote, r)
		//log.Debug("Request Method: " + r.Method)
		pluginRequest := h.createPluginRequest(rw, r, guid)
		log.Debugf("Plugin Request: %+v", pluginRequest)
		resp, err := h.remote.ServeHttp(pluginRequest)
		//resp, err := remote.ServeHttp(&proto.Request{
		//	Request: &proto.HttpRequest{
		//		Body: []byte("test"),
		//	},
		//})
		if h.registry.IsEnabled() {
			pluginDurationLabels := []string{"plugin", filepath.Base(h.plugin.Path), "error", strconv.FormatBool(err != nil), "order", h.plugin.Order}
			h.registry.PluginDurationHistogram().With(pluginDurationLabels...).Observe(time.Since(start).Seconds())
		}
		log.Debugf("Got result from Remote Plugin %+v", resp)
		if err != nil {
			// How to handle errors?
			rw.WriteHeader(http.StatusServiceUnavailable)
			rw.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
			return true
		} else {
			return h.handlePluginResponse(resp, rw, r)
		}
	}
	// nothing was done, so proceed with the next middleware chain
	return false
}

func (h *RemotePluginMiddlewareHandler) createPluginRequest(rw http.ResponseWriter, r *http.Request, guid string) *proto.Request {
	var body []byte
	//if r.Body != http.NoBody {
	if r.GetBody != nil {
		log.Debug("Getting Body Reader")
		bodyReader, _ := r.GetBody()
		log.Debug("Created Body Reader")
		// TODO: Should we handle the error?
		body, _ = ioutil.ReadAll(bodyReader)
		log.Debug("Converted Body to byte[]")
		// TODO: Should we handle the error?
	}

	log.Debug("Creating Remote Plugin Proto Request")
	return &proto.Request{
		RequestUuid: guid,
		Request: &proto.HttpRequest{
			Header:           h.valueList(r.Header),
			Close:            r.Close,
			ContentLength:    r.ContentLength,
			Host:             r.Host,
			Method:           r.Method,
			FormValues:       h.valueList(r.Form),
			PostFormValues:   h.valueList(r.PostForm),
			Proto:            r.Proto,
			ProtoMajor:       int32(r.ProtoMajor),
			ProtoMinor:       int32(r.ProtoMinor),
			RemoteAddr:       r.RemoteAddr,
			RequestUri:       r.RequestURI,
			Trailer:          h.valueList(r.Trailer),
			TransferEncoding: r.TransferEncoding,
			Url:              r.URL.String(),
			Body:             body,
		},
	}
}

// handlePluginResponseAndContinue processes the remote plugin response and returns `false` if "next" middleware in the chain should be executed, otherwise returns `true`
func (h *RemotePluginMiddlewareHandler) handlePluginResponse(pResp *proto.Response, rw http.ResponseWriter, r *http.Request) bool {
	h.syncRequest(pResp.Request, r)
	if pResp.Redirect {
		url, err := url.ParseRequestURI(pResp.Request.Url)
		if err == nil {
			r.URL = url
			r.RequestURI = r.URL.RequestURI()
			fwd, err := forward.New()
			if err == nil {
				fwd.ServeHTTP(rw, r)
				log.Debugf("Forwarded plugin response to %s", pResp.Request.Url)
				return true
			} else {
				log.Errorf("Unable to forward request to %s - %+v", pResp.Request.Url, err)
			}
		}
	}
	if pResp.RenderContent && pResp.Response.Body != nil && len(pResp.Response.Body) > 0 {
		body := pResp.Response.Body
		rw.WriteHeader(int(pResp.Response.StatusCode))
		rw.Write(body)
		log.Debug("Rendered plugin response body")
		return true
	}
	log.Debug("Generic plugin response")
	rw.WriteHeader(int(pResp.Response.StatusCode))
	rw.Write([]byte("Response from plugin"))
	return pResp.StopChain
}

func (h *RemotePluginMiddlewareHandler) syncRequest(src *proto.HttpRequest, dest *http.Request) {
	dest.Close = src.Close
	dest.ContentLength = src.ContentLength
	dest.Form = h.mapOfStrings(src.FormValues)
	dest.Header = h.mapOfStrings(src.Header)
	dest.Host = src.Host
	dest.Method = src.Method
	//dest.MultipartForm
	dest.PostForm = h.mapOfStrings(src.PostFormValues)
	dest.Proto = src.Proto
	dest.ProtoMajor = int(src.ProtoMajor)
	dest.ProtoMinor = int(src.ProtoMinor)
	dest.RemoteAddr = src.RemoteAddr
	dest.RequestURI = src.RequestUri
	//dest.TLS
	dest.Trailer = h.mapOfStrings(src.Trailer)
	dest.TransferEncoding = src.TransferEncoding
	url, err := url.ParseRequestURI(src.Url)
	if err == nil {
		dest.URL = url
	} else {
		log.Errorf("Unable to sync request.url field: %s - %+v", src.Url, err)
	}
}

func (h *RemotePluginMiddlewareHandler) mapOfStrings(values map[string]*proto.ValueList) map[string][]string {
	p := make(map[string][]string)

	for k, v := range values {
		p[k] = v.Value
	}
	return p
}

func (h *RemotePluginMiddlewareHandler) valueList(values map[string][]string) map[string]*proto.ValueList {
	p := make(map[string]*proto.ValueList)

	for k, v := range values {
		p[k] = &proto.ValueList{Value: v}
	}
	return p
}
