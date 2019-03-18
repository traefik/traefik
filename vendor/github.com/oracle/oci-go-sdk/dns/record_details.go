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

// RecordDetails A DNS resource record. For more information about records, see RFC 1034 (https://tools.ietf.org/html/rfc1034#section-3.6).
type RecordDetails struct {

	// The fully qualified domain name where the record can be located.
	Domain *string `mandatory:"true" json:"domain"`

	// The record's data, as whitespace-delimited tokens in
	// type-specific presentation format. All RDATA is normalized and the
	// returned presentation of your RDATA may differ from its initial input.
	// For more information about RDATA, see Supported DNS Resource Record Types (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Reference/supporteddnsresource.htm)
	Rdata *string `mandatory:"true" json:"rdata"`

	// The canonical name for the record's type, such as A or CNAME. For more
	// information, see Resource Record (RR) TYPEs (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4).
	Rtype *string `mandatory:"true" json:"rtype"`

	// The Time To Live for the record, in seconds.
	Ttl *int `mandatory:"true" json:"ttl"`

	// A unique identifier for the record within its zone.
	RecordHash *string `mandatory:"false" json:"recordHash"`

	// A Boolean flag indicating whether or not parts of the record
	// are unable to be explicitly managed.
	IsProtected *bool `mandatory:"false" json:"isProtected"`

	// The latest version of the record's zone in which its RRSet differs
	// from the preceding version.
	RrsetVersion *string `mandatory:"false" json:"rrsetVersion"`
}

func (m RecordDetails) String() string {
	return common.PointerString(m)
}
