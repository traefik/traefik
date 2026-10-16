/*
Copyright 2017 The Kubernetes Authors.

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

package framework

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
)

// Logs returns the log entries of a given Pod.
func Logs(client kubernetes.Interface, namespace, podName string) (string, error) {
	// Logs from jails take a bigger time to get shipped due to the need of tailing them
	Sleep(3 * time.Second)
	logs, err := client.CoreV1().RESTClient().Get().
		Resource("pods").
		Namespace(namespace).
		Name(podName).SubResource("log").
		Param("container", "controller").
		Do(context.TODO()).
		Raw()
	if err != nil {
		return "", err
	}

	return string(logs), nil
}
