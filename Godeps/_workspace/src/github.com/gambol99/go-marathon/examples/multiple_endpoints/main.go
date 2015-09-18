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
	"os"
	"time"

	marathon "github.com/gambol99/go-marathon"

	"github.com/golang/glog"
)

var marathon_url string

func init() {
	flag.StringVar(&marathon_url, "url", "http://127.0.0.1:8080,127.0.0.1:8080", "the url for the marathon endpoint")
}

func Assert(err error) {
	if err != nil {
		glog.Fatalf("Failed, error: %s", err)
	}
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathon_url
	config.LogOutput = os.Stdout
	client, err := marathon.NewClient(config)
	if err != nil {
		glog.Fatalf("Failed to create a client for marathon, error: %s", err)
	}
	for {
		if application, err := client.Applications(nil); err != nil {
			glog.Errorf("Failed to retrieve a list of applications, error: %s", err)
		} else {
			glog.Infof("Retrieved a list of applications, %v", application)
		}
		glog.Infof("Going to sleep for 20 seconds")
		time.Sleep(5 * time.Second)
	}
}
