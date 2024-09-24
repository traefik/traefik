package collector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/redactor"
	"github.com/traefik/traefik/v3/pkg/version"
)

// collectorURL URL where the stats are sent.
const collectorURL = "https://collect.traefik.io/yYaUej3P42cziRVzv6T5w2aYy9po2Mrn"

// Collected data.
type data struct {
	Version       string `json:"version"`
	Codename      string `json:"codename"`
	BuildDate     string `json:"buildDate"`
	Configuration string `json:"configuration"`
	Hash          string `json:"hash"`
}

// Collect anonymous data.
func Collect(staticConfiguration *static.Configuration) error {
	buf, err := createBody(staticConfiguration)
	if err != nil {
		return err
	}

	resp, err := makeHTTPClient().Post(collectorURL, "application/json; charset=utf-8", buf)
	if resp != nil {
		_ = resp.Body.Close()
	}

	return err
}

func createBody(staticConfiguration *static.Configuration) (*bytes.Buffer, error) {
	anonConfig, err := redactor.Anonymize(staticConfiguration)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Anonymous stats sent to %s: %s", collectorURL, anonConfig)

	hashConf, err := hashstructure.Hash(staticConfiguration, nil)
	if err != nil {
		return nil, err
	}

	data := &data{
		Version:       version.Version,
		Codename:      version.Codename,
		BuildDate:     version.BuildDate,
		Hash:          strconv.FormatUint(hashConf, 10),
		Configuration: base64.StdEncoding.EncodeToString([]byte(anonConfig)),
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, err
	}

	return buf, err
}

func makeHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{Transport: transport}
}
