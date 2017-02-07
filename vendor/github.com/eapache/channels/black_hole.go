package channels

// BlackHole implements the InChannel interface and provides an analogue for the "Discard" variable in
// the ioutil package - it never blocks, and simply discards every value it reads. The number of items
// discarded in this way is counted and returned from Len.
type BlackHole struct {
	input  chan interface{}
	length chan int
	count  int
}

func NewBlackHole() *BlackHole {
	ch := &BlackHole{
		input:  make(chan interface{}),
		length: make(chan int),
	}
	go ch.discard()
	return ch
}

func (ch *BlackHole) In() chan<- interface{} {
	return ch.input
}

func (ch *BlackHole) Len() int {
	val, open := <-ch.length
	if open {
		return val
	} else {
		return ch.count
	}
}

func (ch *BlackHole) Cap() BufferCap {
	return Infinity
}

func (ch *BlackHole) Close() {
	close(ch.input)
}

func (ch *BlackHole) discard() {
	for {
		select {
		case _, open := <-ch.input:
			if !open {
				close(ch.length)
				return
			}
			ch.count++
		case ch.length <- ch.count:
		}
	}
}
