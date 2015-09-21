package graceful

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	killTime    = 500 * time.Millisecond
	timeoutTime = 1000 * time.Millisecond
	waitTime    = 100 * time.Millisecond
)

func runQuery(t *testing.T, expected int, shouldErr bool, wg *sync.WaitGroup, once *sync.Once) {
	wg.Add(1)
	defer wg.Done()
	client := http.Client{}
	r, err := client.Get("http://localhost:9654")
	if shouldErr && err == nil {
		once.Do(func() {
			t.Fatal("Expected an error but none was encountered.")
		})
	} else if shouldErr && err != nil {
		if checkErr(t, err, once) {
			return
		}
	}
	if r != nil && r.StatusCode != expected {
		once.Do(func() {
			t.Fatalf("Incorrect status code on response. Expected %d. Got %d", expected, r.StatusCode)
		})
	} else if r == nil {
		once.Do(func() {
			t.Fatal("No response when a response was expected.")
		})
	}
}

func checkErr(t *testing.T, err error, once *sync.Once) bool {
	if err.(*url.Error).Err == io.EOF {
		return true
	}
	errno := err.(*url.Error).Err.(*net.OpError).Err.(syscall.Errno)
	if errno == syscall.ECONNREFUSED {
		return true
	} else if err != nil {
		once.Do(func() {
			t.Fatal("Error on Get:", err)
		})
	}
	return false
}

func createListener(sleep time.Duration) (*http.Server, net.Listener, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(sleep)
		rw.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Addr: ":9654", Handler: mux}
	l, err := net.Listen("tcp", ":9654")
	if err != nil {
		fmt.Println(err)
	}
	return server, l, err
}

func runServer(timeout, sleep time.Duration, c chan os.Signal) error {
	server, l, err := createListener(sleep)
	if err != nil {
		return err
	}

	srv := &Server{Timeout: timeout, Server: server, interrupt: c}
	return srv.Serve(l)
}

func launchTestQueries(t *testing.T, wg *sync.WaitGroup, c chan os.Signal) {
	var once sync.Once
	for i := 0; i < 8; i++ {
		go runQuery(t, http.StatusOK, false, wg, &once)
	}

	time.Sleep(waitTime)
	c <- os.Interrupt
	time.Sleep(waitTime)

	for i := 0; i < 8; i++ {
		go runQuery(t, 0, true, wg, &once)
	}

	wg.Done()
}

func TestGracefulRun(t *testing.T) {
	c := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		runServer(killTime, killTime/2, c)
		wg.Done()
	}()

	wg.Add(1)
	go launchTestQueries(t, &wg, c)
	wg.Wait()
}

func TestGracefulRunTimesOut(t *testing.T) {
	c := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		runServer(killTime, killTime*10, c)
		wg.Done()
	}()

	var once sync.Once
	wg.Add(1)
	go func() {
		for i := 0; i < 8; i++ {
			go runQuery(t, 0, true, &wg, &once)
		}
		time.Sleep(waitTime)
		c <- os.Interrupt
		time.Sleep(waitTime)
		for i := 0; i < 8; i++ {
			go runQuery(t, 0, true, &wg, &once)
		}
		wg.Done()
	}()

	wg.Wait()

}

func TestGracefulRunDoesntTimeOut(t *testing.T) {
	c := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		runServer(0, killTime*2, c)
		wg.Done()
	}()

	wg.Add(1)
	go launchTestQueries(t, &wg, c)
	wg.Wait()
}

func TestGracefulRunNoRequests(t *testing.T) {
	c := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		runServer(0, killTime*2, c)
		wg.Done()
	}()

	c <- os.Interrupt

	wg.Wait()

}

func TestGracefulForwardsConnState(t *testing.T) {
	c := make(chan os.Signal, 1)
	states := make(map[http.ConnState]int)
	var stateLock sync.Mutex

	connState := func(conn net.Conn, state http.ConnState) {
		stateLock.Lock()
		states[state]++
		stateLock.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	expected := map[http.ConnState]int{
		http.StateNew:    8,
		http.StateActive: 8,
		http.StateClosed: 8,
	}

	go func() {
		server, l, _ := createListener(killTime / 2)
		srv := &Server{
			ConnState: connState,
			Timeout:   killTime,
			Server:    server,
			interrupt: c,
		}
		srv.Serve(l)

		wg.Done()
	}()

	wg.Add(1)
	go launchTestQueries(t, &wg, c)
	wg.Wait()

	stateLock.Lock()
	if !reflect.DeepEqual(states, expected) {
		t.Errorf("Incorrect connection state tracking.\n  actual: %v\nexpected: %v\n", states, expected)
	}
	stateLock.Unlock()
}

func TestGracefulExplicitStop(t *testing.T) {
	server, l, err := createListener(1 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{Timeout: killTime, Server: server}

	go func() {
		go srv.Serve(l)
		time.Sleep(waitTime)
		srv.Stop(killTime)
	}()

	// block on the stopChan until the server has shut down
	select {
	case <-srv.StopChan():
	case <-time.After(timeoutTime):
		t.Fatal("Timed out while waiting for explicit stop to complete")
	}
}

func TestGracefulExplicitStopOverride(t *testing.T) {
	server, l, err := createListener(1 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{Timeout: killTime, Server: server}

	go func() {
		go srv.Serve(l)
		time.Sleep(waitTime)
		srv.Stop(killTime / 2)
	}()

	// block on the stopChan until the server has shut down
	select {
	case <-srv.StopChan():
	case <-time.After(killTime):
		t.Fatal("Timed out while waiting for explicit stop to complete")
	}
}

func TestShutdownInitiatedCallback(t *testing.T) {
	server, l, err := createListener(1 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	called := make(chan struct{})
	cb := func() { close(called) }

	srv := &Server{Server: server, ShutdownInitiated: cb}

	go func() {
		go srv.Serve(l)
		time.Sleep(waitTime)
		srv.Stop(killTime)
	}()

	select {
	case <-called:
	case <-time.After(killTime):
		t.Fatal("Timed out while waiting for ShutdownInitiated callback to be called")
	}
}
func hijackingListener(srv *Server) (*http.Server, net.Listener, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		conn, bufrw, err := rw.(http.Hijacker).Hijack()
		if err != nil {
			http.Error(rw, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		defer conn.Close()

		bufrw.WriteString("HTTP/1.1 200 OK\r\n\r\n")
		bufrw.Flush()
	})

	server := &http.Server{Addr: ":9654", Handler: mux}
	l, err := net.Listen("tcp", ":9654")
	return server, l, err
}

func TestNotifyClosed(t *testing.T) {
	c := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	srv := &Server{Timeout: killTime, interrupt: c}
	server, l, err := hijackingListener(srv)
	if err != nil {
		t.Fatal(err)
	}

	srv.Server = server

	go func() {
		srv.Serve(l)
		wg.Done()
	}()

	var once sync.Once
	for i := 0; i < 8; i++ {
		runQuery(t, http.StatusOK, false, &wg, &once)
	}

	srv.Stop(0)

	// block on the stopChan until the server has shut down
	select {
	case <-srv.StopChan():
	case <-time.After(timeoutTime):
		t.Fatal("Timed out while waiting for explicit stop to complete")
	}

	if len(srv.connections) > 0 {
		t.Fatal("hijacked connections should not be managed")
	}

}

func TestStopDeadlock(t *testing.T) {
	c := make(chan struct{})

	server, l, err := createListener(1 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{Server: server, NoSignalHandling: true}

	go func() {
		time.Sleep(waitTime)
		srv.Serve(l)
	}()

	go func() {
		srv.Stop(0)
		close(c)
	}()

	select {
	case <-c:
		l.Close()
	case <-time.After(timeoutTime):
		t.Fatal("Timed out while waiting for explicit stop to complete")
	}
}

// Run with --race
func TestStopRace(t *testing.T) {
	server, l, err := createListener(1 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{Timeout: killTime, Server: server}

	go func() {
		go srv.Serve(l)
		srv.Stop(killTime)
	}()
	srv.Stop(0)
	select {
	case <-srv.StopChan():
	case <-time.After(timeoutTime):
		t.Fatal("Timed out while waiting for explicit stop to complete")
	}
}
