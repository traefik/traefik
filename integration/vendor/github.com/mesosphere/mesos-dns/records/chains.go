package records

import (
	"github.com/mesosphere/mesos-dns/records/labels"
)

// chain is a generation func that consumes record-like strings and does
// something with them
type chain func(...string)

const (
	protocolNone = "" // for readability
	domainNone   = "" // for readability
)

// withProtocol appends `._{protocol}.{framework}` to records. if protocol is "" then
// the protocols "tcp" and "udp" are assumed.
func withProtocol(protocol, framework string, spec labels.Func, gen chain) chain {
	return func(records ...string) {
		protocol = spec(protocol)
		if protocol != protocolNone {
			for i := range records {
				records[i] += "._" + protocol + "." + framework
			}
		} else {
			records = append(records, records...)
			for i, j := 0, len(records)/2; j < len(records); {
				records[i] += "._tcp." + framework
				records[j] += "._udp." + framework
				i++
				j++
			}
		}
		gen(records...)
	}
}

// withSubdomains appends `.{subdomain}` (for each subdomain spec'd) to records.
// the empty subdomain "" indicates to generate records w/o a subdomain fragment.
func withSubdomains(subdomains []string, gen chain) chain {
	if len(subdomains) == 0 {
		return gen
	}
	return func(records ...string) {
		var (
			recordLen = len(records)
			tmp       = make([]string, recordLen*len(subdomains))
			offset    = 0
		)
		for s := range subdomains {
			if subdomains[s] == domainNone {
				copy(tmp[offset:], records)
			} else {
				for i := range records {
					tmp[offset+i] = records[i] + "." + subdomains[s]
				}
			}
			offset += recordLen
		}
		gen(tmp...)
	}
}

// withNamedPort prepends a `_{discoveryInfo port name}.` to records
func withNamedPort(portName string, spec labels.Func, gen chain) chain {
	portName = spec(portName)
	if portName == "" {
		return gen
	}
	return func(records ...string) {
		// generate without port-name prefix
		gen(records...)

		// generate with port-name prefix
		for i := range records {
			records[i] = "_" + portName + "." + records[i]
		}
		gen(records...)
	}
}
