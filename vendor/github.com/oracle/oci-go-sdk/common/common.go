// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.

package common

import (
	"fmt"
	"regexp"
	"strings"
)

//Region type for regions
type Region string

const (
	//RegionSEA region SEA
	RegionSEA Region = "sea"
	//RegionCAToronto1 region for toronto
	RegionCAToronto1 Region = "ca-toronto-1"
	//RegionPHX region PHX
	RegionPHX Region = "us-phoenix-1"
	//RegionIAD region IAD
	RegionIAD Region = "us-ashburn-1"
	//RegionFRA region FRA
	RegionFRA Region = "eu-frankfurt-1"
	//RegionLHR region LHR
	RegionLHR Region = "uk-london-1"

	//RegionUSLangley1 region for langley
	RegionUSLangley1 Region = "us-langley-1"
	//RegionUSLuke1 region for luke
	RegionUSLuke1 Region = "us-luke-1"

	//RegionUSGovAshburn1 region for langley
	RegionUSGovAshburn1 Region = "us-gov-ashburn-1"
	//RegionUSGovChicago1 region for luke
	RegionUSGovChicago1 Region = "us-gov-chicago-1"
	//RegionUSGovPhoenix1 region for luke
	RegionUSGovPhoenix1 Region = "us-gov-phoenix-1"
)

var realm = map[string]string{
	"oc1": "oraclecloud.com",
	"oc2": "oraclegovcloud.com",
	"oc3": "oraclegovcloud.com",
}

var regionRealm = map[Region]string{
	RegionPHX:        "oc1",
	RegionIAD:        "oc1",
	RegionFRA:        "oc1",
	RegionLHR:        "oc1",
	RegionCAToronto1: "oc1",

	RegionUSLangley1: "oc2",
	RegionUSLuke1:    "oc2",

	RegionUSGovAshburn1: "oc3",
	RegionUSGovChicago1: "oc3",
	RegionUSGovPhoenix1: "oc3",
}

// Endpoint returns a endpoint for a service
func (region Region) Endpoint(service string) string {
	return fmt.Sprintf("%s.%s.%s", service, region, region.secondLevelDomain())
}

// EndpointForTemplate returns a endpoint for a service based on template
func (region Region) EndpointForTemplate(service string, serviceEndpointTemplate string) string {
	if serviceEndpointTemplate == "" {
		return region.Endpoint(service)
	}

	// replace service prefix
	endpoint := strings.Replace(serviceEndpointTemplate, "{serviceEndpointPrefix}", service, 1)

	// replace region
	endpoint = strings.Replace(endpoint, "{region}", string(region), 1)

	// replace second level domain
	endpoint = strings.Replace(endpoint, "{secondLevelDomain}", region.secondLevelDomain(), 1)

	return endpoint
}

func (region Region) secondLevelDomain() string {
	if realmID, ok := regionRealm[region]; ok {
		if secondLevelDomain, ok := realm[realmID]; ok {
			return secondLevelDomain
		}
	}

	Debugf("cannot find realm for region : %s, return default realm value.", region)
	return realm["oc1"]
}

//StringToRegion convert a string to Region type
func StringToRegion(stringRegion string) (r Region) {
	switch strings.ToLower(stringRegion) {
	case "sea":
		r = RegionSEA
	case "ca-toronto-1":
		r = RegionCAToronto1
	case "phx", "us-phoenix-1":
		r = RegionPHX
	case "iad", "us-ashburn-1":
		r = RegionIAD
	case "fra", "eu-frankfurt-1":
		r = RegionFRA
	case "lhr", "uk-london-1":
		r = RegionLHR
	case "us-langley-1":
		r = RegionUSLangley1
	case "us-luke-1":
		r = RegionUSLuke1
	case "us-gov-ashburn-1":
		r = RegionUSGovAshburn1
	case "us-gov-chicago-1":
		r = RegionUSGovChicago1
	case "us-gov-phoenix-1":
		r = RegionUSGovPhoenix1
	default:
		r = Region(stringRegion)
		Debugf("region named: %s, is not recognized", stringRegion)
	}
	return
}

// canStringBeRegion test if the string can be a region, if it can, returns the string as is, otherwise it
// returns an error
var blankRegex = regexp.MustCompile("\\s")

func canStringBeRegion(stringRegion string) (region string, err error) {
	if blankRegex.MatchString(stringRegion) || stringRegion == "" {
		return "", fmt.Errorf("region can not be empty or have spaces")
	}
	return stringRegion, nil
}
