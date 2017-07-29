package dns

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/ns1/ns1-go.v2/rest/model/data"
)

func TestUnmarshalZones(t *testing.T) {
	d := []byte(`[
   {  
      "nx_ttl":3600,
      "retry":7200,
      "zone":"test.zone",
      "network_pools":[  
         "p09"
      ],
      "primary":{  
         "enabled":true,
         "secondaries":[  
            {  
               "ip":"1.1.1.1",
               "notify":true,
               "networks":[  

               ],
               "port":53
            },
            {  
               "ip":"2.2.2.2",
               "notify":true,
               "networks":[  

               ],
               "port":53
            }
         ]
      },
      "refresh":43200,
      "expiry":1209600,
      "dns_servers":[  
         "dns1.p09.nsone.net",
         "dns2.p09.nsone.net",
         "dns3.p09.nsone.net",
         "dns4.p09.nsone.net"
      ],
      "meta":{  

      },
      "link":null,
      "serial":1473863358,
      "ttl":3600,
      "id":"57d95da659272400013334de",
      "hostmaster":"hostmaster@nsone.net",
      "networks":[  
         0
      ],
      "pool":"p09"
   },
   {  
	   "nx_ttl":3600,
	   "retry":7200,
	   "zone":"secondary.zone",
	   "network_pools":[  
	      "p09"
	   ],
	   "secondary":{  
	      "status":"pending",
	      "last_xfr":0,
	      "primary_ip":"1.1.1.1",
	      "primary_port":53,
	      "enabled":true,
	      "tsig":{  
		 "enabled":false,
		 "hash":null,
		 "name":null,
		 "key":null
	      },
	      "error":null,
	      "expired":false
	   },
	   "primary":{  
	      "enabled":false,
	      "secondaries":[  

	      ]
	   },
	   "refresh":43200,
	   "expiry":1209600,
	   "dns_servers":[  
	      "dns1.p09.nsone.net",
	      "dns2.p09.nsone.net",
	      "dns3.p09.nsone.net",
	      "dns4.p09.nsone.net"
	   ],
	   "meta":{  

	   },
	   "link":null,
	   "serial":1473868413,
	   "ttl":3600,
	   "id":"57d9727d1c372700011eff6e",
	   "hostmaster":"hostmaster@nsone.net",
	   "networks":[  
	      0
	   ],
	   "pool":"p09"
	},
   {  
      "nx_ttl":3600,
      "retry":7200,
      "zone":"myfailover.com",
      "network_pools":[  
         "p09"
      ],
      "primary":{  
         "enabled":true,
         "secondaries":[  

         ]
      },
      "refresh":43200,
      "expiry":1209600,
      "dns_servers":[  
         "dns1.p09.nsone.net",
         "dns2.p09.nsone.net",
         "dns3.p09.nsone.net",
         "dns4.p09.nsone.net"
      ],
      "meta":{  

      },
      "link":null,
      "serial":1473813629,
      "ttl":3600,
      "id":"57d89c7b1c372700011e0a97",
      "hostmaster":"hostmaster@nsone.net",
      "networks":[  
         0
      ],
      "pool":"p09"
   }
]
`)
	zl := []*Zone{}
	if err := json.Unmarshal(d, &zl); err != nil {
		t.Error(err)
	}
	if len(zl) != 3 {
		fmt.Println(zl)
		t.Error("Do not have 3 zones in list")
	}
	z := zl[0]
	assert.Nil(t, z.Link)
	assert.Nil(t, z.Secondary)
	assert.Equal(t, z.ID, "57d95da659272400013334de", "Wrong zone id")
	assert.Equal(t, z.Zone, "test.zone", "Wrong zone name")
	assert.Equal(t, z.TTL, 3600, "Wrong zone ttl")
	assert.Equal(t, z.NxTTL, 3600, "Wrong zone nxttl")
	assert.Equal(t, z.Retry, 7200, "Wrong zone retry")
	assert.Equal(t, z.Serial, 1473863358, "Wrong zone serial")
	assert.Equal(t, z.Refresh, 43200, "Wrong zone refresh")
	assert.Equal(t, z.Expiry, 1209600, "Wrong zone expiry")
	assert.Equal(t, z.Hostmaster, "hostmaster@nsone.net", "Wrong zone hostmaster")
	assert.Equal(t, z.Pool, "p09", "Wrong zone pool")
	assert.Equal(t, z.NetworkIDs, []int{0}, "Wrong zone network")
	assert.Equal(t, z.NetworkPools, []string{"p09"}, "Wrong zone network pools")
	assert.Equal(t, z.Meta, &data.Meta{}, "Zone meta should be empty")

	dnsServers := []string{
		"dns1.p09.nsone.net",
		"dns2.p09.nsone.net",
		"dns3.p09.nsone.net",
		"dns4.p09.nsone.net",
	}
	assert.Equal(t, z.DNSServers, dnsServers, "Wrong zone dns networks")

	primary := &ZonePrimary{
		Enabled: true,
		Secondaries: []ZoneSecondaryServer{
			ZoneSecondaryServer{
				IP:         "1.1.1.1",
				Port:       53,
				Notify:     true,
				NetworkIDs: []int{},
			},
			ZoneSecondaryServer{
				IP:         "2.2.2.2",
				Port:       53,
				Notify:     true,
				NetworkIDs: []int{},
			},
		},
	}
	assert.Equal(t, z.Primary, primary, "Wrong zone primary")

	// Check zone with secondaries
	secZ := zl[1]
	assert.Nil(t, secZ.Link)
	assert.Equal(t, secZ.Zone, "secondary.zone", "Wrong zone name")
	assert.Equal(t, secZ.Primary, &ZonePrimary{
		Enabled:     false,
		Secondaries: []ZoneSecondaryServer{}}, "Wrong zone secondary primary")

	secondary := secZ.Secondary
	assert.Nil(t, secondary.Error)
	assert.Equal(t, secondary.Status, "pending", "Wrong zone secondary status")
	assert.Equal(t, secondary.LastXfr, 0, "Wrong zone secondary last xfr")
	assert.Equal(t, secondary.PrimaryIP, "1.1.1.1", "Wrong zone secondary primary ip")
	assert.Equal(t, secondary.PrimaryPort, 53, "Wrong zone secondary primary port")
	assert.Equal(t, secondary.Enabled, true, "Wrong zone secondary enabled")

	assert.Equal(t, secondary.TSIG, &TSIG{
		Enabled: false,
		Hash:    "",
		Name:    "",
		Key:     ""}, "Wrong zone secondary tsig")
}
