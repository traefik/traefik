package backoff

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestRetry(t *testing.T) {
	const successOn = 3
	var i = 0

	// This function is successful on "successOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == successOn {
			log.Println("OK")
			return nil
		}

		log.Println("error")
		return errors.New("error")
	}

	err := Retry(f, NewExponentialBackOff())
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryContext(t *testing.T) {
	var cancelOn = 3
	var i = 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This function cancels context on "cancelOn" calls.
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		// cancelling the context in the operation function is not a typical
		// use-case, however it allows to get predictable test results.
		if i == cancelOn {
			cancel()
		}

		log.Println("error")
		return fmt.Errorf("error (%d)", i)
	}

	err := Retry(f, WithContext(NewConstantBackOff(time.Millisecond), ctx))
	if err == nil {
		t.Errorf("error is unexpectedly nil")
	}
	if err.Error() != "error (3)" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != cancelOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestRetryPermenent(t *testing.T) {
	const permanentOn = 3
	var i = 0

	// This function fails permanently after permanentOn tries
	f := func() error {
		i++
		log.Printf("function is called %d. time\n", i)

		if i == permanentOn {
			log.Println("permanent error")
			return Permanent(errors.New("permanent error"))
		}

		log.Println("error")
		return errors.New("error")
	}

	err := Retry(f, NewExponentialBackOff())
	if err == nil || err.Error() != "permanent error" {
		t.Errorf("unexpected error: %s", err)
	}
	if i != permanentOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}
