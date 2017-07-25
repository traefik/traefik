package catalog

type RancherClient struct {
	RancherBaseClient

	ApiVersion      ApiVersionOperations
	Catalog         CatalogOperations
	Template        TemplateOperations
	Question        QuestionOperations
	TemplateVersion TemplateVersionOperations
	Error           ErrorOperations
}

func constructClient(rancherBaseClient *RancherBaseClientImpl) *RancherClient {
	client := &RancherClient{
		RancherBaseClient: rancherBaseClient,
	}

	client.ApiVersion = newApiVersionClient(client)
	client.Catalog = newCatalogClient(client)
	client.Template = newTemplateClient(client)
	client.Question = newQuestionClient(client)
	client.TemplateVersion = newTemplateVersionClient(client)
	client.Error = newErrorClient(client)

	return client
}

func NewRancherClient(opts *ClientOpts) (*RancherClient, error) {
	rancherBaseClient := &RancherBaseClientImpl{
		Types: map[string]Schema{},
	}
	client := constructClient(rancherBaseClient)

	err := setupRancherBaseClient(rancherBaseClient, opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
