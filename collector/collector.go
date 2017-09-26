package collector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/server/uuid"
	"github.com/containous/traefik/version"
)

const containousURL = "https://collect.traefik.io"

// Collected data
type data struct {
	Version       string
	Codename      string
	Configuration string
	UUID          string
}

// Collect anonymous and obfuscated data.
func Collect(globalConfiguration *configuration.GlobalConfiguration) error {

	obfuscatedConfig, err := Obfuscate(globalConfiguration, false)

	log.Debugf("Anonymous collected data: %s", obfuscatedConfig)

	data := &data{
		Version:       version.Version,
		Codename:      version.Codename,
		UUID:          uuid.Get(),
		Configuration: base64.StdEncoding.EncodeToString([]byte(obfuscatedConfig)),
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Post(containousURL, "application/json", bytes.NewBuffer(dataJSON))
	if err != nil {
		return err
	}

	return nil
}
