package requestdecorator

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

type cnameResolv struct {
	TTL    time.Duration
	Record string
}

type byTTL []*cnameResolv

func (a byTTL) Len() int           { return len(a) }
func (a byTTL) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTTL) Less(i, j int) bool { return a[i].TTL > a[j].TTL }

// Resolver used for host resolver.
type Resolver struct {
	CnameFlattening bool
	ResolvConfig    string
	ResolvDepth     int
	cache           *cache.Cache
}

// CNAMEFlatten check if CNAME record exists, flatten if possible.
func (hr *Resolver) CNAMEFlatten(ctx context.Context, host string) string {
	if hr.cache == nil {
		hr.cache = cache.New(30*time.Minute, 5*time.Minute)
	}

	result := host
	request := host

	value, found := hr.cache.Get(host)
	if found {
		return value.(string)
	}

	logger := log.Ctx(ctx)
	cacheDuration := 0 * time.Second
	for depth := 0; depth < hr.ResolvDepth; depth++ {
		resolv, err := cnameResolve(ctx, request, hr.ResolvConfig)
		if err != nil {
			logger.Error().Err(err).Send()
			break
		}
		if resolv == nil {
			break
		}

		result = resolv.Record
		if depth == 0 {
			cacheDuration = resolv.TTL
		}
		request = resolv.Record
	}

	if err := hr.cache.Add(host, result, cacheDuration); err != nil {
		logger.Error().Err(err).Send()
	}

	return result
}

// cnameResolve resolves CNAME if exists, and return with the highest TTL.
func cnameResolve(ctx context.Context, host, resolvPath string) (*cnameResolv, error) {
	config, err := dns.ClientConfigFromFile(resolvPath)
	if err != nil {
		return nil, fmt.Errorf("invalid resolver configuration file: %s", resolvPath)
	}

	client := &dns.Client{Timeout: 30 * time.Second}

	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)

	var result []*cnameResolv
	for _, server := range config.Servers {
		tempRecord, err := getRecord(client, m, server, config.Port)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Failed to resolve host %s", host)
			continue
		}
		result = append(result, tempRecord)
	}

	if len(result) == 0 {
		return nil, nil
	}

	sort.Sort(byTTL(result))
	return result[0], nil
}

func getRecord(client *dns.Client, msg *dns.Msg, server, port string) (*cnameResolv, error) {
	resp, _, err := client.Exchange(msg, net.JoinHostPort(server, port))
	if err != nil {
		return nil, fmt.Errorf("exchange error for server %s: %w", server, err)
	}

	if resp == nil || len(resp.Answer) == 0 {
		return nil, fmt.Errorf("empty answer for server %s", server)
	}

	rr, ok := resp.Answer[0].(*dns.CNAME)
	if !ok {
		return nil, fmt.Errorf("invalid response type for server %s", server)
	}

	return &cnameResolv{
		TTL:    time.Duration(rr.Hdr.Ttl) * time.Second,
		Record: strings.TrimSuffix(rr.Target, "."),
	}, nil
}
