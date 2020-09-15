package auth

import (
	"net"
	"net/http"
)

type ipAllowList struct {
	ipnets                []*net.IPNet
	ips                   []net.IP
	clientIPSourceHeaders []string
}

func newIPAllowList(ips []string, clientIPHeaders []string) ipAllowList {
	list := ipAllowList{
		clientIPSourceHeaders: clientIPHeaders,
	}

	for _, ip := range ips {
		_, ipnet, err := net.ParseCIDR(ip)
		if err == nil {
			list.ipnets = append(list.ipnets, ipnet)
			continue
		}

		if ip := net.ParseIP(ip); ip != nil {
			list.ips = append(list.ips, ip)
		}
	}

	return list
}

func (l *ipAllowList) Check(r *http.Request) bool {
	if len(l.clientIPSourceHeaders) == 0 {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return false
		}

		return l.check(net.ParseIP(host))
	}

	for _, h := range l.clientIPSourceHeaders {
		if ip := r.Header.Get(h); ip != "" {
			if l.check(net.ParseIP(ip)) {
				return true
			}
		}
	}

	return false
}

func (l *ipAllowList) check(ip net.IP) bool {
	if ip == nil {
		return false
	}

	for _, ipNet := range l.ipnets {
		if ipNet.Contains(ip) {
			return true
		}
	}

	for _, ip := range l.ips {
		if ip.Equal(ip) {
			return true
		}
	}

	return false
}
