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
	"log"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

const waitTime = 5 * time.Second

var marathonURL string

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080,127.0.0.1:8080", "the url for the marathon endpoint")
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	client, err := marathon.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create a client for marathon, error: %s", err)
	}
	for {
		if application, err := client.Applications(nil); err != nil {
			log.Fatalf("Failed to retrieve a list of applications, error: %s", err)
		} else {
			log.Printf("Retrieved a list of applications, %v", application)
		}
		log.Printf("Going to sleep for %s\n", waitTime)
		time.Sleep(waitTime)
	}
}
