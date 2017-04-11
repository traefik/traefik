package lib

import (
	"sort"
	"strings"
)

// OS image on Vultr
type OS struct {
	ID        int    `json:"OSID"`
	Name      string `json:"name"`
	Arch      string `json:"arch"`
	Family    string `json:"family"`
	Windows   bool   `json:"windows"`
	Surcharge string `json:"surcharge"`
}

type oses []OS

func (s oses) Len() int           { return len(s) }
func (s oses) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s oses) Less(i, j int) bool { return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name) }

// GetOS returns a list of all available operating systems on Vultr
func (c *Client) GetOS() ([]OS, error) {
	var osMap map[string]OS
	if err := c.get(`os/list`, &osMap); err != nil {
		return nil, err
	}

	var osList []OS
	for _, os := range osMap {
		osList = append(osList, os)
	}
	sort.Sort(oses(osList))
	return osList, nil
}
