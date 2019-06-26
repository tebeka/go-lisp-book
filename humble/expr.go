package main

import (
	"bytes"
	"fmt"
	"strings"
)

// Expression to be computed
type Expression interface {
	Eval(env *Environment) (Object, error)
}

// NumberExpr is a number. e.g. 3.14
type NumberExpr struct {
	value float64
}

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%v", e.value)
}

// Eval evaluates value
func (e *NumberExpr) Eval(env *Environment) (Object, error) {
	return e.value, nil
}

// SymbolExpr is a symbol. e.g. pi
type SymbolExpr struct {
	name string
}

func (e *SymbolExpr) String() string {
	return fmt.Sprintf("%v", e.name)
}

// Eval evaluates value
func (e *SymbolExpr) Eval(env *Environment) (Object, error) {
	env = env.Find(e.name)
	if env == nil {
		return nil, fmt.Errorf("unknown name - %q", e.name)
	}

	return env.Get(e.name), nil
}

// ListExpr is a list expression. e.g. (* 4 5)
type ListExpr struct {
	children []Expression
}

func (e *ListExpr) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(")
	for i, c := range e.children {
		fmt.Fprintf(&buf, "%s", c)
		if i < len(e.children)-1 {
			fmt.Fprintf(&buf, " ")
		}
	}
	fmt.Fprintf(&buf, ")")
	return buf.String()
}

// Eval evaluates value
func (e *ListExpr) Eval(env *Environment) (Object, error) {
	if len(e.children) == 0 {
		return nil, fmt.Errorf("empty list expression")
	}

	rest := e.children[1:]

	// Try special forms first
	ne, ok := e.children[0].(*SymbolExpr)
	if ok {
		op := ne.name

		switch op {
		case "define": // (define n 27)
			return evalDefine(rest, env)
		case "set!": // (set! n 27)
			return evalSet(rest, env)
		case "if": // (if (< x 0) 0 x)
			return evalIf(rest, env)
		case "or": // (or), (or 0 1)
			return evalOr(rest, env)
		case "and": // (and), (and 0 1)
			return evalAnd(rest, env)
		case "lambda": // (lambda (n) (+ n 1))
			return evalLambda(rest, env)
		}
	}

	obj, err := e.children[0].Eval(env)
	if err != nil {
		return nil, err
	}

	c, ok := obj.(Callable)
	if !ok {
		return nil, fmt.Errorf("%s is not callable", obj)
	}

	var params []Object
	for _, e := range rest {
		obj, err := e.Eval(env)
		if err != nil {
			return nil, err
		}
		params = append(params, obj)
	}

	return c.Call(params)
}

// Lambda is a lambda object. e.g. (lambda (n) (+ n 1))
type Lambda struct {
	env    *Environment
	params []string
	body   Expression
}

// Call implements Callable
func (l *Lambda) Call(args []Object) (Object, error) {
	if len(args) != len(l.params) {
		return nil, fmt.Errorf("wrong number of arguments (want %d, got %d)", len(l.params), args)
	}

	m := make(map[string]Object)
	for i, name := range l.params {
		m[name] = args[i]
	}

	env := NewEnvironment(m, l.env)
	return l.body.Eval(env)
}

func (l *Lambda) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(lambda (")
	fmt.Fprint(&buf, strings.Join(l.params, " "))
	fmt.Fprintf(&buf, ") ")
	fmt.Fprintf(&buf, "%s", l.body)
	return buf.String()
}
