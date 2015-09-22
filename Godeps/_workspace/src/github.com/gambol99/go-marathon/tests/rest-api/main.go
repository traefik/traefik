/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
)

type RestMethod struct {
	// the uri of the method
	URI string `yaml:"uri,omitempty"`
	// the http method type (GET|PUT etc)
	Method string `yaml:"method,omitempty"`
	// the content i.e. response
	Content string `yaml:"content,omitempty"`
}

var Options struct {
	// the filename containing the methods
	method_file string
	// port we should be listening on
	port int
}

func init() {
	flag.StringVar(&Options.method_file, "methods", "methods.yml", "the name of the file containing the methods")
	flag.IntVar(&Options.port, "port", 3000, "the port we should be listening on")
}

func main() {
	flag.Parse()
	if Options.method_file == "" {
		glog.Errorf("You have not specified the methods file to import")
		os.Exit(1)
	}

	// step: open and read in the methods yaml
	contents, err := ioutil.ReadFile(Options.method_file)
	if err != nil {
		glog.Errorf("Failed to open|read the methods file: %s, error: %s", Options.method_file, err)
		os.Exit(1)
	}

	// step: unmarshal the yaml
	var methods []*RestMethod
	err = yaml.Unmarshal([]byte(contents), &methods)
	if err != nil {
		glog.Errorf("Failed to unmarshall the method yaml file: %s, error: %s", Options.method_file, err)
		os.Exit(1)
	}

	// step: construct a hash from the methods
	uris := make(map[string]*string, 0)
	for _, method := range methods {
		uris[fmt.Sprintf("%s:%s", method.Method, method.URI)] = &method.Content
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, reader *http.Request) {
		key := fmt.Sprintf("%s:%s", reader.Method, reader.RequestURI)
		glog.V(4).Infof("Request: uri: %s, method: %s", reader.RequestURI, reader.Method)
		if content, found := uris[key]; found {
			glog.V(4).Infof("Content: %s", *content)
			writer.Header().Add("Content-Type", "application/json")
			writer.Write([]byte(*content))
		}
	})
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", Options.port), nil))
}
