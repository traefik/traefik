package govultr

import (
	"context"
	"net/http"
)

// ApplicationService is the interface to interact with the Application endpoint on the Vultr API
// Link: https://www.vultr.com/api/#app
type ApplicationService interface {
	List(ctx context.Context) ([]Application, error)
}

// ApplicationServiceHandler handles interaction with the application methods for the Vultr API
type ApplicationServiceHandler struct {
	client *Client
}

// Application represents a Vultr application
type Application struct {
	AppID      string  `json:"APPID"`
	Name       string  `json:"name"`
	ShortName  string  `json:"short_name"`
	DeployName string  `json:"deploy_name"`
	Surcharge  float64 `json:"surcharge"`
}

// List retrieves a list of available applications that can be launched when creating a Vultr VPS
func (a *ApplicationServiceHandler) List(ctx context.Context) ([]Application, error) {

	uri := "/v1/app/list"
	req, err := a.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	appsMap := make(map[string]Application)

	err = a.client.DoWithContext(ctx, req, &appsMap)
	if err != nil {
		return nil, err
	}

	var apps []Application
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return apps, nil
}
