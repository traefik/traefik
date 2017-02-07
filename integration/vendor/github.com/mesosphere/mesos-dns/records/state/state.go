package state

import (
	"bytes"
	"github.com/mesos/mesos-go/upid"
	"net"
	"strconv"
	"strings"
)

// Resources holds resources as defined in the /state.json Mesos HTTP endpoint.
type Resources struct {
	PortRanges string `json:"ports"`
}

// Ports returns a slice of individual ports expanded from PortRanges.
func (r Resources) Ports() []string {
	if r.PortRanges == "" || r.PortRanges == "[]" {
		return []string{}
	}

	rhs := strings.Split(r.PortRanges, "[")[1]
	lhs := strings.Split(rhs, "]")[0]

	yports := []string{}

	mports := strings.Split(lhs, ",")
	for _, port := range mports {
		tmp := strings.TrimSpace(port)
		pz := strings.Split(tmp, "-")
		lo, _ := strconv.Atoi(pz[0])
		hi, _ := strconv.Atoi(pz[1])

		for t := lo; t <= hi; t++ {
			yports = append(yports, strconv.Itoa(t))
		}
	}
	return yports
}

// Label holds a label as defined in the /state.json Mesos HTTP endpoint.
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Status holds a task status as defined in the /state.json Mesos HTTP endpoint.
type Status struct {
	Timestamp       float64         `json:"timestamp"`
	State           string          `json:"state"`
	Labels          []Label         `json:"labels,omitempty"`
	ContainerStatus ContainerStatus `json:"container_status,omitempty"`
	Healthy         *bool          	`json:"healthy"`
}

// ContainerStatus holds container metadata as defined in the /state.json
// Mesos HTTP endpoint.
type ContainerStatus struct {
	NetworkInfos []NetworkInfo `json:"network_infos,omitempty"`
}

// NetworkInfo holds the network configuration for a single interface
// as defined in the /state.json Mesos HTTP endpoint.
type NetworkInfo struct {
	IPAddresses []IPAddress `json:"ip_addresses,omitempty"`
	// back-compat with 0.25 IPAddress format
	IPAddress string `json:"ip_address,omitempty"`
}

// IPAddress holds a single IP address configured on an interface,
// as defined in the /state.json Mesos HTTP endpoint.
type IPAddress struct {
	IPAddress string `json:"ip_address,omitempty"`
}

// Task holds a task as defined in the /state.json Mesos HTTP endpoint.
type Task struct {
	FrameworkID   string   `json:"framework_id"`
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	SlaveID       string   `json:"slave_id"`
	State         string   `json:"state"`
	Statuses      []Status `json:"statuses"`
	Resources     `json:"resources"`
	DiscoveryInfo DiscoveryInfo `json:"discovery"`

	SlaveIP string  `json:"-"`
	Labels  []Label `json:"labels,omitempty"`
}

// HasDiscoveryInfo return whether the DiscoveryInfo was provided in the state.json
func (t *Task) HasDiscoveryInfo() bool {
	return t.DiscoveryInfo.Name != ""
}

// IP returns the first Task IP found in the given sources.
func (t *Task) IP(srcs ...string) string {
	if ips := t.IPs(srcs...); len(ips) > 0 {
		return ips[0].String()
	}
	return ""
}

// IPs returns a slice of IPs sourced from the given sources with ascending
// priority.
func (t *Task) IPs(srcs ...string) (ips []net.IP) {
	if t == nil {
		return nil
	}
	for i := range srcs {
		if src, ok := sources[srcs[i]]; ok {
			for _, srcIP := range src(t) {
				if ip := net.ParseIP(srcIP); len(ip) > 0 {
					ips = append(ips, ip)
				}
			}
		}
	}
	return ips
}

// sources maps the string representation of IP sources to their functions.
var sources = map[string]func(*Task) []string{
	"host":    hostIPs,
	"mesos":   mesosIPs,
	"docker":  dockerIPs,
	"netinfo": networkInfoIPs,
}

// hostIPs is an IPSource which returns the IP addresses of the slave a Task
// runs on.
func hostIPs(t *Task) []string { return []string{t.SlaveIP} }

// networkInfoIPs returns IP addresses from a given Task's
// []Status.ContainerStatus.[]NetworkInfos.[]IPAddresses.IPAddress
func networkInfoIPs(t *Task) []string {
	return statusIPs(t.Statuses, func(s *Status) []string {
		ips := make([]string, len(s.ContainerStatus.NetworkInfos))
		for _, netinfo := range s.ContainerStatus.NetworkInfos {
			if len(netinfo.IPAddresses) > 0 {
				// In v0.26, we use the IPAddresses field.
				for _, ipAddress := range netinfo.IPAddresses {
					ips = append(ips, ipAddress.IPAddress)
				}
			} else {
				// Fall back to v0.25 syntax of single IPAddress if that's being used.
				if netinfo.IPAddress != "" {
					ips = append(ips, netinfo.IPAddress)
				}
			}
		}
		return ips
	})
}

const (
	// DockerIPLabel is the key of the Label which holds the Docker containerizer IP value.
	DockerIPLabel = "Docker.NetworkSettings.IPAddress"
	// MesosIPLabel is the key of the label which holds the Mesos containerizer IP value.
	MesosIPLabel = "MesosContainerizer.NetworkSettings.IPAddress"
)

// dockerIPs returns IP addresses from the values of all
// Task.[]Status.[]Labels whose keys are equal to "Docker.NetworkSettings.IPAddress".
func dockerIPs(t *Task) []string {
	return statusIPs(t.Statuses, labels(DockerIPLabel))
}

// mesosIPs returns IP addresses from the values of all
// Task.[]Status.[]Labels whose keys are equal to
// "MesosContainerizer.NetworkSettings.IPAddress".
func mesosIPs(t *Task) []string {
	return statusIPs(t.Statuses, labels(MesosIPLabel))
}

// statusIPs returns the latest running status IPs extracted with the given src
func statusIPs(st []Status, src func(*Status) []string) []string {
	// the state.json we extract from mesos makes no guarantees re: the order
	// of the task statuses so we should check the timestamps to avoid problems
	// down the line. we can't rely on seeing the same sequence. (@joris)
	// https://github.com/apache/mesos/blob/0.24.0/src/slave/slave.cpp#L5226-L5238
	ts, j := -1.0, -1
	for i := range st {
		if st[i].State == "TASK_RUNNING" && st[i].Timestamp > ts {
			ts, j = st[i].Timestamp, i
		}
	}
	if j >= 0 {
		return src(&st[j])
	}
	return nil
}

// labels returns all given Status.[]Labels' values whose keys are equal
// to the given key
func labels(key string) func(*Status) []string {
	return func(s *Status) []string {
		vs := make([]string, 0, len(s.Labels))
		for _, l := range s.Labels {
			if l.Key == key {
				vs = append(vs, l.Value)
			}
		}
		return vs
	}
}

// Framework holds a framework as defined in the /state.json Mesos HTTP endpoint.
type Framework struct {
	Tasks    []Task `json:"tasks"`
	PID      PID    `json:"pid"`
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
}

// HostPort returns the hostname and port where a framework's scheduler is
// listening on.
func (f Framework) HostPort() (string, string) {
	if f.PID.UPID != nil {
		return f.PID.Host, f.PID.Port
	}
	return f.Hostname, ""
}

// Slave holds a slave as defined in the /state.json Mesos HTTP endpoint.
type Slave struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	PID      PID    `json:"pid"`
}

// PID holds a Mesos PID and implements the json.Unmarshaler interface.
type PID struct{ *upid.UPID }

// UnmarshalJSON implements the json.Unmarshaler interface for PIDs.
func (p *PID) UnmarshalJSON(data []byte) (err error) {
	p.UPID, err = upid.Parse(string(bytes.Trim(data, `" `)))
	return err
}

// State holds the state defined in the /state.json Mesos HTTP endpoint.
type State struct {
	Frameworks []Framework `json:"frameworks"`
	Slaves     []Slave     `json:"slaves"`
	Leader     string      `json:"leader"`
}

// DiscoveryInfo holds the discovery meta data for a task defined in the /state.json Mesos HTTP endpoint.
type DiscoveryInfo struct {
	Visibilty   string `json:"visibility"`
	Version     string `json:"version,omitempty"`
	Name        string `json:"name,omitempty"`
	Location    string `json:"location,omitempty"`
	Environment string `json:"environment,omitempty"`
	Labels      struct {
			    Labels []Label `json:"labels"`
		    } `json:"labels"`
	Ports Ports `json:"ports"`
}

// Ports holds a list of DiscoveryPort
type Ports struct {
	DiscoveryPorts []DiscoveryPort `json:"ports"`
}

// DiscoveryPort holds a port for a task defined in the /state.json Mesos HTTP endpoint.
type DiscoveryPort struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"`
	Name     string `json:"name"`
}
