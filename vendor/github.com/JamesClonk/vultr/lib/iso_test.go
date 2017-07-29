package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ISO_GetISO_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	iso, err := client.GetISO()
	assert.Nil(t, iso)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_ISO_GetISO_NoISO(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	iso, err := client.GetISO()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, iso)
}

func Test_ISO_GetISO_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"24":{"ISOID": 24,"date_created": "2014-04-01 14:10:09","filename": "CentOS-6.5-x86_64-minimal.iso",
        "size": 9342976,"md5sum": "ec0669895a250f803e1709d0402fc411"},
"37":{"ISOID": 37,"date_created": null,"filename": "ArchLinux-2013-01-01.iso",
        "size": 2345678,"md5sum": "bc583993dcb7aaff88820bc893a778f0"}}`)
	defer server.Close()

	iso, err := client.GetISO()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, iso) {
		assert.Equal(t, 2, len(iso))

		assert.Equal(t, 37, iso[0].ID)
		assert.Equal(t, "ArchLinux-2013-01-01.iso", iso[0].Filename)
		assert.Equal(t, 2345678, iso[0].Size)
		assert.Equal(t, "bc583993dcb7aaff88820bc893a778f0", iso[0].MD5sum)
		assert.Equal(t, "", iso[0].Created)

		assert.Equal(t, 24, iso[1].ID)
		assert.Equal(t, "CentOS-6.5-x86_64-minimal.iso", iso[1].Filename)
		assert.Equal(t, 9342976, iso[1].Size)
		assert.Equal(t, "ec0669895a250f803e1709d0402fc411", iso[1].MD5sum)
		assert.Equal(t, "2014-04-01 14:10:09", iso[1].Created)
	}
}
