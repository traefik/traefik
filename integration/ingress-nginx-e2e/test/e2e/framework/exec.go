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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// GetLbAlgorithm returns algorithm identifier for the given backend
func (f *Framework) GetLbAlgorithm(serviceName string, servicePort int) (string, error) {
	backendName := fmt.Sprintf("%s-%s-%v", f.Namespace, serviceName, servicePort)
	cmd := fmt.Sprintf("/dbg backends get %s", backendName)

	output, err := f.ExecIngressPod(cmd)
	if err != nil {
		return "", err
	}

	var backend map[string]interface{}
	err = json.Unmarshal([]byte(output), &backend)
	if err != nil {
		return "", err
	}

	algorithm, ok := backend["load-balance"].(string)
	if !ok {
		return "", fmt.Errorf("error while accessing load-balance field of backend")
	}

	return algorithm, nil
}

// ExecIngressPod executes a command inside the first container in ingress controller running pod
func (f *Framework) ExecIngressPod(command string) (string, error) {
	return f.ExecCommand(f.pod, command)
}

// ExecCommand executes a command inside a the first container in a running pod
func (f *Framework) ExecCommand(pod *corev1.Pod, command string) (string, error) {
	skipNginxOnly("ExecCommand")

	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	//nolint:gosec // Ignore G204 error
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v exec --namespace %s %s --container controller -- %s", KubectlPath, pod.Namespace, pod.Name, command))
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not execute '%s %s': %v", cmd.Path, cmd.Args, err)
	}

	if execErr.Len() > 0 {
		return "", fmt.Errorf("stderr: %v", execErr.String())
	}

	return execOut.String(), nil
}

// NamespaceContent executes a kubectl command that returns information about
// pods, services, endpoint and deployments inside the current namespace
func (f *Framework) NamespaceContent() (string, error) {
	var (
		execOut bytes.Buffer
		execErr bytes.Buffer
	)

	//nolint:gosec // Ignore G204 error
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%v get pods,services,endpoints,deployments --namespace %s", KubectlPath, f.Namespace))
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not execute '%s %s': %v", cmd.Path, cmd.Args, err)
	}

	eout := strings.TrimSpace(execErr.String())
	if eout != "" {
		return "", fmt.Errorf("stderr: %v", eout)
	}

	return execOut.String(), nil
}

// newIngressController deploys a new NGINX Ingress controller in a namespace
func (f *Framework) newIngressController(namespace, namespaceOverlay string) error {
	return f.deployTraefik(namespace)

	// Creates an nginx deployment
	isChroot, ok := os.LookupEnv("IS_CHROOT")
	if !ok {
		isChroot = "false"
	}

	cmd := exec.Command("./wait-for-nginx.sh", namespace, namespaceOverlay, isChroot)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error waiting for ingress controller deployment: %v.\nLogs:\n%v", err, string(out))
	}

	return nil
}

var proxyRegexp = regexp.MustCompile(`Starting to serve on .*:(\d+)`)

// KubectlProxy creates a proxy to kubernetes apiserver
func (f *Framework) KubectlProxy(port int) (int, *exec.Cmd, error) {
	//nolint:gosec // Ignore G204 error
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s proxy --accept-hosts=.* --address=0.0.0.0 --port=%d", KubectlPath, port))
	stdout, stderr, err := startCmdAndStreamOutput(cmd)
	if err != nil {
		return -1, nil, err
	}

	defer stdout.Close()
	defer stderr.Close()

	buf := make([]byte, 128)
	var n int
	if n, err = stdout.Read(buf); err != nil {
		return -1, cmd, fmt.Errorf("failed to read from kubectl proxy stdout: %v", err)
	}

	output := string(buf[:n])
	match := proxyRegexp.FindStringSubmatch(output)
	if len(match) == 2 {
		if port, err := strconv.Atoi(match[1]); err == nil {
			return port, cmd, nil
		}
	}

	return -1, cmd, fmt.Errorf("failed to parse port from proxy stdout: %s", output)
}

func (f *Framework) UninstallChart() error {
	return f.deleteTraefik()

	//nolint:gosec //Ignore G204 error
	cmd := exec.Command("helm", "uninstall", "--namespace", f.Namespace, "nginx-ingress")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unexpected error uninstalling ingress-nginx release: %v", err)
	}

	return nil
}

func startCmdAndStreamOutput(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}

	Logf("Asynchronously running '%s'", strings.Join(cmd.Args, " "))
	err = cmd.Start()
	return
}
