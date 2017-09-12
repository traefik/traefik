package auth

import (
	"io/ioutil"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
)

// Forward the authentication to a external server
func Forward(forward *types.Forward, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	// Ensure our request client does not follow redirects

	httpClient := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if forward.TLS != nil {
		tlsConfig, err := forward.TLS.CreateTLSConfig()
		if err != nil {
			log.Debugf("Impossible to configure TLS to call %s. Cause %s", forward.Address, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}

	}

	forwardReq, err := http.NewRequest(http.MethodGet, forward.Address, nil)
	if err != nil {
		log.Debugf("Error calling %s. Cause %s", forward.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	forwardReq.Header = r.Header

	forwardResponse, forwardErr := httpClient.Do(forwardReq)
	if forwardErr != nil {
		log.Debugf("Error calling %s. Cause: %s", forward.Address, forwardErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, readError := ioutil.ReadAll(forwardResponse.Body)
	if readError != nil {
		log.Debugf("Error reading body %s. Cause: %s", forward.Address, readError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer forwardResponse.Body.Close()

	// Pass the forward response's body and selected headers if it
	// didn't return a response within the range of [200, 300).

	if forwardResponse.StatusCode < http.StatusOK || forwardResponse.StatusCode >= http.StatusMultipleChoices {
		log.Debugf("Remote error %s. StatusCode: %d", forward.Address, forwardResponse.StatusCode)

		// Grab the location header, if any.

		redirectURL, err := forwardResponse.Location()

		if err != nil && err != http.ErrNoLocation {
			log.Debugf("Error reading response location header %s. Cause: %s", forward.Address, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set the location in our response if one was sent back.

		if err != http.ErrNoLocation && redirectURL.String() != "" {
			w.Header().Add("Location", redirectURL.String())
		}

		// Pass any Set-Cookie headers the forward auth server provides

		cookies := forwardResponse.Cookies()

		for _, cookie := range cookies {
			w.Header().Add("Set-Cookie", cookie.String())
		}

		w.WriteHeader(forwardResponse.StatusCode)
		w.Write(body)
		return
	}

	r.RequestURI = r.URL.RequestURI()
	next(w, r)
}
