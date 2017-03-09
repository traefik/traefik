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

package marathon

import (
	"fmt"
)

// Queue is the definition of marathon queue
type Queue struct {
	Items []Item `json:"queue"`
}

// Item is the definition of element in the queue
type Item struct {
	Count       int         `json:"count"`
	Delay       Delay       `json:"delay"`
	Application Application `json:"app"`
}

// Delay cotains the application postpone infomation
type Delay struct {
	Overdue         bool `json:"overdue"`
	TimeLeftSeconds int  `json:"timeLeftSeconds"`
}

// Queue retrieves content of the marathon launch queue
func (r *marathonClient) Queue() (*Queue, error) {
	var queue *Queue
	err := r.apiGet(marathonAPIQueue, nil, &queue)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

// DeleteQueueDelay resets task launch delay of the specific application
//		appID:		the ID of the application
func (r *marathonClient) DeleteQueueDelay(appID string) error {
	uri := fmt.Sprintf("%s/%s/delay", marathonAPIQueue, trimRootPath(appID))
	err := r.apiDelete(uri, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
