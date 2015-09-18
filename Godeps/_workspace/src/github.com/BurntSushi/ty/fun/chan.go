package fun

import (
	"reflect"

	"github.com/BurntSushi/ty"
)

// AsyncChan has a parametric type:
//
//	func AsyncChan(chan A) (send chan<- A, recv <-chan A)
//
// AsyncChan provides a channel abstraction without a fixed size buffer.
// The input should be a pointer to a channel that has a type without a
// direction, e.g., `new(chan int)`. Two new channels are returned: `send` and
// `recv`. The caller must send data on the `send` channel and receive data on
// the `recv` channel.
//
// Implementation is inspired by Kyle Lemons' work:
// https://github.com/kylelemons/iq/blob/master/iq_slice.go
func AsyncChan(baseChan interface{}) (send, recv interface{}) {
	chk := ty.Check(
		new(func(*chan ty.A) (chan ty.A, chan ty.A)),
		baseChan)

	// We don't care about the baseChan---it is only used to construct
	// the return types.
	tsend, trecv := chk.Returns[0], chk.Returns[1]

	buf := make([]reflect.Value, 0, 10)
	rsend := reflect.MakeChan(tsend, 0)
	rrecv := reflect.MakeChan(trecv, 0)

	go func() {
		defer rrecv.Close()

	BUFLOOP:
		for {
			if len(buf) == 0 {
				rv, ok := rsend.Recv()
				if !ok {
					break BUFLOOP
				}
				buf = append(buf, rv)
			}

			cases := []reflect.SelectCase{
				// case v, ok := <-send
				{
					Dir:  reflect.SelectRecv,
					Chan: rsend,
				},
				// case recv <- buf[0]
				{
					Dir:  reflect.SelectSend,
					Chan: rrecv,
					Send: buf[0],
				},
			}
			choice, rval, rok := reflect.Select(cases)
			switch choice {
			case 0:
				// case v, ok := <-send
				if !rok {
					break BUFLOOP
				}
				buf = append(buf, rval)
			case 1:
				// case recv <- buf[0]
				buf = buf[1:]
			default:
				panic("bug")
			}
		}
		for _, rv := range buf {
			rrecv.Send(rv)
		}
	}()

	// Create the directional channel types.
	tsDir := reflect.ChanOf(reflect.SendDir, tsend.Elem())
	trDir := reflect.ChanOf(reflect.RecvDir, trecv.Elem())
	return rsend.Convert(tsDir).Interface(), rrecv.Convert(trDir).Interface()
}
