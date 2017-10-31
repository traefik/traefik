package egoscale

import (
	"fmt"
)

func (e *Error) Error() error {
	return fmt.Errorf("exoscale API error %d (internal code: %d): %s", e.ErrorCode, e.CSErrorCode, e.ErrorText)
}
