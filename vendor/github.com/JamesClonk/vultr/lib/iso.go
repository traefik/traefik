package lib

import (
	"sort"
	"strings"
)

// ISO image on Vultr
type ISO struct {
	ID       int    `json:"ISOID"`
	Created  string `json:"date_created"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	MD5sum   string `json:"md5sum"`
}

type isos []ISO

func (s isos) Len() int      { return len(s) }
func (s isos) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s isos) Less(i, j int) bool {
	// sort order: filename, created
	if strings.ToLower(s[i].Filename) < strings.ToLower(s[j].Filename) {
		return true
	} else if strings.ToLower(s[i].Filename) > strings.ToLower(s[j].Filename) {
		return false
	}
	return s[i].Created < s[j].Created
}

// GetISO returns a list of all ISO images on Vultr account
func (c *Client) GetISO() ([]ISO, error) {
	var isoMap map[string]ISO
	if err := c.get(`iso/list`, &isoMap); err != nil {
		return nil, err
	}

	var isoList []ISO
	for _, iso := range isoMap {
		isoList = append(isoList, iso)
	}
	sort.Sort(isos(isoList))
	return isoList, nil
}
