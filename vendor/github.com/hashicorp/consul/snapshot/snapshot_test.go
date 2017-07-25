package snapshot

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

// MockFSM is a simple FSM for testing that simply stores its logs in a slice of
// byte slices.
type MockFSM struct {
	sync.Mutex
	logs [][]byte
}

// MockSnapshot is a snapshot sink for testing that encodes the contents of a
// MockFSM using msgpack.
type MockSnapshot struct {
	logs     [][]byte
	maxIndex int
}

// See raft.FSM.
func (m *MockFSM) Apply(log *raft.Log) interface{} {
	m.Lock()
	defer m.Unlock()
	m.logs = append(m.logs, log.Data)
	return len(m.logs)
}

// See raft.FSM.
func (m *MockFSM) Snapshot() (raft.FSMSnapshot, error) {
	m.Lock()
	defer m.Unlock()
	return &MockSnapshot{m.logs, len(m.logs)}, nil
}

// See raft.FSM.
func (m *MockFSM) Restore(in io.ReadCloser) error {
	m.Lock()
	defer m.Unlock()
	defer in.Close()
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(in, &hd)

	m.logs = nil
	return dec.Decode(&m.logs)
}

// See raft.SnapshotSink.
func (m *MockSnapshot) Persist(sink raft.SnapshotSink) error {
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(sink, &hd)
	if err := enc.Encode(m.logs[:m.maxIndex]); err != nil {
		sink.Cancel()
		return err
	}
	sink.Close()
	return nil
}

// See raft.SnapshotSink.
func (m *MockSnapshot) Release() {
}

// makeRaft returns a Raft and its FSM, with snapshots based in the given dir.
func makeRaft(t *testing.T, dir string) (*raft.Raft, *MockFSM) {
	snaps, err := raft.NewFileSnapshotStore(dir, 5, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	fsm := &MockFSM{}
	store := raft.NewInmemStore()
	addr, trans := raft.NewInmemTransport("")

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(fmt.Sprintf("server-%s", addr))

	var members raft.Configuration
	members.Servers = append(members.Servers, raft.Server{
		Suffrage: raft.Voter,
		ID:       config.LocalID,
		Address:  addr,
	})

	err = raft.BootstrapCluster(config, store, store, snaps, trans, members)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	raft, err := raft.NewRaft(config, fsm, store, store, snaps, trans)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	timeout := time.After(10 * time.Second)
	for {
		if raft.Leader() != "" {
			break
		}

		select {
		case <-raft.LeaderCh():
		case <-time.After(1 * time.Second):
			// Need to poll because we might have missed the first
			// go with the leader channel.
		case <-timeout:
			t.Fatalf("timed out waiting for leader")
		}
	}

	return raft, fsm
}

func TestSnapshot(t *testing.T) {
	dir, err := ioutil.TempDir("", "snapshot")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(dir)

	// Make a Raft and populate it with some data. We tee everything we
	// apply off to a buffer for checking post-snapshot.
	var expected []bytes.Buffer
	entries := 64 * 1024
	before, _ := makeRaft(t, path.Join(dir, "before"))
	defer before.Shutdown()
	for i := 0; i < entries; i++ {
		var log bytes.Buffer
		var copy bytes.Buffer
		both := io.MultiWriter(&log, &copy)
		if _, err := io.CopyN(both, rand.Reader, 256); err != nil {
			t.Fatalf("err: %v", err)
		}
		future := before.Apply(log.Bytes(), time.Second)
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
		expected = append(expected, copy)
	}

	// Take a snapshot.
	logger := log.New(os.Stdout, "", 0)
	snap, err := New(logger, before)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer snap.Close()

	// Verify the snapshot. We have to rewind it after for the restore.
	metadata, err := Verify(snap)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := snap.file.Seek(0, 0); err != nil {
		t.Fatalf("err: %v", err)
	}
	if int(metadata.Index) != entries+2 {
		t.Fatalf("bad: %d", metadata.Index)
	}
	if metadata.Term != 2 {
		t.Fatalf("bad: %d", metadata.Index)
	}
	if metadata.Version != raft.SnapshotVersionMax {
		t.Fatalf("bad: %d", metadata.Version)
	}

	// Make a new, independent Raft.
	after, fsm := makeRaft(t, path.Join(dir, "after"))
	defer after.Shutdown()

	// Put some initial data in there that the snapshot should overwrite.
	for i := 0; i < 16; i++ {
		var log bytes.Buffer
		if _, err := io.CopyN(&log, rand.Reader, 256); err != nil {
			t.Fatalf("err: %v", err)
		}
		future := after.Apply(log.Bytes(), time.Second)
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Restore the snapshot.
	if err := Restore(logger, snap, after); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Compare the contents.
	fsm.Lock()
	defer fsm.Unlock()
	if len(fsm.logs) != len(expected) {
		t.Fatalf("bad: %d vs. %d", len(fsm.logs), len(expected))
	}
	for i, _ := range fsm.logs {
		if !bytes.Equal(fsm.logs[i], expected[i].Bytes()) {
			t.Fatalf("bad: log %d doesn't match", i)
		}
	}
}

func TestSnapshot_Nil(t *testing.T) {
	var snap *Snapshot

	if idx := snap.Index(); idx != 0 {
		t.Fatalf("bad: %d", idx)
	}

	n, err := snap.Read(make([]byte, 16))
	if n != 0 || err != io.EOF {
		t.Fatalf("bad: %d %v", n, err)
	}

	if err := snap.Close(); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestSnapshot_BadVerify(t *testing.T) {
	buf := bytes.NewBuffer([]byte("nope"))
	_, err := Verify(buf)
	if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fatalf("err: %v", err)
	}
}

func TestSnapshot_BadRestore(t *testing.T) {
	dir, err := ioutil.TempDir("", "snapshot")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer os.RemoveAll(dir)

	// Make a Raft and populate it with some data.
	before, _ := makeRaft(t, path.Join(dir, "before"))
	defer before.Shutdown()
	for i := 0; i < 16*1024; i++ {
		var log bytes.Buffer
		if _, err := io.CopyN(&log, rand.Reader, 256); err != nil {
			t.Fatalf("err: %v", err)
		}
		future := before.Apply(log.Bytes(), time.Second)
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Take a snapshot.
	logger := log.New(os.Stdout, "", 0)
	snap, err := New(logger, before)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make a new, independent Raft.
	after, fsm := makeRaft(t, path.Join(dir, "after"))
	defer after.Shutdown()

	// Put some initial data in there that should not be harmed by the
	// failed restore attempt.
	var expected []bytes.Buffer
	for i := 0; i < 16; i++ {
		var log bytes.Buffer
		var copy bytes.Buffer
		both := io.MultiWriter(&log, &copy)
		if _, err := io.CopyN(both, rand.Reader, 256); err != nil {
			t.Fatalf("err: %v", err)
		}
		future := after.Apply(log.Bytes(), time.Second)
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
		expected = append(expected, copy)
	}

	// Attempt to restore a truncated version of the snapshot. This is
	// expected to fail.
	err = Restore(logger, io.LimitReader(snap, 512), after)
	if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fatalf("err: %v", err)
	}

	// Compare the contents to make sure the aborted restore didn't harm
	// anything.
	fsm.Lock()
	defer fsm.Unlock()
	if len(fsm.logs) != len(expected) {
		t.Fatalf("bad: %d vs. %d", len(fsm.logs), len(expected))
	}
	for i, _ := range fsm.logs {
		if !bytes.Equal(fsm.logs[i], expected[i].Bytes()) {
			t.Fatalf("bad: log %d doesn't match", i)
		}
	}
}
