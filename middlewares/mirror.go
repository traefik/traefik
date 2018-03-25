package middlewares

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
	"net/url"
)

//Mirror is a middleware that provides the request mirroring capabilities
type Mirror struct {
	backendURL     *url.URL
	requestHeaders map[string]string
	sampleRate     uint64
	mirror         *forward.Forwarder
	counter        uint64
}

//NewMirrorMiddleware initializes the Mirror for the request mirroring
func NewMirrorMiddleware(frontend *types.Frontend, backend *types.Backend) (*Mirror, error) {
	mirror := frontend.Mirror
	// Default sample rate is every request.
	var sampleRate uint64 = 1
	if mirror.SampleRate > 0 {
		sampleRate = uint64(mirror.SampleRate)
	}
	mirrorURL, err := url.Parse(backend.Servers["mirror"].URL)
	if err != nil {
		return nil, err
	}

	forwarder, err := forward.New(
		forward.Stream(true),
		forward.PassHostHeader(frontend.PassHostHeader),
	)
	if err != nil {
		return nil, err
	}
	return &Mirror{backendURL: mirrorURL,
			requestHeaders: mirror.RequestHeaders,
			sampleRate:     sampleRate,
			mirror:         forwarder},
		nil
}

func (m *Mirror) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if m.counter%m.sampleRate == 0 {
		atomic.AddUint64(&m.counter, 1)
		// Copy the request and forward to 'next' and to 'mirror'
		mReq, err := m.copyRequest(req)
		if err != nil {
			log.Errorf("Encountered an error while cloning the request: %v", err)
		}

		// We don't care about the response from the mirror, so we can just launch it in a safe go routine
		// In case of failure, will log the error
		safe.GoWithRecover(func() {
			mRw := createDummyResponseWriter()
			m.mirror.ServeHTTP(mRw, mReq)
		}, func(err interface{}) {
			log.Errorf("Error in request mirroring routine: %s", err)
		})

		next.ServeHTTP(w, req)
	} else {
		// no mirroring
		atomic.AddUint64(&m.counter, 1)
		next.ServeHTTP(w, req)
	}
}

func (m *Mirror) copyRequest(req *http.Request) (*http.Request, error) {
	var body []byte
	o := *req

	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		body = b

		if err != nil {
			return nil, err
		}

		// Create a new reader for the body of the original request
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	// TODO - Is this way better than doing `o := *req` copy?
	//o, err := http.NewRequest(req.Method, utils.CopyURL(req.URL).String(), ioutil.NopCloser(bytes.NewReader(body)))
	//if err != nil {
	//	return nil, err
	//}

	o.ContentLength = int64(len(body))
	// remove TransferEncoding that could have been previously set because we have transformed the request from chunked encoding
	o.TransferEncoding = []string{}
	// http.Transport will close the request body on any error, we are controlling the close process ourselves, so we override the closer here
	if body == nil {
		o.Body = nil
	} else {
		o.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	o.URL = utils.CopyURL(m.backendURL)
	o.Header = make(http.Header)
	utils.CopyHeaders(o.Header, req.Header)

	return &o, nil
}

// dummyResponseWriter captures information from the response and preserves it for
// later analysis.
type dummyResponseWriter struct {
	header      http.Header
	closeNotify chan bool
}

// Header returns empty header
func (rw *dummyResponseWriter) Header() http.Header {

	if rw.header == nil {
		rw.header = make(http.Header)
	}

	return rw.header
}

// WriteHeader does nothing...
func (rw *dummyResponseWriter) WriteHeader(status int) {
}

// Flush does nothing
func (rw *dummyResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (rw *dummyResponseWriter) CloseNotify() <-chan bool {
	return rw.closeNotify
}

func createDummyResponseWriter() http.ResponseWriter {
	return &dummyResponseWriter{closeNotify: make(chan bool, 1)}
}
