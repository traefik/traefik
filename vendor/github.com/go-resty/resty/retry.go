// Copyright (c) 2015-2018 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxRetries  = 3
	defaultWaitTime    = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

type (
	// Option is to create convenient retry options like wait time, max retries, etc.
	Option func(*Options)

	// RetryConditionFunc type is for retry condition function
	RetryConditionFunc func(*Response) (bool, error)

	// Options to hold go-resty retry values
	Options struct {
		maxRetries      int
		waitTime        time.Duration
		maxWaitTime     time.Duration
		retryConditions []RetryConditionFunc
	}
)

// Retries sets the max number of retries
func Retries(value int) Option {
	return func(o *Options) {
		o.maxRetries = value
	}
}

// WaitTime sets the default wait time to sleep between requests
func WaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.waitTime = value
	}
}

// MaxWaitTime sets the max wait time to sleep between requests
func MaxWaitTime(value time.Duration) Option {
	return func(o *Options) {
		o.maxWaitTime = value
	}
}

// RetryConditions sets the conditions that will be checked for retry.
func RetryConditions(conditions []RetryConditionFunc) Option {
	return func(o *Options) {
		o.retryConditions = conditions
	}
}

// Backoff retries with increasing timeout duration up until X amount of retries
// (Default is 3 attempts, Override with option Retries(n))
func Backoff(operation func() (*Response, error), options ...Option) error {
	// Defaults
	opts := Options{
		maxRetries:      defaultMaxRetries,
		waitTime:        defaultWaitTime,
		maxWaitTime:     defaultMaxWaitTime,
		retryConditions: []RetryConditionFunc{},
	}

	for _, o := range options {
		o(&opts)
	}

	var (
		resp *Response
		err  error
	)
	base := float64(opts.waitTime)        // Time to wait between each attempt
	capLevel := float64(opts.maxWaitTime) // Maximum amount of wait time for the retry
	for attempt := 0; attempt < opts.maxRetries; attempt++ {
		resp, err = operation()

		var needsRetry bool
		var conditionErr error
		for _, condition := range opts.retryConditions {
			needsRetry, conditionErr = condition(resp)
			if needsRetry || conditionErr != nil {
				break
			}
		}

		// If the operation returned no error, there was no condition satisfied and
		// there was no error caused by the conditional functions.
		if err == nil && !needsRetry && conditionErr == nil {
			return nil
		}
		// Adding capped exponential backup with jitter
		// See the following article...
		// http://www.awsarchitectureblog.com/2015/03/backoff.html
		temp := math.Min(capLevel, base*math.Exp2(float64(attempt)))
		sleepDuration := time.Duration(int(temp/2) + rand.Intn(int(temp/2)))

		if sleepDuration < opts.waitTime {
			sleepDuration = opts.waitTime
		}
		time.Sleep(sleepDuration)
	}

	return err
}
