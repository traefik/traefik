package hostresolver

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
)

// HostResolver used for host resolver
type HostResolver struct {
	Enabled       bool
	Cache         *cache.Cache
	CacheDuration time.Duration
	ResolvConfig  string
	ResolvDepth   int
}

// NewHostResolver init host resolver
func NewHostResolver(enabled bool, resConfig string, resDepth int, cacheDuration time.Duration) *HostResolver {
	return &HostResolver{
		Enabled:       enabled,
		CacheDuration: cacheDuration,
		ResolvConfig:  resConfig,
		ResolvDepth:   resDepth,
		Cache:         cache.New(cacheDuration, 3*cacheDuration),
	}
}

// NewDefaultHostResolver init default host resolver
func NewDefaultHostResolver() *HostResolver {
	return &HostResolver{
		Enabled:       false,
		CacheDuration: time.Minute * 30,
		ResolvConfig:  "/etc/resolv.conf",
		ResolvDepth:   5,
		Cache:         cache.New(time.Minute*30, 3*time.Minute*30),
	}
}

// CNAMEFlatten check if CNAME records is exist, flatten if possible
func (hr *HostResolver) CNAMEFlatten(host string) (string, string) {
	var result []string
	result = append(result, host)
	if hr.Cache == nil {
		hr.Cache = cache.New(hr.CacheDuration, 3*hr.CacheDuration)
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
		for i := 0; i < hr.ResolvDepth; i++ {
			r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
			if err != nil {
				i--
				continue
			}
			if r != nil && len(r.Answer) > 0 {
				temp := strings.Split(r.Answer[0].String(), "CNAME")
				str := strings.TrimSuffix(strings.TrimSpace(temp[len(temp)-1]), ".")
				result = append(result, str)
				m.SetQuestion(dns.Fqdn(str), dns.TypeCNAME)
			} else {
				break
			}
		}
		hr.Cache.Add(host, strings.Join(result, ","), cache.DefaultExpiration)
	}
	return result[0], result[len(result)-1]
}
