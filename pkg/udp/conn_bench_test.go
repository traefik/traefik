package udp

import (
	"flag"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

var packetSize int

func init() {
	flag.IntVar(&packetSize, "packetSize", 8192, "custom argument for packet size")
}

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

	// Send packets and measure readLoop processing
	packet := make([]byte, packetSize)
	go func() {
		defer udpConn.Close()

		for range b.N {
			n, err := clientConn.Write(packet)
			if err != nil {
				b.Errorf("Failed to send packet: %v", err)
				return
			}
			b.SetBytes(int64(n))
		}

		listener.accepting = false
	}()

	// Goroutine to consume from acceptCh
	go func() {
		for conn := range listener.acceptCh {
			go func(c *Conn) {
				//nolint:revive // drain receiveCh to prevent blocking
				for range c.receiveCh {
				}
			}(conn)
		}
	}()

	runtime.GC()

	b.ReportAllocs()
	b.ResetTimer()

	// Run the actual readLoop - it will process b.N packets then exit
	listener.readLoop()

	b.StopTimer()
}
