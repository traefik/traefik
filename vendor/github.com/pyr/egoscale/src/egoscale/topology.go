package egoscale

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func (exo *Client) GetSecurityGroups() (map[string]SecurityGroup, error) {
	var sgs map[string]SecurityGroup
	params := url.Values{}
	resp, err := exo.Request("listSecurityGroups", params)

	if err != nil {
		return nil, err
	}

	var r ListSecurityGroupsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	sgs = make(map[string]SecurityGroup)
	for _, sg := range r.SecurityGroups {
		sgs[sg.Name] = *sg
	}
	return sgs, nil
}

func (exo *Client) GetSecurityGroupId(name string) (string, error) {
	params := url.Values{}
	resp, err := exo.Request("listSecurityGroups", params); if err != nil {
		return "", err
	}

	var r ListSecurityGroupsResponse
	err = json.Unmarshal(resp, &r); if err != nil {
		return "", err
	}

	for _, sg := range r.SecurityGroups {
		if sg.Name == name {
			return sg.Id, nil
		}
	}

	return "", nil
}

func (exo *Client) GetZones() (map[string]string, error) {
	var zones map[string]string
	params := url.Values{}
	resp, err := exo.Request("listZones", params)

	if err != nil {
		return nil, err
	}

	var r ListZonesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	zones = make(map[string]string)
	for _, zone := range r.Zones {
		zones[zone.Name] = zone.Id
	}
	return zones, nil
}

func (exo *Client) GetProfiles() (map[string]string, error) {

	var profiles map[string]string
	params := url.Values{}
	resp, err := exo.Request("listServiceOfferings", params)

	if err != nil {
		return nil, err
	}

	var r ListServiceOfferingsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	profiles = make(map[string]string)
	for _, offering := range r.ServiceOfferings {
		profiles[strings.ToLower(offering.Name)] = offering.Id
	}

	return profiles, nil
}

func (exo *Client) GetKeypairs() ([]SSHKeyPair, error) {
	params := url.Values{}

	resp, err := exo.Request("listSSHKeyPairs", params)

	if err != nil {
		return nil, err
	}

	var r ListSSHKeyPairsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	var keypairs = make([]SSHKeyPair, r.Count, r.Count)
	for i, keypair := range r.SSHKeyPairs {
		keypairs[i] = *keypair
	}
	return keypairs, nil
}

func (exo *Client) GetAffinityGroups() (map[string]string, error) {
	var affinitygroups map[string]string
	params := url.Values{}

	resp, err := exo.Request("listAffinityGroups", params)

	if err != nil {
		return nil, err
	}

	var r ListAffinityGroupsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	affinitygroups = make(map[string]string)
	for _, affinitygroup := range r.AffinityGroups {
		affinitygroups[affinitygroup.Name] = affinitygroup.Id
	}
	return affinitygroups, nil

}

func (exo *Client) GetImages() (map[string]map[int]string, error) {
	var images map[string]map[int]string
	images = make(map[string]map[int]string)

	params := url.Values{}
	params.Set("templatefilter", "executable")

	resp, err := exo.Request("listTemplates", params)

	if err != nil {
		return nil, err
	}

	var r ListTemplatesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9.]+).*$`)
	for _, template := range r.Templates {
		size := int(template.Size / (1024 * 1024 * 1024))
		submatch := re.FindStringSubmatch(template.Name)
		if len(submatch) > 0 {
			name := strings.Replace(strings.ToLower(submatch[1]), " ", "-", -1)
			version := submatch[2]
			image := fmt.Sprintf("%s-%s", name, version)

			_, present := images[image]
			if !present {
				images[image] = make(map[int]string)
			}
			images[image][size] = template.Id

			images[fmt.Sprintf("%s-%s", name, version)][size] = template.Id
		}
	}
	return images, nil
}

func (exo *Client) GetTopology() (*Topology, error) {

	zones, err := exo.GetZones()
	if err != nil {
		return nil, err
	}
	images, err := exo.GetImages()
	if err != nil {
		return nil, err
	}
	securityGroups, err := exo.GetSecurityGroups()
	if err != nil {
		return nil, err
	}
	groups := make(map[string]string)
	for k,v := range securityGroups {
		groups[k] = v.Id
	}

	keypairs, err := exo.GetKeypairs()
	if err != nil {
		return nil, err
	}

	/* Convert the ssh keypair to contain just the name */
	keynames := make([]string, len(keypairs))
	for i, k := range keypairs {
		keynames[i] = k.Name
	}

	affinitygroups, err := exo.GetAffinityGroups()
	if err != nil {
		return nil, err
	}
	profiles, err := exo.GetProfiles()
	if err != nil {
		return nil, err
	}

	topo := &Topology{
		Zones:          zones,
		Profiles:       profiles,
		Images:         images,
		Keypairs:       keynames,
		AffinityGroups: affinitygroups,
		SecurityGroups: groups,
	}

	return topo, nil
}
