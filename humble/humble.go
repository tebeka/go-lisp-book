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
	s := Scope{
		"+":     Function(Plus),
		"*":     Function(Mul),
		"begin": Function(Begin),
	}

	RegisterBinOp("%", s, func(a, b float64) (Object, error) {
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return float64(int(a) % int(b)), nil
	})
	RegisterBinOp("eq?", s, func(a, b float64) (Object, error) {
		var val float64
		if a == b {
			val = 1
		}
		return val, nil
	})
	// MT: In scheme these get arbitrary number of arguments
	RegisterBinOp("<", s, func(a, b float64) (Object, error) {
		if a < b {
			return 1.0, nil
		}
		return 0.0, nil
	})
	RegisterBinOp("-", s, func(a, b float64) (Object, error) {
		return a - b, nil
	})
	RegisterBinOp("/", s, func(a, b float64) (Object, error) {
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return a / b, nil
	})

	builtins = append(builtins, s)
}

// Token in the language
type Token string

// Tokenize splits the t list of tokens
func Tokenize(code string) []Token {
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	var tokens []Token
	for _, tok := range strings.Fields(code) {
		tokens = append(tokens, Token(tok))
	}
	return tokens
}

// Expression to be computed
type Expression interface {
	Eval(env Environment) (Object, error)
}

// Object in the language
type Object interface{}

// NumberExpr is a number. e.g. 3.14
type NumberExpr struct {
	value float64
}

func (e *NumberExpr) String() string {
	return fmt.Sprintf("%v", e.value)
}

// Eval evaluates value
func (e *NumberExpr) Eval(env Environment) (Object, error) {
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
func (e *SymbolExpr) Eval(env Environment) (Object, error) {
	scope := env.Find(e.name)
	if scope == nil {
		return nil, fmt.Errorf("unknown name - %q", e.name)
	}

	return scope[e.name], nil
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

// Callable object
type Callable interface {
	Call(args []Object) (Object, error)
}

// Function object
type Function func(args []Object) (Object, error)

// Call implements Callable
func (f Function) Call(args []Object) (Object, error) {
	return f(args)
}

// Plus function. e.g. (+ 1 2 3)
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

// Mul function. e.g. (* 1 2 3)
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

// Begin function. e.g. (begin (* 3 4) (/ 5 7))
func Begin(args []Object) (Object, error) {
	if len(args) == 0 {
		return 0.0, nil
	}

	return args[len(args)-1], nil
}

// BinOp is a binary (two argument) operator/function
type BinOp struct {
	name string
	op   func(float64, float64) (Object, error)
}

// Call implement Callable
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

// RegisterBinOp registers a new BinOp
func RegisterBinOp(name string, scope Scope, op func(float64, float64) (Object, error)) {
	scope[name] = &BinOp{name: name, op: op}
}

// Lambda is a lambda object. e.g. (lambda (n) (+ n 1))
type Lambda struct {
	env    Environment
	params []string
	body   Expression
}

// Call implements Callable
func (l *Lambda) Call(args []Object) (Object, error) {
	if len(args) != len(l.params) {
		return nil, fmt.Errorf("wrong number of arguments (want %d, got %d)", len(l.params), args)
	}

	scope := Scope{}
	for i, name := range l.params {
		scope[name] = args[i]
	}

	env := append(l.env, scope)
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

// Scope of variable
type Scope map[string]Object // Not sure about the name Scope

// Environment holds name → values
type Environment []Scope

// Find finds the environment holding name, return nil if not found
func (e Environment) Find(name string) Scope {
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
