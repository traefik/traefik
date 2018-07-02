package tracer

import (
	cryptorand "crypto/rand"
	"log"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

// random holds a thread-safe source of random numbers.
var random *rand.Rand

func init() {
	var seed int64
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
	if err == nil {
		seed = n.Int64()
	} else {
		log.Printf("%scannot generate random seed: %v; using current time\n", errorPrefix, err)
		seed = time.Now().UnixNano()
	}
	random = rand.New(&safeSource{
		source: rand.NewSource(seed),
	})
}

// safeSource holds a thread-safe implementation of rand.Source64.
type safeSource struct {
	source rand.Source
	sync.Mutex
}

func (rs *safeSource) Int63() int64 {
	rs.Lock()
	n := rs.source.Int63()
	rs.Unlock()

	return n
}

func (rs *safeSource) Uint64() uint64 { return uint64(rs.Int63()) }

func (rs *safeSource) Seed(seed int64) {
	rs.Lock()
	rs.Seed(seed)
	rs.Unlock()
}
