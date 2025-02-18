package tcp

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service/tcp"
	tcp2 "github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host host1.localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var cert_host1 = `-----BEGIN CERTIFICATE-----
MIIB/TCCAWagAwIBAgIRAOhIR/mwfLj9ddC2KoBFaOEwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEA7erdwU+ofPewrmC7OYdDQhjkHVRiVtc+4kds4TGQ2CmAVyQrdc7nIQpg
MbZNsmJdYic+FuVREl737h/pp7iXWvlQgSyvgJAQmEK8qeDoweHFrDocNEmgM+oJ
O+Ca5tGjxGVEZxevT7QwEStuuQdEnj7/nMvU7ZDBk4Z/LpOaan0CAwEAAaNRME8w
DgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQC
MAAwGgYDVR0RBBMwEYIPaG9zdDEubG9jYWxob3N0MA0GCSqGSIb3DQEBCwUAA4GB
AGBlOAqoIdA0G80YXuIawkCTk1W+nB0sT7HYKp2v1xxMBsmYnHbYjNlL9hpXEoLw
rMwASUy4Db0Xt20jd6ewQNR+VENQqo6wiqKlwWt6kunuPNfWseuKo2rcDTLwyb5R
jSvmExRD+lBfoZgKa7PUvhUDEvf3jSycHpPqf7MeNq8e
-----END CERTIFICATE-----`

var key_host1 = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAO3q3cFPqHz3sK5g
uzmHQ0IY5B1UYlbXPuJHbOExkNgpgFckK3XO5yEKYDG2TbJiXWInPhblURJe9+4f
6ae4l1r5UIEsr4CQEJhCvKng6MHhxaw6HDRJoDPqCTvgmubRo8RlRGcXr0+0MBEr
brkHRJ4+/5zL1O2QwZOGfy6Tmmp9AgMBAAECgYEAgVX7hTozovPXlYQ6Y3S3yHfV
kmgsKX9LzSD8/JLAZfJxtW2RPrLijOCiGIQ9SqsUjuY8Z5/z6aO87jNlButfQ2X/
apMydU/hkJJp1/My+Qls/gaU7x46cSE/J28juMrHdTZiQFDMrrlnmwdmKDTcfbOJ
S1hXeasQMJkTN6IpXlUCQQDxdtiDyUggglQR5QnYFLa88fd2lEkNiI/eVZ3WHZMa
E2GFcXYZTNztf++dSdMqbedIpFxK0rLfj8UcmV1fLIz/AkEA/D1cj20eV7GASKb5
7PESThQ+WyHXy++i3piHsu6v2plYwEEmA/1DwqeQEbQHqdSiT5KJ3SwI6B8763KR
8L28gwJBAJ+KrQh2aAfC1RV1xglVtmAlaCKbW6Frh9OZsk4VAGsMPzVSgHu7A4aR
L5s3eiTgtR6UKr7tdG6uqch5tO37m7UCQBTeRsAe+PmsV76rAdZWg3suNZJ4lE/s
/X6JBAELukTNlwgg27JMy8RY9JRiXpfwXZVTvFAuCnaZzu1Fx0kxiV0CQDyixsfy
jNmT8Q7KRHpXoB6sY9wZ0AarhK7IosYLIUQrhJqSXJghKicbwLh9OUVHTeKxqZfy
hM9atiJV5XjK1pk=
-----END PRIVATE KEY-----`

// go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host host2.localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var cert_host2 = `-----BEGIN CERTIFICATE-----
MIIB/DCCAWWgAwIBAgIQRVKz4/fbzTfCSgVlA3lCTzANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQCjUNtZ2XWn0hx5E/MKeeG8a1pflbk2Ht323PFRTUd51UqyIEqX5hwOTZVO
o2a+ngldDb/4JpYIICwZPnqzxy4OslhRJ9rxQsD0RbHhZKll+xtcjk+R4x1qFQ0b
vFORIyMIn0NJcmQcurjaogf2pDxktEtAt6x1JVs5umz8BGJyCwIDAQABo1EwTzAO
BgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIw
ADAaBgNVHREEEzARgg9ob3N0Mi5sb2NhbGhvc3QwDQYJKoZIhvcNAQELBQADgYEA
ZXUUsRK3TlbqR0fgSr21PE0haDQ2LFIluoYDHfrVpa5mpZ4wQOGbBZ43dW1PCJiA
lK12aFpXfGR4Rgq0Sbt5sInic8wCxmDetty5/V2nOoZszPMPjAcZy2fPm1pBtMzG
suN1vhy3dlbA8erx5KooqRWhuyKEYm/TEYotf3TG9ig=
-----END CERTIFICATE-----`

var key_host2 = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAKNQ21nZdafSHHkT
8wp54bxrWl+VuTYe3fbc8VFNR3nVSrIgSpfmHA5NlU6jZr6eCV0Nv/gmlgggLBk+
erPHLg6yWFEn2vFCwPRFseFkqWX7G1yOT5HjHWoVDRu8U5EjIwifQ0lyZBy6uNqi
B/akPGS0S0C3rHUlWzm6bPwEYnILAgMBAAECgYBkD/R1lpFJ46hiXuC4eHjgkv3q
Nrgl+r+Qs0p/v9OdSBveC37olqp18P8cEW2wOPAPvY7zIeEm1V9vkCJp6A3FJGbJ
49G4GDuQAdsD/HQvGuf7OdUnnF6/eybePMFmJP626QPyJBAM7T0ZZz8i0MQZf7qH
zat6tXSwTN2d9s5o8QJBAMBt+5KU0STodm9oF/NfEKRSKZa2f4ARFhI+NAvFP943
XI8QdfVYcwjrSXj3r2OWsY2pDz6awBJtXZICV44SYR8CQQDZRK9127G07vHCYKZ4
XbRtp3WXhHUH7Ykhw+Rju+iyTh9b1Nt4cPZolkztlwRMUj5Qv+fmHpiYXwnirrBR
W7WVAkEAnxJMBMBAo+IHBdFm+yh6+VtyRcRXYea9+Bazr4c/ZNMfEKTq3gZgEd9u
vTEDK7BG1nQKxhXm8VS3JRwKhMdswQJBAI8xhpadwcRmyu15951S3Mx8VrMSuHMO
KZgYXFkjCl0hwecrJa5+fNg3XuIj6tBGUA22PSdcOOQLlx9QVKJ6V/UCQFlzbeqy
uQIoOEjHGc2A0hMdC3/sNbFjMYdLqQq6qVDg8qlUMCG6u/SYPhHKk5SaXSKYwyd3
X6QbEQfMGrSqyQE=
-----END PRIVATE KEY-----
`

// go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host default.localhost --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var cert_default = `-----BEGIN CERTIFICATE-----
MIIB/zCCAWigAwIBAgIRAKJPixo/MCa37PCn2PW3bP8wDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEApWTEUs2fepJFD87OzdiBRg77S0vkWYs5xIOjfmFVfxd3td5e/qfJc30S
P9al4+dlXFZJtTWn2sZnlKegufjz1X/MXQaT2PkO/rWsqtqyLOT8wzSxkmrr+cU1
sM26wssYq3ppKBi5O7zSCDCfGlqh6eOq3FamPVa8jN+xypq8TgsCAwEAAaNTMFEw
DgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQC
MAAwHAYDVR0RBBUwE4IRZGVmYXVsdC5sb2NhbGhvc3QwDQYJKoZIhvcNAQELBQAD
gYEAeaOBAwzoxeG40zz1dYS/pUZNkMXY5gumfU6PqfDZXghptNxHCAM+9cs3Gdlm
cQZt66pzcrzcMP4v2qWAFanm39FKi2p16cyWkL8/DZvD0bBhVdevTuv6KmwJ18e7
qtWRP0kBW4sPZNJInIpPIcg4cwSs2XYCkE8X4GbJCU+Nwqo=
-----END CERTIFICATE-----`

var key_default = `-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAKVkxFLNn3qSRQ/O
zs3YgUYO+0tL5FmLOcSDo35hVX8Xd7XeXv6nyXN9Ej/WpePnZVxWSbU1p9rGZ5Sn
oLn489V/zF0Gk9j5Dv61rKrasizk/MM0sZJq6/nFNbDNusLLGKt6aSgYuTu80ggw
nxpaoenjqtxWpj1WvIzfscqavE4LAgMBAAECgYBa1PRc5UBoeFwlSlaZBgY5C5FG
0O8fni6jlgf8KEhj++dqoi1ZfZxNKKsVFDUW7MXl6B2iv0zoAX5xTX4fpHGENR9K
xUvgcB+3VMm/1bavssAYpyWp6yJ/MOLu3GtCfYj9w/DWiQMf6DMZnNdWlUtqgUvs
GrwYDf1rz0DuIuwEwQJBAMzfu9d1GtYu28HhcaGfws/N023eCIfTwct44FI9ibMV
NChj2DGylbzBRjLOsaLRikHIcaCx3W6bRK+jad1qGLcCQQDOqtoqsP7rKvdG3MXx
ullmf0HkFe9kzHdW++uD5T8tNopKWVaL+MIHcPcjTjGpSjOACgnuSSNZRKl0wm7S
3xlNAkEAtk28a7vz1nU57as7nxN3mcxQgGpb8umWgAWeru+9cVLD59D41zhPj/f4
DEvqu7Rzr5e6rMC5Bqw5kYT7NiArvwJBAJWMjNLXwZ/rN4TPvW1uq8K/0655MQJ/
8tu+8G5BNbZCAVBL1ZT0LXO1CyFBNC6MwzekDAuiYTH3vagACrINPwECQEkqiEQz
HB04vhQ+KbcUCS7W5bS0Lod9Sdp9YkBDM+ckE5JHmlaWQNhZ1f+91biazFPhBew2
KaX80/KYL8o/4Jo=
-----END PRIVATE KEY-----`

func TestRuntimeConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		httpServiceConfig map[string]*runtime.ServiceInfo
		httpRouterConfig  map[string]*runtime.RouterInfo
		tcpServiceConfig  map[string]*runtime.TCPServiceInfo
		tcpRouterConfig   map[string]*runtime.TCPRouterInfo
		expectedError     int
	}{
		{
			desc: "No error",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Port:    "8085",
									Address: "127.0.0.1:8085",
								},
								{
									Address: "127.0.0.1:8086",
									Port:    "8086",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bar.foo`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "foo",
						},
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "bar",
						},
					},
				},
			},
			expectedError: 0,
		},
		{
			desc: "Non-ASCII domain error",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Port:    "8085",
									Address: "127.0.0.1:8085",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bàr.foo`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "foo",
						},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "HTTP routers with same domain but different TLS options",
			httpServiceConfig: map[string]*runtime.ServiceInfo{
				"foo-service": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{
								{
									Port: "8085",
									URL:  "127.0.0.1:8085",
								},
								{
									URL:  "127.0.0.1:8086",
									Port: "8086",
								},
							},
						},
					},
				},
			},
			httpRouterConfig: map[string]*runtime.RouterInfo{
				"foo": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "foo",
						},
					},
				},
				"bar": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`) && PathPrefix(`/path`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "bar",
						},
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "HTTP routers with same domain but different TLS store",
			httpServiceConfig: map[string]*runtime.ServiceInfo{
				"foo-service": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{
								{
									Port: "8085",
									URL:  "127.0.0.1:8085",
								},
								{
									URL:  "127.0.0.1:8086",
									Port: "8086",
								},
							},
						},
					},
				},
			},
			httpRouterConfig: map[string]*runtime.RouterInfo{
				"foo": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "foo",
							Store:   "foo",
						},
					},
				},
				"bar": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`) && PathPrefix(`/path`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "foo",
							Store:   "bar",
						},
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "One router with wrong rule",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},

				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "All router with wrong rule",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`foo.bar`)",
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with unknown service",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "wrong-service",
						Rule:        "HostSNI(`bar.foo`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken service",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: nil,
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with priority exceeding the max user-defined priority",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Port:    "8085",
									Address: "127.0.0.1:8085",
								},
								{
									Address: "127.0.0.1:8086",
									Port:    "8086",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
						Priority:    math.MaxInt,
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with HostSNI but no TLS",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bar.foo`)",
					},
				},
			},
			expectedError: 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPoints := []string{"web"}

			conf := &runtime.Configuration{
				Services:    test.httpServiceConfig,
				Routers:     test.httpRouterConfig,
				TCPServices: test.tcpServiceConfig,
				TCPRouters:  test.tcpRouterConfig,
			}
			dialerManager := tcp2.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
			serviceManager := tcp.NewManager(conf, dialerManager)
			tlsManager := traefiktls.NewManager()
			tlsManager.UpdateConfigs(
				context.Background(),
				map[string]traefiktls.Store{},
				map[string]traefiktls.Options{
					"default": {
						MinVersion: "VersionTLS10",
					},
					"foo": {
						MinVersion: "VersionTLS12",
					},
					"bar": {
						MinVersion: "VersionTLS11",
					},
				},
				[]*traefiktls.CertAndStores{})

			middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

			routerManager := NewManager(conf, serviceManager, middlewaresBuilder,
				nil, nil, tlsManager)

			_ = routerManager.BuildHandlers(context.Background(), entryPoints)

			// even though conf was passed by argument to the manager builders above,
			// it's ok to use it as the result we check, because everything worth checking
			// can be accessed by pointers in it.
			var allErrors int
			for _, v := range conf.TCPServices {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range conf.TCPRouters {
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			for _, v := range conf.Services {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range conf.Routers {
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			assert.Equal(t, test.expectedError, allErrors)
		})
	}
}

func TestDomainFronting(t *testing.T) {
	tlsOptionsBase := map[string]traefiktls.Options{
		"default": {
			MinVersion: "VersionTLS10",
		},
		"host1@file": {
			MinVersion: "VersionTLS12",
		},
		"host1@crd": {
			MinVersion: "VersionTLS12",
		},
	}

	entryPoints := []string{"web"}

	tests := []struct {
		desc           string
		routers        map[string]*runtime.RouterInfo
		tlsOptions     map[string]traefiktls.Options
		host           string
		ServerName     string
		expectedStatus int
	}{
		{
			desc: "Request is misdirected when TLS options are different",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS:         &dynamic.RouterTLSConfig{},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is OK when TLS options are the same",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Default TLS options is used when options are ambiguous for the same host",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`) && PathPrefix(`/foo`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "default",
						},
					},
				},
				"router-3@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Default TLS options should not be used when options are the same for the same host",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`) && PathPrefix(`/bar`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-3@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Request is misdirected when TLS options have the same name but from different providers",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@crd": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is OK when TLS options reference from a different provider is the same",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@crd",
						},
					},
				},
				"router-2@crd": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@crd",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Request is misdirected when server name is empty and the host name is an FQDN, but router's rule is not",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@file",
						},
					},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"default": {
					MinVersion: "VersionTLS13",
				},
				"host1@file": {
					MinVersion: "VersionTLS12",
				},
			},
			host:           "host1.local.",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is misdirected when server name is empty and the host name is not FQDN, but router's rule is",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local.`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@file",
						},
					},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"default": {
					MinVersion: "VersionTLS13",
				},
				"host1@file": {
					MinVersion: "VersionTLS12",
				},
			},
			host:           "host1.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conf := &runtime.Configuration{
				Routers: test.routers,
			}

			serviceManager := tcp.NewManager(conf, tcp2.NewDialerManager(nil))

			tlsManager := traefiktls.NewManager()
			tlsManager.UpdateConfigs(context.Background(), map[string]traefiktls.Store{}, test.tlsOptions, []*traefiktls.CertAndStores{})

			httpsHandler := map[string]http.Handler{
				"web": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}),
			}

			middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

			routerManager := NewManager(conf, serviceManager, middlewaresBuilder, nil, httpsHandler, tlsManager)

			routers := routerManager.BuildHandlers(context.Background(), entryPoints)

			router, ok := routers["web"]
			require.True(t, ok)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = test.host
			req.TLS = &tls.ConnectionState{
				ServerName: test.ServerName,
			}

			rw := httptest.NewRecorder()

			router.GetHTTPSHandler().ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
		})
	}
}

func TestStore(t *testing.T) {
	entryPoints := []string{"web"}

	mockBackend, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = mockBackend.Close()
	})

	go func() {
		for {
			conn, err := mockBackend.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	stores := map[string]traefiktls.Store{
		"store1": {
			DefaultCertificate: &traefiktls.Certificate{
				CertFile: types.FileOrContent(cert_default),
				KeyFile:  types.FileOrContent(key_default),
			},
		},
		"store2": {
			DefaultCertificate: &traefiktls.Certificate{
				CertFile: types.FileOrContent(cert_default),
				KeyFile:  types.FileOrContent(key_default),
			},
		},
	}
	certs := []*traefiktls.CertAndStores{
		{
			Stores: []string{"store1"},
			Certificate: traefiktls.Certificate{
				CertFile: types.FileOrContent(cert_host1),
				KeyFile:  types.FileOrContent(key_host1),
			},
		},
		{
			Stores: []string{"store2"},
			Certificate: traefiktls.Certificate{
				CertFile: types.FileOrContent(cert_host2),
				KeyFile:  types.FileOrContent(key_host2),
			},
		},
	}

	tests := []struct {
		desc                      string
		routers                   map[string]*runtime.RouterInfo
		conf                      *runtime.Configuration
		ServerName                string
		expectedCertificateDomain string
		expectedCommonName        string
	}{
		{
			desc: "[HTTP] Use specific certificate store store1 with right certificate in it",
			conf: &runtime.Configuration{
				Routers: map[string]*runtime.RouterInfo{
					"router-1@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host1.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store1",
							},
						},
					},
					"router-2@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host2.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store2",
							},
						},
					},
				},
			},
			ServerName:                "host1.localhost",
			expectedCertificateDomain: "host1.localhost",
		}, {
			desc: "[HTTP] Use specific certificate store store2 with right certificate in it",
			conf: &runtime.Configuration{
				Routers: map[string]*runtime.RouterInfo{
					"router-1@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host1.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store1",
							},
						},
					},
					"router-2@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host2.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store2",
							},
						},
					},
				},
			},
			ServerName:                "host2.localhost",
			expectedCertificateDomain: "host2.localhost",
		},
		{
			desc: "[HTTP] Use specific certificate store without right certificate in it",
			conf: &runtime.Configuration{
				Routers: map[string]*runtime.RouterInfo{
					"router-1@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host1.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store2",
							},
						},
					},
					"router-2@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host2.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store2",
							},
						},
					},
				},
			},
			ServerName:                "host1.localhost",
			expectedCertificateDomain: "default.localhost",
		},
		{
			desc: "[HTTP] Multi-store on same SNI use default store.",
			conf: &runtime.Configuration{
				Routers: map[string]*runtime.RouterInfo{
					"router-1@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host1.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store1",
							},
						},
					},
					"router-2@file": {
						Router: &dynamic.Router{
							EntryPoints: entryPoints,
							Rule:        "Host(`host1.localhost`)",
							TLS: &dynamic.RouterTLSConfig{
								Store: "store2",
							},
						},
					},
				},
			},
			ServerName:         "host1.localhost",
			expectedCommonName: "TRAEFIK DEFAULT CERT",
		},
		{
			desc: "[TCP] Use specific certificate store store1 with right certificate in it",
			conf: &runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"mock@file": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{{Address: mockBackend.Addr().String()}},
							},
						},
					},
				},
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"router-1@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host1.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store1",
							},
							Service: "mock",
						},
					},
					"router-2@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host2.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store2",
							},
							Service: "mock",
						},
					},
				},
			},
			ServerName:                "host1.localhost",
			expectedCertificateDomain: "host1.localhost",
		}, {
			desc: "[TCP] Use specific certificate store store2 with right certificate in it",
			conf: &runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"mock@file": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{{Address: mockBackend.Addr().String()}},
							},
						},
					},
				},
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"router-1@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host1.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store1",
							},
							Service: "mock",
						},
					},
					"router-2@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host2.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store2",
							},
							Service: "mock",
						},
					},
				},
			},
			ServerName:                "host2.localhost",
			expectedCertificateDomain: "host2.localhost",
		},
		{
			desc: "[TCP] Use specific certificate store without right certificate in it",
			conf: &runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"mock@file": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{{Address: mockBackend.Addr().String()}},
							},
						},
					},
				},
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"router-1@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host1.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store2",
							},
							Service: "mock",
						},
					},
					"router-2@file": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: entryPoints,
							Rule:        "HostSNI(`host2.localhost`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Store: "store2",
							},
							Service: "mock",
						},
					},
				},
			},
			ServerName:                "host1.localhost",
			expectedCertificateDomain: "default.localhost",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			dialerManager := tcp2.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{
				"default@internal": &dynamic.TCPServersTransport{},
			})
			serviceManager := tcp.NewManager(test.conf, dialerManager)
			tlsManager := traefiktls.NewManager()
			tlsManager.UpdateConfigs(context.Background(), stores, map[string]traefiktls.Options{
				"default": {
					MinVersion: "VersionTLS13",
				}}, certs)

			httpsHandler := map[string]http.Handler{
				"web": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}),
			}

			middlewaresBuilder := tcpmiddleware.NewBuilder(test.conf.TCPMiddlewares)

			routerManager := NewManager(test.conf, serviceManager, middlewaresBuilder, nil, httpsHandler, tlsManager)

			routers := routerManager.BuildHandlers(context.Background(), entryPoints)
			router, ok := routers["web"]
			require.True(t, ok)

			// serverHTTP handler returns only the "HTTP" value as body for further checks.
			serverHTTP := &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				}),
			}

			ln, err := net.Listen("tcp", ":0")
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = ln.Close()
			})

			forwarder := newHTTPForwarder(ln)
			go func() {
				// defer close(stoppedHTTP)
				_ = serverHTTP.Serve(forwarder)
			}()

			router.SetHTTPSForwarder(forwarder)

			go func() {
				for {
					conn, err := ln.Accept()
					if err != nil {
						return
					}

					tcpConn, ok := conn.(*net.TCPConn)
					if !ok {
						t.Error("not a write closer")
					}

					router.ServeTCP(tcpConn)
				}
			}()

			_, port, err := net.SplitHostPort(ln.Addr().String())
			require.NoError(t, err)

			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM([]byte(cert_host1))
			pool.AppendCertsFromPEM([]byte(cert_host2))
			pool.AppendCertsFromPEM([]byte(cert_default))

			conn, err := tls.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", port), &tls.Config{ServerName: test.ServerName, RootCAs: pool, InsecureSkipVerify: true})
			require.NoError(t, err)

			if len(test.expectedCertificateDomain) > 0 {
				require.Equal(t, test.expectedCertificateDomain, conn.ConnectionState().PeerCertificates[0].DNSNames[0])
			}
			if len(test.expectedCommonName) > 0 {
				require.Equal(t, test.expectedCommonName, conn.ConnectionState().PeerCertificates[0].Subject.CommonName)
			}
		})
	}
}
