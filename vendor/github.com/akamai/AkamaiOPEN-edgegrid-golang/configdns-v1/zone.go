package dns

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
)

type name struct {
	recordType string
	name       string
}

var (
	cnameNames    []name
	nonCnameNames []name
	zoneWriteLock sync.Mutex
)

// Zone represents a DNS zone
type Zone struct {
	Token string `json:"token"`
	Zone  struct {
		Name       string              `json:"name,omitempty"`
		A          []*ARecord          `json:"a,omitempty"`
		Aaaa       []*AaaaRecord       `json:"aaaa,omitempty"`
		Afsdb      []*AfsdbRecord      `json:"afsdb,omitempty"`
		Cname      []*CnameRecord      `json:"cname,omitempty"`
		Dnskey     []*DnskeyRecord     `json:"dnskey,omitempty"`
		Ds         []*DsRecord         `json:"ds,omitempty"`
		Hinfo      []*HinfoRecord      `json:"hinfo,omitempty"`
		Loc        []*LocRecord        `json:"loc,omitempty"`
		Mx         []*MxRecord         `json:"mx,omitempty"`
		Naptr      []*NaptrRecord      `json:"naptr,omitempty"`
		Ns         []*NsRecord         `json:"ns,omitempty"`
		Nsec3      []*Nsec3Record      `json:"nsec3,omitempty"`
		Nsec3param []*Nsec3paramRecord `json:"nsec3param,omitempty"`
		Ptr        []*PtrRecord        `json:"ptr,omitempty"`
		Rp         []*RpRecord         `json:"rp,omitempty"`
		Rrsig      []*RrsigRecord      `json:"rrsig,omitempty"`
		Soa        *SoaRecord          `json:"soa,omitempty"`
		Spf        []*SpfRecord        `json:"spf,omitempty"`
		Srv        []*SrvRecord        `json:"srv,omitempty"`
		Sshfp      []*SshfpRecord      `json:"sshfp,omitempty"`
		Txt        []*TxtRecord        `json:"txt,omitempty"`
	} `json:"zone"`
}

// NewZone creates a new Zone
func NewZone(hostname string) *Zone {
	zone := &Zone{Token: "new"}
	zone.Zone.Soa = NewSoaRecord()
	zone.Zone.Name = hostname
	return zone
}

// GetZone retrieves a DNS Zone for a given hostname
func GetZone(hostname string) (*Zone, error) {
	zone := NewZone(hostname)
	req, err := client.NewRequest(
		Config,
		"GET",
		"/config-dns/v1/zones/"+hostname,
		nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(Config, req)
	if err != nil {
		return nil, err
	}

	if client.IsError(res) && res.StatusCode != 404 {
		return nil, client.NewAPIError(res)
	} else if res.StatusCode == 404 {
		return nil, &ZoneError{zoneName: hostname}
	} else {
		err = client.BodyJSON(res, zone)
		if err != nil {
			return nil, err
		}

		return zone, nil
	}
}

// Save updates the Zone
func (zone *Zone) Save() error {
	// This lock will restrict the concurrency of API calls
	// to 1 save request at a time. This is needed for the Soa.Serial value which
	// is required to be incremented for every subsequent update to a zone
	// so we have to save just one request at a time to ensure this is always
	// incremented properly
	zoneWriteLock.Lock()
	defer zoneWriteLock.Unlock()

	valid, f := zone.validateCnames()
	if valid == false {
		var msg string
		for _, v := range f {
			msg = msg + fmt.Sprintf("\n%s Record '%s' conflicts with CNAME", v.recordType, v.name)
		}
		return &ZoneError{
			zoneName:        zone.Zone.Name,
			apiErrorMessage: "All CNAMEs must be unique in the zone" + msg,
		}
	}

	req, err := client.NewJSONRequest(
		Config,
		"POST",
		"/config-dns/v1/zones/"+zone.Zone.Name,
		zone,
	)
	if err != nil {
		return err
	}

	res, err := client.Do(Config, req)

	// Network error
	if err != nil {
		return &ZoneError{
			zoneName:         zone.Zone.Name,
			httpErrorMessage: err.Error(),
			err:              err,
		}
	}

	// API error
	if client.IsError(res) {
		err := client.NewAPIError(res)
		return &ZoneError{zoneName: zone.Zone.Name, apiErrorMessage: err.Detail, err: err}
	}

	for {
		updatedZone, err := GetZone(zone.Zone.Name)
		if err != nil {
			return err
		}

		if updatedZone.Token != zone.Token {
			*zone = *updatedZone
			break
		}
		time.Sleep(time.Second)
	}

	return nil
}

func (zone *Zone) Delete() error {
	// remove all the records except for SOA
	// which is required and save the zone
	zone.Zone.A = nil
	zone.Zone.Aaaa = nil
	zone.Zone.Afsdb = nil
	zone.Zone.Cname = nil
	zone.Zone.Dnskey = nil
	zone.Zone.Ds = nil
	zone.Zone.Hinfo = nil
	zone.Zone.Loc = nil
	zone.Zone.Mx = nil
	zone.Zone.Naptr = nil
	zone.Zone.Ns = nil
	zone.Zone.Nsec3 = nil
	zone.Zone.Nsec3param = nil
	zone.Zone.Ptr = nil
	zone.Zone.Rp = nil
	zone.Zone.Rrsig = nil
	zone.Zone.Spf = nil
	zone.Zone.Srv = nil
	zone.Zone.Sshfp = nil
	zone.Zone.Txt = nil

	return zone.Save()
}

func (zone *Zone) AddRecord(recordPtr interface{}) error {
	switch recordPtr.(type) {
	case *ARecord:
		zone.addARecord(recordPtr.(*ARecord))
	case *AaaaRecord:
		zone.addAaaaRecord(recordPtr.(*AaaaRecord))
	case *AfsdbRecord:
		zone.addAfsdbRecord(recordPtr.(*AfsdbRecord))
	case *CnameRecord:
		zone.addCnameRecord(recordPtr.(*CnameRecord))
	case *DnskeyRecord:
		zone.addDnskeyRecord(recordPtr.(*DnskeyRecord))
	case *DsRecord:
		zone.addDsRecord(recordPtr.(*DsRecord))
	case *HinfoRecord:
		zone.addHinfoRecord(recordPtr.(*HinfoRecord))
	case *LocRecord:
		zone.addLocRecord(recordPtr.(*LocRecord))
	case *MxRecord:
		zone.addMxRecord(recordPtr.(*MxRecord))
	case *NaptrRecord:
		zone.addNaptrRecord(recordPtr.(*NaptrRecord))
	case *NsRecord:
		zone.addNsRecord(recordPtr.(*NsRecord))
	case *Nsec3Record:
		zone.addNsec3Record(recordPtr.(*Nsec3Record))
	case *Nsec3paramRecord:
		zone.addNsec3paramRecord(recordPtr.(*Nsec3paramRecord))
	case *PtrRecord:
		zone.addPtrRecord(recordPtr.(*PtrRecord))
	case *RpRecord:
		zone.addRpRecord(recordPtr.(*RpRecord))
	case *RrsigRecord:
		zone.addRrsigRecord(recordPtr.(*RrsigRecord))
	case *SoaRecord:
		zone.addSoaRecord(recordPtr.(*SoaRecord))
	case *SpfRecord:
		zone.addSpfRecord(recordPtr.(*SpfRecord))
	case *SrvRecord:
		zone.addSrvRecord(recordPtr.(*SrvRecord))
	case *SshfpRecord:
		zone.addSshfpRecord(recordPtr.(*SshfpRecord))
	case *TxtRecord:
		zone.addTxtRecord(recordPtr.(*TxtRecord))
	}

	return nil
}

func (zone *Zone) RemoveRecord(recordPtr interface{}) error {
	switch recordPtr.(type) {
	case *ARecord:
		return zone.removeARecord(recordPtr.(*ARecord))
	case *AaaaRecord:
		return zone.removeAaaaRecord(recordPtr.(*AaaaRecord))
	case *AfsdbRecord:
		return zone.removeAfsdbRecord(recordPtr.(*AfsdbRecord))
	case *CnameRecord:
		return zone.removeCnameRecord(recordPtr.(*CnameRecord))
	case *DnskeyRecord:
		return zone.removeDnskeyRecord(recordPtr.(*DnskeyRecord))
	case *DsRecord:
		return zone.removeDsRecord(recordPtr.(*DsRecord))
	case *HinfoRecord:
		return zone.removeHinfoRecord(recordPtr.(*HinfoRecord))
	case *LocRecord:
		return zone.removeLocRecord(recordPtr.(*LocRecord))
	case *MxRecord:
		return zone.removeMxRecord(recordPtr.(*MxRecord))
	case *NaptrRecord:
		return zone.removeNaptrRecord(recordPtr.(*NaptrRecord))
	case *NsRecord:
		return zone.removeNsRecord(recordPtr.(*NsRecord))
	case *Nsec3Record:
		return zone.removeNsec3Record(recordPtr.(*Nsec3Record))
	case *Nsec3paramRecord:
		return zone.removeNsec3paramRecord(recordPtr.(*Nsec3paramRecord))
	case *PtrRecord:
		return zone.removePtrRecord(recordPtr.(*PtrRecord))
	case *RpRecord:
		return zone.removeRpRecord(recordPtr.(*RpRecord))
	case *RrsigRecord:
		return zone.removeRrsigRecord(recordPtr.(*RrsigRecord))
	case *SoaRecord:
		return zone.removeSoaRecord(recordPtr.(*SoaRecord))
	case *SpfRecord:
		return zone.removeSpfRecord(recordPtr.(*SpfRecord))
	case *SrvRecord:
		return zone.removeSrvRecord(recordPtr.(*SrvRecord))
	case *SshfpRecord:
		return zone.removeSshfpRecord(recordPtr.(*SshfpRecord))
	case *TxtRecord:
		return zone.removeTxtRecord(recordPtr.(*TxtRecord))
	}

	return nil
}

func (zone *Zone) addARecord(record *ARecord) {
	zone.Zone.A = append(zone.Zone.A, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "A", name: record.Name})
}

func (zone *Zone) addAaaaRecord(record *AaaaRecord) {
	zone.Zone.Aaaa = append(zone.Zone.Aaaa, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "AAAA", name: record.Name})
}

func (zone *Zone) addAfsdbRecord(record *AfsdbRecord) {
	zone.Zone.Afsdb = append(zone.Zone.Afsdb, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "AFSDB", name: record.Name})
}

func (zone *Zone) addCnameRecord(record *CnameRecord) {
	zone.Zone.Cname = append(zone.Zone.Cname, record)
	cnameNames = append(cnameNames, name{recordType: "CNAME", name: record.Name})
}

func (zone *Zone) addDnskeyRecord(record *DnskeyRecord) {
	zone.Zone.Dnskey = append(zone.Zone.Dnskey, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "DNSKEY", name: record.Name})
}

func (zone *Zone) addDsRecord(record *DsRecord) {
	zone.Zone.Ds = append(zone.Zone.Ds, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "DS", name: record.Name})
}

func (zone *Zone) addHinfoRecord(record *HinfoRecord) {
	zone.Zone.Hinfo = append(zone.Zone.Hinfo, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "HINFO", name: record.Name})
}

func (zone *Zone) addLocRecord(record *LocRecord) {
	zone.Zone.Loc = append(zone.Zone.Loc, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "LOC", name: record.Name})
}

func (zone *Zone) addMxRecord(record *MxRecord) {
	zone.Zone.Mx = append(zone.Zone.Mx, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "MX", name: record.Name})
}

func (zone *Zone) addNaptrRecord(record *NaptrRecord) {
	zone.Zone.Naptr = append(zone.Zone.Naptr, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "NAPTR", name: record.Name})
}

func (zone *Zone) addNsRecord(record *NsRecord) {
	zone.Zone.Ns = append(zone.Zone.Ns, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "NS", name: record.Name})
}

func (zone *Zone) addNsec3Record(record *Nsec3Record) {
	zone.Zone.Nsec3 = append(zone.Zone.Nsec3, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "NSEC3", name: record.Name})
}

func (zone *Zone) addNsec3paramRecord(record *Nsec3paramRecord) {
	zone.Zone.Nsec3param = append(zone.Zone.Nsec3param, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "NSEC3PARAM", name: record.Name})
}

func (zone *Zone) addPtrRecord(record *PtrRecord) {
	zone.Zone.Ptr = append(zone.Zone.Ptr, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "PTR", name: record.Name})
}

func (zone *Zone) addRpRecord(record *RpRecord) {
	zone.Zone.Rp = append(zone.Zone.Rp, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "RP", name: record.Name})
}

func (zone *Zone) addRrsigRecord(record *RrsigRecord) {
	zone.Zone.Rrsig = append(zone.Zone.Rrsig, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "RRSIG", name: record.Name})
}

func (zone *Zone) addSoaRecord(record *SoaRecord) {
	// Only one SOA records is allowed
	zone.Zone.Soa = record
}

func (zone *Zone) addSpfRecord(record *SpfRecord) {
	zone.Zone.Spf = append(zone.Zone.Spf, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "SPF", name: record.Name})
}

func (zone *Zone) addSrvRecord(record *SrvRecord) {
	zone.Zone.Srv = append(zone.Zone.Srv, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "SRV", name: record.Name})
}

func (zone *Zone) addSshfpRecord(record *SshfpRecord) {
	zone.Zone.Sshfp = append(zone.Zone.Sshfp, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "SSHFP", name: record.Name})
}

func (zone *Zone) addTxtRecord(record *TxtRecord) {
	zone.Zone.Txt = append(zone.Zone.Txt, record)
	nonCnameNames = append(nonCnameNames, name{recordType: "TXT", name: record.Name})
}

func (zone *Zone) removeARecord(record *ARecord) error {
	for key, r := range zone.Zone.A {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.A[:key]
			if len(zone.Zone.A) > key {
				if len(zone.Zone.A) > key {
					zone.Zone.A = append(records, zone.Zone.A[key+1:]...)
				} else {
					zone.Zone.A = records
				}
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("A Record not found")
}

func (zone *Zone) removeAaaaRecord(record *AaaaRecord) error {
	for key, r := range zone.Zone.Aaaa {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Aaaa[:key]
			if len(zone.Zone.Aaaa) > key {
				zone.Zone.Aaaa = append(records, zone.Zone.Aaaa[key+1:]...)
			} else {
				zone.Zone.Aaaa = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("AAAA Record not found")
}

func (zone *Zone) removeAfsdbRecord(record *AfsdbRecord) error {
	for key, r := range zone.Zone.Afsdb {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Afsdb[:key]
			if len(zone.Zone.Afsdb) > key {
				zone.Zone.Afsdb = append(records, zone.Zone.Afsdb[key+1:]...)
			} else {
				zone.Zone.Afsdb = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Afsdb Record not found")
}

func (zone *Zone) removeCnameRecord(record *CnameRecord) error {
	for key, r := range zone.Zone.Cname {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Cname[:key]
			if len(zone.Zone.Cname) > key {
				zone.Zone.Cname = append(records, zone.Zone.Cname[key+1:]...)
			} else {
				zone.Zone.Cname = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Cname Record not found")

	zone.removeCnameName(record.Name)

	return nil
}

func (zone *Zone) removeDnskeyRecord(record *DnskeyRecord) error {
	for key, r := range zone.Zone.Dnskey {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Dnskey[:key]
			if len(zone.Zone.Dnskey) > key {
				zone.Zone.Dnskey = append(records, zone.Zone.Dnskey[key+1:]...)
			} else {
				zone.Zone.Dnskey = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Dnskey Record not found")
}

func (zone *Zone) removeDsRecord(record *DsRecord) error {
	for key, r := range zone.Zone.Ds {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Ds[:key]
			if len(zone.Zone.Ds) > key {
				zone.Zone.Ds = append(records, zone.Zone.Ds[key+1:]...)
			} else {
				zone.Zone.Ds = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Ds Record not found")
}

func (zone *Zone) removeHinfoRecord(record *HinfoRecord) error {
	for key, r := range zone.Zone.Hinfo {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Hinfo[:key]
			if len(zone.Zone.Hinfo) > key {
				zone.Zone.Hinfo = append(records, zone.Zone.Hinfo[key+1:]...)
			} else {
				zone.Zone.Hinfo = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Hinfo Record not found")
}

func (zone *Zone) removeLocRecord(record *LocRecord) error {
	for key, r := range zone.Zone.Loc {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Loc[:key]
			if len(zone.Zone.Loc) > key {
				zone.Zone.Loc = append(records, zone.Zone.Loc[key+1:]...)
			} else {
				zone.Zone.Loc = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Loc Record not found")
}

func (zone *Zone) removeMxRecord(record *MxRecord) error {
	for key, r := range zone.Zone.Mx {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Mx[:key]
			if len(zone.Zone.Mx) > key {
				zone.Zone.Mx = append(records, zone.Zone.Mx[key+1:]...)
			} else {
				zone.Zone.Mx = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Mx Record not found")
}

func (zone *Zone) removeNaptrRecord(record *NaptrRecord) error {
	for key, r := range zone.Zone.Naptr {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Naptr[:key]
			if len(zone.Zone.Naptr) > key {
				zone.Zone.Naptr = append(records, zone.Zone.Naptr[key+1:]...)
			} else {
				zone.Zone.Naptr = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Naptr Record not found")
}

func (zone *Zone) removeNsRecord(record *NsRecord) error {
	for key, r := range zone.Zone.Ns {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Ns[:key]
			if len(zone.Zone.Ns) > key {
				zone.Zone.Ns = append(records, zone.Zone.Ns[key+1:]...)
			} else {
				zone.Zone.Ns = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Ns Record not found")
}

func (zone *Zone) removeNsec3Record(record *Nsec3Record) error {
	for key, r := range zone.Zone.Nsec3 {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Nsec3[:key]
			if len(zone.Zone.Nsec3) > key {
				zone.Zone.Nsec3 = append(records, zone.Zone.Nsec3[key+1:]...)
			} else {
				zone.Zone.Nsec3 = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Nsec3 Record not found")
}

func (zone *Zone) removeNsec3paramRecord(record *Nsec3paramRecord) error {
	for key, r := range zone.Zone.Nsec3param {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Nsec3param[:key]
			if len(zone.Zone.Nsec3param) > key {
				zone.Zone.Nsec3param = append(records, zone.Zone.Nsec3param[key+1:]...)
			} else {
				zone.Zone.Nsec3param = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Nsec3param Record not found")
}

func (zone *Zone) removePtrRecord(record *PtrRecord) error {
	for key, r := range zone.Zone.Ptr {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Ptr[:key]
			if len(zone.Zone.Ptr) > key {
				zone.Zone.Ptr = append(records, zone.Zone.Ptr[key+1:]...)
			} else {
				zone.Zone.Ptr = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Ptr Record not found")
}

func (zone *Zone) removeRpRecord(record *RpRecord) error {
	for key, r := range zone.Zone.Rp {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Rp[:key]
			if len(zone.Zone.Rp) > key {
				zone.Zone.Rp = append(records, zone.Zone.Rp[key+1:]...)
			} else {
				zone.Zone.Rp = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Rp Record not found")
}

func (zone *Zone) removeRrsigRecord(record *RrsigRecord) error {
	for key, r := range zone.Zone.Rrsig {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Rrsig[:key]
			if len(zone.Zone.Rrsig) > key {
				zone.Zone.Rrsig = append(records, zone.Zone.Rrsig[key+1:]...)
			} else {
				zone.Zone.Rrsig = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Rrsig Record not found")
}

func (zone *Zone) removeSoaRecord(record *SoaRecord) error {
	if reflect.DeepEqual(zone.Zone.Soa, record) {
		zone.Zone.Soa = nil

		return nil
	}

	return errors.New("SOA Record does not match")
}

func (zone *Zone) removeSpfRecord(record *SpfRecord) error {
	for key, r := range zone.Zone.Spf {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Spf[:key]
			if len(zone.Zone.Spf) > key {
				zone.Zone.Spf = append(records, zone.Zone.Spf[key+1:]...)
			} else {
				zone.Zone.Spf = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Spf Record not found")
}

func (zone *Zone) removeSrvRecord(record *SrvRecord) error {
	for key, r := range zone.Zone.Srv {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Srv[:key]
			if len(zone.Zone.Srv) > key {
				zone.Zone.Srv = append(records, zone.Zone.Srv[key+1:]...)
			} else {
				zone.Zone.Srv = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Srv Record not found")
}

func (zone *Zone) removeSshfpRecord(record *SshfpRecord) error {
	for key, r := range zone.Zone.Sshfp {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Sshfp[:key]
			if len(zone.Zone.Sshfp) > key {
				zone.Zone.Sshfp = append(records, zone.Zone.Sshfp[key+1:]...)
			} else {
				zone.Zone.Sshfp = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Sshfp Record not found")
}

func (zone *Zone) removeTxtRecord(record *TxtRecord) error {
	for key, r := range zone.Zone.Txt {
		if reflect.DeepEqual(r, record) {
			records := zone.Zone.Txt[:key]
			if len(zone.Zone.Txt) > key {
				zone.Zone.Txt = append(records, zone.Zone.Txt[key+1:]...)
			} else {
				zone.Zone.Txt = records
			}
			zone.removeNonCnameName(record.Name)

			return nil
		}
	}

	return errors.New("Txt Record not found")
}

func (zone *Zone) PostUnmarshalJSON() error {
	if zone.Zone.Soa.Serial > 0 {
		zone.Zone.Soa.originalSerial = zone.Zone.Soa.Serial
	}
	return nil
}

func (zone *Zone) PreMarshalJSON() error {
	if zone.Zone.Soa.Serial == 0 {
		zone.Zone.Soa.Serial = uint(time.Now().Unix())
	} else if zone.Zone.Soa.Serial == zone.Zone.Soa.originalSerial {
		zone.Zone.Soa.Serial = zone.Zone.Soa.Serial + 1
	}
	return nil
}

func (zone *Zone) validateCnames() (bool, []name) {
	var valid bool = true
	var failedRecords []name
	for _, v := range cnameNames {
		for _, vv := range nonCnameNames {
			if v.name == vv.name {
				valid = false
				failedRecords = append(failedRecords, vv)
			}
		}
	}
	return valid, failedRecords
}

func (zone *Zone) removeCnameName(host string) {
	var ncn []name
	for _, v := range cnameNames {
		if v.name != host {
			ncn =append(ncn, v)
		}
	}
	cnameNames = ncn
}


func (zone *Zone) removeNonCnameName(host string) {
	var ncn []name
	for _, v := range nonCnameNames {
		if v.name != host {
			ncn =append(ncn, v)
		}
	}
	nonCnameNames = ncn
}

func (zone *Zone) FindRecords(recordType string, options map[string]interface{}) []DNSRecord {
	switch strings.ToUpper(recordType) {
	case "A":
		return zone.findARecord(options)
	case "AAAA":
		return zone.findAaaaRecord(options)
	case "AFSDB":
		return zone.findAfsdbRecord(options)
	case "CNAME":
		return zone.findCnameRecord(options)
	case "DNSKEY":
		return zone.findDnskeyRecord(options)
	case "DS":
		return zone.findDsRecord(options)
	case "HINFO":
		return zone.findHinfoRecord(options)
	case "LOC":
		return zone.findLocRecord(options)
	case "MX":
		return zone.findMxRecord(options)
	case "NAPTR":
		return zone.findNaptrRecord(options)
	case "NS":
		return zone.findNsRecord(options)
	case "NSEC3":
		return zone.findNsec3Record(options)
	case "NSEC3PARAM":
		return zone.findNsec3paramRecord(options)
	case "PTR":
		return zone.findPtrRecord(options)
	case "RP":
		return zone.findRpRecord(options)
	case "RRSIG":
		return zone.findRrsigRecord(options)
	case "SPF":
		return zone.findSpfRecord(options)
	case "SRV":
		return zone.findSrvRecord(options)
	case "SSHFP":
		return zone.findSshfpRecord(options)
	case "TXT":
		return zone.findTxtRecord(options)
	}

	return make([]DNSRecord, 0)
}

func (zone *Zone) findARecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.A {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}

	return found
}

func (zone *Zone) findAaaaRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Aaaa {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}

	return found
}

func (zone *Zone) findAfsdbRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Afsdb {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if subtype, ok := options["subtype"]; ok && record.Subtype == subtype.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findCnameRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Cname {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findDnskeyRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Dnskey {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if flags, ok := options["flags"]; ok && record.Flags == flags.(int) {
			matchCounter++
		}

		if protocol, ok := options["protocol"]; ok && record.Protocol == protocol.(int) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if key, ok := options["key"]; ok && record.Key == key.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findDsRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Ds {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if keytag, ok := options["keytag"]; ok && record.Keytag == keytag.(int) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if digesttype, ok := options["digesttype"]; ok && record.DigestType == digesttype.(int) {
			matchCounter++
		}

		if digest, ok := options["digest"]; ok && record.Digest == digest.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findHinfoRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Hinfo {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if hardware, ok := options["hardware"]; ok && record.Hardware == hardware.(string) {
			matchCounter++
		}

		if software, ok := options["software"]; ok && record.Software == software.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findLocRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Loc {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findMxRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Mx {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if priority, ok := options["priority"]; ok && record.Priority == priority.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findNaptrRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Naptr {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if order, ok := options["order"]; ok && record.Order == order.(uint16) {
			matchCounter++
		}

		if preference, ok := options["preference"]; ok && record.Preference == preference.(uint16) {
			matchCounter++
		}

		if flags, ok := options["flags"]; ok && record.Flags == flags.(string) {
			matchCounter++
		}

		if service, ok := options["service"]; ok && record.Service == service.(string) {
			matchCounter++
		}

		if regexp, ok := options["regexp"]; ok && record.Regexp == regexp.(string) {
			matchCounter++
		}

		if replacement, ok := options["replacement"]; ok && record.Replacement == replacement.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findNsRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Ns {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findNsec3Record(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Nsec3 {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if flags, ok := options["flags"]; ok && record.Flags == flags.(int) {
			matchCounter++
		}

		if iterations, ok := options["iterations"]; ok && record.Iterations == iterations.(int) {
			matchCounter++
		}

		if salt, ok := options["salt"]; ok && record.Salt == salt.(string) {
			matchCounter++
		}

		if nextHashedOwnerName, ok := options["nextHashedOwnerName"]; ok && record.NextHashedOwnerName == nextHashedOwnerName.(string) {
			matchCounter++
		}

		if typeBitmaps, ok := options["typeBitmaps"]; ok && record.TypeBitmaps == typeBitmaps.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findNsec3paramRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Nsec3param {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if flags, ok := options["flags"]; ok && record.Flags == flags.(int) {
			matchCounter++
		}

		if iterations, ok := options["iterations"]; ok && record.Iterations == iterations.(int) {
			matchCounter++
		}

		if salt, ok := options["salt"]; ok && record.Salt == salt.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findPtrRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Ptr {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findRpRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Rp {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if mailbox, ok := options["mailbox"]; ok && record.Mailbox == mailbox.(string) {
			matchCounter++
		}

		if txt, ok := options["txt"]; ok && record.Txt == txt.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findRrsigRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Rrsig {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if typeCovered, ok := options["typeCovered"]; ok && record.TypeCovered == typeCovered.(string) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if originalTTL, ok := options["originalTTL"]; ok && record.OriginalTTL == originalTTL.(int) {
			matchCounter++
		}

		if expiration, ok := options["expiration"]; ok && record.Expiration == expiration.(string) {
			matchCounter++
		}

		if inception, ok := options["inception"]; ok && record.Inception == inception.(string) {
			matchCounter++
		}

		if keytag, ok := options["keytag"]; ok && record.Keytag == keytag.(int) {
			matchCounter++
		}

		if signer, ok := options["signer"]; ok && record.Signer == signer.(string) {
			matchCounter++
		}

		if signature, ok := options["signature"]; ok && record.Signature == signature.(string) {
			matchCounter++
		}

		if labels, ok := options["labels"]; ok && record.Labels == labels.(int) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findSpfRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Spf {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findSrvRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Srv {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if priority, ok := options["priority"]; ok && record.Priority == priority.(int) {
			matchCounter++
		}

		if weight, ok := options["weight"]; ok && record.Weight == weight.(uint16) {
			matchCounter++
		}

		if port, ok := options["port"]; ok && record.Port == port.(uint16) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findSshfpRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Sshfp {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if algorithm, ok := options["algorithm"]; ok && record.Algorithm == algorithm.(int) {
			matchCounter++
		}

		if fingerprintType, ok := options["fingerprintType"]; ok && record.FingerprintType == fingerprintType.(int) {
			matchCounter++
		}

		if fingerprint, ok := options["fingerprint"]; ok && record.Fingerprint == fingerprint.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}

func (zone *Zone) findTxtRecord(options map[string]interface{}) []DNSRecord {
	found := make([]DNSRecord, 0)
	matchesNeeded := len(options)
	for _, record := range zone.Zone.Txt {
		matchCounter := 0
		if name, ok := options["name"]; ok && record.Name == name.(string) {
			matchCounter++
		}

		if active, ok := options["active"]; ok && record.Active == active.(bool) {
			matchCounter++
		}

		if ttl, ok := options["ttl"]; ok && record.TTL == ttl.(int) {
			matchCounter++
		}

		if target, ok := options["target"]; ok && record.Target == target.(string) {
			matchCounter++
		}

		if matchCounter >= matchesNeeded {
			found = append(found, record)
		}
	}
	return found
}
