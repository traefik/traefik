package servicefabric

import (
	sfsdk "github.com/jjcollinge/servicefabric"
)

type clientMock struct {
	applications *sfsdk.ApplicationItemsPage
	services     *sfsdk.ServiceItemsPage
	partitions   *sfsdk.PartitionItemsPage
	replicas     *sfsdk.ReplicaItemsPage
	instances    *sfsdk.InstanceItemsPage
	labels       map[string]string
}

func (c *clientMock) GetApplications() (*sfsdk.ApplicationItemsPage, error) {
	return c.applications, nil
}

func (c *clientMock) GetServices(appName string) (*sfsdk.ServiceItemsPage, error) {
	return c.services, nil
}

func (c *clientMock) GetPartitions(appName, serviceName string) (*sfsdk.PartitionItemsPage, error) {
	return c.partitions, nil
}

func (c *clientMock) GetReplicas(appName, serviceName, partitionName string) (*sfsdk.ReplicaItemsPage, error) {
	return c.replicas, nil
}

func (c *clientMock) GetInstances(appName, serviceName, partitionName string) (*sfsdk.InstanceItemsPage, error) {
	return c.instances, nil
}

func (c *clientMock) GetServiceExtension(appType, applicationVersion, serviceTypeName, extensionKey string, response interface{}) error {
	return nil
}
func (c *clientMock) GetProperties(name string) (bool, map[string]string, error) {
	return true, c.labels, nil
}

func (c *clientMock) GetServiceLabels(service *sfsdk.ServiceItem, app *sfsdk.ApplicationItem, prefix string) (map[string]string, error) {
	return c.labels, nil
}
