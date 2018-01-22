package trace

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteError sets up HTTP error response and writes it to writer w
func WriteError(w http.ResponseWriter, err error) {
	if IsAggregate(err) {
		for i := 0; i < maxHops; i++ {
			var aggErr Aggregate
			var ok bool
			if aggErr, ok = Unwrap(err).(Aggregate); !ok {
				break
			}
			errors := aggErr.Errors()
			if len(errors) == 0 {
				break
			}
			err = errors[0]
		}
	}
	writeError(w, err)
}

// ErrorToCode returns an appropriate HTTP status code based on the provided error type
func ErrorToCode(err error) int {
	if IsNotFound(err) {
		return http.StatusNotFound
	} else if IsBadParameter(err) || IsOAuth2(err) {
		return http.StatusBadRequest
	} else if IsCompareFailed(err) {
		return http.StatusPreconditionFailed
	} else if IsAccessDenied(err) {
		return http.StatusForbidden
	} else if IsAlreadyExists(err) {
		return http.StatusConflict
	} else if IsLimitExceeded(err) {
		return http.StatusTooManyRequests
	} else if IsConnectionProblem(err) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

func writeError(w http.ResponseWriter, err error) {
	replyJSON(w, ErrorToCode(err), err)
}

// ReadError converts http error to internal error type
// based on HTTP response code and HTTP body contents
// if status code does not indicate error, it will return nil
func ReadError(statusCode int, re []byte) error {
	var e error
	switch statusCode {
	case http.StatusNotFound:
		e = &NotFoundError{Message: string(re)}
	case http.StatusBadRequest:
		e = &BadParameterError{Message: string(re)}
	case http.StatusPreconditionFailed:
		e = &CompareFailedError{Message: string(re)}
	case http.StatusForbidden:
		e = &AccessDeniedError{Message: string(re)}
	case http.StatusConflict:
		e = &AlreadyExistsError{Message: string(re)}
	case http.StatusTooManyRequests:
		e = &LimitExceededError{Message: string(re)}
	case http.StatusGatewayTimeout:
		e = &ConnectionProblemError{Message: string(re)}
	default:
		if statusCode < 200 || statusCode > 299 {
			return Errorf(string(re))
		}
		return nil
	}
	return unmarshalError(e, re)
}

func replyJSON(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var out []byte
	if IsDebug() {
		// trace error can marshal itself,
		// otherwise capture error message and marshal it explicitly
		var obj interface{} = err
		if _, ok := err.(*TraceErr); !ok {
			obj = message{Message: err.Error()}
		}
		out, err = json.MarshalIndent(obj, "", "    ")
		if err != nil {
			out = []byte(fmt.Sprintf(`{"message": "internal marshal error: %v"}`, err))
		}
	} else {
		innerError := err
		if terr, ok := err.(Error); ok {
			innerError = terr.OrigError()
		}
		out, err = json.Marshal(message{Message: innerError.Error()})
	}
	w.Write(out)
}

type message struct {
	Message string `json:"message"`
}

func unmarshalError(err error, responseBody []byte) error {
	if len(responseBody) == 0 {
		return err
	}
	var raw RawTrace
	if err2 := json.Unmarshal(responseBody, &raw); err2 != nil {
		return err
	}
	if len(raw.Traces) != 0 && len(raw.Err) != 0 {
		// try to capture traces, if there are any
		err2 := json.Unmarshal(raw.Err, err)
		if err2 != nil {
			return err
		}
		return &TraceErr{Traces: raw.Traces, Err: err, Message: raw.Message}
	}
	json.Unmarshal(responseBody, err)
	return err
}
