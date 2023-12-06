package httputil

import "sync"

const bufferSize = 32 * 1024

type bufferPool struct {
	pool sync.Pool
}

func newBufferPool() *bufferPool {
	b := &bufferPool{
		pool: sync.Pool{},
	}

	b.pool.New = func() interface{} {
		return make([]byte, bufferSize)
	}

	return b
}

func (b *bufferPool) Get() []byte {
	return b.pool.Get().([]byte)
}

func (b *bufferPool) Put(bytes []byte) {
	b.pool.Put(bytes)
}
