package channels

import "testing"

func TestSharedBufferSingleton(t *testing.T) {
	buf := NewSharedBuffer(3)

	ch := buf.NewChannel()
	for i := 0; i < 5; i++ {
		ch.In() <- (*int)(nil)
		ch.In() <- (*int)(nil)
		ch.In() <- (*int)(nil)
		select {
		case ch.In() <- (*int)(nil):
			t.Error("Wrote to full shared-buffer")
		default:
		}

		<-ch.Out()
		<-ch.Out()
		<-ch.Out()
		select {
		case <-ch.Out():
			t.Error("Read from empty shared-buffer")
		default:
		}
	}

	ch.Close()
	buf.Close()
}

func TestSharedBufferMultiple(t *testing.T) {
	buf := NewSharedBuffer(3)

	ch1 := buf.NewChannel()
	ch2 := buf.NewChannel()

	ch1.In() <- (*int)(nil)
	ch1.In() <- (*int)(nil)
	ch1.In() <- (*int)(nil)

	select {
	case ch2.In() <- (*int)(nil):
		t.Error("Wrote to full shared-buffer")
	case <-ch2.Out():
		t.Error("Read from empty channel")
	default:
	}

	<-ch1.Out()

	for i := 0; i < 10; i++ {
		ch2.In() <- (*int)(nil)

		select {
		case ch1.In() <- (*int)(nil):
			t.Error("Wrote to full shared-buffer")
		case ch2.In() <- (*int)(nil):
			t.Error("Wrote to full shared-buffer")
		default:
		}

		<-ch2.Out()
	}

	<-ch1.Out()
	<-ch1.Out()

	ch1.Close()
	ch2.Close()
	buf.Close()
}

func TestSharedBufferConcurrent(t *testing.T) {
	const threads = 10
	const iters = 200

	buf := NewSharedBuffer(3)
	done := make(chan bool)

	for i := 0; i < threads; i++ {
		go func() {
			ch := buf.NewChannel()
			for i := 0; i < iters; i++ {
				ch.In() <- i
				val := <-ch.Out()
				if val.(int) != i {
					t.Error("Mismatched value out of channel")
				}
			}
			ch.Close()
			done <- true
		}()
	}

	for i := 0; i < threads; i++ {
		<-done
	}
	close(done)
	buf.Close()
}

func ExampleSharedBuffer() {
	// never more than 3 elements in the pipeline at once
	buf := NewSharedBuffer(3)

	ch1 := buf.NewChannel()
	ch2 := buf.NewChannel()

	// or, instead of a straight pipe, implement your pipeline step
	Pipe(ch1, ch2)

	// inputs
	go func() {
		for i := 0; i < 20; i++ {
			ch1.In() <- i
		}
		ch1.Close()
	}()

	for _ = range ch2.Out() {
		// outputs
	}

	buf.Close()
}
