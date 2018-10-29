package doapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/iij/doapi/protocol"
)

func execTemplate(name string, tmplstr string, arg interface{}) string {
	tmpl, err := template.New(name).Parse(tmplstr)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, arg); err != nil {
		panic(err)
	}
	return buf.String()
}

// GetPath APIのURIのパス部分を求める
func GetPath(arg protocol.CommonArg) string {
	return "/r/" + APIVersion + execTemplate(arg.APIName(), arg.URI(), arg)
}

// GetParam APIのクエリストリング部分を求める
func GetParam(api API, arg protocol.CommonArg) *url.URL {
	param, err := url.Parse(api.Endpoint)
	if err != nil {
		panic(err)
	}
	param.Path = GetPath(arg)
	q := param.Query()
	_, toQuery, _ := ArgumentListType(arg)
	typs := reflect.TypeOf(arg)
	vals := reflect.ValueOf(arg)
	for _, key := range toQuery {
		if _, flag := typs.FieldByName(key); !flag {
			log.Info("no field", key)
			continue
		}
		if val := vals.FieldByName(key).String(); len(val) != 0 {
			q.Set(key, val)
		}
	}
	param.RawQuery = q.Encode()
	return param
}

// GetBody API呼び出しのリクエストボディ(JSON文字列)を求める
func GetBody(arg protocol.CommonArg) string {
	b, err := json.Marshal(arg)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Call API呼び出しを実行し、レスポンスを得る
func Call(api API, arg protocol.CommonArg, resp interface{}) (err error) {
	if err = Validate(arg); err != nil {
		return
	}
	var res *http.Response
	param := GetParam(api, arg)
	log.Debug("method", arg.Method(), "param", param, "arg", arg)
	if res, err = api.PostSome(arg.Method(), *param, arg); err != nil {
		log.Error("PostSome", err)
		return
	}
	log.Debug("res", res, "err", err)
	var b []byte
	if b, err = ioutil.ReadAll(res.Body); err != nil {
		log.Error("ioutil.ReadAll", err)
		return
	}
	log.Debug("data", string(b))
	if err = json.Unmarshal(b, &resp); err != nil {
		log.Error("json.Unmarshal", err)
		return
	}
	var cresp = protocol.CommonResponse{}
	err = json.Unmarshal(b, &cresp)
	if err == nil && cresp.ErrorResponse.ErrorType != "" {
		return fmt.Errorf("%s: %s", cresp.ErrorResponse.ErrorType, cresp.ErrorResponse.ErrorMessage)
	}
	return
}

// Validate APIの必須引数が入っているかどうかをチェック
func Validate(arg protocol.CommonArg) error {
	var res []string
	typs := reflect.TypeOf(arg)
	vals := reflect.ValueOf(arg)
	for i := 0; i < typs.NumField(); i++ {
		fld := typs.Field(i)
		tagstrJSON := fld.Tag.Get("json")
		tagstrP2 := fld.Tag.Get("p2pub")
		if strings.Contains(tagstrJSON, "omitempty") {
			// optional argument
			continue
		}
		if strings.Contains(tagstrP2, "query") {
			// optional argument
			continue
		}
		if val := vals.Field(i).String(); len(val) == 0 {
			res = append(res, fld.Name)
		}
	}
	if len(res) != 0 {
		return fmt.Errorf("missing arguments: %+v", res)
	}
	return nil
}

// ArgumentList API引数のリストを求める。必須とオプションに分類
func ArgumentList(arg protocol.CommonArg) (required, optional []string) {
	typs := reflect.TypeOf(arg)
	for i := 0; i < typs.NumField(); i++ {
		fld := typs.Field(i)
		tagstrJSON := fld.Tag.Get("json")
		tagstrP2 := fld.Tag.Get("p2pub")
		if strings.Contains(tagstrJSON, "omitempty") || strings.Contains(tagstrP2, "query") {
			optional = append(optional, fld.Name)
		} else {
			required = append(required, fld.Name)
		}
	}
	return
}

// ArgumentListType API引数のリストを求める。URI埋め込み、クエリストリング、JSONに分類
func ArgumentListType(arg protocol.CommonArg) (toURI, toQuery, toJSON []string) {
	typs := reflect.TypeOf(arg)
	for i := 0; i < typs.NumField(); i++ {
		fld := typs.Field(i)
		tagstrJSON := fld.Tag.Get("json")
		tagstrP2 := fld.Tag.Get("p2pub")
		if strings.Contains(tagstrP2, "query") {
			toQuery = append(toQuery, fld.Name)
		} else if strings.HasPrefix(tagstrJSON, "-") {
			toURI = append(toURI, fld.Name)
		} else {
			toJSON = append(toJSON, fld.Name)
		}
	}
	return
}

func argumentAltKeyList(arg protocol.CommonArg) (toAltQuery, toAltJSON map[string]string) {
	toAltQuery = make(map[string]string)
	toAltJSON = make(map[string]string)
	typs := reflect.TypeOf(arg)
	for i := 0; i < typs.NumField(); i++ {
		fld := typs.Field(i)
		tagstrJSON := fld.Tag.Get("json")
		tagstrP2 := fld.Tag.Get("p2pub")
		altKey := strings.Split(tagstrJSON, ",")[0]
		if altKey == "" || altKey == "-" {
			continue
		}
		if strings.Contains(tagstrP2, "query") {
			toAltQuery[fld.Name] = altKey
		} else {
			toAltJSON[fld.Name] = altKey
		}
	}
	return
}

// ValidateMap APIの必須引数が入っているかどうかをチェック
func ValidateMap(name string, data map[string]string) error {
	var res []string
	typs := protocol.TypeMap[name]
	for i := 0; i < typs.NumField(); i++ {
		fld := typs.Field(i)
		tagstrJSON := fld.Tag.Get("json")
		tagstrP2 := fld.Tag.Get("p2pub")
		if strings.Contains(tagstrJSON, "omitempty") {
			// optional argument
			continue
		}
		if strings.Contains(tagstrP2, "query") {
			// optional argument
			continue
		}
		if data[fld.Name] == "" {
			res = append(res, fld.Name)
		}
	}
	if len(res) != 0 {
		return fmt.Errorf("missing arguments: %+v", res)
	}
	return nil
}

// CallWithMap API呼び出しを実行する。引数と戻り値が構造体ではなくmap
func CallWithMap(api API, name string, data map[string]string, resp map[string]interface{}) error {
	if err := ValidateMap(name, data); err != nil {
		return err
	}
	argt := protocol.TypeMap[name]
	arg := reflect.Zero(argt).Interface().(protocol.CommonArg)
	var res *http.Response
	param, err := url.Parse(api.Endpoint)
	if err != nil {
		panic(err)
	}
	param.Path = "/r/" + APIVersion + execTemplate(name, arg.URI(), data)
	q := param.Query()
	_, toQuery, toJSON := ArgumentListType(arg)
	log.Debug("query", toQuery, "json", toJSON, "path", param.Path)
	toAltQuery, toAltJSON := argumentAltKeyList(arg)
	log.Debug("query altkey - ", toAltQuery, " json altkey - ", toAltJSON)
	var jsonmap = map[string]interface{}{}
	for _, v := range toJSON {
		if len(data[v]) != 0 {
			if altkey, ok := toAltJSON[v]; ok {
				jsonmap[altkey] = data[v]
			} else {
				jsonmap[v] = data[v]
			}
		}
	}
	for _, v := range toQuery {
		if len(data[v]) != 0 {
			if altkey, ok := toAltQuery[v]; ok {
				q.Set(altkey, data[v])
			} else {
				q.Set(v, data[v])
			}
		}
	}
	param.RawQuery = q.Encode()
	if res, err = api.PostSome(arg.Method(), *param, jsonmap); err != nil {
		log.Error("PostSome", err)
		return err
	}
	log.Debug("res", res)
	var b []byte
	if b, err = ioutil.ReadAll(res.Body); err != nil {
		log.Error("ioutil.ReadAll", err)
		return err
	}
	log.Debug("data", string(b))
	if err = json.Unmarshal(b, &resp); err != nil {
		log.Error("json.Unmarshal", err)
		return err
	}
	if val, ok := resp["ErrorResponse"]; ok {
		errstr := execTemplate("ErrorResponse", "{{.ErrorType}}: {{.ErrorMessage}}", val)
		return errors.New(errstr)
	}
	return nil
}
