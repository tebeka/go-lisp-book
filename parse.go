package main

import (
	"fmt"
	"io"
	"strconv"
)

type Expr interface {
	Eval() (Expr, error)
}

type Number float64

func (n Number) Eval() (Expr, error) {
	return n, nil
}

type List []Expr

func (l List) Eval() (Expr, error) {
	if len(l) != 3 {
		return nil, fmt.Errorf("wrong number of arguments")
	}

	op, lhs, rhs := l[0], l[1], l[2]

	lval, err := lhs.Eval()
	if err != nil {
		return nil, err
	}

	nlval, ok := lval.(Number)
	if !ok {
		return nil, fmt.Errorf("bad left hand value")
	}

	rval, err := rhs.Eval()
	if err != nil {
		return nil, err
	}

	nrval, ok := rval.(Number)
	if !ok {
		return nil, fmt.Errorf("bad right hand value")
	}

	sym, ok := op.(Symbol)
	if !ok {
		return nil, fmt.Errorf("bad operator - %v", op)
	}

	switch sym {
	case "+":
		return Number(nlval * nrval), nil
	case "-":
		return Number(nlval - nrval), nil
	case "*":
		return Number(nlval * nrval), nil
	case "/":
		return Number(nlval * nrval), nil
	}

	return nil, fmt.Errorf("unknown operator - %v", sym)

}

type Symbol string

func (s Symbol) Eval() (Expr, error) {
	return s, nil
}

func parse(code string) (Expr, error) {
	lex := NewLexer(code)
	return readFromTokens(lex)
}

func readFromTokens(lex *Lexer) (Expr, error) {
	if !lex.Next() {
		return nil, io.EOF
	}

	switch lex.Token() {
	case "(":
		var expr List
		for lex.Peek() != ")" {
			sub, err := readFromTokens(lex)
			if err != nil {
				return nil, err
			}
			expr = append(expr, sub)
		}
		if !lex.Next() { // Pop )
			return nil, fmt.Errorf("missing )")
		}
		return expr, nil
	case ")":
		return nil, fmt.Errorf("unexpected )")
	default:
		val, err := strconv.ParseFloat(lex.Token(), 64)
		if err == nil {
			return Number(val), nil
		}
		return Symbol(lex.Token()), nil
	}
}
