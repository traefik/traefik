package predicate

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
)

func NewParser(d Def) (Parser, error) {
	return &predicateParser{d: d}, nil
}

type predicateParser struct {
	d Def
}

func (p *predicateParser) Parse(in string) (interface{}, error) {
	expr, err := parser.ParseExpr(in)
	if err != nil {
		return nil, err
	}
	return p.parseNode(expr)
}

func (p *predicateParser) parseNode(node ast.Node) (interface{}, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		return literalToValue(n)
	case *ast.BinaryExpr:
		x, err := p.parseNode(n.X)
		if err != nil {
			return nil, err
		}
		y, err := p.parseNode(n.Y)
		if err != nil {
			return nil, err
		}
		return p.joinPredicates(n.Op, x, y)
	case *ast.CallExpr:
		// We expect function that will return predicate
		name, err := getIdentifier(n.Fun)
		if err != nil {
			return nil, err
		}
		fn, err := p.getFunction(name)
		if err != nil {
			return nil, err
		}
		arguments, err := collectLiterals(n.Args)
		if err != nil {
			return nil, err
		}
		return callFunction(fn, arguments)
	case *ast.ParenExpr:
		return p.parseNode(n.X)
	}
	return nil, fmt.Errorf("unsupported %T", node)
}

func (p *predicateParser) getFunction(name string) (interface{}, error) {
	v, ok := p.d.Functions[name]
	if !ok {
		return nil, fmt.Errorf("unsupported function: %s", name)
	}
	return v, nil
}

func (p *predicateParser) joinPredicates(op token.Token, a, b interface{}) (interface{}, error) {
	joinFn, err := p.getJoinFunction(op)
	if err != nil {
		return nil, err
	}
	return callFunction(joinFn, []interface{}{a, b})
}

func (p *predicateParser) getJoinFunction(op token.Token) (interface{}, error) {
	var fn interface{}
	switch op {
	case token.LAND:
		fn = p.d.Operators.AND
	case token.LOR:
		fn = p.d.Operators.OR
	case token.GTR:
		fn = p.d.Operators.GT
	case token.GEQ:
		fn = p.d.Operators.GE
	case token.LSS:
		fn = p.d.Operators.LT
	case token.LEQ:
		fn = p.d.Operators.LE
	case token.EQL:
		fn = p.d.Operators.EQ
	case token.NEQ:
		fn = p.d.Operators.NEQ
	}
	if fn == nil {
		return nil, fmt.Errorf("%v is not supported", op)
	}
	return fn, nil
}

func getIdentifier(node ast.Node) (string, error) {
	sexpr, ok := node.(*ast.SelectorExpr)
	if ok {
		id, ok := sexpr.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("expected selector identifier, got: %T", sexpr.X)
		}
		return fmt.Sprintf("%s.%s", id.Name, sexpr.Sel.Name), nil
	}

	id, ok := node.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("expected identifier, got: %T", node)
	}
	return id.Name, nil
}

func collectLiterals(nodes []ast.Expr) ([]interface{}, error) {
	out := make([]interface{}, len(nodes))
	for i, n := range nodes {
		l, ok := n.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("expected literal, got %T", n)
		}
		val, err := literalToValue(l)
		if err != nil {
			return nil, err
		}
		out[i] = val
	}
	return out, nil
}

func literalToValue(a *ast.BasicLit) (interface{}, error) {
	switch a.Kind {
	case token.FLOAT:
		value, err := strconv.ParseFloat(a.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse argument: %s, error: %s", a.Value, err)
		}
		return value, nil
	case token.INT:
		value, err := strconv.Atoi(a.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse argument: %s, error: %s", a.Value, err)
		}
		return value, nil
	case token.STRING:
		value, err := strconv.Unquote(a.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse argument: %s, error: %s", a.Value, err)
		}
		return value, nil
	}
	return nil, fmt.Errorf("unsupported function argument type: '%v'", a.Kind)
}

func callFunction(f interface{}, args []interface{}) (v interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()
	arguments := make([]reflect.Value, len(args))
	for i, a := range args {
		arguments[i] = reflect.ValueOf(a)
	}
	fn := reflect.ValueOf(f)

	ret := fn.Call(arguments)
	switch len(ret) {
	case 1:
		return ret[0].Interface(), nil
	case 2:
		v, e := ret[0].Interface(), ret[1].Interface()
		if e == nil {
			return v, nil
		}
		err, ok := e.(error)
		if !ok {
			return nil, fmt.Errorf("expected error as a second return value, got %T", e)
		}
		return v, err
	}
	return nil, fmt.Errorf("expected at least one return argument for '%v'", fn)
}
