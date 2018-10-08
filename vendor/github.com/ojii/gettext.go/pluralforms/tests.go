package pluralforms

type equal struct {
	value uint32
}

func (e equal) test(n uint32) bool {
	return n == e.value
}

type notequal struct {
	value uint32
}

func (e notequal) test(n uint32) bool {
	return n != e.value
}

type gt struct {
	value   uint32
	flipped bool
}

func (e gt) test(n uint32) bool {
	if e.flipped {
		return e.value > n
	} else {
		return n > e.value
	}
}

type lt struct {
	value   uint32
	flipped bool
}

func (e lt) test(n uint32) bool {
	if e.flipped {
		return e.value < n
	} else {
		return n < e.value
	}
}

type gte struct {
	value   uint32
	flipped bool
}

func (e gte) test(n uint32) bool {
	if e.flipped {
		return e.value >= n
	} else {
		return n >= e.value
	}
}

type lte struct {
	value   uint32
	flipped bool
}

func (e lte) test(n uint32) bool {
	if e.flipped {
		return e.value <= n
	} else {
		return n <= e.value
	}
}

type and struct {
	left  test
	right test
}

func (e and) test(n uint32) bool {
	if !e.left.test(n) {
		return false
	} else {
		return e.right.test(n)
	}
}

type or struct {
	left  test
	right test
}

func (e or) test(n uint32) bool {
	if e.left.test(n) {
		return true
	} else {
		return e.right.test(n)
	}
}

type pipe struct {
	modifier math
	action   test
}

func (e pipe) test(n uint32) bool {
	return e.action.test(e.modifier.calc(n))
}
