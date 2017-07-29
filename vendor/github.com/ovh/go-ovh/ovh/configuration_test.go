package ovh

import (
	"io/ioutil"
	"os"
	"testing"
)

//
// Utils
//

var home string

func setup() {
	systemConfigPath = "./ovh.unittest.global.conf"
	userConfigPath = "/.ovh.unittest.user.conf"
	localConfigPath = "./ovh.unittest.local.conf"
	home, _ = currentUserHome()
}

func teardown() {
	os.Remove(systemConfigPath)
	os.Remove(home + userConfigPath)
	os.Remove(localConfigPath)
}

//
// Tests
//

func TestConfigFromFiles(t *testing.T) {
	// Write each parameter to one different configuration file
	// This is a simple way to test precedence

	// Prepare
	ioutil.WriteFile(systemConfigPath, []byte(`
[ovh-eu]
application_key=system
application_secret=system
consumer_key=system
`), 0660)

	ioutil.WriteFile(home+userConfigPath, []byte(`
[ovh-eu]
application_secret=user
consumer_key=user
`), 0660)

	ioutil.WriteFile(localConfigPath, []byte(`
[ovh-eu]
consumer_key=local
`), 0660)

	// Clear
	defer ioutil.WriteFile(systemConfigPath, []byte(``), 0660)
	defer ioutil.WriteFile(home+userConfigPath, []byte(``), 0660)
	defer ioutil.WriteFile(localConfigPath, []byte(``), 0660)

	// Test
	client := Client{}
	err := client.loadConfig("ovh-eu")

	// Validate
	if err != nil {
		t.Fatalf("loadConfig failed with: '%v'", err)
	}
	if client.AppKey != "system" {
		t.Fatalf("client.AppKey should be 'system'. Got '%s'", client.AppKey)
	}
	if client.AppSecret != "user" {
		t.Fatalf("client.AppSecret should be 'user'. Got '%s'", client.AppSecret)
	}
	if client.ConsumerKey != "local" {
		t.Fatalf("client.ConsumerKey should be 'local'. Got '%s'", client.ConsumerKey)
	}
}

func TestConfigFromOnlyOneFile(t *testing.T) {
	// ini package has a bug causing it to ignore all subsequent configuration
	// files if any could not be loaded. Make sure that workaround... works.

	// Prepare
	os.Remove(systemConfigPath)
	ioutil.WriteFile(home+userConfigPath, []byte(`
[ovh-eu]
application_key=user
application_secret=user
consumer_key=user
`), 0660)

	// Clear
	defer ioutil.WriteFile(home+userConfigPath, []byte(``), 0660)

	// Test
	client := Client{}
	err := client.loadConfig("ovh-eu")

	// Validate
	if err != nil {
		t.Fatalf("loadConfig failed with: '%v'", err)
	}
	if client.AppKey != "user" {
		t.Fatalf("client.AppKey should be 'user'. Got '%s'", client.AppKey)
	}
	if client.AppSecret != "user" {
		t.Fatalf("client.AppSecret should be 'user'. Got '%s'", client.AppSecret)
	}
	if client.ConsumerKey != "user" {
		t.Fatalf("client.ConsumerKey should be 'user'. Got '%s'", client.ConsumerKey)
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Prepare
	ioutil.WriteFile(systemConfigPath, []byte(`
[ovh-eu]
application_key=fail
application_secret=fail
consumer_key=fail
`), 0660)

	defer ioutil.WriteFile(systemConfigPath, []byte(``), 0660)
	os.Setenv("OVH_ENDPOINT", "ovh-eu")
	os.Setenv("OVH_APPLICATION_KEY", "env")
	os.Setenv("OVH_APPLICATION_SECRET", "env")
	os.Setenv("OVH_CONSUMER_KEY", "env")

	// Clear
	defer os.Unsetenv("OVH_ENDPOINT")
	defer os.Unsetenv("OVH_APPLICATION_KEY")
	defer os.Unsetenv("OVH_APPLICATION_SECRET")
	defer os.Unsetenv("OVH_CONSUMER_KEY")

	// Test
	client := Client{}
	err := client.loadConfig("")

	// Validate
	if err != nil {
		t.Fatalf("loadConfig failed with: '%v'", err)
	}
	if client.endpoint != OvhEU {
		t.Fatalf("client.AppKey should be 'env'. Got '%s'", client.AppKey)
	}
	if client.AppKey != "env" {
		t.Fatalf("client.AppKey should be 'env'. Got '%s'", client.AppKey)
	}
	if client.AppSecret != "env" {
		t.Fatalf("client.AppSecret should be 'env'. Got '%s'", client.AppSecret)
	}
	if client.ConsumerKey != "env" {
		t.Fatalf("client.ConsumerKey should be 'env'. Got '%s'", client.ConsumerKey)
	}
}

func TestConfigFromArgs(t *testing.T) {
	// Test
	client := Client{
		AppKey:      "param",
		AppSecret:   "param",
		ConsumerKey: "param",
	}
	err := client.loadConfig("ovh-eu")

	// Validate
	if err != nil {
		t.Fatalf("loadConfig failed with: '%v'", err)
	}
	if client.endpoint != OvhEU {
		t.Fatalf("client.AppKey should be 'param'. Got '%s'", client.AppKey)
	}
	if client.AppKey != "param" {
		t.Fatalf("client.AppKey should be 'param'. Got '%s'", client.AppKey)
	}
	if client.AppSecret != "param" {
		t.Fatalf("client.AppSecret should be 'param'. Got '%s'", client.AppSecret)
	}
	if client.ConsumerKey != "param" {
		t.Fatalf("client.ConsumerKey should be 'param'. Got '%s'", client.ConsumerKey)
	}
}

func TestEndpoint(t *testing.T) {
	// Prepare
	ioutil.WriteFile(systemConfigPath, []byte(`
[ovh-eu]
application_key=ovh
application_secret=ovh
consumer_key=ovh

[https://api.example.com:4242]
application_key=example.com
application_secret=example.com
consumer_key=example.com
`), 0660)

	// Clear
	defer ioutil.WriteFile(systemConfigPath, []byte(``), 0660)

	// Test: by name
	client := Client{}
	err := client.loadConfig("ovh-eu")
	if err != nil {
		t.Fatalf("loadConfig should not fail for endpoint 'ovh-eu'. Got '%v'", err)
	}
	if client.AppKey != "ovh" {
		t.Fatalf("configured value should be 'ovh' for endpoint 'ovh-eu'. Got '%s'", client.AppKey)
	}

	// Test: by URL
	client = Client{}
	err = client.loadConfig("https://api.example.com:4242")
	if err != nil {
		t.Fatalf("loadConfig should not fail for endpoint 'https://api.example.com:4242'. Got '%v'", err)
	}
	if client.AppKey != "example.com" {
		t.Fatalf("configured value should be 'example.com' for endpoint 'https://api.example.com:4242'. Got '%s'", client.AppKey)
	}

}

func TestMissingParam(t *testing.T) {
	// Setup
	var err error
	client := Client{
		AppKey:      "param",
		AppSecret:   "param",
		ConsumerKey: "param",
	}

	// Test
	client.endpoint = ""
	if err = client.loadConfig(""); err == nil {
		t.Fatalf("loadConfig should fail when client.endpoint is missing. Got '%s'", client.endpoint)
	}

	client.AppKey = ""
	if err = client.loadConfig("ovh-eu"); err == nil {
		t.Fatalf("loadConfig should fail when client.AppKey is missing. Got '%s'", client.AppKey)
	}
	client.AppKey = "param"

	client.AppSecret = ""
	if err = client.loadConfig("ovh-eu"); err == nil {
		t.Fatalf("loadConfig should fail when client.AppSecret is missing. Got '%s'", client.AppSecret)
	}
	client.AppSecret = "param"
}

//
// Main
//

// TestMain changes the location of configuration files. We need
// this to avoid any interference with existing configuration
// and non-root users
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
