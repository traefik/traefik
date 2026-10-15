package udp

import "errors"

// Constructor constructs a UDP middleware around the next handler.
type Constructor func(Handler) (Handler, error)

// Chain is an immutable list of UDP middleware constructors.
type Chain struct {
	constructors []Constructor
}

// NewChain creates a UDP middleware chain.
func NewChain(constructors ...Constructor) Chain {
	return Chain{constructors: constructors}
}

// Then adds a handler at the end of the chain.
func (c Chain) Then(handler Handler) (Handler, error) {
	if handler == nil {
		return nil, errors.New("cannot add a nil handler to the chain")
	}
	for index := range c.constructors {
		var err error
		handler, err = c.constructors[len(c.constructors)-1-index](handler)
		if err != nil {
			return nil, err
		}
	}
	return handler, nil
}

// Append adds constructors to the request flow.
func (c Chain) Append(constructors ...Constructor) Chain {
	values := make([]Constructor, 0, len(c.constructors)+len(constructors))
	values = append(values, c.constructors...)
	values = append(values, constructors...)
	return Chain{constructors: values}
}

// Extend appends another chain.
func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.constructors...)
}
