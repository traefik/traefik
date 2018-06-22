package hostresolver

import (
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/containous/traefik/log"
	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

// HostResolver used for host resolver
type HostResolver struct {
	Enabled      bool
	Cache        *cache.Cache
	ResolvConfig string
	ResolvDepth  int
}

// CNAMEResolv used for storing CNAME result
type CNAMEResolv struct {
	TTL    int
	Record string
}

// CNAMEFlatten check if CNAME records is exists, flatten if possible
func (hr *HostResolver) CNAMEFlatten(host string) (string, string) {
	var result []string
	result = append(result, host)
	request := host
	if hr.Cache == nil {
		hr.Cache = cache.New(30*time.Minute, 5*time.Minute)
	}
	rst, found := hr.Cache.Get(host)
	if found {
		result = strings.Split(rst.(string), ",")
	} else {
		var cacheDuration = 0 * time.Second
		for i := 0; i < hr.ResolvDepth; i++ {
			r := hr.CNAMEResolve(request)
			if r != nil {
				result = append(result, r.Record)
				if i == 0 {
					cacheDuration = time.Duration(r.TTL) * time.Second
				}
				request = r.Record
			} else {
				break
			}
		}
		hr.Cache.Add(host, strings.Join(result, ","), cacheDuration)
	}
	return result[0], result[len(result)-1]
}

// CNAMEResolve resolve CNAME if exists, and return with the highest TTL
func (hr *HostResolver) CNAMEResolve(host string) *CNAMEResolv {
	config, _ := dns.ClientConfigFromFile(hr.ResolvConfig)
	c := dns.Client{Timeout: 30 * time.Second}
	c.Timeout = 30 * time.Second
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)
	var result []*CNAMEResolv
	for i := 0; i < len(config.Servers); i++ {
		r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[i], config.Port))
		if err != nil {
			log.Errorf("Failed to resolve host %s with server %s", host, config.Servers[i])
		}
		if r != nil && len(r.Answer) > 0 {
			temp := strings.Fields(r.Answer[0].String())
			ttl, _ := strconv.Atoi(temp[1])
			tempRecord := &CNAMEResolv{
				TTL:    ttl,
				Record: strings.TrimSuffix(strings.TrimSpace(temp[len(temp)-1]), "."),
			}
			result = append(result, tempRecord)
		}
	}
	if len(result) > 0 {
		sort.Slice(result, func(i, j int) bool {
			return result[i].TTL > result[j].TTL
		})
		return result[0]
	}
	return nil
}
