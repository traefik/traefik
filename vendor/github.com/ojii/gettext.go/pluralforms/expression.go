package pluralforms

// Expression is a plurfalforms expression. Eval evaluates the expression for
// a given n value. Use pluralforms.Compile to generate Expression instances.
type Expression interface {
	Eval(n uint32) int
}

type const_value struct {
	value int
}

func (c const_value) Eval(n uint32) int {
	return c.value
}

type test interface {
	test(n uint32) bool
}

type ternary struct {
	test       test
	true_expr  Expression
	false_expr Expression
}

func (t ternary) Eval(n uint32) int {
	if t.test.test(n) {
		if t.true_expr == nil {
			return -1
		}
		return t.true_expr.Eval(n)
	} else {
		if t.false_expr == nil {
			return -1
		}
		return t.false_expr.Eval(n)
	}
}
