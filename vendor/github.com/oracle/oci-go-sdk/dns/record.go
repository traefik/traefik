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

// Record A DNS resource record. For more information about DNS records, see RFC 1034 (https://tools.ietf.org/html/rfc1034#section-3.6).
type Record struct {

	// The fully qualified domain name where the record can be located.
	Domain *string `mandatory:"false" json:"domain"`

	// A unique identifier for the record within its zone.
	RecordHash *string `mandatory:"false" json:"recordHash"`

	// A Boolean flag indicating whether or not parts of the record
	// are unable to be explicitly managed.
	IsProtected *bool `mandatory:"false" json:"isProtected"`

	// The record's data, as whitespace-delimited tokens in
	// type-specific presentation format. All RDATA is normalized and the
	// returned presentation of your RDATA may differ from its initial input.
	// For more information about RDATA, see Supported DNS Resource Record Types (https://docs.us-phoenix-1.oraclecloud.com/iaas/Content/DNS/Reference/supporteddnsresource.htm)
	Rdata *string `mandatory:"false" json:"rdata"`

	// The latest version of the record's zone in which its RRSet differs
	// from the preceding version.
	RrsetVersion *string `mandatory:"false" json:"rrsetVersion"`

	// The canonical name for the record's type, such as A or CNAME. For more
	// information, see Resource Record (RR) TYPEs (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4).
	Rtype *string `mandatory:"false" json:"rtype"`

	// The Time To Live for the record, in seconds.
	Ttl *int `mandatory:"false" json:"ttl"`
}

func (m Record) String() string {
	return common.PointerString(m)
}
