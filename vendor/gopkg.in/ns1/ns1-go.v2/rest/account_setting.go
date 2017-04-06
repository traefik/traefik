package rest

import (
	"net/http"

	"gopkg.in/ns1/ns1-go.v2/rest/model/account"
)

// SettingsService handles 'account/settings' endpoint.
type SettingsService service

// Get returns the basic contact details associated with the account.
//
// NS1 API docs: https://ns1.com/api/#settings-get
func (s *SettingsService) Get() (*account.Setting, *http.Response, error) {
	req, err := s.client.NewRequest("GET", "account/settings", nil)
	if err != nil {
		return nil, nil, err
	}

	var us account.Setting
	resp, err := s.client.Do(req, &us)
	if err != nil {
		return nil, resp, err
	}

	return &us, resp, nil
}

// Update changes most of the basic contact details, except customerid.
//
// NS1 API docs: https://ns1.com/api/#settings-post
func (s *SettingsService) Update(us *account.Setting) (*http.Response, error) {
	req, err := s.client.NewRequest("POST", "account/settings", &us)
	if err != nil {
		return nil, err
	}

	// Update usagewarnings fields with data from api(ensure consistent)
	resp, err := s.client.Do(req, &us)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
