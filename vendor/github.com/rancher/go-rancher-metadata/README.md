# Go bindings for Rancher-metadata

This library is incomplete, but implements a variety of calls against  [rancher-metadata](https://github.com/rancher/rancher-metadata) service

#Example usage

```go
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
			logrus.Debugf("Metadata Version has been changed. Old version: %s. New version: %s.", version, newVersion)
			version = newVersion
		}
		time.Sleep(5 * time.Second)
	}
}
```
