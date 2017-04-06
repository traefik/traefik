package detect

import (
	"encoding/binary"
	"net"
	"strconv"
	"unsafe"

	"github.com/mesos/mesos-go/detector"
	mesos "github.com/mesos/mesos-go/mesosproto"

	"github.com/mesosphere/mesos-dns/logging"
)

var (
	_ detector.MasterChanged = (*Masters)(nil)
	_ detector.AllMasters    = (*Masters)(nil)
)

// Masters detects changes of leader and/or master elections
// and sends these changes to a channel.
type Masters struct {
	// current masters list,
	// 1st item represents the leader,
	// the rest remaining masters
	masters []string

	// the channel leader/master changes are being sent to
	changed chan<- []string
}

// NewMasters returns a new Masters detector with the given initial masters
// and the given changed channel to which master changes will be sent to.
// Initially the leader is unknown which is represented by
// setting the first item of the sent masters slice to be empty.
func NewMasters(masters []string, changed chan<- []string) *Masters {
	return &Masters{
		masters: append([]string{""}, masters...),
		changed: changed,
	}
}

// OnMasterChanged sets the given MasterInfo as the current leader
// leaving the remaining masters unchanged and emits the current masters state.
// It implements the detector.MasterChanged interface.
func (ms *Masters) OnMasterChanged(leader *mesos.MasterInfo) {
	logging.VeryVerbose.Println("Updated leader: ", leader)
	ms.masters = ordered(masterAddr(leader), ms.masters[1:])
	emit(ms.changed, ms.masters)
}

// UpdatedMasters sets the given slice of MasterInfo as the current remaining masters
// leaving the current leader unchanged and emits the current masters state.
// It implements the detector.AllMasters interface.
func (ms *Masters) UpdatedMasters(infos []*mesos.MasterInfo) {
	logging.VeryVerbose.Println("Updated masters: ", infos)
	masters := make([]string, 0, len(infos))
	for _, info := range infos {
		if addr := masterAddr(info); addr != "" {
			masters = append(masters, addr)
		}
	}
	ms.masters = ordered(ms.masters[0], masters)
	emit(ms.changed, ms.masters)
}

func emit(ch chan<- []string, s []string) {
	ch <- append(make([]string, 0, len(s)), s...)
}

// ordered returns a slice of masters with the given leader in the first position
func ordered(leader string, masters []string) []string {
	ms := append(make([]string, 0, len(masters)+1), leader)
	for _, m := range masters {
		if m != leader {
			ms = append(ms, m)
		}
	}
	return ms
}

// masterAddr returns an address (ip:port) from the given *mesos.MasterInfo or
// an empty string if it nil.
//
// BUG(tsenart): The byte order of the `ip` field in MasterInfo is platform
// dependent. We assume that Mesos is compiled with the same architecture as
// Mesos-DNS and hence same byte order. If this isn't the case, the address
// returned will be wrong. This only affects Mesos versions < 0.24.0
func masterAddr(info *mesos.MasterInfo) string {
	if info == nil {
		return ""
	}
	ip, port := "", int64(0)
	if addr := info.GetAddress(); addr != nil { // Mesos >= 0.24.0
		ip, port = addr.GetIp(), int64(addr.GetPort())
	} else { // Mesos < 0.24.0
		ipv4 := make([]byte, net.IPv4len)
		byteOrder.PutUint32(ipv4, info.GetIp())
		ip, port = net.IP(ipv4).String(), int64(info.GetPort())
	}
	return net.JoinHostPort(ip, strconv.FormatInt(port, 10))
}

// byteOrder is instantiated at package initialization time to the
// binary.ByteOrder of the running process.
// https://groups.google.com/d/msg/golang-nuts/zmh64YkqOV8/iJe-TrTTeREJ
var byteOrder = func() binary.ByteOrder {
	switch x := uint32(0x01020304); *(*byte)(unsafe.Pointer(&x)) {
	case 0x01:
		return binary.BigEndian
	case 0x04:
		return binary.LittleEndian
	}
	panic("unknown byte order")
}()
