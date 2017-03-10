package models

// AXFRResourceRecordSet is a representation of record name -> string
type AXFRResourceRecordSet map[string][]string

// Even though it would be nice to have the model map to something better than a string
// (a la an actual SRV, or A struct).
// This is the internal structure of how mesos-dns works today and the transformation of string -> DNS Struct
// happens on actual query time. Why this logic happens at query time? Who knows.

// AXFRRecords are the As, and SRVs that actually make up the Mesos-DNS zone
type AXFRRecords struct {
	As   AXFRResourceRecordSet
	SRVs AXFRResourceRecordSet
}

// AXFR is a rough representation of a "transfer" of the Mesos-DNS data
type AXFR struct {
	TTL            int32  // DNS TTL according to config
	Serial         uint32 // Current DNS zone version / serial number
	RefreshSeconds int    // How often we try to poll Mesos for updates -- minimum downstream poll interval
	Mname          string // primary name server
	Rname          string // email of admin esponsible
	Domain         string // Domain: name of the domain used (default "mesos", ie .mesos domain)
	Records        AXFRRecords
}
