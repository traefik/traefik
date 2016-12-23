package acme

import (
	"encoding/base64"
	"github.com/xenolf/lego/acme"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

func TestDomainsSet(t *testing.T) {
	checkMap := map[string]Domains{
		"":                                   {},
		"foo.com":                            {Domain{Main: "foo.com", SANs: []string{}}},
		"foo.com,bar.net":                    {Domain{Main: "foo.com", SANs: []string{"bar.net"}}},
		"foo.com,bar1.net,bar2.net,bar3.net": {Domain{Main: "foo.com", SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	for in, check := range checkMap {
		ds := Domains{}
		ds.Set(in)
		if !reflect.DeepEqual(check, ds) {
			t.Errorf("Expected %+v\nGot %+v", check, ds)
		}
	}
}

func TestDomainsSetAppend(t *testing.T) {
	inSlice := []string{
		"",
		"foo1.com",
		"foo2.com,bar.net",
		"foo3.com,bar1.net,bar2.net,bar3.net",
	}
	checkSlice := []Domains{
		{},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}},
			Domain{Main: "foo3.com",
				SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	ds := Domains{}
	for i, in := range inSlice {
		ds.Set(in)
		if !reflect.DeepEqual(checkSlice[i], ds) {
			t.Errorf("Expected  %s %+v\nGot %+v", in, checkSlice[i], ds)
		}
	}
}

func TestCertificatesRenew(t *testing.T) {
	domainsCertificates := DomainsCertificates{
		lock: sync.RWMutex{},
		Certs: []*DomainsCertificate{
			{
				Domains: Domain{
					Main: "foo1.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo1.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA6OqHGdwGy20+3Jcz9IgfN4IR322X2Hhwk6n8Hss/Ws7FeTZo
PvXW8uHeI1bmQJsy9C6xo3odzO64o7prgMZl5eDw5fk1mmUij3J3nM3gwtc/Cc+8
ADXGldauASdHBFTRvWQge0Pv/Q5U0fyL2VCHoR9mGv4CQ7nRNKPus0vYJMbXoTbO
8z4sIbNz3Ov9o/HGMRb8D0rNPTMdC62tHSbiO1UoxLXr9dcBOGt786AsiRTJ8bq9
GCVQgzd0Wftb8z6ddW2YuWrmExlkHdfC4oG0D5SU1QB4ldPyl7fhVWlfHwC1NX+c
RnDSEeYkAcdvvIekdM/yH+z62XhwToM0E9TCzwIDAQABAoIBACq3EC3S50AZeeTU
qgeXizoP1Z1HKQjfFa5PB1jSZ30M3LRdIQMi7NfASo/qmPGSROb5RUS42YxC34PP
ZXXJbNiaxzM13/m/wHXURVFxhF3XQc1X1p+nPRMvutulS2Xk9E4qdbaFgBbFsRKN
oUwqc6U97+jVWq72/gIManNhXnNn1n1SRLBEkn+WStMPn6ZvWRlpRMjhy0c1mpwg
u6em92HvMvfKPQ60naUhdKp+q0rsLp2YKWjiytos9ENSYI5gAGLIDhKeqiD8f92E
4FGPmNRipwxCE2SSvZFlM26tRloWVcBPktRN79hUejE8iopiqVS0+4h/phZ2wG0D
18cqVpECgYEA+qmagnhm0LLvwVkUN0B2nRARQEFinZDM4Hgiv823bQvc9I8dVTqJ
aIQm5y4Y5UA3xmyDsRoO7GUdd0oVeh9GwTONzMRCOny/mOuOC51wXPhKHhI0O22u
sfbOHszl+bxl6ZQMUJa2/I8YIWBLU5P+fTgrfNwBEgZ3YPwUV5tyHNcCgYEA7eAv
pjQkbJNRq/fv/67sojN7N9QoH84egN5cZFh5d8PJomnsvy5JDV4WaG1G6mJpqjdD
YRVdFw5oZ4L8yCVdCeK9op896Uy51jqvfSe3+uKmNqE0qDHgaLubQNI8yYc5sacW
fYJBmDR6rNIeE7Q2240w3CdKfREuXdDnhyTTEskCgYBFeAnFTP8Zqe2+hSSQJ4J4
BwLw7u4Yww+0yja/N5E1XItRD/TOMRnx6GYrvd/ScVjD2kEpLRKju2ZOMC8BmHdw
hgwvitjcAsTK6cWFPI3uhjVsXhkxuzUmR0Naz+iQrQEFmi1LjGmMV1AVt+1IbYSj
SZTr1sFJMJeXPmWY3hDjIwKBgQC4H9fCJoorIL0PB5NVreishHzT8fw84ibqSTPq
2DDtazcf6C3AresN1c4ydqN1uUdg4fXdp9OujRBzTwirQ4CIrmFrBye89g7CrBo6
Hgxivh06G/3OUw0JBG5f9lvnAiy+Pj9CVxi+36A1NU7ioZP0zY0MW71koW/qXlFY
YkCfQQKBgBqwND/c3mPg7iY4RMQ9XjrKfV9o6FMzA51lAinjujHlNgsBmqiR951P
NA3kWZQ73D3IxeLEMaGHpvS7andPN3Z2qPhe+FbJKcF6ZZNTrFQkh/Fpz3wmYPo1
GIL4+09kNgMRWapaROqI+/3+qJQ+GVJZIPfYC0poJOO6vYqifWe8
-----END RSA PRIVATE KEY-----
`),
					Certificate: []byte(`
-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIJAK78ukR/Qu4rMA0GCSqGSIb3DQEBBQUAMBMxETAPBgNV
BAMMCGZvbzEuY29tMB4XDTE2MDYxOTIyMDMyM1oXDTI2MDYxNzIyMDMyM1owEzER
MA8GA1UEAwwIZm9vMS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDo6ocZ3AbLbT7clzP0iB83ghHfbZfYeHCTqfweyz9azsV5Nmg+9dby4d4jVuZA
mzL0LrGjeh3M7rijumuAxmXl4PDl+TWaZSKPcneczeDC1z8Jz7wANcaV1q4BJ0cE
VNG9ZCB7Q+/9DlTR/IvZUIehH2Ya/gJDudE0o+6zS9gkxtehNs7zPiwhs3Pc6/2j
8cYxFvwPSs09Mx0Lra0dJuI7VSjEtev11wE4a3vzoCyJFMnxur0YJVCDN3RZ+1vz
Pp11bZi5auYTGWQd18LigbQPlJTVAHiV0/KXt+FVaV8fALU1f5xGcNIR5iQBx2+8
h6R0z/If7PrZeHBOgzQT1MLPAgMBAAGjUDBOMB0GA1UdDgQWBBRFLH1wF6BT51uq
yWNqBnCrPFIglzAfBgNVHSMEGDAWgBRFLH1wF6BT51uqyWNqBnCrPFIglzAMBgNV
HRMEBTADAQH/MA0GCSqGSIb3DQEBBQUAA4IBAQAr7aH3Db6TeAZkg4Zd7SoF2q11
erzv552PgQUyezMZcRBo2q1ekmUYyy2600CBiYg51G+8oUqjJKiKnBuaqbMX7pFa
FsL7uToZCGA57cBaVejeB+p24P5bxoJGKCMeZcEBe5N93Tqu5WBxNEX7lQUo6TSs
gSN2Olf3/grNKt5V4BduSIQZ+YHlPUWLTaz5B1MXKSUqjmabARP9lhjO14u9USvi
dMBDFskJySQ6SUfz3fyoXELoDOVbRZETuSodpw+aFCbEtbcQCLT3A0FG+BEPayZH
tt19zKUlr6e+YFpyjQPGZ7ZkY7iMgHEkhKrXx2DiZ1+cif3X1xfXWQr0S5+E
-----END CERTIFICATE-----
`),
				},
			},
			{
				Domains: Domain{
					Main: "foo2.com",
					SANs: []string{}},
				Certificate: &Certificate{
					Domain:        "foo2.com",
					CertURL:       "url",
					CertStableURL: "url",
					PrivateKey: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA7rIVuSrZ3FfYXhR3qaWwfVcgiqKS//yXFzNqkJS6mz9nRCNT
lPawvrCFIRKdR7UO7xD7A5VTcbrGOAaTvrEaH7mB/4FGL+gN4AiTbVFpKXngAYEW
A3//zeBZ7XUSWaQ+CNC+l796JeoDvQD++KwCke4rVD1pGN1hpVEeGhwzyKOYPKLo
4+AGVe1LFWw4U/v8Iil1/gBBehZBILuhASpXy4W132LJPl76/EbGqh0nVz2UlFqU
HRxO+2U2ba4YIpI+0/VOQ9Cq/TzHSUdTTLfBHE/Qb+aDBfptMWTRvAngLqUglOcZ
Fi6SAljxEkJO6z6btmoVUWsoKBpbIHDC5++dZwIDAQABAoIBAAD8rYhRfAskNdnV
vdTuwXcTOCg6md8DHWDULpmgc9EWhwfKGZthFcQEGNjVKd9VCVXFvTP7lxe+TPmI
VW4Rb2k4LChxUWf7TqthfbKTBptMTLfU39Ft4xHn3pdTx5qlSjhhHJimCwxDFnbe
nS9MDsqpsHYtttSKfc/gMP6spS4sNPZ/r9zseT3eWkBEhn+FQABxJiuPcQ7q7S+Q
uOghmr7f3FeYvizQOhBtULsLrK/hsmQIIB4amS1QlpNWKbIoiUPNPjCA5PVQyAER
waYjuc7imBbeD98L/z8bRTlEskSKjtPSEXGVHa9OYdBU+02Ci6TjKztUp6Ho7JE9
tcHj+eECgYEA+9Ntv6RqIdpT/4/52JYiR+pOem3U8tweCOmUqm/p/AWyfAJTykqt
cJ8RcK1MfM+uoa5Sjm8hIcA2XPVEqH2J50PC4w04Q3xtfsz3xs7KJWXQCoha8D0D
ZIFNroEPnld0qOuJzpIIteXTrCLhSu17ZhN+Wk+5gJ7Ewu/QMM5OPjECgYEA8qbw
zfwSjE6jkrqO70jzqSxgi2yjo0vMqv+BNBuhxhDTBXnKQI1KsHoiS0FkSLSJ9+DS
CT3WEescD2Lumdm2s9HXvaMmnDSKBY58NqCGsNzZifSgmj1H/yS9FX8RXfSjXcxq
RDvTbD52/HeaCiOxHZx8JjmJEb+ZKJC4MDvjtxcCgYBM516GvgEjYXdxfliAiijh
6W4Z+Vyk5g/ODPc3rYG5U0wUjuljx7Z7xDghPusy2oGsIn5XvRxTIE35yXU0N1Jb
69eiWzEpeuA9bv7kGdal4RfNf6K15wwYL1y3w/YvFuorg/LLwNEkK5Ge6e//X9Ll
c2KM1fgCjXntRitAHGDMoQKBgDnkgodioLpA+N3FDN0iNqAiKlaZcOFA8G/LzfO0
tAAhe3dO+2YzT6KTQSNbUqXWDSTKytHRowVbZrJ1FCA4xVJZunNQPaH/Fv8EY7ZU
zk3cIzq61qZ2AHtrNIGwc2BLQb7bSm9FJsgojxLlJidNJLC/6Q7lo0JMyCnZfVhk
sYu5AoGAZt/MfyFTKm674UddSNgGEt86PyVYbLMnRoAXOaNB38AE12kaYHPil1tL
FnL8OQLpbX5Qo2JGgeZRlpMJ4Jxw2zzvUKr/n+6khaLxHmtX48hMu2QM7ZvnkZCs
Kkgz6v+Wcqm94ugtl3HSm+u9xZzVQxN6gu/jZQv3VpQiAZHjPYc=
-----END RSA PRIVATE KEY-----
`),
					Certificate: []byte(`
-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIJAK25/Z9Jz6IBMA0GCSqGSIb3DQEBBQUAMBMxETAPBgNV
BAMMCGZvbzIuY29tMB4XDTE2MDYyMDA5MzUyNloXDTI2MDYxODA5MzUyNlowEzER
MA8GA1UEAwwIZm9vMi5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDushW5KtncV9heFHeppbB9VyCKopL//JcXM2qQlLqbP2dEI1OU9rC+sIUhEp1H
tQ7vEPsDlVNxusY4BpO+sRofuYH/gUYv6A3gCJNtUWkpeeABgRYDf//N4FntdRJZ
pD4I0L6Xv3ol6gO9AP74rAKR7itUPWkY3WGlUR4aHDPIo5g8oujj4AZV7UsVbDhT
+/wiKXX+AEF6FkEgu6EBKlfLhbXfYsk+Xvr8RsaqHSdXPZSUWpQdHE77ZTZtrhgi
kj7T9U5D0Kr9PMdJR1NMt8EcT9Bv5oMF+m0xZNG8CeAupSCU5xkWLpICWPESQk7r
Ppu2ahVRaygoGlsgcMLn751nAgMBAAGjUDBOMB0GA1UdDgQWBBQ6FZWqB9qI4NN+
2jFY6xH8uoUTnTAfBgNVHSMEGDAWgBQ6FZWqB9qI4NN+2jFY6xH8uoUTnTAMBgNV
HRMEBTADAQH/MA0GCSqGSIb3DQEBBQUAA4IBAQCRhuf2dQhIEOmSOGgtRELF2wB6
NWXt0lCty9x4u+zCvITXV8Z0C34VQGencO3H2bgyC3ZxNpPuwZfEc2Pxe8W6bDc/
OyLckk9WLo00Tnr2t7rDOeTjEGuhXFZkhIbJbKdAH8cEXrxKR8UXWtZgTv/b8Hv/
g6tbeH6TzBsdMoFtUCsyWxygYwnLU+quuYvE2s9FiCegf2mdYTCh/R5J5n/51gfB
uC+NakKMfaCvNg3mOAFSYC/0r0YcKM/5ldKGTKTCVJAMhnmBnyRc/70rKkVRFy2g
iIjUFs+9aAgfCiL0WlyyXYAtIev2gw4FHUVlcT/xKks+x8Kgj6e5LTIrRRwW
-----END CERTIFICATE-----
`),
				},
			},
		},
	}

	newCertificate := &Certificate{
		Domain:        "foo1.com",
		CertURL:       "url",
		CertStableURL: "url",
		PrivateKey: []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA1OdSuXK2zeSLf0UqgrI4pjkpaqhra++pnda4Li4jXo151svi
Sn7DSynJOoq1jbfRJAoyDhxsBC4S4RuD54U5elJ4wLPZXmHRsvb+NwiHs9VmDqwu
It21btuqeNMebkab5cnDnC6KKufMhXRcRAlluYXyCkQe/+N+LlUQd6Js34TixMpk
eQOX4/OVrokSyVRnIq4u+o0Ufe7z5+41WVH63tcy7Hwi7244aLUzZCs+QQa2Dw6f
qEwjbonr974fM68UxDjTZEQy9u24yDzajhDBp1OTAAklh7U+li3g9dSyNVBFXqEu
nW2fyBvLqeJOSTihqfcrACB/YYhYOX94vMXELQIDAQABAoIBAFYK3t3fxI1VTiMz
WsjTKh3TgC+AvVkz1ILbojfXoae22YS7hUrCDD82NgMYx+LsZPOBw1T8m5Lc4/hh
3F8W8nHDHtYSWUjRk6QWOgsXwXAmUEahw0uH+qlA0ZZfDC9ZDexCLHHURTat03Qj
4J4GhjwCLB2GBlk4IWisLCmNVR7HokrpfIw4oM1aB5E21Tl7zh/x7ikRijEkUsKw
7YhaMeLJqBnMnAdV63hhF7FaDRjl8P2s/3octz/6pqDIABrDrUW3KAkNYCZIWdhF
Kk0wRMbZ/WrYT9GIGoJe7coQC7ezTrlrEkAFEIPGHCLkgXB/0TyuSy0yY59e4zmi
VvHoWUECgYEA/rOL2KJ/p+TZW7+YbsUzs0+F+M+G6UCr0nWfYN9MKmNtmns3eLDG
+pIpBMc5mjqeJR/sCCdkD8OqHC202Y8e4sr0pKSBeBofh2BmXtpyu3QQ50Pa63RS
SK6mYUrFqPmFFDbNGpFI4sIeI+Vf6hm96FQPnyPtUTGqk39m0RbWM/UCgYEA1f04
Nf3wbqwqIHZjYpPmymfjleyMn3hGUjpi7pmI6inXGMk3nkeG1cbOhnfPxL5BWD12
3RqHI2B4Z4r0BMyjctDNb1TxhMIpm5+PKm5KeeKfoYA85IS0mEeq6VdMm3mL1x/O
3LYvcUvAEVf6pWX/+ZFLMudqhF3jbTrdNOC6ZFkCgYBKpEeJdyW+CD0CvEVpwPUD
yXxTjE3XMZKpHLtWYlop2fWW3iFFh1jouci3k8L3xdHuw0oioZibXhYOJ/7l+yFs
CVpknakrj0xKGiAmEBKriLojbClN80rh7fzoakc+29D6OY0mCgm4GndGwcO4EU8s
NOZXFupHbyy0CRQSloSzuQKBgQC1Z/MtIlefGuijmHlsakGuuR+gS2ZzEj1bHBAe
gZ4mFM46PuqdjblqpR0TtaI3AarXqVOI4SJLBU9NR+jR4MF3Zjeh9/q/NvKa8Usn
B1Svu0TkXphAiZenuKnVIqLY8tNvzZFKXlAd1b+/dDwR10SHR3rebnxINmfEg7Bf
UVvyEQKBgAEjI5O6LSkLNpbVn1l2IO8u8D2RkFqs/Sbx78uFta3f9Gddzb4wMnt3
jVzymghCLp4Qf1ump/zC5bcQ8L97qmnjJ+H8X9HwmkqetuI362JNnz+12YKVDIWi
wI7SJ8BwDqYMrLw6/nE+degn39KedGDH8gz5cZcdlKTZLjbuBOfU
-----END RSA PRIVATE KEY-----
`),
		Certificate: []byte(`
-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIJAPQiOiQcwYaRMA0GCSqGSIb3DQEBBQUAMBMxETAPBgNV
BAMMCGZvbzEuY29tMB4XDTE2MDYxOTIyMTE1NFoXDTI2MDYxNzIyMTE1NFowEzER
MA8GA1UEAwwIZm9vMS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIB
AQDU51K5crbN5It/RSqCsjimOSlqqGtr76md1rguLiNejXnWy+JKfsNLKck6irWN
t9EkCjIOHGwELhLhG4PnhTl6UnjAs9leYdGy9v43CIez1WYOrC4i3bVu26p40x5u
RpvlycOcLooq58yFdFxECWW5hfIKRB7/434uVRB3omzfhOLEymR5A5fj85WuiRLJ
VGciri76jRR97vPn7jVZUfre1zLsfCLvbjhotTNkKz5BBrYPDp+oTCNuiev3vh8z
rxTEONNkRDL27bjIPNqOEMGnU5MACSWHtT6WLeD11LI1UEVeoS6dbZ/IG8up4k5J
OKGp9ysAIH9hiFg5f3i8xcQtAgMBAAGjUDBOMB0GA1UdDgQWBBQPfkS5ehpstmSb
8CGJE7GxSCxl2DAfBgNVHSMEGDAWgBQPfkS5ehpstmSb8CGJE7GxSCxl2DAMBgNV
HRMEBTADAQH/MA0GCSqGSIb3DQEBBQUAA4IBAQA99A+itS9ImdGRGgHZ5fSusiEq
wkK5XxGyagL1S0f3VM8e78VabSvC0o/xdD7DHVg6Az8FWxkkksH6Yd7IKfZZUzvs
kXQhlOwWpxgmguSmAs4uZTymIoMFRVj3nG664BcXkKu4Yd9UXKNOWP59zgvrCJMM
oIsmYiq5u0MFpM31BwfmmW3erqIcfBI9OJrmr1XDzlykPZNWtUSSfVuNQ8d4bim9
XH8RfVLeFbqDydSTCHIFvYthH/ESbpRCiGJHoJ8QLfOkhD1k2fI0oJZn5RVtG2W8
bZME3gHPYCk1QFZUptriMCJ5fMjCgxeOTR+FAkstb/lTRuCc4UyILJguIMar
-----END CERTIFICATE-----
`),
	}

	err := domainsCertificates.renewCertificates(
		newCertificate,
		Domain{
			Main: "foo1.com",
			SANs: []string{}})
	if err != nil {
		t.Errorf("Error in renewCertificates :%v", err)
	}
	if len(domainsCertificates.Certs) != 2 {
		t.Errorf("Expected domainsCertificates length %d %+v\nGot %+v", 2, domainsCertificates.Certs, len(domainsCertificates.Certs))
	}
	if !reflect.DeepEqual(domainsCertificates.Certs[0].Certificate, newCertificate) {
		t.Errorf("Expected new certificate %+v \nGot %+v", newCertificate, domainsCertificates.Certs[0].Certificate)
	}
}

func TestNoPreCheckOverride(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
	err := dnsOverrideDelay(0)
	if err != nil {
		t.Errorf("Error in dnsOverrideDelay :%v", err)
	}
	if acme.PreCheckDNS != nil {
		t.Errorf("Unexpected change to acme.PreCheckDNS when leaving DNS verification as is.")
	}
}

func TestSillyPreCheckOverride(t *testing.T) {
	err := dnsOverrideDelay(-5)
	if err == nil {
		t.Errorf("Missing expected error in dnsOverrideDelay!")
	}
}

func TestPreCheckOverride(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
	err := dnsOverrideDelay(5)
	if err != nil {
		t.Errorf("Error in dnsOverrideDelay :%v", err)
	}
	if acme.PreCheckDNS == nil {
		t.Errorf("No change to acme.PreCheckDNS when meant to be adding enforcing override function.")
	}
}

func TestAcmeClientCreation(t *testing.T) {
	acme.PreCheckDNS = nil // Irreversable - but not expecting real calls into this during testing process
	// Lengthy setup to avoid external web requests - oh for easier golang testing!
	account := &Account{Email: "f@f"}
	account.PrivateKey, _ = base64.StdEncoding.DecodeString(`
MIIBPAIBAAJBAMp2Ni92FfEur+CAvFkgC12LT4l9D53ApbBpDaXaJkzzks+KsLw9zyAxvlrfAyTCQ
7tDnEnIltAXyQ0uOFUUdcMCAwEAAQJAK1FbipATZcT9cGVa5x7KD7usytftLW14heQUPXYNV80r/3
lmnpvjL06dffRpwkYeN8DATQF/QOcy3NNNGDw/4QIhAPAKmiZFxA/qmRXsuU8Zhlzf16WrNZ68K64
asn/h3qZrAiEA1+wFR3WXCPIolOvd7AHjfgcTKQNkoMPywU4FYUNQ1AkCIQDv8yk0qPjckD6HVCPJ
llJh9MC0svjevGtNlxJoE3lmEQIhAKXy1wfZ32/XtcrnENPvi6lzxI0T94X7s5pP3aCoPPoJAiEAl
cijFkALeQp/qyeXdFld2v9gUN3eCgljgcl0QweRoIc=---`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
"new-authz": "https://foo/acme/new-authz",
"new-cert": "https://foo/acme/new-cert",
"new-reg": "https://foo/acme/new-reg",
"revoke-cert": "https://foo/acme/revoke-cert"
}`))
	}))
	defer ts.Close()
	a := ACME{DNSProvider: "manual", DelayDontCheckDNS: 10, CAServer: ts.URL}

	client, err := a.buildACMEClient(account)
	if err != nil {
		t.Errorf("Error in buildACMEClient: %v", err)
	}
	if client == nil {
		t.Errorf("No client from buildACMEClient!")
	}
	if acme.PreCheckDNS == nil {
		t.Errorf("No change to acme.PreCheckDNS when meant to be adding enforcing override function.")
	}
}
