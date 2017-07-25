/*
Copyright 2016 Rohith All rights reserved.

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
	"log"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

var marathonURL string

func init() {
	flag.StringVar(&marathonURL, "url", "http://127.0.0.1:8080", "the url for the marathon endpoint")
}

func main() {
	config := marathon.NewDefaultConfig()
	config.URL = marathonURL
	client, err := marathon.NewClient(config)
	if err != nil {
		log.Fatalf("Make new marathon client error: %v", err)
	}

	app := marathon.Application{}
	app.ID = "queue-test"
	app.Command("sleep 5")
	app.Count(1)
	app.Memory(32)
	fmt.Println("Creating/updating app.")
	// Update application will either create or update the app.
	_, err = client.UpdateApplication(&app, false)
	if err != nil {
		log.Fatalf("Update application error: %v", err)
	}
	// wait until marathon will launch tasks
	err = client.WaitOnApplication(app.ID, 10*time.Second)
	if err != nil {
		log.Fatalln("Application deploy failure, timeout.")
	}
	fmt.Println("Application was deployed.")

	// get marathon queue by chance
	for i := 0; i < 30; i++ {
		// Avoid shadowing err from outer scope.
		var queue *marathon.Queue
		queue, err = client.Queue()
		if err != nil {
			log.Fatalf("Get queue error: %v\n", err)
		}
		if len(queue.Items) > 0 {
			fmt.Println(queue)
			break
		}
		fmt.Printf("Queue is blank now, retry(%d)...\n", 30-i)
		time.Sleep(time.Second)
	}

	// delete marathon queue delay
	err = client.DeleteQueueDelay(app.ID)
	if err != nil {
		log.Fatalf("Delete queue delay error: %v\n", err)
	}
	fmt.Println("Queue delay deleted.")

	return
}
