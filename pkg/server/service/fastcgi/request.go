package fastcgi

import (
	"fmt"
	"io"
	"strconv"
)

type Request struct {
	params     map[string]string
	body       io.Reader
	role       uint16
	httpMethod string
}

func NewRequest(body io.Reader, params map[string]string) (*Request, error) {
	if contentLen, ok := params["CONTENT_LENGTH"]; !ok {
		return nil, fmt.Errorf("fastcgi request must have content length parameter")
	} else if _, err := strconv.ParseUint(contentLen, 10, 64); err != nil {
		return nil, fmt.Errorf("fastcgi request content length must be positive int")
	}

	return &Request{
		params: params,
		body:   body,
		role:   FastCgiRoleResponder,
	}, nil
}
