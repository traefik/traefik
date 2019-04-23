package sakuracloud

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
)

const sacloudAPILockKey = "lego/dns/sacloud"

func (d *DNSProvider) addTXTRecord(fqdn, domain, value string, ttl int) error {
	sacloud.LockByKey(sacloudAPILockKey)
	defer sacloud.UnlockByKey(sacloudAPILockKey)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	name := d.extractRecordName(fqdn, zone.Name)

	zone.AddRecord(zone.CreateNewRecord(name, "TXT", value, ttl))
	_, err = d.client.Update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}

	return nil
}

func (d *DNSProvider) cleanupTXTRecord(fqdn, domain string) error {
	sacloud.LockByKey(sacloudAPILockKey)
	defer sacloud.UnlockByKey(sacloudAPILockKey)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	records := d.findTxtRecords(fqdn, zone)

	for _, record := range records {
		var updRecords []sacloud.DNSRecordSet
		for _, r := range zone.Settings.DNS.ResourceRecordSets {
			if !(r.Name == record.Name && r.Type == record.Type && r.RData == record.RData) {
				updRecords = append(updRecords, r)
			}
		}
		zone.Settings.DNS.ResourceRecordSets = updRecords
	}

	_, err = d.client.Update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (*sacloud.DNS, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return nil, err
	}

	zoneName := dns01.UnFqdn(authZone)

	res, err := d.client.Reset().WithNameLike(zoneName).Find()
	if err != nil {
		if notFound, ok := err.(api.Error); ok && notFound.ResponseCode() == http.StatusNotFound {
			return nil, fmt.Errorf("zone %s not found on SakuraCloud DNS: %v", zoneName, err)
		}
		return nil, fmt.Errorf("API call failed: %v", err)
	}

	for _, zone := range res.CommonServiceDNSItems {
		if zone.Name == zoneName {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found", zoneName)
}

func (d *DNSProvider) findTxtRecords(fqdn string, zone *sacloud.DNS) []sacloud.DNSRecordSet {
	recordName := d.extractRecordName(fqdn, zone.Name)

	var res []sacloud.DNSRecordSet
	for _, record := range zone.Settings.DNS.ResourceRecordSets {
		if record.Name == recordName && record.Type == "TXT" {
			res = append(res, record)
		}
	}
	return res
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
