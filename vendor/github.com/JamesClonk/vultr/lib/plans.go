package lib

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

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

type plans []Plan

func (p plans) Len() int      { return len(p) }
func (p plans) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p plans) Less(i, j int) bool {
	pa, _ := strconv.ParseFloat(strings.TrimSpace(p[i].Price), 64)
	pb, _ := strconv.ParseFloat(strings.TrimSpace(p[j].Price), 64)
	ra, _ := strconv.ParseInt(strings.TrimSpace(p[i].RAM), 10, 64)
	rb, _ := strconv.ParseInt(strings.TrimSpace(p[j].RAM), 10, 64)
	da, _ := strconv.ParseInt(strings.TrimSpace(p[i].Disk), 10, 64)
	db, _ := strconv.ParseInt(strings.TrimSpace(p[j].Disk), 10, 64)

	// sort order: price, vcpu, ram, disk
	if pa < pb {
		return true
	} else if pa > pb {
		return false
	}

	if p[i].VCpus < p[j].VCpus {
		return true
	} else if p[i].VCpus > p[j].VCpus {
		return false
	}

	if ra < rb {
		return true
	} else if ra > rb {
		return false
	}

	return da < db
}

// GetPlans returns a list of all available plans on Vultr account
func (c *Client) GetPlans() ([]Plan, error) {
	var planMap map[string]Plan
	if err := c.get(`plans/list`, &planMap); err != nil {
		return nil, err
	}

	var p plans
	for _, plan := range planMap {
		p = append(p, plan)
	}

	sort.Sort(plans(p))
	return p, nil
}

// GetAvailablePlansForRegion returns available plans for specified region
func (c *Client) GetAvailablePlansForRegion(id int) (planIDs []int, err error) {
	if err := c.get(fmt.Sprintf(`regions/availability?DCID=%v`, id), &planIDs); err != nil {
		return nil, err
	}
	return
}
