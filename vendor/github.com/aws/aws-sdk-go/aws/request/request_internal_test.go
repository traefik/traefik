package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	handlers := Handlers{}
	op := &Operation{}
	op.HTTPMethod = "Foo"
	req := &Request{}
	req.Operation = op
	req.Handlers = handlers

	r := req.copy()
	assert.NotEqual(t, req, r)
	assert.Equal(t, req.Operation.HTTPMethod, r.Operation.HTTPMethod)
}
