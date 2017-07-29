package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StartupScripts_GetStartupScripts_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	scripts, err := client.GetStartupScripts()
	assert.Nil(t, scripts)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_StartupScripts_GetStartupScripts_NoScripts(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	scripts, err := client.GetStartupScripts()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, scripts)
}

func Test_StartupScripts_GetStartupScripts_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"5": {"SCRIPTID": "5","date_created": "2014-08-22 15:27:18","date_modified": "2014-09-22 15:27:18","name": "beta","type": "pxe",
    "script": "#!ipxe\necho Hello World\nshell"},
"3": {"SCRIPTID": "3","date_created": "2014-05-21 15:27:18","date_modified": "2014-05-21 15:27:18","name": "alpha","type": "boot",
    "script": "#!/bin/bash echo Hello World > /root/hello"}}`)
	defer server.Close()

	scripts, err := client.GetStartupScripts()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, scripts) {
		assert.Equal(t, 2, len(scripts))

		assert.Equal(t, "3", scripts[0].ID)
		assert.Equal(t, "alpha", scripts[0].Name)
		assert.Equal(t, "boot", scripts[0].Type)
		assert.Equal(t, "#!/bin/bash echo Hello World > /root/hello", scripts[0].Content)

		assert.Equal(t, "5", scripts[1].ID)
		assert.Equal(t, "beta", scripts[1].Name)
		assert.Equal(t, "pxe", scripts[1].Type)
		assert.Equal(t, "#!ipxe\necho Hello World\nshell", scripts[1].Content)
	}
}

func Test_StartupScripts_GetStartupScript_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	_, err := client.GetStartupScript("5")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_StartupScripts_GetStartupScript_NoScript(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	script, err := client.GetStartupScript("5")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, StartupScript{}, script)
}

func Test_StartupScripts_GetStartupScript_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"3": {"SCRIPTID": "3","date_created": "2014-05-21 15:27:18","date_modified": "2014-05-21 15:27:18","name": "alpha","type": "boot",
    "script": "#!/bin/bash echo Hello World > /root/hello"},
"5": {"SCRIPTID": "5","date_created": "2014-08-22 15:27:18","date_modified": "2014-09-22 15:27:18","name": "beta","type": "pxe",
    "script": "#!ipxe\necho Hello World\nshell"}}`)
	defer server.Close()

	script, err := client.GetStartupScript("5")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, script) {
		assert.Equal(t, "beta", script.Name)
		assert.Equal(t, "pxe", script.Type)
		assert.Equal(t, "#!ipxe\necho Hello World\nshell", script.Content)
	}
}

func Test_StartupScripts_CreateStartupScript_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	script, err := client.CreateStartupScript("delta", "#!/bin/bash echo Hello World > /root/hello", "boot")
	assert.Equal(t, StartupScript{}, script)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_StartupScripts_CreateStartupScript_NoKey(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	script, err := client.CreateStartupScript("delta", "#!/bin/bash echo Hello World > /root/hello", "boot")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "", script.ID)
}

func Test_StartupScripts_CreateStartupScript_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"SCRIPTID":"123"}`)
	defer server.Close()

	script, err := client.CreateStartupScript("delta", "#!/bin/bash echo Hello World > /root/hello", "boot")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, script) {
		assert.Equal(t, "123", script.ID)
		assert.Equal(t, "delta", script.Name)
		assert.Equal(t, "boot", script.Type)
		assert.Equal(t, "#!/bin/bash echo Hello World > /root/hello", script.Content)
	}
}

func Test_StartupScripts_UpdateStartupScript_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.UpdateStartupScript(StartupScript{"o1", "omega", "oooo", "2012-12-12 12:12:12"})
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_StartupScripts_UpdateStartupScript_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	script := StartupScript{"12345", "omega", "pxe", "#!ipxe\necho Hello World\nshell"}
	if err := client.UpdateStartupScript(script); err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, script) {
		assert.Equal(t, "12345", script.ID)
		assert.Equal(t, "omega", script.Name)
		assert.Equal(t, "pxe", script.Type)
		assert.Equal(t, "#!ipxe\necho Hello World\nshell", script.Content)
	}
}

func Test_StartupScripts_DeleteStartupScript_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DeleteStartupScript("7")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_StartupScripts_DeleteStartupScript_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DeleteStartupScript("7"))
}
