package detect

import (
	"net"
	"reflect"
	"testing"

	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesosphere/mesos-dns/logging"
)

func init() {
	// TODO(tsenart): Refactor the logging package
	logging.VerboseFlag = true
	logging.SetupLogs()
}

func TestMasters_UpdatedMasters(t *testing.T) {
	// create a new masters detector with an unknown leader and no masters
	ch := make(chan []string, 1)
	m := NewMasters([]string{}, ch)

	for i, tt := range []struct {
		masters []*mesos.MasterInfo
		want    []string
	}{
		{
			// update a single master
			// leave the unknown leader "" unchanged
			masterInfos(masterInfo(ip("1.1.1.1"))),
			[]string{"", "1.1.1.1:5050"},
		},
		{
			// update additional masters,
			// expect them to be appended with the default port number,
			// leave the unknown leader "" unchanged
			masterInfos(
				masterInfo(ip("1.1.1.1")),
				masterInfo(ip("1.1.1.2")),
				masterInfo(ip("1.1.1.3")),
			),
			[]string{"", "1.1.1.1:5050", "1.1.1.2:5050", "1.1.1.3:5050"},
		},
		{
			// update additional masters with an empty slice
			// expect empty masters
			masterInfos(),
			[]string{""},
		},
		{
			// update masters with a niladic value
			// expect empty masters
			nil,
			[]string{""},
		},
	} {
		m.UpdatedMasters(tt.masters)

		if got := recv(ch); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test #%d: got %#v, want: %#v", i, got, tt.want)
		}
	}
}

func TestMasters_OnMasterChanged(t *testing.T) {
	// create a new masters detector with an unknown leader
	// and two initial masters "1.1.1.1:5050", "1.1.1.2:5050"
	ch := make(chan []string, 1)
	m := NewMasters([]string{"1.1.1.1:5050", "1.1.1.2:5050"}, ch)

	for i, tt := range []struct {
		leader *mesos.MasterInfo
		want   []string
	}{
		{
			// update new leader "1.1.1.1",
			// expect an appended port number
			// leaving "1.1.1.2:5050" as the only additional master
			masterInfo(ip("1.1.1.1")),
			[]string{"1.1.1.1:5050", "1.1.1.2:5050"},
		},
		{
			// update new leader "1.1.1.3"
			// replacing "1.1.1.1:5050"
			masterInfo(ip("1.1.1.3")),
			[]string{"1.1.1.3:5050", "1.1.1.2:5050"},
		},
		{
			// update new leader "1.1.1.2"
			// replacing "1.1.1.3"
			masterInfo(ip("1.1.1.2")),
			[]string{"1.1.1.2:5050"},
		},
		{
			// update new leader with a niladic value
			// expect empty leader
			nil,
			[]string{""},
		},
	} {
		m.OnMasterChanged(tt.leader)

		if got := recv(ch); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test #%d: got %#v, want: %#v", i, got, tt.want)
		}
	}
}

func TestMasterAddr(t *testing.T) {
	for i, tt := range []struct {
		*mesos.MasterInfo
		addr string
	}{
		{nil, ""},
		{masterInfo(addr("", 1234)), ":1234"},
		{masterInfo(addr("1.1.1.1", 0)), "1.1.1.1:0"},
		{masterInfo(addr("1.1.1.1", 1234)), "1.1.1.1:1234"},
		{masterInfo(port(1234)), "0.0.0.0:1234"},
		{masterInfo(ip("1.1.1.1")), "1.1.1.1:5050"},
		{masterInfo(ip("1.1.1.1"), port(1234)), "1.1.1.1:1234"},
	} {
		if got, want := masterAddr(tt.MasterInfo), tt.addr; got != want {
			t.Errorf("test #%d: got %v, want %v", i, got, want)
		}
	}
}

// recv receives from a channel in a non-blocking way, returning the received value or nil.
func recv(ch <-chan []string) []string {
	select {
	case val := <-ch:
		return val
	default:
		return nil
	}
}

// masterInfoOpt is a functional option type for *mesos.MasterInfo structs
type masterInfoOpt func(*mesos.MasterInfo)

// masterInfo returns a *mesos.MasterInfo with the given opts applied.
func masterInfo(opts ...masterInfoOpt) *mesos.MasterInfo {
	var info mesos.MasterInfo
	for _, opt := range opts {
		opt(&info)
	}
	return &info
}

// masterInfos is an utility function that simply returns the given masterInfos.
func masterInfos(infos ...*mesos.MasterInfo) []*mesos.MasterInfo { return infos }

// ip returns a masterInfoOpt that sets the Ip field to the given value.
func ip(a string) masterInfoOpt {
	return func(info *mesos.MasterInfo) {
		ipv4 := byteOrder.Uint32(net.ParseIP(a).To4())
		info.Ip = &ipv4
	}
}

// port returns a masterInfoOpt that sets the Port field to the given value.
func port(n uint32) masterInfoOpt {
	return func(info *mesos.MasterInfo) {
		info.Port = &n
	}
}

// addr returns a masterInfoOpt that sets the Address field with the given
// ip and port.
func addr(ip string, port int32) masterInfoOpt {
	return func(info *mesos.MasterInfo) {
		info.Address = &mesos.Address{Ip: &ip, Port: &port}
	}
}
