// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DNS API
//
// API for the DNS service. Use this API to manage DNS zones, records, and other DNS resources.
// For more information, see Overview of the DNS Service (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Concepts/dnszonemanagement.htm).
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ExternalMaster An external master name server used as the source of zone data.
type ExternalMaster struct {

	// The server's IP address (IPv4 or IPv6).
	Address *string `mandatory:"true" json:"address"`

	// The server's port. Port value must be a value of 53, otherwise omit
	// the port value.
	Port *int `mandatory:"false" json:"port"`

	Tsig *Tsig `mandatory:"false" json:"tsig"`
}

func (m ExternalMaster) String() string {
	return common.PointerString(m)
}
