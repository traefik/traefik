package snapshot

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/raft"
)

func TestArchive(t *testing.T) {
	// Create some fake snapshot data.
	metadata := raft.SnapshotMeta{
		Index: 2005,
		Term:  2011,
		Configuration: raft.Configuration{
			Servers: []raft.Server{
				raft.Server{
					Suffrage: raft.Voter,
					ID:       raft.ServerID("hello"),
					Address:  raft.ServerAddress("127.0.0.1:8300"),
				},
			},
		},
		Size: 1024,
	}
	var snap bytes.Buffer
	var expected bytes.Buffer
	both := io.MultiWriter(&snap, &expected)
	if _, err := io.Copy(both, io.LimitReader(rand.Reader, 1024)); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Write out the snapshot.
	var archive bytes.Buffer
	if err := write(&archive, &metadata, &snap); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Read the snapshot back.
	var newMeta raft.SnapshotMeta
	var newSnap bytes.Buffer
	if err := read(&archive, &newMeta, &newSnap); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the contents.
	if !reflect.DeepEqual(newMeta, metadata) {
		t.Fatalf("bad: %#v", newMeta)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, &newSnap); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), expected.Bytes()) {
		t.Fatalf("snapshot contents didn't match")
	}
}

func TestArchive_BadData(t *testing.T) {
	cases := []struct {
		Name  string
		Error string
	}{
		{"../test/snapshot/empty.tar", "failed checking integrity of snapshot"},
		{"../test/snapshot/extra.tar", "unexpected file \"nope\""},
		{"../test/snapshot/missing-meta.tar", "hash check failed for \"meta.json\""},
		{"../test/snapshot/missing-state.tar", "hash check failed for \"state.bin\""},
		{"../test/snapshot/missing-sha.tar", "file missing"},
		{"../test/snapshot/corrupt-meta.tar", "hash check failed for \"meta.json\""},
		{"../test/snapshot/corrupt-state.tar", "hash check failed for \"state.bin\""},
		{"../test/snapshot/corrupt-sha.tar", "list missing hash for \"nope\""},
	}
	for i, c := range cases {
		f, err := os.Open(c.Name)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer f.Close()

		var metadata raft.SnapshotMeta
		err = read(f, &metadata, ioutil.Discard)
		if err == nil || !strings.Contains(err.Error(), c.Error) {
			t.Fatalf("case %d (%s): %v", i, c.Name, err)
		}
	}
}

func TestArchive_hashList(t *testing.T) {
	hl := newHashList()
	for i := 0; i < 16; i++ {
		h := hl.Add(fmt.Sprintf("file-%d", i))
		if _, err := io.CopyN(h, rand.Reader, 32); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Do a normal round trip.
	var buf bytes.Buffer
	if err := hl.Encode(&buf); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := hl.DecodeAndVerify(&buf); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Have a local hash that isn't in the file.
	buf.Reset()
	if err := hl.Encode(&buf); err != nil {
		t.Fatalf("err: %v", err)
	}
	hl.Add("nope")
	err := hl.DecodeAndVerify(&buf)
	if err == nil || !strings.Contains(err.Error(), "file missing for \"nope\"") {
		t.Fatalf("err: %v", err)
	}

	// Have a hash in the file that we haven't seen locally.
	buf.Reset()
	if err := hl.Encode(&buf); err != nil {
		t.Fatalf("err: %v", err)
	}
	delete(hl.hashes, "nope")
	err = hl.DecodeAndVerify(&buf)
	if err == nil || !strings.Contains(err.Error(), "list missing hash for \"nope\"") {
		t.Fatalf("err: %v", err)
	}
}
