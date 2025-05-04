package fastcgi

import (
	"io"
)

type Request struct {
	params     map[string]string
	body       io.Reader
	role       uint16
	httpMethod string
}
