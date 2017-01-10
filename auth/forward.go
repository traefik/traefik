package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/stretchr/stew/objects"
)

// Forward the authentication to a external server
func Forward(forward *types.Forward, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	client := http.Client{}

	forwardReq, err := http.NewRequest("GET", forward.Address, nil)
	if err != nil {
		log.Debugf("Error calling %s. Cause %s", forward.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if forward.ForwardAllHeaders {
		forwardReq.Header = r.Header
	}

	forwardReq.Header.Add("Accept", "application/json")
	forwardQuery := forwardReq.URL.Query()
	rQuery := r.URL.Query()
	for _, reqParam := range forward.RequestParameters {
		switch reqParam.In {
		case "parameter", "":
			paramValue := rQuery.Get(reqParam.Name)
			forwardQuery.Add(reqParam.As, paramValue)
		case "header":
			headerValues := r.Header.Get(reqParam.Name)
			forwardReq.Header.Add(reqParam.As, headerValues)
		}
	}
	forwardReq.URL.RawQuery = forwardQuery.Encode()
	forwardResponse, forwardErr := client.Do(forwardReq)
	if forwardErr != nil {
		log.Debugf("Error calling %s. Cause: %s", forward.Address, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer forwardResponse.Body.Close()
	body, readError := ioutil.ReadAll(forwardResponse.Body)
	if readError != nil {
		log.Debugf("Error reading body %s. Cause: %s", forward.Address, readError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if forwardResponse.StatusCode != 200 {
		log.Debugf("Remote error %s. StatusCode: %s", forward.Address, forwardResponse.StatusCode)
		w.WriteHeader(forwardResponse.StatusCode)
		w.Write(body)
		return
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(body, &data); err != nil {
		log.Debugf("Error Auth failed %s", err)
		return
	}

	objects := objects.Map(data)
	for _, replay := range forward.ResponseReplayFields {
		object := objects.Get(replay.Path)
		var value string
		switch v := object.(type) {
		case string:
			value = v
		default:
			byteValue, err := json.Marshal(object)
			if err != nil {
				log.Debugf("Error failed to Marshal object: %v. Cause: %s", object, err)
				return
			}
			value = string(byteValue)
		}

		switch replay.In {
		case "parameter":
			rQuery.Add(replay.As, value)

		case "header", "":
			r.Header.Add(replay.As, value)
		}

	}

	r.URL.RawQuery = rQuery.Encode()
	r.RequestURI = r.URL.RequestURI()
	next(w, r)
}
