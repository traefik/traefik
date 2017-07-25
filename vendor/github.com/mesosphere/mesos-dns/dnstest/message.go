package dnstest

import (
	"net"

	"github.com/miekg/dns"
)

// A MsgOpt is functional option for dns.Msgs
type MsgOpt func(*dns.Msg)

// Message returns a dns.Msg with the given MsgOpts applied to it.
func Message(opts ...MsgOpt) *dns.Msg {
	var m dns.Msg
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

// Header returns a MsgOpt that sets a dns.Msg's MsgHdr with the given arguments
// and some hard-coded defaults.
func Header(auth bool, rcode int) MsgOpt {
	return func(m *dns.Msg) {
		m.Authoritative = auth
		m.Response = true
		m.Rcode = rcode
		m.Compress = true
	}
}

// Question returns a MsgOpt that sets the Question section in a Msg.
func Question(name string, qtype uint16) MsgOpt {
	return func(m *dns.Msg) { m.SetQuestion(name, qtype) }
}

// Answers returns a MsgOpt that appends the given dns.RRs to a Msg's Answer
// section.
func Answers(rrs ...dns.RR) MsgOpt {
	return func(m *dns.Msg) { m.Answer = append(m.Answer, rrs...) }
}

// NSs returns a MsgOpt that appends the given dns.RRs to a Msg's Ns section.
func NSs(rrs ...dns.RR) MsgOpt {
	return func(m *dns.Msg) { m.Ns = append(m.Ns, rrs...) }
}

// Extras returns a MsgOpt that appends the given dns.RRs to a Msg's Extras
// section.
func Extras(rrs ...dns.RR) MsgOpt {
	return func(m *dns.Msg) { m.Extra = append(m.Extra, rrs...) }
}

// RRHeader returns a dns.RR_Header with the given arguments set as well as a
// few hard-coded defaults.
func RRHeader(name string, rrtype uint16, ttl uint32) dns.RR_Header {
	return dns.RR_Header{
		Name:   name,
		Rrtype: rrtype,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
}

// A returns an A record set with the given arguments.
func A(hdr dns.RR_Header, ip net.IP) *dns.A {
	return &dns.A{
		Hdr: hdr,
		A:   ip.To4(),
	}
}

// SRV returns a SRV record set with the given arguments.
func SRV(hdr dns.RR_Header, target string, port, priority, weight uint16) *dns.SRV {
	return &dns.SRV{
		Hdr:      hdr,
		Target:   target,
		Port:     port,
		Priority: priority,
		Weight:   weight,
	}
}

// NS returns a NS record set with the given arguments.
func NS(hdr dns.RR_Header, ns string) *dns.NS {
	return &dns.NS{
		Hdr: hdr,
		Ns:  ns,
	}
}

// SOA returns an SOA records set with the given arguments and some hard-coded
// defaults.
func SOA(hdr dns.RR_Header, ns, mbox string, minttl uint32) *dns.SOA {
	return &dns.SOA{
		Hdr:     hdr,
		Ns:      ns,
		Mbox:    mbox,
		Minttl:  minttl,
		Refresh: 60,
		Retry:   600,
		Expire:  86400,
	}
}
