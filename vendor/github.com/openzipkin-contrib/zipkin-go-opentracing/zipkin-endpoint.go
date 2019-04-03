package zipkintracer

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"

	"github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

// makeEndpoint takes the hostport and service name that represent this Zipkin
// service, and returns an endpoint that's embedded into the Zipkin core Span
// type. It will return a nil endpoint if the input parameters are malformed.
func makeEndpoint(hostport, serviceName string) (ep *zipkincore.Endpoint) {
	ep = zipkincore.NewEndpoint()

	// Set the ServiceName
	ep.ServiceName = serviceName

	if strings.IndexByte(hostport, ':') < 0 {
		// "<host>" becomes "<host>:0"
		hostport = hostport + ":0"
	}

	// try to parse provided "<host>:<port>"
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		// if unparsable, return as "undefined:0"
		return
	}

	// try to set port number
	p, _ := strconv.ParseUint(port, 10, 16)
	ep.Port = int16(p)

	// if <host> is a domain name, look it up
	addrs, err := net.LookupIP(host)
	if err != nil {
		// return as "undefined:<port>"
		return
	}

	var addr4, addr16 net.IP
	for i := range addrs {
		addr := addrs[i].To4()
		if addr == nil {
			// IPv6
			if addr16 == nil {
				addr16 = addrs[i].To16() // IPv6 - 16 bytes
			}
		} else {
			// IPv4
			if addr4 == nil {
				addr4 = addr // IPv4 - 4 bytes
			}
		}
		if addr16 != nil && addr4 != nil {
			// IPv4 & IPv6 have been set, we can stop looking further
			break
		}
	}
	// default to 0 filled 4 byte array for IPv4 if IPv6 only host was found
	if addr4 == nil {
		addr4 = make([]byte, 4)
	}

	// set IPv4 and IPv6 addresses
	ep.Ipv4 = (int32)(binary.BigEndian.Uint32(addr4))
	ep.Ipv6 = []byte(addr16)
	return
}
