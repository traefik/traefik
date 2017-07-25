package storage

// TableServiceClient contains operations for Microsoft Azure Table Storage
// Service.
type TableServiceClient struct {
	client Client
	auth   authentication
}

// GetServiceProperties gets the properties of your storage account's table service.
// See: https://docs.microsoft.com/en-us/rest/api/storageservices/fileservices/get-table-service-properties
func (c *TableServiceClient) GetServiceProperties() (*ServiceProperties, error) {
	return c.client.getServiceProperties(tableServiceName, c.auth)
}

// SetServiceProperties sets the properties of your storage account's table service.
// See: https://docs.microsoft.com/en-us/rest/api/storageservices/fileservices/set-table-service-properties
func (c *TableServiceClient) SetServiceProperties(props ServiceProperties) error {
	return c.client.setServiceProperties(props, tableServiceName, c.auth)
}
