package server

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShutdownHTTP(t *testing.T) {
	entryPoint, err := NewTCPEntryPoint(context.Background(), &static.EntryPoint{
		Address: ":0",
		Transport: &static.EntryPointsTransport{
			LifeCycle: &static.LifeCycle{
				RequestAcceptGraceTimeout: 0,
				GraceTimeOut:              5000000000,
			},
		},
		ForwardedHeaders: &static.ForwardedHeaders{},
	})
	require.NoError(t, err)

	go entryPoint.startTCP(context.Background())

	router := &tcp.Router{}
	router.HTTPHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Second * 1)
		rw.WriteHeader(http.StatusOK)
	}))
	entryPoint.switchRouter(router)

	conn, err := net.Dial("tcp", entryPoint.listener.Addr().String())
	require.NoError(t, err)

	go entryPoint.Shutdown(context.Background())

	request, err := http.NewRequest("GET", "http://127.0.0.1:8082", nil)
	require.NoError(t, err)

	err = request.Write(conn)
	require.NoError(t, err)

	resp, err := http.ReadResponse(bufio.NewReader(conn), request)
	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestShutdownHTTPHijacked(t *testing.T) {
	entryPoint, err := NewTCPEntryPoint(context.Background(), &static.EntryPoint{
		Address: ":0",
		Transport: &static.EntryPointsTransport{
			LifeCycle: &static.LifeCycle{
				RequestAcceptGraceTimeout: 0,
				GraceTimeOut:              5000000000,
			},
		},
		ForwardedHeaders: &static.ForwardedHeaders{},
	})
	require.NoError(t, err)

	go entryPoint.startTCP(context.Background())

	router := &tcp.Router{}
	router.HTTPHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, _, err := rw.(http.Hijacker).Hijack()
		require.NoError(t, err)
		time.Sleep(time.Second * 1)

		resp := http.Response{StatusCode: http.StatusOK}
		err = resp.Write(conn)
		require.NoError(t, err)

	}))
	entryPoint.switchRouter(router)

	conn, err := net.Dial("tcp", entryPoint.listener.Addr().String())
	require.NoError(t, err)

	go entryPoint.Shutdown(context.Background())

	request, err := http.NewRequest("GET", "http://127.0.0.1:8082", nil)
	require.NoError(t, err)

	err = request.Write(conn)
	require.NoError(t, err)

	resp, err := http.ReadResponse(bufio.NewReader(conn), request)
	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestShutdownTCPConn(t *testing.T) {
	entryPoint, err := NewTCPEntryPoint(context.Background(), &static.EntryPoint{
		Address: ":0",
		Transport: &static.EntryPointsTransport{
			LifeCycle: &static.LifeCycle{
				RequestAcceptGraceTimeout: 0,
				GraceTimeOut:              5000000000,
			},
		},
		ForwardedHeaders: &static.ForwardedHeaders{},
	})
	require.NoError(t, err)

	go entryPoint.startTCP(context.Background())

	router := &tcp.Router{}
	router.AddCatchAllNoTLS(tcp.HandlerFunc(func(conn net.Conn) {
		_, err := http.ReadRequest(bufio.NewReader(conn))
		require.NoError(t, err)
		time.Sleep(time.Second * 1)

		resp := http.Response{StatusCode: http.StatusOK}
		err = resp.Write(conn)
		require.NoError(t, err)
	}))

	entryPoint.switchRouter(router)

	conn, err := net.Dial("tcp", entryPoint.listener.Addr().String())
	require.NoError(t, err)

	go entryPoint.Shutdown(context.Background())

	request, err := http.NewRequest("GET", "http://127.0.0.1:8082", nil)
	require.NoError(t, err)

	err = request.Write(conn)
	require.NoError(t, err)

	resp, err := http.ReadResponse(bufio.NewReader(conn), request)
	require.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
