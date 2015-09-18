package backoff

import "log"

func ExampleRetry() error {
	operation := func() error {
		// An operation that might fail.
		return nil // or return errors.New("some error")
	}

	err := Retry(operation, NewExponentialBackOff())
	if err != nil {
		// Handle error.
		return err
	}

	// Operation is successful.
	return nil
}

func ExampleTicker() error {
	operation := func() error {
		// An operation that might fail
		return nil // or return errors.New("some error")
	}

	b := NewExponentialBackOff()
	ticker := NewTicker(b)

	var err error

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for _ = range ticker.C {
		if err = operation(); err != nil {
			log.Println(err, "will retry...")
			continue
		}

		ticker.Stop()
		break
	}

	if err != nil {
		// Operation has failed.
		return err
	}

	// Operation is successful.
	return nil
}
