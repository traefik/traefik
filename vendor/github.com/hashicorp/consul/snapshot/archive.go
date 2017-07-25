// The archive utilities manage the internal format of a snapshot, which is a
// tar file with the following contents:
//
// meta.json  - JSON-encoded snapshot metadata from Raft
// state.bin  - Encoded snapshot data from Raft
// SHA256SUMS - SHA-256 sums of the above two files
//
// The integrity information is automatically created and checked, and a failure
// there just looks like an error to the caller.
package snapshot

import (
	"archive/tar"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/hashicorp/raft"
)

// hashList manages a list of filenames and their hashes.
type hashList struct {
	hashes map[string]hash.Hash
}

// newHashList returns a new hashList.
func newHashList() *hashList {
	return &hashList{
		hashes: make(map[string]hash.Hash),
	}
}

// Add creates a new hash for the given file.
func (hl *hashList) Add(file string) hash.Hash {
	if existing, ok := hl.hashes[file]; ok {
		return existing
	}

	h := sha256.New()
	hl.hashes[file] = h
	return h
}

// Encode takes the current sum of all the hashes and saves the hash list as a
// SHA256SUMS-style text file.
func (hl *hashList) Encode(w io.Writer) error {
	for file, h := range hl.hashes {
		if _, err := fmt.Fprintf(w, "%x  %s\n", h.Sum([]byte{}), file); err != nil {
			return err
		}
	}
	return nil
}

// DecodeAndVerify reads a SHA256SUMS-style text file and checks the results
// against the current sums for all the hashes.
func (hl *hashList) DecodeAndVerify(r io.Reader) error {
	// Read the file and make sure everything in there has a matching hash.
	seen := make(map[string]struct{})
	s := bufio.NewScanner(r)
	for s.Scan() {
		sha := make([]byte, sha256.Size)
		var file string
		if _, err := fmt.Sscanf(s.Text(), "%x  %s", &sha, &file); err != nil {
			return err
		}

		h, ok := hl.hashes[file]
		if !ok {
			return fmt.Errorf("list missing hash for %q", file)
		}
		if !bytes.Equal(sha, h.Sum([]byte{})) {
			return fmt.Errorf("hash check failed for %q", file)
		}
		seen[file] = struct{}{}
	}
	if err := s.Err(); err != nil {
		return err
	}

	// Make sure everything we had a hash for was seen.
	for file, _ := range hl.hashes {
		if _, ok := seen[file]; !ok {
			return fmt.Errorf("file missing for %q", file)
		}
	}

	return nil
}

// write takes a writer and creates an archive with the snapshot metadata,
// the snapshot itself, and adds some integrity checking information.
func write(out io.Writer, metadata *raft.SnapshotMeta, snap io.Reader) error {
	// Start a new tarball.
	now := time.Now()
	archive := tar.NewWriter(out)

	// Create a hash list that we will use to write a SHA256SUMS file into
	// the archive.
	hl := newHashList()

	// Encode the snapshot metadata, which we need to feed back during a
	// restore.
	metaHash := hl.Add("meta.json")
	var metaBuffer bytes.Buffer
	enc := json.NewEncoder(&metaBuffer)
	if err := enc.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode snapshot metadata: %v", err)
	}
	if err := archive.WriteHeader(&tar.Header{
		Name:    "meta.json",
		Mode:    0600,
		Size:    int64(metaBuffer.Len()),
		ModTime: now,
	}); err != nil {
		return fmt.Errorf("failed to write snapshot metadata header: %v", err)
	}
	if _, err := io.Copy(archive, io.TeeReader(&metaBuffer, metaHash)); err != nil {
		return fmt.Errorf("failed to write snapshot metadata: %v", err)
	}

	// Copy the snapshot data given the size from the metadata.
	snapHash := hl.Add("state.bin")
	if err := archive.WriteHeader(&tar.Header{
		Name:    "state.bin",
		Mode:    0600,
		Size:    metadata.Size,
		ModTime: now,
	}); err != nil {
		return fmt.Errorf("failed to write snapshot data header: %v", err)
	}
	if _, err := io.CopyN(archive, io.TeeReader(snap, snapHash), metadata.Size); err != nil {
		return fmt.Errorf("failed to write snapshot metadata: %v", err)
	}

	// Create a SHA256SUMS file that we can use to verify on restore.
	var shaBuffer bytes.Buffer
	if err := hl.Encode(&shaBuffer); err != nil {
		return fmt.Errorf("failed to encode snapshot hashes: %v", err)
	}
	if err := archive.WriteHeader(&tar.Header{
		Name:    "SHA256SUMS",
		Mode:    0600,
		Size:    int64(shaBuffer.Len()),
		ModTime: now,
	}); err != nil {
		return fmt.Errorf("failed to write snapshot hashes header: %v", err)
	}
	if _, err := io.Copy(archive, &shaBuffer); err != nil {
		return fmt.Errorf("failed to write snapshot metadata: %v", err)
	}

	// Finalize the archive.
	if err := archive.Close(); err != nil {
		return fmt.Errorf("failed to finalize snapshot: %v", err)
	}

	return nil
}

// read takes a reader and extracts the snapshot metadata and the snapshot
// itself, and also checks the integrity of the data. You must arrange to call
// Close() on the returned object or else you will leak a temporary file.
func read(in io.Reader, metadata *raft.SnapshotMeta, snap io.Writer) error {
	// Start a new tar reader.
	archive := tar.NewReader(in)

	// Create a hash list that we will use to compare with the SHA256SUMS
	// file in the archive.
	hl := newHashList()

	// Populate the hashes for all the files we expect to see. The check at
	// the end will make sure these are all present in the SHA256SUMS file
	// and that the hashes match.
	metaHash := hl.Add("meta.json")
	snapHash := hl.Add("state.bin")

	// Look through the archive for the pieces we care about.
	var shaBuffer bytes.Buffer
	for {
		hdr, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading snapshot: %v", err)
		}

		switch hdr.Name {
		case "meta.json":
			dec := json.NewDecoder(io.TeeReader(archive, metaHash))
			if err := dec.Decode(&metadata); err != nil {
				return fmt.Errorf("failed to decode snapshot metadata: %v", err)
			}

		case "state.bin":
			if _, err := io.Copy(io.MultiWriter(snap, snapHash), archive); err != nil {
				return fmt.Errorf("failed to read or write snapshot data: %v", err)
			}

		case "SHA256SUMS":
			if _, err := io.Copy(&shaBuffer, archive); err != nil {
				return fmt.Errorf("failed to read snapshot hashes: %v", err)
			}

		default:
			return fmt.Errorf("unexpected file %q in snapshot", hdr.Name)
		}

	}

	// Verify all the hashes.
	if err := hl.DecodeAndVerify(&shaBuffer); err != nil {
		return fmt.Errorf("failed checking integrity of snapshot: %v", err)
	}

	return nil
}
