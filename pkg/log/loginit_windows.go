// +build windows

package log

import (
	"github.com/Microsoft/go-winio/pkg/etwlogrus"
	"github.com/sirupsen/logrus"
)

func InitNativeTracer() {
	// Provider ID: {63fae199-b614-503b-8fb6-c0be6dbe3fe5}
	// GUID is generated based on name - see Microsoft/go-winio/tools/etw-provider-gen.
	if hook, err := etwlogrus.NewHook("traefik"); err == nil {
		logrus.AddHook(hook)
	}
	return
}
