package metadata

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
)

func (m *client) OnChangeWithError(intervalSeconds int, do func(string)) error {
	return m.onChangeFromVersionWithError("init", intervalSeconds, do)
}

func (m *client) onChangeFromVersionWithError(version string, intervalSeconds int, do func(string)) error {
	for {
		newVersion, err := m.waitVersion(intervalSeconds, version)
		if err != nil {
			return err
		} else if version == newVersion {
			logrus.Debug("No changes in metadata version")
		} else {
			logrus.Debugf("Metadata Version has been changed. Old version: %s. New version: %s.", version, newVersion)
			version = newVersion
			do(newVersion)
		}
	}

	return nil
}

func (m *client) OnChange(intervalSeconds int, do func(string)) {
	version := "init"
	updateVersionAndDo := func(v string) {
		version = v
		do(version)
	}
	interval := time.Duration(intervalSeconds)
	for {
		if err := m.onChangeFromVersionWithError(version, intervalSeconds, updateVersionAndDo); err != nil {
			logrus.Errorf("Error reading metadata version: %v", err)
		}
		time.Sleep(interval * time.Second)
	}
}

type timeout interface {
	Timeout() bool
}

func (m *client) waitVersion(maxWait int, version string) (string, error) {
	for {
		resp, err := m.SendRequest(fmt.Sprintf("/version?wait=true&value=%s&maxWait=%d", version, maxWait))
		if err != nil {
			t, ok := err.(timeout)
			if ok && t.Timeout() {
				continue
			}
			return "", err
		}
		err = json.Unmarshal(resp, &version)
		return version, err
	}
}
