package middlewares

import (
	"encoding/json"
	"log"
	"time"
	"net"
	"net/http"

	"github.com/gorilla/mux"
    "github.com/miekg/dns"
     "github.com/karlseguin/ccache"
)

// Routes holds the gorilla mux routes (for the API & co).
// It also does CNAME flattening, so Traefik can handle requests that come from other domains that use
// Traefik as the canonical domain
type Routes struct {
	router *mux.Router
    dnsCache *ccache.Cache
    dnsClient *dns.Client
    nameServerAddress string
}

// NewRoutes return a Routes based on the given router.
func NewRoutes(router *mux.Router, resolvConf string) *Routes {
    config, err := dns.ClientConfigFromFile(resolvConf)
    if err != nil {
		log.Fatal("Error reading resolv.conf file. Could not configure nameservers: ", err)
    }
    nameServerAddress := net.JoinHostPort(config.Servers[0], config.Port)
    dnsCache := ccache.New(ccache.Configure().MaxSize(1000).ItemsToPrune(100))
	return &Routes{router, dnsCache, new(dns.Client), nameServerAddress}
}

func (router *Routes) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	routeMatch := mux.RouteMatch{}
	if router.router.Match(r, &routeMatch) {
		json, _ := json.Marshal(routeMatch.Handler)
		log.Println("Request match route ", json)
    } else {
        host := router.dnsCache.Get(r.Host)
        if (host == nil || host.TTL() < 0) {
            newName, ttl, err := router.lookupCNAME(r.Host)
            if err == nil {
                if newName != nil {
                    router.dnsCache.Set(r.Host, newName, ttl)
                } else {
                    router.dnsCache.Set(r.Host, nil, 60 * time.Second) // 60 second TTL for domains with no CNAME
                }
            }
        }
        if (host != nil) {
            r.Host = host.Value().(string) // Rewrite the host header
        }
    }
	next(rw, r)
}

func (router *Routes) lookupCNAME(host string) (*string, time.Duration, error) {
    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)
    m.RecursionDesired = true

    r, _, err := router.dnsClient.Exchange(m, router.nameServerAddress)
    if r == nil {
        log.Println("DNS lookup failed for %s: %s\n", host, err.Error())
        return nil, 0, err
    }
    if r.Rcode != dns.RcodeSuccess {
        log.Println("Invalid answer in CNAME query for %s\n", host)
        return nil, 0, err
    }
    // Stuff must be in the answer section
    for _, ans := range r.Answer {
        return &ans.Header().Name, time.Duration(ans.Header().Ttl) * time.Second, nil
    }
    return nil, 0, nil
}
