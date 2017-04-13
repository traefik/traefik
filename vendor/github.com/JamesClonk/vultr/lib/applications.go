package lib

import (
	"sort"
	"strings"
)

// Application on Vultr
type Application struct {
	ID         string  `json:"APPID"`
	Name       string  `json:"name"`
	ShortName  string  `json:"short_name"`
	DeployName string  `json:"deploy_name"`
	Surcharge  float64 `json:"surcharge"`
}

type applications []Application

func (s applications) Len() int      { return len(s) }
func (s applications) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s applications) Less(i, j int) bool {
	return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name)
}

// GetApplications returns a list of all available applications on Vultr
func (c *Client) GetApplications() ([]Application, error) {
	var appMap map[string]Application
	if err := c.get(`app/list`, &appMap); err != nil {
		return nil, err
	}

	var appList []Application
	for _, app := range appMap {
		appList = append(appList, app)
	}
	sort.Sort(applications(appList))
	return appList, nil
}
