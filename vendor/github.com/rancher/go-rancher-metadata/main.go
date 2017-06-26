package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	metadataUrl = "http://rancher-metadata/2015-12-19"
)

func main() {
	m := metadata.NewClient(metadataUrl)

	version := "init"

	for {
		newVersion, err := m.GetVersion()
		if err != nil {
			logrus.Errorf("Error reading metadata version: %v", err)
		} else if version == newVersion {
			logrus.Debug("No changes in metadata version")
		} else {
			logrus.Debugf("Metadata version has changed, oldVersion=[%s], newVersion=[%s]", version, newVersion)
			version = newVersion
		}
		time.Sleep(5 * time.Second)
	}
}
