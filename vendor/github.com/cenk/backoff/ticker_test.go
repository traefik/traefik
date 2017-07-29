package backoff

import (
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestTicker(t *testing.T) {
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

	b := NewExponentialBackOff()
	ticker := NewTicker(b)

	var err error
	for _ = range ticker.C {
		if err = f(); err != nil {
			t.Log(err)
			continue
		}

		break
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if i != successOn {
		t.Errorf("invalid number of retries: %d", i)
	}
}

func TestTickerContext(t *testing.T) {
	const cancelOn = 3
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

	b := WithContext(NewConstantBackOff(time.Millisecond), ctx)
	ticker := NewTicker(b)

	var err error
	for _ = range ticker.C {
		if err = f(); err != nil {
			t.Log(err)
			continue
		}

		break
	}
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
