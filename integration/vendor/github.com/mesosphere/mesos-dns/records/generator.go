// Package records contains functions to generate resource records from
// mesos master states to serve through a dns server
package records

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mesosphere/mesos-dns/errorutil"
	"github.com/mesosphere/mesos-dns/logging"
	"github.com/mesosphere/mesos-dns/models"
	"github.com/mesosphere/mesos-dns/records/labels"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/tv42/zbase32"
)

// Map host/service name to DNS answer
// REFACTOR - when discoveryinfo is integrated
// Will likely become map[string][]discoveryinfo
// Effectively we're (ab)using the map type as a set
// It used to have the type: rrs map[string][]string
type rrs map[string]map[string]struct{}

func (r rrs) add(name, host string) bool {
	if host == "" {
		return false
	}
	v, ok := r[name]
	if !ok {
		v = make(map[string]struct{})
		r[name] = v
	} else {
		// don't overwrite existing values
		_, ok = v[host]
		if ok {
			return false
		}
	}
	v[host] = struct{}{}
	return true
}

func (r rrs) First(name string) (string, bool) {
	for host := range r[name] {
		return host, true
	}
	return "", false
}

// Transform the record set into something exportable via the REST API
func (r rrs) ToAXFRResourceRecordSet() models.AXFRResourceRecordSet {
	ret := make(models.AXFRResourceRecordSet, len(r))
	for host, values := range r {
		ret[host] = make([]string, 0, len(values))
		for record := range values {
			ret[host] = append(ret[host], record)
		}
	}
	return ret
}

type rrsKind string

const (
	// A record types
	A rrsKind = "A"
	// SRV record types
	SRV = "SRV"
)

func (kind rrsKind) rrs(rg *RecordGenerator) rrs {
	switch kind {
	case A:
		return rg.As
	case SRV:
		return rg.SRVs
	default:
		return nil
	}
}

// RecordGenerator contains DNS records and methods to access and manipulate
// them. TODO(kozyraki): Refactor when discovery id is available.
type RecordGenerator struct {
	As         rrs
	SRVs       rrs
	SlaveIPs   map[string]string
	EnumData   EnumerationData
	httpClient http.Client
}

// EnumerableRecord is the lowest level object, and should map 1:1 with DNS records
type EnumerableRecord struct {
	Name  string `json:"name"`
	Host  string `json:"host"`
	Rtype string `json:"rtype"`
}

// EnumerableTask consists of the records derived from a task
type EnumerableTask struct {
	Name    string             `json:"name"`
	ID      string             `json:"id"`
	Records []EnumerableRecord `json:"records"`
}

// EnumerableFramework is consistent of enumerable tasks, and include the name of the framework
type EnumerableFramework struct {
	Tasks []*EnumerableTask `json:"tasks"`
	Name  string            `json:"name"`
}

// EnumerationData is the top level container pointing to the
// enumerable frameworks containing enumerable tasks
type EnumerationData struct {
	Frameworks []*EnumerableFramework `json:"frameworks"`
}

// NewRecordGenerator returns a RecordGenerator that's been configured with a timeout.
func NewRecordGenerator(httpTimeout time.Duration) *RecordGenerator {
	enumData := EnumerationData{
		Frameworks: []*EnumerableFramework{},
	}
	rg := &RecordGenerator{
		httpClient: http.Client{Timeout: httpTimeout},
		EnumData:   enumData,
	}
	return rg
}

// ParseState retrieves and parses the Mesos master /state.json and converts it
// into DNS records.
func (rg *RecordGenerator) ParseState(c Config, masters ...string) error {
	// find master -- return if error
	sj, err := rg.FindMaster(masters...)
	if err != nil {
		logging.Error.Println("no master")
		return err
	}
	if sj.Leader == "" {
		logging.Error.Println("Unexpected error")
		err = errors.New("empty master")
		return err
	}

	hostSpec := labels.RFC1123
	if c.EnforceRFC952 {
		hostSpec = labels.RFC952
	}

	return rg.InsertState(sj, c.Domain, c.SOARname, c.Listener, masters, c.IPSources, hostSpec)
}

// Tries each master and looks for the leader
// if no leader responds it errors
func (rg *RecordGenerator) FindMaster(masters ...string) (state.State, error) {
	var sj state.State
	var leader string

	if len(masters) > 0 {
		leader, masters = masters[0], masters[1:]
	}

	// Check if ZK leader is correct
	if leader != "" {
		logging.VeryVerbose.Println("Zookeeper says the leader is: ", leader)
		ip, port, err := getProto(leader)
		if err != nil {
			logging.Error.Println(err)
		}

		if sj, err = rg.loadWrap(ip, port); err == nil && sj.Leader != "" {
			return sj, nil
		}
		logging.Verbose.Println("Warning: Zookeeper is wrong about leader")
		if len(masters) == 0 {
			return sj, errors.New("no master")
		}
		logging.Verbose.Println("Warning: falling back to Masters config field: ", masters)
	}

	// try each listed mesos master before dying
	for i, master := range masters {
		ip, port, err := getProto(master)
		if err != nil {
			logging.Error.Println(err)
		}

		if sj, err = rg.loadWrap(ip, port); err == nil && sj.Leader == "" {
			logging.VeryVerbose.Println("Warning: not a leader - trying next one")
			if len(masters)-1 == i {
				return sj, errors.New("no master")
			}
		} else {
			return sj, nil
		}

	}

	return sj, errors.New("no master")
}

// Loads state.json from mesos master
func (rg *RecordGenerator) loadFromMaster(ip string, port string) (state.State, error) {
	// REFACTOR: state.json security

	var sj state.State
	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(ip, port),
		Path:   "/master/state.json",
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		logging.Error.Println(err)
		return state.State{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := rg.httpClient.Do(req)
	if err != nil {
		logging.Error.Println(err)
		return state.State{}, err
	}

	defer errorutil.Ignore(resp.Body.Close)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Error.Println(err)
		return state.State{}, err
	}

	err = json.Unmarshal(body, &sj)
	if err != nil {
		logging.Error.Println(err)
		return state.State{}, err
	}

	return sj, nil
}

// Catches an attempt to load state.json from a mesos master
// attempts can fail from down server or mesos master secondary
// it also reloads from a different master if the master it attempted to
// load from was not the leader
func (rg *RecordGenerator) loadWrap(ip string, port string) (state.State, error) {
	var err error
	var sj state.State

	logging.VeryVerbose.Println("reloading from master " + ip)
	sj, err = rg.loadFromMaster(ip, port)
	if err != nil {
		return state.State{}, err
	}
	if rip := leaderIP(sj.Leader); rip != ip {
		logging.VeryVerbose.Println("Warning: master changed to " + ip)
		sj, err = rg.loadFromMaster(rip, port)
		return sj, err
	}
	return sj, nil
}

// hashes a given name using a truncated sha1 hash
// 5 characters extracted from the zbase32 encoded hash provides
// enough entropy to avoid collisions
// zbase32: http://philzimmermann.com/docs/human-oriented-base-32-encoding.txt
// is used to promote human-readable names
func hashString(s string) string {
	hash := sha1.Sum([]byte(s))
	return zbase32.EncodeToString(hash[:])[:5]
}

// attempt to translate the hostname into an IPv4 address. logs an error if IP
// lookup fails. if an IP address cannot be found, returns the same hostname
// that was given. upon success returns the IP address as a string.
func hostToIP4(hostname string) (string, bool) {
	ip := net.ParseIP(hostname)
	if ip == nil {
		t, err := net.ResolveIPAddr("ip4", hostname)
		if err != nil {
			logging.Error.Printf("cannot translate hostname %q into an ip4 address", hostname)
			return hostname, false
		}
		ip = t.IP
	}
	return ip.String(), true
}

// InsertState transforms a StateJSON into RecordGenerator RRs
func (rg *RecordGenerator) InsertState(sj state.State, domain, ns, listener string, masters, ipSources []string, spec labels.Func) error {

	rg.SlaveIPs = map[string]string{}
	rg.SRVs = rrs{}
	rg.As = rrs{}
	rg.frameworkRecords(sj, domain, spec)
	rg.slaveRecords(sj, domain, spec)
	rg.listenerRecord(listener, ns)
	rg.masterRecord(domain, masters, sj.Leader)
	rg.taskRecords(sj, domain, spec, ipSources)

	return nil
}

// frameworkRecords injects A and SRV records into the generator store:
//     frameworkname.domain.                 // resolves to IPs of each framework
//     _framework._tcp.frameworkname.domain. // resolves to the driver port and IP of each framework
func (rg *RecordGenerator) frameworkRecords(sj state.State, domain string, spec labels.Func) {
	for _, f := range sj.Frameworks {
		fname := labels.DomainFrag(f.Name, labels.Sep, spec)
		host, port := f.HostPort()
		if address, ok := hostToIP4(host); ok {
			a := fname + "." + domain + "."
			rg.insertRR(a, address, A)
			if port != "" {
				srvAddress := net.JoinHostPort(a, port)
				rg.insertRR("_framework._tcp."+a, srvAddress, SRV)
			}
		}
	}
}

// slaveRecords injects A and SRV records into the generator store:
//     slave.domain.      // resolves to IPs of all slaves
//     _slave._tc.domain. // resolves to the driver port and IP of all slaves
func (rg *RecordGenerator) slaveRecords(sj state.State, domain string, spec labels.Func) {
	for _, slave := range sj.Slaves {
		address, ok := hostToIP4(slave.PID.Host)
		if ok {
			a := "slave." + domain + "."
			rg.insertRR(a, address, A)
			srv := net.JoinHostPort(a, slave.PID.Port)
			rg.insertRR("_slave._tcp."+domain+".", srv, SRV)
		} else {
			logging.VeryVerbose.Printf("string '%q' for slave with id %q is not a valid IP address", address, slave.ID)
			address = labels.DomainFrag(address, labels.Sep, spec)
		}
		rg.SlaveIPs[slave.ID] = address
	}
}

// masterRecord injects A and SRV records into the generator store:
//     master.domain.  // resolves to IPs of all masters
//     masterN.domain. // one IP address for each master
//     leader.domain.  // one IP address for the leading master
//
// The current func implementation makes an assumption about the order of masters:
// it's the order in which you expect the enumerated masterN records to be created.
// This is probably important: if a new leader is elected, you may not want it to
// become master0 simply because it's the leader. You probably want your DNS records
// to change as little as possible. And this func should have the least impact on
// enumeration order, or name/IP mappings - it's just creating the records. So let
// the caller do the work of ordering/sorting (if desired) the masters list if a
// different outcome is desired.
//
// Another consequence of the current overall mesos-dns app implementation is that
// the leader may not even be in the masters list at some point in time. masters is
// really fallback-masters (only consider these to be masters if I can't find a
// leader via ZK). At some point in time, they may not actually be masters any more.
// Consider a cluster of 3 nodes that suffers the loss of a member, and gains a new
// member (VM crashed, was replaced by another VM). And the cycle repeats several
// times. You end up with a set of running masters (and leader) that's different
// than the set of statically configured fallback masters.
//
// So the func tries to index the masters as they're listed and begrudgingly assigns
// the leading master an index out-of-band if it's not actually listed in the masters
// list. There are probably better ways to do it.
func (rg *RecordGenerator) masterRecord(domain string, masters []string, leader string) {
	// create records for leader
	// A records
	h := strings.Split(leader, "@")
	if len(h) < 2 {
		logging.Error.Println(leader)
		return // avoid a panic later
	}
	leaderAddress := h[1]
	ip, port, err := getProto(leaderAddress)
	if err != nil {
		logging.Error.Println(err)
		return
	}
	arec := "leader." + domain + "."
	rg.insertRR(arec, ip, A)
	arec = "master." + domain + "."
	rg.insertRR(arec, ip, A)

	// SRV records
	tcp := "_leader._tcp." + domain + "."
	udp := "_leader._udp." + domain + "."
	host := "leader." + domain + "." + ":" + port
	rg.insertRR(tcp, host, SRV)
	rg.insertRR(udp, host, SRV)

	// if there is a list of masters, insert that as well
	addedLeaderMasterN := false
	idx := 0
	for _, master := range masters {
		masterIP, _, err := getProto(master)
		if err != nil {
			logging.Error.Println(err)
			continue
		}

		// A records (master and masterN)
		if master != leaderAddress {
			arec := "master." + domain + "."
			added := rg.insertRR(arec, masterIP, A)
			if !added {
				// duplicate master?!
				continue
			}
		}

		if master == leaderAddress && addedLeaderMasterN {
			// duplicate leader in masters list?!
			continue
		}

		arec := "master" + strconv.Itoa(idx) + "." + domain + "."
		rg.insertRR(arec, masterIP, A)
		idx++

		if master == leaderAddress {
			addedLeaderMasterN = true
		}
	}
	// flake: we ended up with a leader that's not in the list of all masters?
	if !addedLeaderMasterN {
		// only a flake if there were fallback masters configured
		if len(masters) > 0 {
			logging.Error.Printf("warning: leader %q is not in master list", leader)
		}
		arec = "master" + strconv.Itoa(idx) + "." + domain + "."
		rg.insertRR(arec, ip, A)
	}
}

// A record for mesos-dns (the name is listed in SOA replies)
func (rg *RecordGenerator) listenerRecord(listener string, ns string) {
	if listener == "0.0.0.0" {
		rg.setFromLocal(listener, ns)
	} else if listener == "127.0.0.1" {
		rg.insertRR(ns, "127.0.0.1", A)
	} else {
		rg.insertRR(ns, listener, A)
	}
}

func (rg *RecordGenerator) taskRecords(sj state.State, domain string, spec labels.Func, ipSources []string) {
	for _, f := range sj.Frameworks {
		enumerableFramework := &EnumerableFramework{
			Name:  f.Name,
			Tasks: []*EnumerableTask{},
		}
		rg.EnumData.Frameworks = append(rg.EnumData.Frameworks, enumerableFramework)

		for _, task := range f.Tasks {
			var ok bool
			task.SlaveIP, ok = rg.SlaveIPs[task.SlaveID]

			// only do running and discoverable tasks
			if ok && (task.State == "TASK_RUNNING") {
				rg.taskRecord(task, f, domain, spec, ipSources, enumerableFramework)
			}
		}
	}
}

type context struct {
	taskName,
	taskID,
	slaveID,
	taskIP,
	slaveIP string
}

func (rg *RecordGenerator) taskRecord(task state.Task, f state.Framework, domain string, spec labels.Func, ipSources []string, enumFW *EnumerableFramework) {

	newTask := &EnumerableTask{ID: task.ID, Name: task.Name}

	enumFW.Tasks = append(enumFW.Tasks, newTask)

	// define context
	ctx := context{
		spec(task.Name),
		hashString(task.ID),
		slaveIDTail(task.SlaveID),
		task.IP(ipSources...),
		task.SlaveIP,
	}

	// use DiscoveryInfo name if defined instead of task name
	if task.HasDiscoveryInfo() {
		// LEGACY TODO: REMOVE
		ctx.taskName = task.DiscoveryInfo.Name
		rg.taskContextRecord(ctx, task, f, domain, spec, newTask)
		// LEGACY, TODO: REMOVE

		ctx.taskName = spec(task.DiscoveryInfo.Name)
		rg.taskContextRecord(ctx, task, f, domain, spec, newTask)
	} else {
		rg.taskContextRecord(ctx, task, f, domain, spec, newTask)
	}

}
func (rg *RecordGenerator) taskContextRecord(ctx context, task state.Task, f state.Framework, domain string, spec labels.Func, enumTask *EnumerableTask) {
	fname := labels.DomainFrag(f.Name, labels.Sep, spec)

	tail := "." + domain + "."

	// insert canonical A records
	canonical := ctx.taskName + "-" + ctx.taskID + "-" + ctx.slaveID + "." + fname
	arec := ctx.taskName + "." + fname

	rg.insertTaskRR(arec+tail, ctx.taskIP, A, enumTask)
	rg.insertTaskRR(canonical+tail, ctx.taskIP, A, enumTask)

	rg.insertTaskRR(arec+".slave"+tail, ctx.slaveIP, A, enumTask)
	rg.insertTaskRR(canonical+".slave"+tail, ctx.slaveIP, A, enumTask)

	// recordName generates records for ctx.taskName, given some generation chain
	recordName := func(gen chain) { gen("_" + ctx.taskName) }

	// asSRV is always the last link in a chain, it must insert RR's
	asSRV := func(target string) chain {
		return func(records ...string) {
			for i := range records {
				name := records[i] + tail
				rg.insertTaskRR(name, target, SRV, enumTask)
			}
		}
	}

	// Add RFC 2782 SRV records
	var subdomains []string
	if task.HasDiscoveryInfo() {
		subdomains = []string{"slave"}
	} else {
		subdomains = []string{"slave", domainNone}
	}

	slaveHost := canonical + ".slave" + tail
	for _, port := range task.Ports() {
		slaveTarget := slaveHost + ":" + port
		recordName(withProtocol(protocolNone, fname, spec,
			withSubdomains(subdomains, asSRV(slaveTarget))))
	}

	if !task.HasDiscoveryInfo() {
		return
	}

	for _, port := range task.DiscoveryInfo.Ports.DiscoveryPorts {
		target := canonical + tail + ":" + strconv.Itoa(port.Number)
		recordName(withProtocol(port.Protocol, fname, spec,
			withNamedPort(port.Name, spec, asSRV(target))))
	}
}

// A records for each local interface
// If this causes problems you should explicitly set the
// listener address in config.json
func (rg *RecordGenerator) setFromLocal(host string, ns string) {

	ifaces, err := net.Interfaces()
	if err != nil {
		logging.Error.Println(err)
	}

	// handle err
	for _, i := range ifaces {

		addrs, err := i.Addrs()
		if err != nil {
			logging.Error.Println(err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
			}

			rg.insertRR(ns, ip.String(), A)
		}
	}
}

// insertRR adds a record to the appropriate record map for the given name/host pair,
// but only if the pair is unique. returns true if added, false otherwise.
// TODO(???): REFACTOR when storage is updated
func (rg *RecordGenerator) insertTaskRR(name, host string, kind rrsKind, enumTask *EnumerableTask) bool {
	if rg.insertRR(name, host, kind) {
		enumRecord := EnumerableRecord{Name: name, Host: host, Rtype: string(kind)}
		enumTask.Records = append(enumTask.Records, enumRecord)
		return true
	}
	return false
}

func (rg *RecordGenerator) insertRR(name, host string, kind rrsKind) (added bool) {
	if rrs := kind.rrs(rg); rrs != nil {
		if added = rrs.add(name, host); added {
			logging.VeryVerbose.Println("[" + string(kind) + "]\t" + name + ": " + host)
		}
	}
	return
}

// leaderIP returns the ip for the mesos master
// input format master@ip:port
func leaderIP(leader string) string {
	pair := strings.Split(leader, "@")[1]
	return strings.Split(pair, ":")[0]
}

// return the slave number from a Mesos slave id
func slaveIDTail(slaveID string) string {
	fields := strings.Split(slaveID, "-")
	return strings.ToLower(fields[len(fields)-1])
}

// should be able to accept
// ip:port
// zk://host1:port1,host2:port2,.../path
// zk://username:password@host1:port1,host2:port2,.../path
// file:///path/to/file (where file contains one of the above)
func getProto(pair string) (string, string, error) {
	h := strings.SplitN(pair, ":", 2)
	if len(h) != 2 {
		return "", "", fmt.Errorf("unable to parse proto from %q", pair)
	}
	return h[0], h[1], nil
}
