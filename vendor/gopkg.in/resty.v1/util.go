// Copyright (c) 2015-2019 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Helper methods
//___________________________________

// IsStringEmpty method tells whether given string is empty or not
func IsStringEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// DetectContentType method is used to figure out `Request.Body` content type for request header
func DetectContentType(body interface{}) string {
	contentType := plainTextType
	kind := kindOf(body)
	switch kind {
	case reflect.Struct, reflect.Map:
		contentType = jsonContentType
	case reflect.String:
		contentType = plainTextType
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = jsonContentType
		}
	}

	return contentType
}

// IsJSONType method is to check JSON content type or not
func IsJSONType(ct string) bool {
	return jsonCheck.MatchString(ct)
}

// IsXMLType method is to check XML content type or not
func IsXMLType(ct string) bool {
	return xmlCheck.MatchString(ct)
}

// Unmarshal content into object from JSON or XML
// Deprecated: kept for backward compatibility
func Unmarshal(ct string, b []byte, d interface{}) (err error) {
	if IsJSONType(ct) {
		err = json.Unmarshal(b, d)
	} else if IsXMLType(ct) {
		err = xml.Unmarshal(b, d)
	}

	return
}

// Unmarshalc content into object from JSON or XML
func Unmarshalc(c *Client, ct string, b []byte, d interface{}) (err error) {
	if IsJSONType(ct) {
		err = c.JSONUnmarshal(b, d)
	} else if IsXMLType(ct) {
		err = xml.Unmarshal(b, d)
	}

	return
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// RequestLog and ResponseLog type
//___________________________________

// RequestLog struct is used to collected information from resty request
// instance for debug logging. It sent to request log callback before resty
// actually logs the information.
type RequestLog struct {
	Header http.Header
	Body   string
}

// ResponseLog struct is used to collected information from resty response
// instance for debug logging. It sent to response log callback before resty
// actually logs the information.
type ResponseLog struct {
	Header http.Header
	Body   string
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Package Unexported methods
//___________________________________

// way to disable the HTML escape as opt-in
func jsonMarshal(c *Client, r *Request, d interface{}) ([]byte, error) {
	if !r.jsonEscapeHTML {
		return noescapeJSONMarshal(d)
	} else if !c.jsonEscapeHTML {
		return noescapeJSONMarshal(d)
	}
	return c.JSONMarshal(d)
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if !IsStringEmpty(s) {
			return s
		}
	}
	return ""
}

func getLogger(w io.Writer) *log.Logger {
	return log.New(w, "RESTY ", log.LstdFlags)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func createMultipartHeader(param, fileName, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		escapeQuotes(param), escapeQuotes(fileName)))
	hdr.Set("Content-Type", contentType)
	return hdr
}

func addMultipartFormField(w *multipart.Writer, mf *MultipartField) error {
	partWriter, err := w.CreatePart(createMultipartHeader(mf.Param, mf.FileName, mf.ContentType))
	if err != nil {
		return err
	}

	_, err = io.Copy(partWriter, mf.Reader)
	return err
}

func writeMultipartFormFile(w *multipart.Writer, fieldName, fileName string, r io.Reader) error {
	// Auto detect actual multipart content type
	cbuf := make([]byte, 512)
	size, err := r.Read(cbuf)
	if err != nil {
		return err
	}

	partWriter, err := w.CreatePart(createMultipartHeader(fieldName, fileName, http.DetectContentType(cbuf)))
	if err != nil {
		return err
	}

	if _, err = partWriter.Write(cbuf[:size]); err != nil {
		return err
	}

	_, err = io.Copy(partWriter, r)
	return err
}

func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer closeq(file)
	return writeMultipartFormFile(w, fieldName, filepath.Base(path), file)
}

func addFileReader(w *multipart.Writer, f *File) error {
	return writeMultipartFormFile(w, f.ParamName, f.Name, f.Reader)
}

func getPointer(v interface{}) interface{} {
	vv := valueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}

func isPayloadSupported(m string, allowMethodGet bool) bool {
	return !(m == MethodHead || m == MethodOptions || (m == MethodGet && !allowMethodGet))
}

func typeOf(i interface{}) reflect.Type {
	return indirect(valueOf(i)).Type()
}

func valueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}

func indirect(v reflect.Value) reflect.Value {
	return reflect.Indirect(v)
}

func kindOf(v interface{}) reflect.Kind {
	return typeOf(v).Kind()
}

func createDirectory(dir string) (err error) {
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); err != nil {
				return
			}
		}
	}
	return
}

func canJSONMarshal(contentType string, kind reflect.Kind) bool {
	return IsJSONType(contentType) && (kind == reflect.Struct || kind == reflect.Map || kind == reflect.Slice)
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func acquireBuffer() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		bufPool.Put(buf)
	}
}

func closeq(v interface{}) {
	if c, ok := v.(io.Closer); ok {
		sliently(c.Close())
	}
}

func sliently(_ ...interface{}) {}

func composeHeaders(hdrs http.Header) string {
	var str []string
	for _, k := range sortHeaderKeys(hdrs) {
		str = append(str, fmt.Sprintf("%25s: %s", k, strings.Join(hdrs[k], ", ")))
	}
	return strings.Join(str, "\n")
}

func sortHeaderKeys(hdrs http.Header) []string {
	var keys []string
	for key := range hdrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func copyHeaders(hdrs http.Header) http.Header {
	nh := http.Header{}
	for k, v := range hdrs {
		nh[k] = v
	}
	return nh
}
