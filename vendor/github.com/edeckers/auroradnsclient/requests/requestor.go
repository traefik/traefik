package requests

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	request_errors "github.com/edeckers/auroradnsclient/requests/errors"
	"github.com/edeckers/auroradnsclient/tokens"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

// AuroraRequestor performs actual requests to API
type AuroraRequestor struct {
	endpoint string
	userID   string
	key      string
}

// NewAuroraRequestor instantiates a new requestor
func NewAuroraRequestor(endpoint string, userID string, key string) (*AuroraRequestor, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("Aurora endpoint missing")
	}

	if userID == "" || key == "" {
		return nil, fmt.Errorf("Aurora credentials missing")
	}

	return &AuroraRequestor{endpoint: endpoint, userID: userID, key: key}, nil
}

func (requestor *AuroraRequestor) buildRequest(relativeURL string, method string, body []byte) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", requestor.endpoint, relativeURL)

	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		logrus.Errorf("Failed to build request: %s", err)

		return request, err
	}

	timestamp := time.Now().UTC()
	fmtTime := timestamp.Format("20060102T150405Z")

	token := tokens.NewToken(requestor.userID, requestor.key, method, fmt.Sprintf("/%s", relativeURL), timestamp)

	request.Header.Set("X-AuroraDNS-Date", fmtTime)
	request.Header.Set("Authorization", fmt.Sprintf("AuroraDNSv1 %s", token))

	request.Header.Set("Content-Type", "application/json")

	rawRequest, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		logrus.Errorf("Failed to dump request: %s", err)
	}

	logrus.Debugf("Built request:\n%s", rawRequest)

	return request, err
}

func (requestor *AuroraRequestor) testInvalidResponse(resp *http.Response, response []byte) ([]byte, error) {
	if resp.StatusCode < 400 {
		return response, nil
	}

	logrus.Errorf("Received invalid status code %d:\n%s", resp.StatusCode, response)

	content := errors.New(string(response))

	statusCodeErrorMap := map[int]error{
		400: request_errors.BadRequest(content),
		401: request_errors.Unauthorized(content),
		403: request_errors.Forbidden(content),
		404: request_errors.NotFound(content),
		500: request_errors.ServerError(content),
	}

	mappedError := statusCodeErrorMap[resp.StatusCode]

	if mappedError == nil {
		return nil, request_errors.InvalidStatusCodeError(content)
	}

	return nil, mappedError
}

// Request builds and executues a request to the API
func (requestor *AuroraRequestor) Request(relativeURL string, method string, body []byte) ([]byte, error) {
	req, err := requestor.buildRequest(relativeURL, method, body)

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Failed request: %s", err)
		return nil, err
	}

	defer resp.Body.Close()

	rawResponse, err := httputil.DumpResponse(resp, true)
	logrus.Debugf("Received raw response:\n%s", rawResponse)
	if err != nil {
		logrus.Errorf("Failed to dump response: %s", err)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read response: %s", response)
		return nil, err
	}

	response, err = requestor.testInvalidResponse(resp, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
