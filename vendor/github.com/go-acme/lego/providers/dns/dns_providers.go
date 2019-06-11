package dns

import (
	"fmt"

	"github.com/go-acme/lego/challenge"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/providers/dns/acmedns"
	"github.com/go-acme/lego/providers/dns/alidns"
	"github.com/go-acme/lego/providers/dns/auroradns"
	"github.com/go-acme/lego/providers/dns/azure"
	"github.com/go-acme/lego/providers/dns/bindman"
	"github.com/go-acme/lego/providers/dns/bluecat"
	"github.com/go-acme/lego/providers/dns/cloudflare"
	"github.com/go-acme/lego/providers/dns/cloudns"
	"github.com/go-acme/lego/providers/dns/cloudxns"
	"github.com/go-acme/lego/providers/dns/conoha"
	"github.com/go-acme/lego/providers/dns/designate"
	"github.com/go-acme/lego/providers/dns/digitalocean"
	"github.com/go-acme/lego/providers/dns/dnsimple"
	"github.com/go-acme/lego/providers/dns/dnsmadeeasy"
	"github.com/go-acme/lego/providers/dns/dnspod"
	"github.com/go-acme/lego/providers/dns/dode"
	"github.com/go-acme/lego/providers/dns/dreamhost"
	"github.com/go-acme/lego/providers/dns/duckdns"
	"github.com/go-acme/lego/providers/dns/dyn"
	"github.com/go-acme/lego/providers/dns/easydns"
	"github.com/go-acme/lego/providers/dns/exec"
	"github.com/go-acme/lego/providers/dns/exoscale"
	"github.com/go-acme/lego/providers/dns/fastdns"
	"github.com/go-acme/lego/providers/dns/gandi"
	"github.com/go-acme/lego/providers/dns/gandiv5"
	"github.com/go-acme/lego/providers/dns/gcloud"
	"github.com/go-acme/lego/providers/dns/glesys"
	"github.com/go-acme/lego/providers/dns/godaddy"
	"github.com/go-acme/lego/providers/dns/hostingde"
	"github.com/go-acme/lego/providers/dns/httpreq"
	"github.com/go-acme/lego/providers/dns/iij"
	"github.com/go-acme/lego/providers/dns/inwx"
	"github.com/go-acme/lego/providers/dns/joker"
	"github.com/go-acme/lego/providers/dns/lightsail"
	"github.com/go-acme/lego/providers/dns/linode"
	"github.com/go-acme/lego/providers/dns/linodev4"
	"github.com/go-acme/lego/providers/dns/mydnsjp"
	"github.com/go-acme/lego/providers/dns/namecheap"
	"github.com/go-acme/lego/providers/dns/namedotcom"
	"github.com/go-acme/lego/providers/dns/netcup"
	"github.com/go-acme/lego/providers/dns/nifcloud"
	"github.com/go-acme/lego/providers/dns/ns1"
	"github.com/go-acme/lego/providers/dns/oraclecloud"
	"github.com/go-acme/lego/providers/dns/otc"
	"github.com/go-acme/lego/providers/dns/ovh"
	"github.com/go-acme/lego/providers/dns/pdns"
	"github.com/go-acme/lego/providers/dns/rackspace"
	"github.com/go-acme/lego/providers/dns/rfc2136"
	"github.com/go-acme/lego/providers/dns/route53"
	"github.com/go-acme/lego/providers/dns/sakuracloud"
	"github.com/go-acme/lego/providers/dns/selectel"
	"github.com/go-acme/lego/providers/dns/stackpath"
	"github.com/go-acme/lego/providers/dns/transip"
	"github.com/go-acme/lego/providers/dns/vegadns"
	"github.com/go-acme/lego/providers/dns/vscale"
	"github.com/go-acme/lego/providers/dns/vultr"
	"github.com/go-acme/lego/providers/dns/zoneee"
)

// NewDNSChallengeProviderByName Factory for DNS providers
func NewDNSChallengeProviderByName(name string) (challenge.Provider, error) {
	switch name {
	case "acme-dns":
		return acmedns.NewDNSProvider()
	case "alidns":
		return alidns.NewDNSProvider()
	case "azure":
		return azure.NewDNSProvider()
	case "auroradns":
		return auroradns.NewDNSProvider()
	case "bindman":
		return bindman.NewDNSProvider()
	case "bluecat":
		return bluecat.NewDNSProvider()
	case "cloudflare":
		return cloudflare.NewDNSProvider()
	case "cloudns":
		return cloudns.NewDNSProvider()
	case "cloudxns":
		return cloudxns.NewDNSProvider()
	case "conoha":
		return conoha.NewDNSProvider()
	case "designate":
		return designate.NewDNSProvider()
	case "digitalocean":
		return digitalocean.NewDNSProvider()
	case "dnsimple":
		return dnsimple.NewDNSProvider()
	case "dnsmadeeasy":
		return dnsmadeeasy.NewDNSProvider()
	case "dnspod":
		return dnspod.NewDNSProvider()
	case "dode":
		return dode.NewDNSProvider()
	case "dreamhost":
		return dreamhost.NewDNSProvider()
	case "duckdns":
		return duckdns.NewDNSProvider()
	case "dyn":
		return dyn.NewDNSProvider()
	case "fastdns":
		return fastdns.NewDNSProvider()
	case "easydns":
		return easydns.NewDNSProvider()
	case "exec":
		return exec.NewDNSProvider()
	case "exoscale":
		return exoscale.NewDNSProvider()
	case "gandi":
		return gandi.NewDNSProvider()
	case "gandiv5":
		return gandiv5.NewDNSProvider()
	case "glesys":
		return glesys.NewDNSProvider()
	case "gcloud":
		return gcloud.NewDNSProvider()
	case "godaddy":
		return godaddy.NewDNSProvider()
	case "hostingde":
		return hostingde.NewDNSProvider()
	case "httpreq":
		return httpreq.NewDNSProvider()
	case "iij":
		return iij.NewDNSProvider()
	case "inwx":
		return inwx.NewDNSProvider()
	case "joker":
		return joker.NewDNSProvider()
	case "lightsail":
		return lightsail.NewDNSProvider()
	case "linode":
		return linode.NewDNSProvider()
	case "linodev4":
		return linodev4.NewDNSProvider()
	case "manual":
		return dns01.NewDNSProviderManual()
	case "mydnsjp":
		return mydnsjp.NewDNSProvider()
	case "namecheap":
		return namecheap.NewDNSProvider()
	case "namedotcom":
		return namedotcom.NewDNSProvider()
	case "netcup":
		return netcup.NewDNSProvider()
	case "nifcloud":
		return nifcloud.NewDNSProvider()
	case "ns1":
		return ns1.NewDNSProvider()
	case "oraclecloud":
		return oraclecloud.NewDNSProvider()
	case "otc":
		return otc.NewDNSProvider()
	case "ovh":
		return ovh.NewDNSProvider()
	case "pdns":
		return pdns.NewDNSProvider()
	case "rackspace":
		return rackspace.NewDNSProvider()
	case "route53":
		return route53.NewDNSProvider()
	case "rfc2136":
		return rfc2136.NewDNSProvider()
	case "sakuracloud":
		return sakuracloud.NewDNSProvider()
	case "stackpath":
		return stackpath.NewDNSProvider()
	case "selectel":
		return selectel.NewDNSProvider()
	case "transip":
		return transip.NewDNSProvider()
	case "vegadns":
		return vegadns.NewDNSProvider()
	case "vultr":
		return vultr.NewDNSProvider()
	case "vscale":
		return vscale.NewDNSProvider()
	case "zoneee":
		return zoneee.NewDNSProvider()
	default:
		return nil, fmt.Errorf("unrecognised DNS provider: %s", name)
	}
}
