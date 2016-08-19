package utils

import (
	"github.com/cenkalti/backoff"
	"time"
)

const (
	minLongJobInterval = 30 * time.Second
)

// RetryNotifyJob calls notify function with the error and wait duration
// for each failed attempt before sleep.
func RetryNotifyJob(operation backoff.Operation, b backoff.BackOff, notify backoff.Notify) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		before := time.Now()
		if err = operation(); err == nil {
			return nil
		}
		elapsed := time.Since(before)

		// If long job, we reset the backoff
		if elapsed >= minLongJobInterval {
			b.Reset()
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}

		time.Sleep(next)
	}
}
