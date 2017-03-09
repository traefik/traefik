package lib

import "fmt"

// Plan on Vultr
type Plan struct {
	ID        int    `json:"VPSPLANID,string"`
	Name      string `json:"name"`
	VCpus     int    `json:"vcpu_count,string"`
	RAM       string `json:"ram"`
	Disk      string `json:"disk"`
	Bandwidth string `json:"bandwidth"`
	Price     string `json:"price_per_month"`
	Regions   []int  `json:"available_locations"`
}

// GetPlans returns a list of all available plans on Vultr account
func (c *Client) GetPlans() ([]Plan, error) {
	var planMap map[string]Plan
	if err := c.get(`plans/list`, &planMap); err != nil {
		return nil, err
	}

	var planList []Plan
	for _, plan := range planMap {
		planList = append(planList, plan)
	}
	return planList, nil
}

// GetAvailablePlansForRegion returns available plans for specified region
func (c *Client) GetAvailablePlansForRegion(id int) (planIDs []int, err error) {
	if err := c.get(fmt.Sprintf(`regions/availability?DCID=%v`, id), &planIDs); err != nil {
		return nil, err
	}
	return
}
