// snapshot manages the interactions between Consul and Raft in order to take
// and restore snapshots for disaster recovery. The internal format of a
// snapshot is simply a tar file, as described in archive.go.
package snapshot

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

// Snapshot is a structure that holds state about a temporary file that is used
// to hold a snapshot. By using an intermediate file we avoid holding everything
// in memory.
type Snapshot struct {
	file  *os.File
	index uint64
}

// New takes a state snapshot of the given Raft instance into a temporary file
// and returns an object that gives access to the file as an io.Reader. You must
// arrange to call Close() on the returned object or else you will leak a
// temporary file.
func New(logger *log.Logger, r *raft.Raft) (*Snapshot, error) {
	// Take the snapshot.
	future := r.Snapshot()
	if err := future.Error(); err != nil {
		return nil, fmt.Errorf("Raft error when taking snapshot: %v", err)
	}

	// Open up the snapshot.
	metadata, snap, err := future.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open snapshot: %v:", err)
	}
	defer func() {
		if err := snap.Close(); err != nil {
			logger.Printf("[ERR] snapshot: Failed to close Raft snapshot: %v", err)
		}
	}()

	// Make a scratch file to receive the contents so that we don't buffer
	// everything in memory. This gets deleted in Close() since we keep it
	// around for re-reading.
	archive, err := ioutil.TempFile("", "snapshot")
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot file: %v", err)
	}

	// If anything goes wrong after this point, we will attempt to clean up
	// the temp file. The happy path will disarm this.
	var keep bool
	defer func() {
		if keep {
			return
		}

		if err := os.Remove(archive.Name()); err != nil {
			logger.Printf("[ERR] snapshot: Failed to clean up temp snapshot: %v", err)
		}
	}()

	// Wrap the file writer in a gzip compressor.
	compressor := gzip.NewWriter(archive)

	// Write the archive.
	if err := write(compressor, metadata, snap); err != nil {
		return nil, fmt.Errorf("failed to write snapshot file: %v", err)
	}

	// Finish the compressed stream.
	if err := compressor.Close(); err != nil {
		return nil, fmt.Errorf("failed to compress snapshot file: %v", err)
	}

	// Sync the compressed file and rewind it so it's ready to be streamed
	// out by the caller.
	if err := archive.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync snapshot: %v", err)
	}
	if _, err := archive.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind snapshot: %v", err)
	}

	keep = true
	return &Snapshot{archive, metadata.Index}, nil
}

// Index returns the index of the snapshot. This is safe to call on a nil
// snapshot, it will just return 0.
func (s *Snapshot) Index() uint64 {
	if s == nil {
		return 0
	}
	return s.index
}

// Read passes through to the underlying snapshot file. This is safe to call on
// a nil snapshot, it will just return an EOF.
func (s *Snapshot) Read(p []byte) (n int, err error) {
	if s == nil {
		return 0, io.EOF
	}
	return s.file.Read(p)
}

// Close closes the snapshot and removes any temporary storage associated with
// it. You must arrange to call this whenever NewSnapshot() has been called
// successfully. This is safe to call on a nil snapshot.
func (s *Snapshot) Close() error {
	if s == nil {
		return nil
	}

	if err := s.file.Close(); err != nil {
		return err
	}
	return os.Remove(s.file.Name())
}

// Verify takes the snapshot from the reader and verifies its contents.
func Verify(in io.Reader) (*raft.SnapshotMeta, error) {
	// Wrap the reader in a gzip decompressor.
	decomp, err := gzip.NewReader(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress snapshot: %v", err)
	}
	defer decomp.Close()

	// Read the archive, throwing away the snapshot data.
	var metadata raft.SnapshotMeta
	if err := read(decomp, &metadata, ioutil.Discard); err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %v", err)
	}
	return &metadata, nil
}

// Restore takes the snapshot from the reader and attempts to apply it to the
// given Raft instance.
func Restore(logger *log.Logger, in io.Reader, r *raft.Raft) error {
	// Wrap the reader in a gzip decompressor.
	decomp, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("failed to decompress snapshot: %v", err)
	}
	defer func() {
		if err := decomp.Close(); err != nil {
			logger.Printf("[ERR] snapshot: Failed to close snapshot decompressor: %v", err)
		}
	}()

	// Make a scratch file to receive the contents of the snapshot data so
	// we can avoid buffering in memory.
	snap, err := ioutil.TempFile("", "snapshot")
	if err != nil {
		return fmt.Errorf("failed to create temp snapshot file: %v", err)
	}
	defer func() {
		if err := snap.Close(); err != nil {
			logger.Printf("[ERR] snapshot: Failed to close temp snapshot: %v", err)
		}
		if err := os.Remove(snap.Name()); err != nil {
			logger.Printf("[ERR] snapshot: Failed to clean up temp snapshot: %v", err)
		}
	}()

	// Read the archive.
	var metadata raft.SnapshotMeta
	if err := read(decomp, &metadata, snap); err != nil {
		return fmt.Errorf("failed to read snapshot file: %v", err)
	}

	// Sync and rewind the file so it's ready to be read again.
	if err := snap.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp snapshot: %v", err)
	}
	if _, err := snap.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind temp snapshot: %v", err)
	}

	// Feed the snapshot into Raft.
	if err := r.Restore(&metadata, snap, 60*time.Second); err != nil {
		return fmt.Errorf("Raft error when restoring snapshot: %v", err)
	}

	return nil
}
