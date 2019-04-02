package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type timeout interface {
	Timeout() bool
}

func (m *client) OnChangeWithError(intervalSeconds int, do func(string)) error {
	return m.onChangeFromVersionWithError("init", intervalSeconds, do)
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

func (m *client) OnChangeCtx(ctx context.Context, intervalSeconds int, do func(string)) {
	m.onChangeFromVersionWithErrorCtx(ctx, "init", intervalSeconds, do)
}

func (m *client) onChangeFromVersionWithErrorCtx(ctx context.Context, version string, intervalSeconds int, do func(string)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		newVersion, err := m.waitVersionCtx(ctx, intervalSeconds, version)
		if err != nil {
			t, ok := err.(timeout)
			if !ok || !t.Timeout() {
				logrus.Errorf("Error reading metadata version: %v", err)
				time.Sleep(time.Duration(intervalSeconds) * time.Second)
			}
			continue
		}

		if version == newVersion {
			logrus.Debug("No changes in metadata version")
		} else {
			logrus.Debugf("Metadata Version has been changed. Old version: %s. New version: %s.", version, newVersion)
			version = newVersion
			do(newVersion)
		}
	}
}

func (m *client) waitVersionCtx(ctx context.Context, maxWait int, version string) (string, error) {
	resp, err := m.SendRequestCtx(ctx, fmt.Sprintf("/version?wait=true&value=%s&maxWait=%d", version, maxWait))
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(resp, &version)
	return version, err
}
