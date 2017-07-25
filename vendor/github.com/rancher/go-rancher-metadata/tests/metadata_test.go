package metadata_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	rmd "github.com/rancher/rancher-metadata"
	//"reflect"
	"testing"
)

const (
	mdVersion = "2015-12-19"

	listenPort1       = ":33001"
	listenReloadPort1 = ":33011"
	metadataURL1      = "http://localhost" + listenPort1 + "/" + mdVersion
	answers1          = "./answers.host1.yml"

	listenPort2       = ":33002"
	listenReloadPort2 = ":33012"
	metadataURL2      = "http://localhost" + listenPort2 + "/" + mdVersion
	answers2          = "./answers.host2.yml"

	listenPort3       = ":33003"
	listenReloadPort3 = ":33013"
	metadataURL3      = "http://localhost" + listenPort3 + "/" + mdVersion
	answers3          = "./answers.host3.yml"
)

func init() {

	logrus.SetLevel(logrus.DebugLevel)
	runAllTestMetadataServers()
}

func runAllTestMetadataServers() {
	runTestMetadataServer1()
	runTestMetadataServer2()
	runTestMetadataServer3()
}

func runTestMetadataServer1() {
	runTestMetadataServer(answers1, metadataURL1, listenPort1, listenReloadPort1)
}

func runTestMetadataServer2() {
	runTestMetadataServer(answers2, metadataURL2, listenPort2, listenReloadPort2)
}

func runTestMetadataServer3() {
	runTestMetadataServer(answers3, metadataURL3, listenPort3, listenReloadPort3)
}

func runTestMetadataServer(answers, url, listenPort, listenReloadPort string) {

	logrus.Debugf("Starting Test Metadata Server")

	sc := rmd.NewServerConfig(
		answers,
		listenPort,
		listenReloadPort,
		true,
	)

	go func() { sc.Start() }()
}

func TestGetContainers(t *testing.T) {

	clientIP1 := "10.42.73.81"

	mc1, err := metadata.NewClientWithIPAndWait(metadataURL1, clientIP1)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}

	containers, err := mc1.GetContainers()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	expectedContainersLength := 12
	actualContainersLength := len(containers)
	if actualContainersLength != expectedContainersLength {
		t.Error("expectedContainersLength: %v actualContainersLength: %v", expectedContainersLength, actualContainersLength)
	}

}

func TestGetSelfHost(t *testing.T) {

	clientIP1 := "10.42.73.81"
	mc1, err := metadata.NewClientWithIPAndWait(metadataURL1, clientIP1)
	logrus.Debugf("mc1: %v", mc1)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}
	selfHost1, err := mc1.GetSelfHost()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	expectedSelfHost1Name := "aa-leo-tmp-10acre-1"
	if selfHost1.Name != expectedSelfHost1Name {
		t.Error("expected: %s, actual: %s", expectedSelfHost1Name, selfHost1.Name)
	}

	logrus.Debugf("selfHost1: %v", selfHost1)

	clientIP2 := "10.42.128.187"
	mc2, err := metadata.NewClientWithIPAndWait(metadataURL2, clientIP2)
	logrus.Debugf("mc2: %v", mc2)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}
	selfHost2, err := mc2.GetSelfHost()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	expectedSelfHost2Name := "aa-leo-tmp-10acre-2"
	if selfHost2.Name != expectedSelfHost2Name {
		t.Error("expected: %s, actual: %s", expectedSelfHost2Name, selfHost2.Name)
	}
}

func TestGetSelfContainer(t *testing.T) {

	clientIP1 := "10.42.73.81"
	mc1, err := metadata.NewClientWithIPAndWait(metadataURL1, clientIP1)
	logrus.Debugf("mc1: %v", mc1)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}
	c1, err := mc1.GetSelfContainer()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	logrus.Debugf("c1: %v", c1)

	clientIP2 := "10.42.128.187"
	mc2, err := metadata.NewClientWithIPAndWait(metadataURL2, clientIP2)
	logrus.Debugf("mc2: %v", mc2)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}
	c2, err := mc2.GetSelfContainer()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	if c2.PrimaryIp != clientIP2 {
		t.Error("expected: %s, actual: %c", clientIP2, c2.PrimaryIp)
	}
	logrus.Debugf("c2: %+v", c2)
}

func TestGetServices(t *testing.T) {

	clientIP1 := "10.42.73.81"
	mc1, err := metadata.NewClientWithIPAndWait(metadataURL1, clientIP1)
	logrus.Debugf("mc1: %v", mc1)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}

	services, err := mc1.GetServices()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}

	expectedServicesLength := 9
	actualServicesLength := len(services)
	if actualServicesLength != expectedServicesLength {
		t.Error("expectedServicesLength: %v actualServicesLength: %v", expectedServicesLength, actualServicesLength)
	}
}

func TestGetSelfService(t *testing.T) {

	clientIP1 := "10.42.73.81"
	mc1, err := metadata.NewClientWithIPAndWait(metadataURL1, clientIP1)
	logrus.Debugf("mc1: %v", mc1)
	if err != nil {
		logrus.Errorf("couldn't create metadata client")
	}

	selfService, err := mc1.GetSelfService()
	if err != nil {
		t.Error("not expecting error, got: %v", err)
	}
	logrus.Debugf("selfService: %+v", selfService)

	actualNumOfServiceContainers := len(selfService.Containers)
	expectedNumOfServiceContainers := 1
	if actualNumOfServiceContainers != expectedNumOfServiceContainers {
		t.Error("expected %v got %v", expectedNumOfServiceContainers, actualNumOfServiceContainers)
	}
}
