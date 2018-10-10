// Copyright (c) 2015-2018 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/url"
	"reflect"
	"strings"
)

// SRVRecord holds the data to query the SRV record for the following service
type SRVRecord struct {
	Service string
	Domain  string
}

// SetHeader method is to set a single header field and its value in the current request.
// Example: To set `Content-Type` and `Accept` as `application/json`.
// 		resty.R().
//			SetHeader("Content-Type", "application/json").
//			SetHeader("Accept", "application/json")
//
// Also you can override header value, which was set at client instance level.
//
func (r *Request) SetHeader(header, value string) *Request {
	r.Header.Set(header, value)
	return r
}

// SetHeaders method sets multiple headers field and its values at one go in the current request.
// Example: To set `Content-Type` and `Accept` as `application/json`
//
// 		resty.R().
//			SetHeaders(map[string]string{
//				"Content-Type": "application/json",
//				"Accept": "application/json",
//			})
// Also you can override header value, which was set at client instance level.
//
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for h, v := range headers {
		r.SetHeader(h, v)
	}

	return r
}

// SetQueryParam method sets single parameter and its value in the current request.
// It will be formed as query string for the request.
// Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
// 		resty.R().
//			SetQueryParam("search", "kitchen papers").
//			SetQueryParam("size", "large")
// Also you can override query params value, which was set at client instance level
//
func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParam.Set(param, value)
	return r
}

// SetQueryParams method sets multiple parameters and its values at one go in the current request.
// It will be formed as query string for the request.
// Example: `search=kitchen%20papers&size=large` in the URL after `?` mark.
// 		resty.R().
//			SetQueryParams(map[string]string{
//				"search": "kitchen papers",
//				"size": "large",
//			})
// Also you can override query params value, which was set at client instance level
//
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for p, v := range params {
		r.SetQueryParam(p, v)
	}

	return r
}

// SetMultiValueQueryParams method appends multiple parameters with multi-value
// at one go in the current request. It will be formed as query string for the request.
// Example: `status=pending&status=approved&status=open` in the URL after `?` mark.
// 		resty.R().
//			SetMultiValueQueryParams(url.Values{
//				"status": []string{"pending", "approved", "open"},
//			})
// Also you can override query params value, which was set at client instance level
//
func (r *Request) SetMultiValueQueryParams(params url.Values) *Request {
	for p, v := range params {
		for _, pv := range v {
			r.QueryParam.Add(p, pv)
		}
	}

	return r
}

// SetQueryString method provides ability to use string as an input to set URL query string for the request.
//
// Using String as an input
// 		resty.R().
//			SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more")
//
func (r *Request) SetQueryString(query string) *Request {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		for p, v := range params {
			for _, pv := range v {
				r.QueryParam.Add(p, pv)
			}
		}
	} else {
		r.client.Log.Printf("ERROR %v", err)
	}
	return r
}

// SetFormData method sets Form parameters and their values in the current request.
// It's applicable only HTTP method `POST` and `PUT` and requests content type would be set as
// `application/x-www-form-urlencoded`.
// 		resty.R().
// 			SetFormData(map[string]string{
//				"access_token": "BC594900-518B-4F7E-AC75-BD37F019E08F",
//				"user_id": "3455454545",
//			})
// Also you can override form data value, which was set at client instance level
//
func (r *Request) SetFormData(data map[string]string) *Request {
	for k, v := range data {
		r.FormData.Set(k, v)
	}

	return r
}

// SetMultiValueFormData method appends multiple form parameters with multi-value
// at one go in the current request.
// 		resty.R().
//			SetMultiValueFormData(url.Values{
//				"search_criteria": []string{"book", "glass", "pencil"},
//			})
// Also you can override form data value, which was set at client instance level
//
func (r *Request) SetMultiValueFormData(params url.Values) *Request {
	for k, v := range params {
		for _, kv := range v {
			r.FormData.Add(k, kv)
		}
	}

	return r
}

// SetBody method sets the request body for the request. It supports various realtime need easy.
// We can say its quite handy or powerful. Supported request body data types is `string`, `[]byte`,
// `struct` and `map`. Body value can be pointer or non-pointer. Automatic marshalling
// for JSON and XML content type, if it is `struct` or `map`.
//
// Example:
//
// Struct as a body input, based on content type, it will be marshalled.
//		resty.R().
//			SetBody(User{
//				Username: "jeeva@myjeeva.com",
//				Password: "welcome2resty",
//			})
//
// Map as a body input, based on content type, it will be marshalled.
//		resty.R().
//			SetBody(map[string]interface{}{
//				"username": "jeeva@myjeeva.com",
//				"password": "welcome2resty",
//				"address": &Address{
//					Address1: "1111 This is my street",
//					Address2: "Apt 201",
//					City: "My City",
//					State: "My State",
//					ZipCode: 00000,
//				},
//			})
//
// String as a body input. Suitable for any need as a string input.
//		resty.R().
//			SetBody(`{
//				"username": "jeeva@getrightcare.com",
//				"password": "admin"
//			}`)
//
// []byte as a body input. Suitable for raw request such as file upload, serialize & deserialize, etc.
// 		resty.R().
//			SetBody([]byte("This is my raw request, sent as-is"))
//
func (r *Request) SetBody(body interface{}) *Request {
	r.Body = body
	return r
}

// SetResult method is to register the response `Result` object for automatic unmarshalling in the RESTful mode
// if response status code is between 200 and 299 and content type either JSON or XML.
//
// Note: Result object can be pointer or non-pointer.
//		resty.R().SetResult(&AuthToken{})
//		// OR
//		resty.R().SetResult(AuthToken{})
//
// Accessing a result value
//		response.Result().(*AuthToken)
//
func (r *Request) SetResult(res interface{}) *Request {
	r.Result = getPointer(res)
	return r
}

// SetError method is to register the request `Error` object for automatic unmarshalling in the RESTful mode
// if response status code is greater than 399 and content type either JSON or XML.
//
// Note: Error object can be pointer or non-pointer.
// 		resty.R().SetError(&AuthError{})
//		// OR
//		resty.R().SetError(AuthError{})
//
// Accessing a error value
//		response.Error().(*AuthError)
//
func (r *Request) SetError(err interface{}) *Request {
	r.Error = getPointer(err)
	return r
}

// SetFile method is to set single file field name and its path for multipart upload.
//	resty.R().
//		SetFile("my_file", "/Users/jeeva/Gas Bill - Sep.pdf")
//
func (r *Request) SetFile(param, filePath string) *Request {
	r.isMultiPart = true
	r.FormData.Set("@"+param, filePath)

	return r
}

// SetFiles method is to set multiple file field name and its path for multipart upload.
//	resty.R().
//		SetFiles(map[string]string{
//				"my_file1": "/Users/jeeva/Gas Bill - Sep.pdf",
//				"my_file2": "/Users/jeeva/Electricity Bill - Sep.pdf",
//				"my_file3": "/Users/jeeva/Water Bill - Sep.pdf",
//			})
//
func (r *Request) SetFiles(files map[string]string) *Request {
	r.isMultiPart = true

	for f, fp := range files {
		r.FormData.Set("@"+f, fp)
	}

	return r
}

// SetFileReader method is to set single file using io.Reader for multipart upload.
//	resty.R().
//		SetFileReader("profile_img", "my-profile-img.png", bytes.NewReader(profileImgBytes)).
//		SetFileReader("notes", "user-notes.txt", bytes.NewReader(notesBytes))
//
func (r *Request) SetFileReader(param, fileName string, reader io.Reader) *Request {
	r.isMultiPart = true

	r.multipartFiles = append(r.multipartFiles, &File{
		Name:      fileName,
		ParamName: param,
		Reader:    reader,
	})

	return r
}

// SetMultipartField method is to set custom data using io.Reader for multipart upload.
func (r *Request) SetMultipartField(param, fileName, contentType string, reader io.Reader) *Request {
	r.isMultiPart = true

	r.multipartFields = append(r.multipartFields, &multipartField{
		Param:       param,
		FileName:    fileName,
		ContentType: contentType,
		Reader:      reader,
	})

	return r
}

// SetContentLength method sets the HTTP header `Content-Length` value for current request.
// By default go-resty won't set `Content-Length`. Also you have an option to enable for every
// request. See `resty.SetContentLength`
// 		resty.R().SetContentLength(true)
//
func (r *Request) SetContentLength(l bool) *Request {
	r.setContentLength = true

	return r
}

// SetBasicAuth method sets the basic authentication header in the current HTTP request.
// For Header example:
//		Authorization: Basic <base64-encoded-value>
//
// To set the header for username "go-resty" and password "welcome"
// 		resty.R().SetBasicAuth("go-resty", "welcome")
//
// This method overrides the credentials set by method `resty.SetBasicAuth`.
//
func (r *Request) SetBasicAuth(username, password string) *Request {
	r.UserInfo = &User{Username: username, Password: password}
	return r
}

// SetAuthToken method sets bearer auth token header in the current HTTP request. Header example:
// 		Authorization: Bearer <auth-token-value-comes-here>
//
// Example: To set auth token BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F
//
// 		resty.R().SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F")
//
// This method overrides the Auth token set by method `resty.SetAuthToken`.
//
func (r *Request) SetAuthToken(token string) *Request {
	r.Token = token
	return r
}

// SetOutput method sets the output file for current HTTP request. Current HTTP response will be
// saved into given file. It is similar to `curl -o` flag. Absolute path or relative path can be used.
// If is it relative path then output file goes under the output directory, as mentioned
// in the `Client.SetOutputDirectory`.
// 		resty.R().
// 			SetOutput("/Users/jeeva/Downloads/ReplyWithHeader-v5.1-beta.zip").
// 			Get("http://bit.ly/1LouEKr")
//
// Note: In this scenario `Response.Body` might be nil.
func (r *Request) SetOutput(file string) *Request {
	r.outputFile = file
	r.isSaveResponse = true
	return r
}

// SetSRV method sets the details to query the service SRV record and execute the
// request.
// 		resty.R().
//			SetSRV(SRVRecord{"web", "testservice.com"}).
//			Get("/get")
func (r *Request) SetSRV(srv *SRVRecord) *Request {
	r.SRV = srv
	return r
}

// SetDoNotParseResponse method instructs `Resty` not to parse the response body automatically.
// Resty exposes the raw response body as `io.ReadCloser`. Also do not forget to close the body,
// otherwise you might get into connection leaks, no connection reuse.
//
// Please Note: Response middlewares are not applicable, if you use this option. Basically you have
// taken over the control of response parsing from `Resty`.
func (r *Request) SetDoNotParseResponse(parse bool) *Request {
	r.notParseResponse = parse
	return r
}

// SetPathParams method sets multiple URL path key-value pairs at one go in the
// resty current request instance.
// 		resty.R().SetPathParams(map[string]string{
// 		   "userId": "sample@sample.com",
// 		   "subAccountId": "100002",
// 		})
//
// 		Result:
// 		   URL - /v1/users/{userId}/{subAccountId}/details
// 		   Composed URL - /v1/users/sample@sample.com/100002/details
// It replace the value of the key while composing request URL. Also you can
// override Path Params value, which was set at client instance level.
func (r *Request) SetPathParams(params map[string]string) *Request {
	for p, v := range params {
		r.pathParams[p] = v
	}
	return r
}

// ExpectContentType method allows to provide fallback `Content-Type` for automatic unmarshalling
// when `Content-Type` response header is unavailable.
func (r *Request) ExpectContentType(contentType string) *Request {
	r.fallbackContentType = contentType
	return r
}

// SetJSONEscapeHTML method is to enable/disable the HTML escape on JSON marshal.
//
// NOTE: This option only applicable to standard JSON Marshaller.
func (r *Request) SetJSONEscapeHTML(b bool) *Request {
	r.jsonEscapeHTML = b
	return r
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// HTTP verb method starts here
//___________________________________

// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (r *Request) Get(url string) (*Response, error) {
	return r.Execute(MethodGet, url)
}

// Head method does HEAD HTTP request. It's defined in section 4.3.2 of RFC7231.
func (r *Request) Head(url string) (*Response, error) {
	return r.Execute(MethodHead, url)
}

// Post method does POST HTTP request. It's defined in section 4.3.3 of RFC7231.
func (r *Request) Post(url string) (*Response, error) {
	return r.Execute(MethodPost, url)
}

// Put method does PUT HTTP request. It's defined in section 4.3.4 of RFC7231.
func (r *Request) Put(url string) (*Response, error) {
	return r.Execute(MethodPut, url)
}

// Delete method does DELETE HTTP request. It's defined in section 4.3.5 of RFC7231.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Execute(MethodDelete, url)
}

// Options method does OPTIONS HTTP request. It's defined in section 4.3.7 of RFC7231.
func (r *Request) Options(url string) (*Response, error) {
	return r.Execute(MethodOptions, url)
}

// Patch method does PATCH HTTP request. It's defined in section 2 of RFC5789.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Execute(MethodPatch, url)
}

// Execute method performs the HTTP request with given HTTP method and URL
// for current `Request`.
// 		resp, err := resty.R().Execute(resty.GET, "http://httpbin.org/get")
//
func (r *Request) Execute(method, url string) (*Response, error) {
	var addrs []*net.SRV
	var err error

	if r.isMultiPart && !(method == MethodPost || method == MethodPut) {
		return nil, fmt.Errorf("multipart content is not allowed in HTTP verb [%v]", method)
	}

	if r.SRV != nil {
		_, addrs, err = net.LookupSRV(r.SRV.Service, "tcp", r.SRV.Domain)
		if err != nil {
			return nil, err
		}
	}

	r.Method = method
	r.URL = r.selectAddr(addrs, url, 0)

	if r.client.RetryCount == 0 {
		return r.client.execute(r)
	}

	var resp *Response
	attempt := 0
	_ = Backoff(
		func() (*Response, error) {
			attempt++

			r.URL = r.selectAddr(addrs, url, attempt)

			resp, err = r.client.execute(r)
			if err != nil {
				r.client.Log.Printf("ERROR %v, Attempt %v", err, attempt)
				if r.isContextCancelledIfAvailable() {
					// stop Backoff from retrying request if request has been
					// canceled by context
					return resp, nil
				}
			}

			return resp, err
		},
		Retries(r.client.RetryCount),
		WaitTime(r.client.RetryWaitTime),
		MaxWaitTime(r.client.RetryMaxWaitTime),
		RetryConditions(r.client.RetryConditions),
	)

	return resp, err
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Request Unexported methods
//___________________________________

func (r *Request) fmtBodyString() (body string) {
	body = "***** NO CONTENT *****"
	if isPayloadSupported(r.Method, r.client.AllowGetMethodPayload) {
		if _, ok := r.Body.(io.Reader); ok {
			body = "***** BODY IS io.Reader *****"
			return
		}

		// multipart or form-data
		if r.isMultiPart || r.isFormData {
			body = r.bodyBuf.String()
			return
		}

		// request body data
		if r.Body == nil {
			return
		}
		var prtBodyBytes []byte
		var err error

		contentType := r.Header.Get(hdrContentTypeKey)
		kind := kindOf(r.Body)
		if canJSONMarshal(contentType, kind) {
			prtBodyBytes, err = json.MarshalIndent(&r.Body, "", "   ")
		} else if IsXMLType(contentType) && (kind == reflect.Struct) {
			prtBodyBytes, err = xml.MarshalIndent(&r.Body, "", "   ")
		} else if b, ok := r.Body.(string); ok {
			if IsJSONType(contentType) {
				bodyBytes := []byte(b)
				out := acquireBuffer()
				defer releaseBuffer(out)
				if err = json.Indent(out, bodyBytes, "", "   "); err == nil {
					prtBodyBytes = out.Bytes()
				}
			} else {
				body = b
				return
			}
		} else if b, ok := r.Body.([]byte); ok {
			body = base64.StdEncoding.EncodeToString(b)
		}

		if prtBodyBytes != nil && err == nil {
			body = string(prtBodyBytes)
		}
	}

	return
}

func (r *Request) selectAddr(addrs []*net.SRV, path string, attempt int) string {
	if addrs == nil {
		return path
	}

	idx := attempt % len(addrs)
	domain := strings.TrimRight(addrs[idx].Target, ".")
	path = strings.TrimLeft(path, "/")

	return fmt.Sprintf("%s://%s:%d/%s", r.client.scheme, domain, addrs[idx].Port, path)
}
