package lib

// OS image on Vultr
type OS struct {
	ID        int    `json:"OSID"`
	Name      string `json:"name"`
	Arch      string `json:"arch"`
	Family    string `json:"family"`
	Windows   bool   `json:"windows"`
	Surcharge string `json:"surcharge"`
}

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
	return osList, nil
}
