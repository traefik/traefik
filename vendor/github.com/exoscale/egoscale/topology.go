package egoscale

import (
	"fmt"
	"regexp"
	"strings"
)

// GetSecurityGroups returns all security groups
//
// Deprecated: do it yourself
func (exo *Client) GetSecurityGroups() (map[string]SecurityGroup, error) {
	var sgs map[string]SecurityGroup
	resp, err := exo.Request(&ListSecurityGroups{})
	if err != nil {
		return nil, err
	}

	sgs = make(map[string]SecurityGroup)
	for _, sg := range resp.(*ListSecurityGroupsResponse).SecurityGroup {
		sgs[sg.Name] = sg
	}
	return sgs, nil
}

// GetSecurityGroupID returns security group by name
//
// Deprecated: do it yourself
func (exo *Client) GetSecurityGroupID(name string) (string, error) {
	resp, err := exo.Request(&ListSecurityGroups{SecurityGroupName: name})
	if err != nil {
		return "", err
	}

	for _, sg := range resp.(*ListSecurityGroupsResponse).SecurityGroup {
		if sg.Name == name {
			return sg.ID, nil
		}
	}

	return "", nil
}

// GetAllZones returns all the zone id by name
//
// Deprecated: do it yourself
func (exo *Client) GetAllZones() (map[string]string, error) {
	var zones map[string]string
	resp, err := exo.Request(&ListZones{})
	if err != nil {
		return zones, err
	}

	zones = make(map[string]string)
	for _, zone := range resp.(*ListZonesResponse).Zone {
		zones[strings.ToLower(zone.Name)] = zone.ID
	}
	return zones, nil
}

// GetProfiles returns a mapping of the service offerings by name
//
// Deprecated: do it yourself
func (exo *Client) GetProfiles() (map[string]string, error) {
	profiles := make(map[string]string)
	resp, err := exo.Request(&ListServiceOfferings{})
	if err != nil {
		return profiles, nil
	}

	for _, offering := range resp.(*ListServiceOfferingsResponse).ServiceOffering {
		profiles[strings.ToLower(offering.Name)] = offering.ID
	}

	return profiles, nil
}

// GetKeypairs returns the list of SSH keyPairs
//
// Deprecated: do it yourself
func (exo *Client) GetKeypairs() ([]SSHKeyPair, error) {
	var keypairs []SSHKeyPair

	resp, err := exo.Request(&ListSSHKeyPairs{})
	if err != nil {
		return keypairs, err
	}

	r := resp.(*ListSSHKeyPairsResponse)
	keypairs = make([]SSHKeyPair, r.Count)
	for i, keypair := range r.SSHKeyPair {
		keypairs[i] = keypair
	}
	return keypairs, nil
}

// GetAffinityGroups returns a mapping of the (anti-)affinity groups
//
// Deprecated: do it yourself
func (exo *Client) GetAffinityGroups() (map[string]string, error) {
	var affinitygroups map[string]string

	resp, err := exo.Request(&ListAffinityGroups{})
	if err != nil {
		return affinitygroups, err
	}

	affinitygroups = make(map[string]string)
	for _, affinitygroup := range resp.(*ListAffinityGroupsResponse).AffinityGroup {
		affinitygroups[affinitygroup.Name] = affinitygroup.ID
	}
	return affinitygroups, nil
}

// GetImages list the available featured images and group them by name, then size.
//
// Deprecated: do it yourself
func (exo *Client) GetImages() (map[string]map[int64]string, error) {
	var images map[string]map[int64]string
	images = make(map[string]map[int64]string)
	re := regexp.MustCompile(`^Linux (?P<name>.+?) (?P<version>[0-9.]+)\b`)

	resp, err := exo.Request(&ListTemplates{
		TemplateFilter: "featured",
		ZoneID:         "1", // XXX: Hack to list only CH-GVA
	})
	if err != nil {
		return images, err
	}

	for _, template := range resp.(*ListTemplatesResponse).Template {
		size := int64(template.Size >> 30) // B to GiB

		fullname := strings.ToLower(template.Name)

		if _, present := images[fullname]; !present {
			images[fullname] = make(map[int64]string)
		}
		images[fullname][size] = template.ID

		submatch := re.FindStringSubmatch(template.Name)
		if len(submatch) > 0 {
			name := strings.Replace(strings.ToLower(submatch[1]), " ", "-", -1)
			version := submatch[2]
			image := fmt.Sprintf("%s-%s", name, version)

			if _, present := images[image]; !present {
				images[image] = make(map[int64]string)
			}
			images[image][size] = template.ID
		}
	}
	return images, nil
}

// GetTopology returns an big, yet incomplete view of the world
//
// Deprecated: will go away in the future
func (exo *Client) GetTopology() (*Topology, error) {
	zones, err := exo.GetAllZones()
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
	for k, v := range securityGroups {
		groups[k] = v.ID
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
		Images:         images,
		Keypairs:       keynames,
		Profiles:       profiles,
		AffinityGroups: affinitygroups,
		SecurityGroups: groups,
	}

	return topo, nil
}
