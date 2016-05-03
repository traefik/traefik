package main

import (
	"github.com/cocap10/flaeg"
	"reflect"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	var configuration GlobalConfiguration
	defaultConfiguration := NewGlobalConfiguration()
	args := []string{
		// "-h",
		"--docker",
		"--file",
		"--web",
		"--marathon",
		"--consul",
		"--consulcatalog",
		"--etcd",
		"--zookeeper",
		"--boltdb",
	}
	if err := flaeg.Load(&configuration, defaultConfiguration, args); err != nil {
		t.Fatalf("Error: %s", err)
	}
	// fmt.Printf("result : \n%+v\n", configuration)
	if !reflect.DeepEqual(configuration, *defaultConfiguration) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v", *defaultConfiguration, configuration)
	}
}

func TestLoadWithParsers(t *testing.T) {
	var configuration GlobalConfiguration
	defaultConfiguration := NewGlobalConfiguration()
	args := []string{
		// "-h",
		"--docker",
		// "--file",
		"--web.address=:8888",
		"--marathon",
		"--consul",
		"--consulcatalog",
		"--etcd.tls.insecureskipverify",
		"--zookeeper",
		"--boltdb",
		"--accesslogsfile=log2/access.log",
		"--entrypoints=Name:http Address::8000 Redirect.EntryPoint:https",
		"--entrypoints=Name:https Address::8443 Redirect.EntryPoint:http",
		"--defaultentrypoints=https",
		"--defaultentrypoints=ssh",
		"--providersthrottleduration=4s",
	}
	parsers := map[reflect.Type]flaeg.Parser{}
	var defaultEntryPointsParser DefaultEntryPoints
	parsers[reflect.TypeOf(DefaultEntryPoints{})] = &defaultEntryPointsParser
	entryPointsParser := EntryPoints{}
	parsers[reflect.TypeOf(EntryPoints{})] = &entryPointsParser

	if err := flaeg.LoadWithParsers(&configuration, defaultConfiguration, args, parsers); err != nil {
		t.Fatalf("Error: %s", err)
	}
	// fmt.Printf("result : \n%+v\n", configuration)

	//Check
	check := *defaultConfiguration
	check.File = nil
	check.Web.Address = ":8888"
	check.AccessLogsFile = "log2/access.log"
	check.Etcd.TLS.InsecureSkipVerify = true
	check.EntryPoints = make(map[string]*EntryPoint)
	check.EntryPoints["http"] = &EntryPoint{
		Address: ":8000",
		Redirect: &Redirect{
			EntryPoint: "https",
		},
	}
	check.EntryPoints["https"] = &EntryPoint{
		Address: ":8443",
		Redirect: &Redirect{
			EntryPoint: "http",
		},
	}
	check.DefaultEntryPoints = []string{"https", "ssh"}
	check.ProvidersThrottleDuration = time.Duration(4 * time.Second)

	if !reflect.DeepEqual(&configuration, &check) {
		t.Fatalf("\nexpected\t: %+v\ngot\t\t\t: %+v", check, configuration)
	}
}
