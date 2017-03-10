package lib

// ISO image on Vultr
type ISO struct {
	ID       int    `json:"ISOID"`
	Created  string `json:"date_created"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	MD5sum   string `json:"md5sum"`
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
	return isoList, nil
}
