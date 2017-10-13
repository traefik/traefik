package auth

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// Forward the authentication to a external server
func Forward(config *types.Forward, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	httpClient := http.Client{}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig()
		if err != nil {
			log.Debugf("Impossible to configure TLS to call %s. Cause %s", config.Address, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	forwardReq, err := http.NewRequest(http.MethodGet, config.Address, nil)
	if err != nil {
		log.Debugf("Error calling %s. Cause %s", config.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHeader(r, forwardReq, config.TrustForwardHeader)

	forwardResponse, forwardErr := httpClient.Do(forwardReq)
	if forwardErr != nil {
		log.Debugf("Error calling %s. Cause: %s", config.Address, forwardErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, readError := ioutil.ReadAll(forwardResponse.Body)
	if readError != nil {
		log.Debugf("Error reading body %s. Cause: %s", config.Address, readError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer forwardResponse.Body.Close()

	if forwardResponse.StatusCode < http.StatusOK || forwardResponse.StatusCode >= http.StatusMultipleChoices {
		log.Debugf("Remote error %s. StatusCode: %d", config.Address, forwardResponse.StatusCode)
		w.WriteHeader(forwardResponse.StatusCode)
		w.Write(body)
		return
	}

	r.RequestURI = r.URL.RequestURI()
	next(w, r)
}

func writeHeader(req *http.Request, forwardReq *http.Request, trustForwardHeader bool) {
	utils.CopyHeaders(forwardReq.Header, req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if trustForwardHeader {
			if prior, ok := req.Header[forward.XForwardedFor]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
		}
		forwardReq.Header.Set(forward.XForwardedFor, clientIP)
	}

	if xfp := req.Header.Get(forward.XForwardedProto); xfp != "" && trustForwardHeader {
		forwardReq.Header.Set(forward.XForwardedProto, xfp)
	} else if req.TLS != nil {
		forwardReq.Header.Set(forward.XForwardedProto, "https")
	} else {
		forwardReq.Header.Set(forward.XForwardedProto, "http")
	}

	if xfp := req.Header.Get(forward.XForwardedPort); xfp != "" && trustForwardHeader {
		forwardReq.Header.Set(forward.XForwardedPort, xfp)
	}

	if xfh := req.Header.Get(forward.XForwardedHost); xfh != "" && trustForwardHeader {
		forwardReq.Header.Set(forward.XForwardedHost, xfh)
	} else if req.Host != "" {
		forwardReq.Header.Set(forward.XForwardedHost, req.Host)
	} else {
		forwardReq.Header.Del(forward.XForwardedHost)
	}
}
