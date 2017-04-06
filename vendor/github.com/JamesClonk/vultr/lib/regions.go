package lib

// Region on Vultr
type Region struct {
	ID           int    `json:"DCID,string"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	Continent    string `json:"continent"`
	State        string `json:"state"`
	Ddos         bool   `json:"ddos_protection"`
	BlockStorage bool   `json:"block_storage"`
	Code         string `json:"regioncode"`
}

// GetRegions returns a list of all available Vultr regions
func (c *Client) GetRegions() ([]Region, error) {
	var regionMap map[string]Region
	if err := c.get(`regions/list`, &regionMap); err != nil {
		return nil, err
	}

	var regionList []Region
	for _, os := range regionMap {
		regionList = append(regionList, os)
	}
	return regionList, nil
}
