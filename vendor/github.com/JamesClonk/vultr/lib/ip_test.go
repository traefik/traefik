package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IP_ListIPv4_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	list, err := client.ListIPv4("123456789")
	assert.Nil(t, list)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_ListIPv4_NoList(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[]}`)
	defer server.Close()

	list, err := client.ListIPv4("123456789")
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, list)
}

func Test_IP_ListIPv4_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[
{"ip":"123.123.123.124","netmask":"255.255.255.248","gateway":"123.123.123.1","type":"secondary_ip","reverse":"host2.example.com"},
{"ip":"10.99.0.10","netmask":"255.255.0.0","gateway":"","type":"private","reverse":""},
{"ip":"123.123.123.123","netmask":"255.255.255.248","gateway":"123.123.123.1","type":"main_ip","reverse":"host1.example.com"}]}`)
	defer server.Close()

	list, err := client.ListIPv4("123456789")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, list) {
		assert.Equal(t, 3, len(list))

		assert.Equal(t, "123.123.123.123", list[0].IP)
		assert.Equal(t, "255.255.255.248", list[0].Netmask)
		assert.Equal(t, "main_ip", list[0].Type)
		assert.Equal(t, "host1.example.com", list[0].ReverseDNS)

		assert.Equal(t, "10.99.0.10", list[1].IP)
		assert.Equal(t, "255.255.0.0", list[1].Netmask)
		assert.Equal(t, "", list[1].Gateway)
		assert.Equal(t, "private", list[1].Type)
		assert.Equal(t, "", list[1].ReverseDNS)

		assert.Equal(t, "123.123.123.124", list[2].IP)
		assert.Equal(t, "123.123.123.1", list[2].Gateway)
		assert.Equal(t, "secondary_ip", list[2].Type)
		assert.Equal(t, "host2.example.com", list[2].ReverseDNS)
	}
}

func Test_IP_ListIPv6_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	list, err := client.ListIPv6("123456789")
	assert.Nil(t, list)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_ListIPv6_NoList(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[]}`)
	defer server.Close()

	list, err := client.ListIPv6("123456789")
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, list)
}

func Test_IP_ListIPv6_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[
{"ip":"2002:DB9:9001::200","network":"2001:DB8:1000::","network_size":"64","type":"secondary_ip"},
{"ip":"2001:DB8:1000::100","network":"2001:DB8:1000::","network_size":"64","type":"main_ip"}]}`)
	defer server.Close()

	list, err := client.ListIPv6("123456789")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, list) {
		assert.Equal(t, 2, len(list))

		assert.Equal(t, "2001:DB8:1000::100", list[0].IP)
		assert.Equal(t, "2001:DB8:1000::", list[0].Network)
		assert.Equal(t, "main_ip", list[0].Type)

		assert.Equal(t, "2002:DB9:9001::200", list[1].IP)
		assert.Equal(t, "64", list[1].NetworkSize)
		assert.Equal(t, "secondary_ip", list[1].Type)
	}
}

func Test_IP_ListIPv6ReverseDNS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	list, err := client.ListIPv6ReverseDNS("123456789")
	assert.Nil(t, list)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_ListIPv6ReverseDNS_NoEntries(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[]}`)
	defer server.Close()

	list, err := client.ListIPv6ReverseDNS("123456789")
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, list)
}

func Test_IP_ListIPv6ReverseDNS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"576965":[
{"ip":"2002:DB9:9001::200","reverse":"host2.example.com"},
{"ip":"2001:DB8:1000::100","reverse":"host1.example.com"}]}`)
	defer server.Close()

	list, err := client.ListIPv6ReverseDNS("123456789")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, list) {
		assert.Equal(t, 2, len(list))

		assert.Equal(t, "2001:DB8:1000::100", list[0].IP)
		assert.Equal(t, "host1.example.com", list[0].ReverseDNS)
		assert.Equal(t, "2002:DB9:9001::200", list[1].IP)
		assert.Equal(t, "host2.example.com", list[1].ReverseDNS)
	}
}

func Test_IP_DeleteIPv6ReverseDNS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DeleteIPv6ReverseDNS("123456789", "AAAA:BBBB:CCCC")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_DeleteIPv6ReverseDNS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DeleteIPv6ReverseDNS("123456789", "AAAA:BBBB:CCCC"))
}

func Test_IP_SetIPv6ReverseDNS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.SetIPv6ReverseDNS("123456789", "AAAA:BBBB:CCCC", "host1.example.com")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_SetIPv6ReverseDNS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.SetIPv6ReverseDNS("123456789", "AAAA:BBBB:CCCC", "host1.example.com"))
}

func Test_IP_DefaultIPv4ReverseDNS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DefaultIPv4ReverseDNS("123456789", "123.456.789.0")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_DefaultIPv4ReverseDNS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DefaultIPv4ReverseDNS("123456789", "123.456.789.0"))
}

func Test_IP_SetIPv4ReverseDNS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.SetIPv4ReverseDNS("123456789", "123.456.789.0", "host1.example.com")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_IP_SetIPv4ReverseDNS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.SetIPv4ReverseDNS("123456789", "123.456.789.0", "host1.example.com"))
}
