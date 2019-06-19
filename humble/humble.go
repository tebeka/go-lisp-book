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
	builtins Environment
)

func init() {
	builtins = append(builtins, map[string]Object{
		"+":     Function(Plus),
		"*":     Function(Mul),
		"begin": Function(Begin),
		"%": &BinOp{
			name: "%",
			op: func(a, b float64) (Object, error) {
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
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
		"<": &BinOp{
			name: "<",
			op: func(a, b float64) (Object, error) {
				if a < b {
					return 1.0, nil
				}
				return 0.0, nil
			},
		},
		"/": &BinOp{
			name: "/",
			op: func(a, b float64) (Object, error) {
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
		},
		"-": &BinOp{
			name: "-",
			op: func(a, b float64) (Object, error) {
				return a - b, nil
			},
		},
	})
}

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
	Eval(env Environment) (Object, error)
}

type Object interface{}

type NumberExpr struct {
	value float64
}

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e *NumberExpr) Eval(env Environment) (Object, error) {
	return e.value, nil
}

type SymbolExpr struct {
	name string
}

func (e *SymbolExpr) String() string {
	return fmt.Sprintf("%v", e.name)
}

func (e *SymbolExpr) Eval(env Environment) (Object, error) {
	scope := env.Find(e.name)
	if scope == nil {
		return nil, fmt.Errorf("unknown name - %q", e.name)
	}

	return scope[e.name], nil
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

func (e *ListExpr) Eval(env Environment) (Object, error) {
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
		return nil, fmt.Errorf("%s is not callabled", obj)
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

func evalDefine(args []Expression, env Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	s, ok := args[0].(*SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("bad name in 'define'")
	}

	val, err := args[1].Eval(env)
	if err != nil {
		return nil, err
	}
	env[len(env)-1][s.name] = val
	return val, nil
}

func evalSet(args []Expression, env Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	s, ok := args[0].(*SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("bad name in 'define'")
	}

	scope := env.Find(s.name)
	if scope == nil {
		return nil, fmt.Errorf("unknown name - %s", s.name)
	}

	val, err := args[1].Eval(env)
	if err != nil {
		return nil, err
	}

	scope[s.name] = val
	return val, nil
}

func evalIf(args []Expression, env Environment) (Object, error) {
	if len(args) != 3 { // TODO: if without else
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	cond, err := args[0].Eval(env)
	if err != nil {
		return nil, err
	}

	if cond == 1.0 {
		return args[1].Eval(env)
	}
	return args[2].Eval(env)
}

func evalOr(args []Expression, env Environment) (Object, error) {
	for _, e := range args {
		obj, err := e.Eval(env)
		if err != nil {
			return nil, err
		}

		val, ok := obj.(float64)
		if !ok {
			return nil, fmt.Errorf("or - %v bad type %T", val, val)
		}

		if val != 0.0 {
			return val, nil
		}
	}

	return 0.0, nil
}

func evalAnd(args []Expression, env Environment) (Object, error) {
	val, ok := 1.0, false
	for _, e := range args {
		obj, err := e.Eval(env)
		if err != nil {
			return nil, err
		}

		val, ok = obj.(float64)
		if !ok {
			return nil, fmt.Errorf("or - %v bad type %T", val, val)
		}

		if val == 0.0 {
			return val, nil
		}
	}

	return val, nil
}

func evalLambda(args []Expression, env Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("malformed lambda")
	}

	le, ok := args[0].(*ListExpr)
	if !ok {
		return nil, fmt.Errorf("malformed lambda")
	}

	params := make([]string, len(le.children))
	for i, e := range le.children {
		s, ok := e.(*SymbolExpr)
		if !ok {
			return nil, fmt.Errorf("malformed lambda")
		}
		params[i] = s.name
	}
	obj := &Lambda{
		env:    env,
		params: params,
		body:   args[1],
	}
	return obj, nil
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

func Mul(args []Object) (Object, error) {
	total := 1.0
	for i, arg := range args {
		fval, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("%d bad argument: %v of %T", i, args, arg)
		}
		total *= fval
	}

	return total, nil
}

func Begin(args []Object) (Object, error) {
	if len(args) == 0 {
		return 0.0, nil
	}

	return args[len(args)-1], nil
}

type BinOp struct {
	name string
	op   func(float64, float64) (Object, error)
}

func (bo *BinOp) Call(args []Object) (Object, error) {
	if len(args) != 2 {
		return nil, bo.errorf("wrong number of arguments (want 2, got %d)", len(args))
	}

	a, ok := args[0].(float64)
	if !ok {
		return nil, bo.errorf("bad type for first argument - %T", args[0])
	}

	b, ok := args[1].(float64)
	if !ok {
		return nil, bo.errorf("bad type for second argument - %T", args[0])
	}

	val, err := bo.op(a, b)
	if err != nil {
		return nil, bo.errorf("%s", err)
	}

	return val, nil
}

func (bo *BinOp) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s - %s", bo.name, msg)
}

type Lambda struct {
	env    Environment
	params []string
	body   Expression
}

func (l *Lambda) Call(args []Object) (Object, error) {
	if len(args) != len(l.params) {
		return nil, fmt.Errorf("wrong number of arguments (want %d, got %d)", len(l.params), args)
	}

	scope := map[string]Object{}
	for i, name := range l.params {
		scope[name] = args[i]
	}

	env := append(l.env, scope)
	return l.body.Eval(env)
}

func (l *Lambda) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(lambda (")
	fmt.Fprintf(&buf, strings.Join(l.params, " "))
	fmt.Fprintf(&buf, ") ")
	fmt.Fprintf(&buf, "%s", l.body)
	return buf.String()
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
	return &SymbolExpr{lit}, tokens, nil // name
}

type Environment []map[string]Object

func (e Environment) Find(name string) map[string]Object {
	for i := len(e) - 1; i >= 0; i-- {
		if _, ok := e[i][name]; ok {
			return e[i]
		}
	}
	return nil
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
