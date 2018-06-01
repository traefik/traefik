package hostresolver

import (
	"net"
	"strconv"
	"strings"
	"time"

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

// NewHostResolver init host resolver
func NewHostResolver(enabled bool, resConfig string, resDepth int) *HostResolver {
	return &HostResolver{
		Enabled:      enabled,
		ResolvConfig: resConfig,
		ResolvDepth:  resDepth,
		Cache:        cache.New(cache.DefaultExpiration, 5*time.Minute),
	}
}

// NewDefaultHostResolver init default host resolver
func NewDefaultHostResolver() *HostResolver {
	return &HostResolver{
		Enabled:      false,
		ResolvConfig: "/etc/resolv.conf",
		ResolvDepth:  5,
		Cache:        cache.New(cache.DefaultExpiration, 5*time.Minute),
	}
}

// CNAMEFlatten check if CNAME records is exist, flatten if possible
func (hr *HostResolver) CNAMEFlatten(host string) (string, string) {
	var result []string
	result = append(result, host)
	if hr.Cache == nil {
		hr.Cache = cache.New(cache.DefaultExpiration, 5*time.Minute)
	}
	rst, found := hr.Cache.Get(host)
	if found {
		result = strings.Split(rst.(string), ",")
	} else {
		config, _ := dns.ClientConfigFromFile(hr.ResolvConfig)
		c := new(dns.Client)
		c.Timeout = 30 * time.Second
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)
		var cacheDuration = 0 * time.Second
		for i := 0; i < hr.ResolvDepth; i++ {
			r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
			if err != nil {
				i--
				continue
			}
			if r != nil && len(r.Answer) > 0 {
				temp := strings.Fields(r.Answer[0].String())
				str := strings.TrimSuffix(strings.TrimSpace(temp[len(temp)-1]), ".")
				result = append(result, str)
				if i == 0 {
					ttl, err := strconv.Atoi(temp[1])
					if err != nil {
						cacheDuration = time.Duration(ttl) * time.Second
					}
				}
				m.SetQuestion(dns.Fqdn(str), dns.TypeCNAME)
			} else {
				break
			}
		}
		hr.Cache.Add(host, strings.Join(result, ","), cacheDuration)
	}
	return result[0], result[len(result)-1]
}
