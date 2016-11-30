package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/stretchr/stew/objects"
)

// Forward the authentication to a external server
func Forward(forward *types.Forward, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	client := http.Client{}

	forwardReq, _ := http.NewRequest("GET", forward.Address, nil)
	forwardReq.Header.Add("Accept", "application/json")
	forwardQuery := forwardReq.URL.Query()
	rQuery := r.URL.Query()
	for _, reqParam := range forward.RequestParameters {
		paramValue := rQuery.Get(reqParam.Name)
		forwardQuery.Add(reqParam.As, paramValue)
	}
	forwardReq.URL.RawQuery = forwardQuery.Encode()
	forwardResponse, forwardErr := client.Do(forwardReq)
	if forwardErr != nil || forwardResponse.StatusCode != 200 {
		log.Debugf("Auth failed...")
		return
	}

	data := make(map[string]interface{})

	defer forwardResponse.Body.Close()
	if err := json.NewDecoder(forwardResponse.Body).Decode(&data); err != nil {
		log.Debugf("Auth failed...")
		return
	}

	objects := objects.Map(data)
	for _, replay := range forward.ResponseReplayFields {
		object := objects.Get(replay.Path)
		value := fmt.Sprintf("%v", object)
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
