package server

import (
	"testing"
	"time"

	"github.com/containous/traefik/types"
)

func TestThrottleProviderConfigReload(t *testing.T) {
	throttleDuration := 30 * time.Millisecond
	publishConfig := make(chan types.ConfigMessage)
	providerConfig := make(chan types.ConfigMessage)
	stop := make(chan bool)
	defer func() {
		stop <- true
	}()

	throttler := newProviderUpdateThrottler(nil, throttleDuration)
	go throttler.throttleProviderConfigReload(publishConfig, providerConfig, stop)

	publishedConfigCount := 0
	stopConsumeConfigs := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-stopConsumeConfigs:
				return
			case <-publishConfig:
				publishedConfigCount++
			}
		}
	}()

	// publish 5 new configs, one new config each 10 milliseconds
	for i := 0; i < 5; i++ {
		providerConfig <- types.ConfigMessage{}
		time.Sleep(10 * time.Millisecond)
	}

	// after 50 milliseconds 5 new configs were published
	// with a throttle duration of 30 milliseconds this means, we should have received 2 new configs
	wantPublishedConfigCount := 2
	if publishedConfigCount != wantPublishedConfigCount {
		t.Errorf("%d times configs were published, want %d times", publishedConfigCount, wantPublishedConfigCount)
	}

	stopConsumeConfigs <- true

	select {
	case <-publishConfig:
		// There should be exactly one more message that we receive after ~60 milliseconds since the start of the test.
		select {
		case <-publishConfig:
			t.Error("extra config publication found")
		case <-time.After(100 * time.Millisecond):
			return
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Last config was not published in time")
	}
}
