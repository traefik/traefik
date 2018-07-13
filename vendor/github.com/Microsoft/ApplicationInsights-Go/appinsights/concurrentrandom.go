package appinsights

import (
	"encoding/base64"
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"
)

type concurrentRandom struct {
	channel chan string
	random  *rand.Rand
}

var randomGenerator *concurrentRandom

func newConcurrentRandom() *concurrentRandom {
	source := rand.NewSource(time.Now().UnixNano())
	return &concurrentRandom{
		channel: make(chan string, 4),
		random:  rand.New(source),
	}
}

func (generator *concurrentRandom) run() {
	buf := make([]byte, 8)
	for {
		generator.random.Read(buf)
		generator.channel <- base64.StdEncoding.EncodeToString(buf)
	}
}

func randomId() string {
	if randomGenerator == nil {
		r := newConcurrentRandom()
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&randomGenerator)), unsafe.Pointer(nil), unsafe.Pointer(r)) {
			go r.run()
		} else {
			close(r.channel)
		}
	}

	return <-randomGenerator.channel
}
