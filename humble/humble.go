package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	builtins = map[string]Object{
		"+": Function(Plus),
		"mod": &BinOp{
			name: "mod",
			op: func(a, b float64) (Object, error) {
				if b == 0 {
					return nil, fmt.Errorf("mod: zero division")
				}
				return float64(int(a) % int(b)), nil
			},
		},
		"eq?": &BinOp{
			name: "eq?",
			op: func(a, b float64) (Object, error) {
				var val float64
				if a == b {
					val = 1
				}
				return val, nil
			},
		},
	}
)

// Token in the language
type Token string

func tokenize(code string) []Token {
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	var tokens []Token
	for _, tok := range strings.Fields(code) {
		tokens = append(tokens, Token(tok))
	}
	return tokens
}

type Expression interface {
	Eval() (Object, error)
}

type Object interface{}

type NumberExpr struct {
	value float64
}

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e *NumberExpr) Eval() (Object, error) {
	return e.value, nil
}

type NameExpr struct {
	name string
}

func (e *NameExpr) String() string {
	return fmt.Sprintf("%v", e.name)
}

func (e *NameExpr) Eval() (Object, error) {
	val, ok := builtins[e.name]
	if !ok {
		return nil, fmt.Errorf("unknown name - %q", e.name)
	}

	return val, nil
}

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

func (e *ListExpr) Eval() (Object, error) {
	if len(e.children) == 0 {
		return nil, fmt.Errorf("empty list expression")
	}

	ne, ok := e.children[0].(*NameExpr)
	if !ok {
		return nil, fmt.Errorf("%v starting list expression", e.children[0])
	}

	op := ne.name
	args := e.children[1:]

	switch op {
	case "define":
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments for 'define'")
		}

		arg0, ok := args[0].(*NameExpr)
		if !ok {
			return nil, fmt.Errorf("bad name in 'define'")
		}

		val, err := args[1].Eval()
		if err != nil {
			return nil, err
		}
		builtins[arg0.name] = val
		return val, nil
	}

	obj, ok := builtins[op]
	if !ok {
		return nil, fmt.Errorf("unknown name - %s", op)
	}

	c, ok := obj.(Callable)
	if !ok {
		return nil, fmt.Errorf("%s (%T) is not callabled", op, obj)
	}

	var params []Object
	for _, e := range args {
		obj, err := e.Eval()
		if err != nil {
			return nil, err
		}
		params = append(params, obj)
	}

	return c.Call(params)
}

type Callable interface {
	Call(args []Object) (Object, error)
}

type Function func(args []Object) (Object, error)

func (f Function) Call(args []Object) (Object, error) {
	return f(args)
}

func Plus(args []Object) (Object, error) {
	total := 0.0
	for i, arg := range args {
		fval, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("%d bad argument: %v of %T", i, args, arg)
		}
		total += fval
	}

	return total, nil
}

type BinOp struct {
	name string
	op   func(float64, float64) (Object, error)
}

func (bo *BinOp) Call(args []Object) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s: wrong number of arguments (want 2, got %d)", bo.name, len(args))
	}

	a, ok := args[0].(float64)
	if !ok {
		return nil, fmt.Errorf("%s: bad type for first argument - %T", bo.name, args[0])
	}

	b, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("%s: bad type for second argument - %T", bo.name, args[0])
	}

	return bo.op(a, b)
}

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
	return &NameExpr{lit}, tokens, nil // name
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

		tokens := tokenize(text)
		// fmt.Println("tokens →", tokens)

		expr, _, err := ReadExpr(tokens)
		if err != nil {
			fmt.Println("ERROR: ", err)
			continue
		}
		//fmt.Printf("expr → %s\n", expr)

		out, err := expr.Eval()
		if err != nil {
			fmt.Println("ERROR: ", err)
		} else {
			fmt.Println(out)
		}
	}
}

// rlwrap go run humble.go
func main() {
	fmt.Println("Welcome to Hubmle lisp (hit CTRL-D to quit)")
	repl()
	fmt.Println("\nkthxbai ☺")
}
