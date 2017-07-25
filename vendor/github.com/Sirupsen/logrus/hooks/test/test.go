package test

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

// test.Hook is a hook designed for dealing with logs in test scenarios.
type Hook struct {
	Entries []*logrus.Entry
}

// Installs a test hook for the global logger.
func NewGlobal() *Hook {

	hook := new(Hook)
	logrus.AddHook(hook)

	return hook

}

// Installs a test hook for a given local logger.
func NewLocal(logger *logrus.Logger) *Hook {

	hook := new(Hook)
	logger.Hooks.Add(hook)

	return hook

}

// Creates a discarding logger and installs the test hook.
func NewNullLogger() (*logrus.Logger, *Hook) {

	logger := logrus.New()
	logger.Out = ioutil.Discard

	return logger, NewLocal(logger)

}

func (t *Hook) Fire(e *logrus.Entry) error {
	t.Entries = append(t.Entries, e)
	return nil
}

func (t *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// LastEntry returns the last entry that was logged or nil.
func (t *Hook) LastEntry() (l *logrus.Entry) {

	if i := len(t.Entries) - 1; i < 0 {
		return nil
	} else {
		return t.Entries[i]
	}

}

// Reset removes all Entries from this test hook.
func (t *Hook) Reset() {
	t.Entries = make([]*logrus.Entry, 0)
}
