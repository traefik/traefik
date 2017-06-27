package metadata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client interface {
	OnChangeWithError(int, func(string)) error
	OnChange(int, func(string))
	SendRequest(string) ([]byte, error)
	GetVersion() (string, error)
	GetSelfHost() (Host, error)
	GetSelfContainer() (Container, error)
	GetSelfServiceByName(string) (Service, error)
	GetSelfService() (Service, error)
	GetSelfStack() (Stack, error)
	GetServices() ([]Service, error)
	GetStacks() ([]Stack, error)
	GetContainers() ([]Container, error)
	GetServiceContainers(string, string) ([]Container, error)
	GetHosts() ([]Host, error)
	GetHost(string) (Host, error)
	GetNetworks() ([]Network, error)
}

type client struct {
	url    string
	ip     string
	client *http.Client
}

func newClient(url, ip string) *client {
	return &client{url, ip, &http.Client{Timeout: 10 * time.Second}}
}

func NewClient(url string) Client {
	ip := ""
	return newClient(url, ip)
}

func NewClientWithIPAndWait(url, ip string) (Client, error) {
	client := newClient(url, ip)

	if err := testConnection(client); err != nil {
		return nil, err
	}

	return client, nil
}

func NewClientAndWait(url string) (Client, error) {
	ip := ""
	client := newClient(url, ip)

	if err := testConnection(client); err != nil {
		return nil, err
	}

	return client, nil
}

func (m *client) SendRequest(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", m.url+path, nil)
	req.Header.Add("Accept", "application/json")
	if m.ip != "" {
		req.Header.Add("X-Forwarded-For", m.ip)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error %v accessing %v path", resp.StatusCode, path)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (m *client) GetVersion() (string, error) {
	resp, err := m.SendRequest("/version")
	if err != nil {
		return "", err
	}
	return string(resp[:]), nil
}

func (m *client) GetSelfHost() (Host, error) {
	resp, err := m.SendRequest("/self/host")
	var host Host
	if err != nil {
		return host, err
	}

	if err = json.Unmarshal(resp, &host); err != nil {
		return host, err
	}

	return host, nil
}

func (m *client) GetSelfContainer() (Container, error) {
	resp, err := m.SendRequest("/self/container")
	var container Container
	if err != nil {
		return container, err
	}

	if err = json.Unmarshal(resp, &container); err != nil {
		return container, err
	}

	return container, nil
}

func (m *client) GetSelfServiceByName(name string) (Service, error) {
	resp, err := m.SendRequest("/self/stack/services/" + name)
	var service Service
	if err != nil {
		return service, err
	}

	if err = json.Unmarshal(resp, &service); err != nil {
		return service, err
	}

	return service, nil
}

func (m *client) GetSelfService() (Service, error) {
	resp, err := m.SendRequest("/self/service")
	var service Service
	if err != nil {
		return service, err
	}

	if err = json.Unmarshal(resp, &service); err != nil {
		return service, err
	}

	return service, nil
}

func (m *client) GetSelfStack() (Stack, error) {
	resp, err := m.SendRequest("/self/stack")
	var stack Stack
	if err != nil {
		return stack, err
	}

	if err = json.Unmarshal(resp, &stack); err != nil {
		return stack, err
	}

	return stack, nil
}

func (m *client) GetServices() ([]Service, error) {
	resp, err := m.SendRequest("/services")
	var services []Service
	if err != nil {
		return services, err
	}

	if err = json.Unmarshal(resp, &services); err != nil {
		return services, err
	}
	return services, nil
}

func (m *client) GetStacks() ([]Stack, error) {
	resp, err := m.SendRequest("/stacks")
	var stacks []Stack
	if err != nil {
		return stacks, err
	}

	if err = json.Unmarshal(resp, &stacks); err != nil {
		return stacks, err
	}
	return stacks, nil
}

func (m *client) GetContainers() ([]Container, error) {
	resp, err := m.SendRequest("/containers")
	var containers []Container
	if err != nil {
		return containers, err
	}

	if err = json.Unmarshal(resp, &containers); err != nil {
		return containers, err
	}
	return containers, nil
}

func (m *client) GetServiceContainers(serviceName string, stackName string) ([]Container, error) {
	var serviceContainers = []Container{}
	containers, err := m.GetContainers()
	if err != nil {
		return serviceContainers, err
	}

	for _, container := range containers {
		if container.StackName == stackName && container.ServiceName == serviceName {
			serviceContainers = append(serviceContainers, container)
		}
	}

	return serviceContainers, nil
}

func (m *client) GetHosts() ([]Host, error) {
	resp, err := m.SendRequest("/hosts")
	var hosts []Host
	if err != nil {
		return hosts, err
	}

	if err = json.Unmarshal(resp, &hosts); err != nil {
		return hosts, err
	}
	return hosts, nil
}

func (m *client) GetHost(UUID string) (Host, error) {
	var host Host
	hosts, err := m.GetHosts()
	if err != nil {
		return host, err
	}
	for _, host := range hosts {
		if host.UUID == UUID {
			return host, nil
		}
	}

	return host, fmt.Errorf("could not find host by UUID %v", UUID)
}

func (m *client) GetNetworks() ([]Network, error) {
	resp, err := m.SendRequest("/networks")
	var networks []Network
	if err != nil {
		return networks, err
	}

	if err = json.Unmarshal(resp, &networks); err != nil {
		return networks, err
	}

	return networks, nil
}
