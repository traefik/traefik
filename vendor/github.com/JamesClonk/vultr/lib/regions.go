package lib

import "sort"

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

type regions []Region

func (s regions) Len() int      { return len(s) }
func (s regions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s regions) Less(i, j int) bool {
	// sort order: continent, name
	if s[i].Continent < s[j].Continent {
		return true
	} else if s[i].Continent > s[j].Continent {
		return false
	}
	return s[i].Name < s[j].Name
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
	sort.Sort(regions(regionList))
	return regionList, nil
}
