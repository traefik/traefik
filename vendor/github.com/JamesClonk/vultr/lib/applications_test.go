package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Applications_GetApplications_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	apps, err := client.GetApplications()
	assert.Nil(t, apps)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Applications_GetApplications_NoApplication(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	apps, err := client.GetApplications()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, apps)
}

func Test_Applications_GetApplications_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"2": {"APPID": "2","name": "WordPress","short_name": "wordpress","deploy_name": "WordPress on CentOS 6 x64","surcharge": 0},
"1": {"APPID": "1","name": "LEMP","short_name": "lemp","deploy_name": "LEMP on CentOS 6 x64","surcharge": 5}
}`)
	defer server.Close()

	apps, err := client.GetApplications()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, apps) {
		assert.Equal(t, 2, len(apps))

		assert.Equal(t, "1", apps[0].ID)
		assert.Equal(t, "LEMP", apps[0].Name)
		assert.Equal(t, "lemp", apps[0].ShortName)
		assert.Equal(t, "LEMP on CentOS 6 x64", apps[0].DeployName)
		assert.Equal(t, float64(5), apps[0].Surcharge)

		assert.Equal(t, "2", apps[1].ID)
		assert.Equal(t, "WordPress", apps[1].Name)
		assert.Equal(t, "wordpress", apps[1].ShortName)
		assert.Equal(t, "WordPress on CentOS 6 x64", apps[1].DeployName)
	}
}
