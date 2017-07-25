package command

import (
	"github.com/codegangsta/cli"
	"github.com/vulcand/vulcand/engine"
)

func getTLSFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{Name: "tlsSkipVerify", Usage: "insecure: skip certificate verification"},
		cli.BoolFlag{Name: "tlsPreferServerCS", Usage: "prefer server cipher suites, recommended on for listener settings"},
		cli.BoolFlag{Name: "tlsSessionTicketsOff", Usage: "turns off TLS session tickets"},
		cli.StringFlag{Name: "tlsMinV", Usage: "minimum supported TLS version"},
		cli.StringFlag{Name: "tlsMaxV", Usage: "maximum supported TLS version"},
		cli.StringFlag{Name: "tlsSessionCache", Usage: "session cache type"},
		cli.IntFlag{Name: "tlsSessionCacheCapacity", Usage: "session cache capacity"},
		cli.StringSliceFlag{Name: "tlsCS", Usage: "optional list of preferred cipher suites", Value: &cli.StringSlice{}},
	}
}

func getTLSSettings(c *cli.Context) (*engine.TLSSettings, error) {
	s := &engine.TLSSettings{
		InsecureSkipVerify:       c.Bool("tlsSkipVerify"),
		PreferServerCipherSuites: c.Bool("tlsPreferServerCS"),
		SessionTicketsDisabled:   c.Bool("tlsSessionTicketsOff"),
		MinVersion:               c.String("tlsMinV"),
		MaxVersion:               c.String("tlsMaxV"),
		CipherSuites:             c.StringSlice("tlsCS"),
	}
	s.SessionCache.Type = c.String("tlsSessionCache")
	if s.SessionCache.Type == engine.LRUCacheType {
		s.SessionCache.Settings = &engine.LRUSessionCacheSettings{
			Capacity: c.Int("tlsSessionCacheCapacity"),
		}
	}
	if _, err := engine.NewTLSConfig(s); err != nil {
		return nil, err
	}
	return s, nil
}
