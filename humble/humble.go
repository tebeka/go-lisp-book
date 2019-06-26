package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	builtins *Environment
)

func init() {
	m := map[string]Object{
		"+":     Function(Plus),
		"*":     Function(Mul),
		"begin": Function(Begin),
		"%": &BinOp{"%", func(a, b float64) (Object, error) {
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return float64(int(a) % int(b)), nil
		}},
		"eq?": &BinOp{"eq?", func(a, b float64) (Object, error) {
			var val float64
			if a == b {
				val = 1
			}
			return val, nil
		}},
		// MT: In scheme these get arbitrary number of arguments
		"<": &BinOp{"<", func(a, b float64) (Object, error) {
			if a < b {
				return 1.0, nil
			}
			return 0.0, nil
		}},
		"-": &BinOp{"-", func(a, b float64) (Object, error) {
			return a - b, nil
		}},
		"/": &BinOp{"/", func(a, b float64) (Object, error) {
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return a / b, nil
		}},
	}

	builtins = NewEnvironment(m, nil)
}

// Object in the language
type Object interface{}

// Callable object
type Callable interface {
	Call(args []Object) (Object, error)
}

// ReadExpr reads an expression from slice of tokens
func ReadExpr(tokens []Token) (Expression, []Token, error) {
	var err error
	if len(tokens) == 0 {
		return nil, nil, io.EOF
	}

	tok, tokens := tokens[0], tokens[1:]
	if tok == "(" {
		var children []Expression
		for len(tokens) > 0 && tokens[0] != ")" {
			var child Expression
			child, tokens, err = ReadExpr(tokens)
			if err != nil {
				return nil, nil, err
			}
			children = append(children, child)
		}

		if len(tokens) == 0 {
			return nil, nil, fmt.Errorf("unbalanced expression")
		}

		tokens = tokens[1:] // remove closing ')'
		return &ListExpr{children}, tokens, nil
	}

	switch tok {
	case ")": // TODO: file:line
		return nil, nil, fmt.Errorf("unexpected ')'")
	}

	lit := string(tok)
	val, err := strconv.ParseFloat(lit, 64)
	if err == nil {
		return &NumberExpr{val}, tokens, nil
	}
	return &SymbolExpr{lit}, tokens, nil // name
}

func repl() {
	rdr := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("» ")
		text, err := rdr.ReadString('\n')
		if err != nil {
			break
		}

		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}

		tokens := Tokenize(text)
		// fmt.Println("tokens →", tokens)

		expr, _, err := ReadExpr(tokens)
		if err != nil {
			printError(err)
			continue
		}
		//fmt.Printf("expr → %s\n", expr)

		out, err := expr.Eval(builtins)
		if err != nil {
			printError(err)
			continue
		}
		fmt.Println(out)
	}
}

func printError(err error) {
	fmt.Printf("\033[31mERROR: %s\033[0m\n", err)
}

// rlwrap go run humble.go
func main() {
	fmt.Println("Welcome to Hubmle lisp (hit CTRL-D to quit)")
	repl()
	fmt.Println("\nkthxbai ☺")
}
