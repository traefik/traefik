package instana

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	seededIDGen  = rand.New(rand.NewSource(time.Now().UnixNano()))
	seededIDLock sync.Mutex
)

func randomID() int64 {
	seededIDLock.Lock()
	defer seededIDLock.Unlock()
	return int64(seededIDGen.Int63())
}

// ID2Header converts an Instana ID to a value that can be used in
// context propagation (such as HTTP headers).  More specifically,
// this converts a signed 64 bit integer into an unsigned hex string.
func ID2Header(id int64) (string, error) {
	// FIXME: We're assuming LittleEndian here

	// Write out _signed_ 64bit integer to byte buffer
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, id); err == nil {
		// Read bytes back into _unsigned_ 64 bit integer
		var unsigned uint64
		if err = binary.Read(buf, binary.LittleEndian, &unsigned); err == nil {
			// Convert uint64 to hex string equivalent and return that
			return strconv.FormatUint(unsigned, 16), nil
		}
		log.debug(err)
	} else {
		log.debug(err)
	}
	return "", errors.New("context corrupted; could not convert value")
}

// Header2ID converts an header context value into an Instana ID.  More
// specifically, this converts an unsigned 64 bit hex value into a signed
// 64bit integer.
func Header2ID(header string) (int64, error) {
	// FIXME: We're assuming LittleEndian here

	// Parse unsigned 64 bit hex string into unsigned 64 bit base 10 integer
	if unsignedID, err := strconv.ParseUint(header, 16, 64); err == nil {
		// Write out _unsigned_ 64bit integer to byte buffer
		buf := new(bytes.Buffer)
		if err = binary.Write(buf, binary.LittleEndian, unsignedID); err == nil {
			// Read bytes back into _signed_ 64 bit integer
			var signedID int64
			if err = binary.Read(buf, binary.LittleEndian, &signedID); err == nil {
				// The success case
				return signedID, nil
			}
			log.debug(err)
		} else {
			log.debug(err)
		}
	} else {
		log.debug(err)
	}
	return int64(0), errors.New("context corrupted; could not convert value")
}

func getCommandLine() (string, []string) {
	var cmdlinePath string = "/proc/" + strconv.Itoa(os.Getpid()) + "/cmdline"

	cmdline, err := ioutil.ReadFile(cmdlinePath)

	if err != nil {
		log.debug("No /proc.  Returning OS reported cmdline")
		return os.Args[0], os.Args[1:]
	}

	parts := strings.FieldsFunc(string(cmdline), func(c rune) bool {
		if c == '\u0000' {
			return true
		}
		return false
	})
	log.debug("cmdline says:", parts[0], parts[1:])
	return parts[0], parts[1:]
}

func abs(x int64) int64 {
	y := x >> 63
	return (x + y) ^ y
}
