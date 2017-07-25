package main_test

import (
	"bytes"
	crypto "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/boltdb/bolt/cmd/bolt"
)

// Ensure the "info" command can print information about a database.
func TestInfoCommand_Run(t *testing.T) {
	db := MustOpen(0666, nil)
	db.DB.Close()
	defer db.Close()

	// Run the info command.
	m := NewMain()
	if err := m.Run("info", db.Path); err != nil {
		t.Fatal(err)
	}
}

// Ensure the "stats" command executes correctly with an empty database.
func TestStatsCommand_Run_EmptyDatabase(t *testing.T) {
	// Ignore
	if os.Getpagesize() != 4096 {
		t.Skip("system does not use 4KB page size")
	}

	db := MustOpen(0666, nil)
	defer db.Close()
	db.DB.Close()

	// Generate expected result.
	exp := "Aggregate statistics for 0 buckets\n\n" +
		"Page count statistics\n" +
		"\tNumber of logical branch pages: 0\n" +
		"\tNumber of physical branch overflow pages: 0\n" +
		"\tNumber of logical leaf pages: 0\n" +
		"\tNumber of physical leaf overflow pages: 0\n" +
		"Tree statistics\n" +
		"\tNumber of keys/value pairs: 0\n" +
		"\tNumber of levels in B+tree: 0\n" +
		"Page size utilization\n" +
		"\tBytes allocated for physical branch pages: 0\n" +
		"\tBytes actually used for branch data: 0 (0%)\n" +
		"\tBytes allocated for physical leaf pages: 0\n" +
		"\tBytes actually used for leaf data: 0 (0%)\n" +
		"Bucket statistics\n" +
		"\tTotal number of buckets: 0\n" +
		"\tTotal number on inlined buckets: 0 (0%)\n" +
		"\tBytes used for inlined buckets: 0 (0%)\n"

	// Run the command.
	m := NewMain()
	if err := m.Run("stats", db.Path); err != nil {
		t.Fatal(err)
	} else if m.Stdout.String() != exp {
		t.Fatalf("unexpected stdout:\n\n%s", m.Stdout.String())
	}
}

// Ensure the "stats" command can execute correctly.
func TestStatsCommand_Run(t *testing.T) {
	// Ignore
	if os.Getpagesize() != 4096 {
		t.Skip("system does not use 4KB page size")
	}

	db := MustOpen(0666, nil)
	defer db.Close()

	if err := db.Update(func(tx *bolt.Tx) error {
		// Create "foo" bucket.
		b, err := tx.CreateBucket([]byte("foo"))
		if err != nil {
			return err
		}
		for i := 0; i < 10; i++ {
			if err := b.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i))); err != nil {
				return err
			}
		}

		// Create "bar" bucket.
		b, err = tx.CreateBucket([]byte("bar"))
		if err != nil {
			return err
		}
		for i := 0; i < 100; i++ {
			if err := b.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i))); err != nil {
				return err
			}
		}

		// Create "baz" bucket.
		b, err = tx.CreateBucket([]byte("baz"))
		if err != nil {
			return err
		}
		if err := b.Put([]byte("key"), []byte("value")); err != nil {
			return err
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
	db.DB.Close()

	// Generate expected result.
	exp := "Aggregate statistics for 3 buckets\n\n" +
		"Page count statistics\n" +
		"\tNumber of logical branch pages: 0\n" +
		"\tNumber of physical branch overflow pages: 0\n" +
		"\tNumber of logical leaf pages: 1\n" +
		"\tNumber of physical leaf overflow pages: 0\n" +
		"Tree statistics\n" +
		"\tNumber of keys/value pairs: 111\n" +
		"\tNumber of levels in B+tree: 1\n" +
		"Page size utilization\n" +
		"\tBytes allocated for physical branch pages: 0\n" +
		"\tBytes actually used for branch data: 0 (0%)\n" +
		"\tBytes allocated for physical leaf pages: 4096\n" +
		"\tBytes actually used for leaf data: 1996 (48%)\n" +
		"Bucket statistics\n" +
		"\tTotal number of buckets: 3\n" +
		"\tTotal number on inlined buckets: 2 (66%)\n" +
		"\tBytes used for inlined buckets: 236 (11%)\n"

	// Run the command.
	m := NewMain()
	if err := m.Run("stats", db.Path); err != nil {
		t.Fatal(err)
	} else if m.Stdout.String() != exp {
		t.Fatalf("unexpected stdout:\n\n%s", m.Stdout.String())
	}
}

// Main represents a test wrapper for main.Main that records output.
type Main struct {
	*main.Main
	Stdin  bytes.Buffer
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// NewMain returns a new instance of Main.
func NewMain() *Main {
	m := &Main{Main: main.NewMain()}
	m.Main.Stdin = &m.Stdin
	m.Main.Stdout = &m.Stdout
	m.Main.Stderr = &m.Stderr
	return m
}

// MustOpen creates a Bolt database in a temporary location.
func MustOpen(mode os.FileMode, options *bolt.Options) *DB {
	// Create temporary path.
	f, _ := ioutil.TempFile("", "bolt-")
	f.Close()
	os.Remove(f.Name())

	db, err := bolt.Open(f.Name(), mode, options)
	if err != nil {
		panic(err.Error())
	}
	return &DB{DB: db, Path: f.Name()}
}

// DB is a test wrapper for bolt.DB.
type DB struct {
	*bolt.DB
	Path string
}

// Close closes and removes the database.
func (db *DB) Close() error {
	defer os.Remove(db.Path)
	return db.DB.Close()
}

func TestCompactCommand_Run(t *testing.T) {
	var s int64
	if err := binary.Read(crypto.Reader, binary.BigEndian, &s); err != nil {
		t.Fatal(err)
	}
	rand.Seed(s)

	dstdb := MustOpen(0666, nil)
	dstdb.Close()

	// fill the db
	db := MustOpen(0666, nil)
	if err := db.Update(func(tx *bolt.Tx) error {
		n := 2 + rand.Intn(5)
		for i := 0; i < n; i++ {
			k := []byte(fmt.Sprintf("b%d", i))
			b, err := tx.CreateBucketIfNotExists(k)
			if err != nil {
				return err
			}
			if err := b.SetSequence(uint64(i)); err != nil {
				return err
			}
			if err := fillBucket(b, append(k, '.')); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		t.Fatal(err)
	}

	// make the db grow by adding large values, and delete them.
	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("large_vals"))
		if err != nil {
			return err
		}
		n := 5 + rand.Intn(5)
		for i := 0; i < n; i++ {
			v := make([]byte, 1000*1000*(1+rand.Intn(5)))
			_, err := crypto.Read(v)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(fmt.Sprintf("l%d", i)), v); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("large_vals")).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if err := c.Delete(); err != nil {
				return err
			}
		}
		return tx.DeleteBucket([]byte("large_vals"))
	}); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.DB.Close()
	defer db.Close()
	defer dstdb.Close()

	dbChk, err := chkdb(db.Path)
	if err != nil {
		t.Fatal(err)
	}

	m := NewMain()
	if err := m.Run("compact", "-o", dstdb.Path, db.Path); err != nil {
		t.Fatal(err)
	}

	dbChkAfterCompact, err := chkdb(db.Path)
	if err != nil {
		t.Fatal(err)
	}

	dstdbChk, err := chkdb(dstdb.Path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(dbChk, dbChkAfterCompact) {
		t.Error("the original db has been touched")
	}
	if !bytes.Equal(dbChk, dstdbChk) {
		t.Error("the compacted db data isn't the same than the original db")
	}
}

func fillBucket(b *bolt.Bucket, prefix []byte) error {
	n := 10 + rand.Intn(50)
	for i := 0; i < n; i++ {
		v := make([]byte, 10*(1+rand.Intn(4)))
		_, err := crypto.Read(v)
		if err != nil {
			return err
		}
		k := append(prefix, []byte(fmt.Sprintf("k%d", i))...)
		if err := b.Put(k, v); err != nil {
			return err
		}
	}
	// limit depth of subbuckets
	s := 2 + rand.Intn(4)
	if len(prefix) > (2*s + 1) {
		return nil
	}
	n = 1 + rand.Intn(3)
	for i := 0; i < n; i++ {
		k := append(prefix, []byte(fmt.Sprintf("b%d", i))...)
		sb, err := b.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := fillBucket(sb, append(k, '.')); err != nil {
			return err
		}
	}
	return nil
}

func chkdb(path string) ([]byte, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var buf bytes.Buffer
	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return walkBucket(b, name, nil, &buf)
		})
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func walkBucket(parent *bolt.Bucket, k []byte, v []byte, w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%d:%x=%x\n", parent.Sequence(), k, v); err != nil {
		return err
	}

	// not a bucket, exit.
	if v != nil {
		return nil
	}
	return parent.ForEach(func(k, v []byte) error {
		if v == nil {
			return walkBucket(parent.Bucket(k), k, nil, w)
		}
		return walkBucket(parent, k, v, w)
	})
}
