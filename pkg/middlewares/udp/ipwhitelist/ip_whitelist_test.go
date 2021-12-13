package udpipwhitelist

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/udp"
)

func TestNewIPWhiteLister(t *testing.T) {
	testCases := []struct {
		desc          string
		whiteList     dynamic.UDPIPWhiteList
		expectedError bool
	}{
		{
			desc:          "Empty config",
			whiteList:     dynamic.UDPIPWhiteList{},
			expectedError: true,
		},
		{
			desc: "invalid IP",
			whiteList: dynamic.UDPIPWhiteList{
				SourceRange: []string{"foo"},
			},
			expectedError: true,
		},
		{
			desc: "valid IP",
			whiteList: dynamic.UDPIPWhiteList{
				SourceRange: []string{"10.10.10.10"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := udp.HandlerFunc(func(conn *udp.Conn) {})
			whiteLister, err := New(context.Background(), next, test.whiteList, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, whiteLister)
			}
		})
	}
}

func TestIPWhiteLister_ServeUDP(t *testing.T) {
	testCases := []struct {
		desc       string
		whiteList  dynamic.UDPIPWhiteList
		listenAddr string
		expected   string
	}{
		{
			desc: "authorized with remote address",
			whiteList: dynamic.UDPIPWhiteList{
				SourceRange: []string{"127.0.0.1"},
			},
			listenAddr: "127.0.0.1:20200",
			expected:   "OK",
		},
		{
			desc: "non authorized with remote address",
			whiteList: dynamic.UDPIPWhiteList{
				SourceRange: []string{"10.10.10.10"},
			},
			listenAddr: "127.0.0.1:30300",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := udp.HandlerFunc(func(conn *udp.Conn) {
				n, err := conn.Write([]byte("OK"))
				require.NoError(t, err)
				assert.Equal(t, 2, n)

				err = conn.Close()
				require.NoError(t, err)
			})

			whiteLister, err := New(context.Background(), next, test.whiteList, "traefikTest")
			require.NoError(t, err)

			listenAddr, err := net.ResolveUDPAddr("udp", test.listenAddr)
			require.NoError(t, err)

			ln, err := udp.Listen("udp", listenAddr, time.Second)
			require.NoError(t, err)

			go func() {
				rConn, err := ln.Accept()
				require.NoError(t, err)
				whiteLister.ServeUDP(rConn)
			}()

			lConn, err := net.DialUDP("udp", nil, listenAddr)
			require.NoError(t, err)

			write := []byte("connect")
			n, err := lConn.Write(write)
			require.NoError(t, err)
			assert.Equal(t, len(write), n)

			readCh := make(chan []byte)
			go func() {
				read := make([]byte, 10)
				n, _, err = lConn.ReadFromUDP(read)
				require.NoError(t, err)
				readCh <- read[:n]
			}()

			select {
			case read := <-readCh:
				assert.Equal(t, test.expected, string(read))
			case <-time.After(5 * time.Second):
				assert.Empty(t, test.expected)
			}
		})
	}
}
