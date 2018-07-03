package vegadns2client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Token - struct to hold token information
type Token struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	ExpiresAt time.Time
}

func (t Token) valid() error {
	if time.Now().UTC().After(t.ExpiresAt) {
		return errors.New("Token Expired")
	}
	return nil
}

func (vega *VegaDNSClient) getBearer() string {
	if vega.token.valid() != nil {
		vega.getAuthToken()
	}
	return vega.token.formatBearer()
}

func (t Token) formatBearer() string {
	return fmt.Sprintf("Bearer %s", t.Token)
}

func (vega *VegaDNSClient) getAuthToken() {
	tokenEndpoint := vega.getURL("token")
	v := url.Values{}
	v.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(v.Encode()))
	if err != nil {
		log.Fatalf("Error forming POST to getAuthToken: %s", err)
	}
	req.SetBasicAuth(vega.APIKey, vega.APISecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	issueTime := time.Now().UTC()
	resp, err := vega.client.Do(req)
	if err != nil {
		log.Fatalf("Error sending POST to getAuthToken: %s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response from POST to getAuthToken: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Got bad answer from VegaDNS on getAuthToken. Code: %d. Message: %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, &vega.token); err != nil {
		log.Fatalf("Error unmarshalling body of POST to getAuthToken: %s", err)
	}

	if vega.token.TokenType != "bearer" {
		log.Fatal("We don't support anything except bearer tokens")
	}
	vega.token.ExpiresAt = issueTime.Add(time.Duration(vega.token.ExpiresIn) * time.Second)
}
