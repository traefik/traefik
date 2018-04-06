package dns

import (
	"fmt"
)

type ConfigDNSError interface {
	error
	Network() bool
	NotFound() bool
	FailedToSave() bool
	ValidationFailed() bool
}

func IsConfigDNSError(e error) bool {
	_, ok := e.(ConfigDNSError)
	return ok
}

type ZoneError struct {
	zoneName         string
	httpErrorMessage string
	apiErrorMessage  string
	err              error
}

func (e *ZoneError) Network() bool {
	if e.httpErrorMessage != "" {
		return true
	}
	return false
}

func (e *ZoneError) NotFound() bool {
	if e.err == nil && e.httpErrorMessage == "" && e.apiErrorMessage == "" {
		return true
	}
	return false
}

func (e *ZoneError) FailedToSave() bool {
	return false
}

func (e *ZoneError) ValidationFailed() bool {
	if e.apiErrorMessage != "" {
		return true
	}
	return false
}

func (e *ZoneError) Error() string {
	if e.Network() {
		return fmt.Sprintf("Zone \"%s\" network error: [%s]", e.zoneName, e.httpErrorMessage)
	}

	if e.NotFound() {
		return fmt.Sprintf("Zone \"%s\" not found.", e.zoneName)
	}

	if e.FailedToSave() {
		return fmt.Sprintf("Zone \"%s\" failed to save: [%s]", e.zoneName, e.err.Error())
	}

	if e.ValidationFailed() {
		return fmt.Sprintf("Zone \"%s\" validation failed: [%s]", e.zoneName, e.apiErrorMessage)
	}

	if e.err != nil {
		return e.err.Error()
	}

	return "<nil>"
}

type RecordError struct {
	fieldName        string
	httpErrorMessage string
	err              error
}

func (e *RecordError) Network() bool {
	if e.httpErrorMessage != "" {
		return true
	}
	return false
}

func (e *RecordError) NotFound() bool {
	return false
}

func (e *RecordError) FailedToSave() bool {
	if e.fieldName == "" {
		return true
	}
	return false
}

func (e *RecordError) ValidationFailed() bool {
	if e.fieldName != "" {
		return true
	}
	return false
}

func (e *RecordError) Error() string {
	if e.Network() {
		return fmt.Sprintf("Record network error: [%s]", e.httpErrorMessage)
	}

	if e.NotFound() {
		return fmt.Sprintf("Record not found.")
	}

	if e.FailedToSave() {
		return fmt.Sprintf("Record failed to save: [%s]", e.err.Error())
	}

	if e.ValidationFailed() {
		return fmt.Sprintf("Record validation failed for field [%s]", e.fieldName)
	}

	return "<nil>"
}
