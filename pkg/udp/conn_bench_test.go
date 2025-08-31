package udp

import (
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func BenchmarkReadLoopAllocations(b *testing.B) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		b.Fatalf("Failed to resolve UDP address: %v", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		b.Fatalf("Failed to create UDP connection: %v", err)
	}
	b.Cleanup(func() { udpConn.Close() })

	listener := &Listener{
		pConn:     udpConn,
		acceptCh:  make(chan *Conn, b.N), // Buffer for all expected connections
		conns:     make(map[string]*Conn),
		accepting: true,
		timeout:   3 * time.Second,
		readBufferPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, maxDatagramSize)
			},
		},
	}

	clientConn, err := net.Dial("udp", listener.Addr().String())
	if err != nil {
		b.Fatalf("Failed to create client connection: %v", err)
	}
	b.Cleanup(func() { clientConn.Close() })

	// Goroutine to consume from acceptCh
	go func() {
		for conn := range listener.acceptCh {
			// Drain receiveCh to prevent blocking
			go func(c *Conn) {
				for range c.receiveCh {
				}
			}(conn)
		}
	}()

	// Send packets and measure readLoop processing
	go func() {
		defer udpConn.Close()

		for i := 0; i < b.N; i++ {
			_, err := clientConn.Write([]byte("test"))
			if err != nil {
				b.Errorf("Failed to send packet: %v", err)
				return
			}
			time.Sleep(time.Microsecond) // Small delay between packets
		}

		listener.accepting = false
	}()

	runtime.GC()

	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()

	// Run the actual readLoop - it will process b.N packets then exit
	listener.readLoop()

	b.StopTimer()
}
