package location

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/management"
)

//LocationClient is used to perform operations on Azure Locations
type LocationClient struct {
	client management.Client
}

type ListLocationsResponse struct {
	XMLName   xml.Name   `xml:"Locations"`
	Locations []Location `xml:"Location"`
}

type Location struct {
	Name                    string
	DisplayName             string
	AvailableServices       []string `xml:"AvailableServices>AvailableService"`
	WebWorkerRoleSizes      []string `xml:"ComputeCapabilities>WebWorkerRoleSizes>RoleSize"`
	VirtualMachineRoleSizes []string `xml:"ComputeCapabilities>VirtualMachinesRoleSizes>RoleSize"`
}

func (ll ListLocationsResponse) String() string {
	var buf bytes.Buffer
	for _, l := range ll.Locations {
		fmt.Fprintf(&buf, "%s, ", l.Name)
	}

	return strings.Trim(buf.String(), ", ")
}
