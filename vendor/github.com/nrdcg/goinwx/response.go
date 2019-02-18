package goinwx

import "fmt"

// Response is a INWX API response. This wraps the standard http.Response returned from INWX.
type Response struct {
	Code         int                    `xmlrpc:"code"`
	Message      string                 `xmlrpc:"msg"`
	ReasonCode   string                 `xmlrpc:"reasonCode"`
	Reason       string                 `xmlrpc:"reason"`
	ResponseData map[string]interface{} `xmlrpc:"resData"`
}

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	Code       int    `xmlrpc:"code"`
	Message    string `xmlrpc:"msg"`
	ReasonCode string `xmlrpc:"reasonCode"`
	Reason     string `xmlrpc:"reason"`
}

func (r *ErrorResponse) Error() string {
	if r.Reason != "" {
		return fmt.Sprintf("(%d) %s. Reason: (%s) %s",
			r.Code, r.Message, r.ReasonCode, r.Reason)
	}
	return fmt.Sprintf("(%d) %s", r.Code, r.Message)
}
