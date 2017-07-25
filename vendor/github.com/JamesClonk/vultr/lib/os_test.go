package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OS_GetOS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	os, err := client.GetOS()
	assert.Nil(t, os)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_OS_GetOS_NoOS(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	os, err := client.GetOS()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, os)
}

func Test_OS_GetOS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"124":{"OSID":124,"name":"Windows 2012 R2 x64","arch":"x64","family":"windows","windows":true},
"127":{"OSID":127,"name":"CentOS 6 x64","arch":"x64","family":"centos","windows":false},
"179":{"OSID":179,"name":"CoreOS Stable","arch":"x64","family":"coreos","windows":false}}`)
	defer server.Close()

	os, err := client.GetOS()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, os) {
		assert.Equal(t, 3, len(os))

		assert.Equal(t, 127, os[0].ID)
		assert.Equal(t, "CentOS 6 x64", os[0].Name)
		assert.Equal(t, "x64", os[0].Arch)
		assert.Equal(t, "centos", os[0].Family)

		assert.Equal(t, 179, os[1].ID)
		assert.Equal(t, "coreos", os[1].Family)
		assert.Equal(t, "CoreOS Stable", os[1].Name)
		assert.Equal(t, false, os[1].Windows)

		assert.Equal(t, 124, os[2].ID)
		assert.Equal(t, "windows", os[2].Family)
		assert.Equal(t, "Windows 2012 R2 x64", os[2].Name)
		assert.Equal(t, true, os[2].Windows)
	}
}
