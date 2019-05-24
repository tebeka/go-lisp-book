package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	commentRe = regexp.MustCompile(";.*$")
	builtins  envList
)

func init() {
	env := map[string]interface{}{
		"+":   &binOp{"+", func(a, b float64) interface{} { return a + b }},
		"-":   &binOp{"-", func(a, b float64) interface{} { return a - b }},
		"*":   &binOp{"*", func(a, b float64) interface{} { return a * b }},
		"/":   &binOp{"/", func(a, b float64) interface{} { return a / b }},
		">":   &binOp{">", func(a, b float64) interface{} { return a > b }},
		">=":  &binOp{">=", func(a, b float64) interface{} { return a >= b }},
		"<":   &binOp{"<", func(a, b float64) interface{} { return a < b }},
		"<=":  &binOp{"<=", func(a, b float64) interface{} { return a <= b }},
		"=":   &binOp{"=", func(a, b float64) interface{} { return a == b }},
		"!=":  &binOp{"!=", func(a, b float64) interface{} { return a != b }},
		"if":  ifExpr{},
		"or":  orExpr{},
		"and": andExpr{},
	}

	builtins = envList{env}
}

type callable interface {
	call([]interface{}, envList) (interface{}, error)
}

type binOp struct {
	name string
	fn   func(float64, float64) interface{}
}

func (b *binOp) call(args []interface{}, env envList) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments")
	}

	lhs, ok := args[0].(float64)
	if !ok {
		return nil, fmt.Errorf("lhs type error, got %v of type %T", args[0], args[0])
	}

	rhs, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("rhs type error, got %v of type %T", args[1], args[1])
	}

	return b.fn(lhs, rhs), nil
}

func (b *binOp) String() string {
	return b.name
}

func tokenize(code string) []string {
	code = commentRe.ReplaceAllString(code, "")
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	return strings.Fields(code)
}

type ifExpr struct{}

// (if (> 2 1) 10 20)
// (if (> 2 1) 10)
func (e ifExpr) call(args []interface{}, env envList) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("malformed if")
	}

	val, err := eval(args[0], env)
	if err != nil {
		return nil, err
	}

	test, ok := val.(bool)
	if !ok {
		return nil, fmt.Errorf("no bool values as test")
	}
	if test {
		return eval(args[1], env)
	}

	if len(args) == 3 { // has else part
		return eval(args[2], env)
	}

	return nil, nil
}

func (e ifExpr) String() string {
	return "if"
}

type orExpr struct{}

// (or (> 1 2) (> 3 4))
func (e orExpr) call(args []interface{}, env envList) (interface{}, error) {
	for _, expr := range args {
		out, err := eval(expr, env)
		if err != nil {
			return nil, err
		}

		val, ok := out.(bool)
		if !ok {
			return nil, fmt.Errorf("bad boolean in or: %#v", out)
		}
		if val {
			return true, nil
		}
	}

	return false, nil
}

func (e orExpr) String() string {
	return "or"
}

type andExpr struct{}

// (and (> 1 2) (> 3 4))
func (e andExpr) call(args []interface{}, env envList) (interface{}, error) {
	for _, expr := range args {
		out, err := eval(expr, env)
		if err != nil {
			return nil, err
		}

		val, ok := out.(bool)
		if !ok {
			return nil, fmt.Errorf("bad boolean in or: %#v", out)
		}
		if !val {
			return false, nil
		}
	}

	return true, nil
}

func (e andExpr) String() string {
	return "and"
}

func nameVal(op string, args []interface{}, env envList) (string, interface{}, error) {
	if len(args) != 2 {
		return "", nil, fmt.Errorf("malformed %s", op)
	}

	name, ok := args[0].(string)
	if !ok {
		return "", nil, fmt.Errorf("can't assign to non-name - %v", args[0])
	}

	val, err := eval(args[1], env)
	if err != nil {
		return "", nil, err
	}

	return name, val, nil
}

// (define a 1)
func evalDefine(args []interface{}, env envList) (interface{}, error) {
	name, val, err := nameVal("define", args, env)
	if err != nil {
		return nil, err
	}

	env[0][name] = val
	return nil, nil
}

// (set! a 1)
func evalSet(args []interface{}, env envList) (interface{}, error) {
	name, val, err := nameVal("set!", args, env)
	if err != nil {
		return nil, err
	}

	e := findEnv(name, env)
	if e == nil {
		return nil, fmt.Errorf("unknown variable - %s", name)
	}

	e[name] = val
	return nil, nil
}

type lambdaExpr struct {
	params []string
	body   interface{}
	env    envList
}

func (e *lambdaExpr) call(args []interface{}, env envList) (interface{}, error) {
	if len(args) != len(e.params) {
		return nil, fmt.Errorf("wrong number of arguments")
	}

	locals := make(map[string]interface{})
	for i, param := range e.params {
		locals[param] = args[i]
	}
	env = append(env, locals)
	return eval(e.body, env)
}

// (lambda (a b) (+ a b))
func makeLambda(args []interface{}, env envList) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("malformed lambda")
	}
	body := args[1]

	arg0, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("malformed lambda")
	}

	var params []string
	for _, v := range arg0 {
		param, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("malformed lambda")
		}
		params = append(params, param)
	}

	lambda := &lambdaExpr{
		params: params,
		body:   body,
		env:    env,
	}
	return lambda, nil
}

func readSExpr(tokens []string) (interface{}, []string, error) {
	var err error
	if len(tokens) == 0 {
		return nil, nil, io.EOF
	}
	tok, tokens := tokens[0], tokens[1:]
	if tok == "(" {
		var sexpr []interface{}
		for tokens[0] != ")" {
			var child interface{}
			child, tokens, err = readSExpr(tokens)
			if err != nil {
				return nil, nil, err
			}
			sexpr = append(sexpr, child)
		}
		tokens = tokens[1:] // remove closing ')'
		return sexpr, tokens, nil
	}

	// TODO: file:line
	if tok == ")" {
		return nil, nil, fmt.Errorf("unexpected ')'")
	}

	/*
	   if tok in {'#t', '#f'}:
	       return tok == '#t'
	*/

	val, err := strconv.ParseFloat(tok, 64)
	if err == nil {
		return val, tokens, nil
	}
	return tok, tokens, nil // name
}

func eval(sexpr interface{}, env envList) (interface{}, error) {
	if name, ok := sexpr.(string); ok { // name
		e := findEnv(name, env)
		if e != nil {
			return e[name], nil
		}
		return nil, fmt.Errorf("unknown name - %s", name)
	}

	list, ok := sexpr.([]interface{})
	if !ok {
		return sexpr, nil // atom
	}

	op, rest := list[0], list[1:]
	name, ok := op.(string)

	// Special cases
	if ok {
		switch name {
		case "define":
			return evalDefine(rest, env)
		case "lambda":
			return makeLambda(rest, env)
		case "set!":
			return evalSet(rest, env)
		}
	}

	val, err := eval(op, env)
	if err != nil {
		return nil, err
	}
	// function invocation
	fn, ok := val.(callable)
	if !ok {
		return nil, fmt.Errorf("%v is not a function", val)
	}

	var args []interface{}
	for _, expr := range rest {
		arg, err := eval(expr, env)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return fn.call(args, env)
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
		sexpr, _, err := readSExpr(tokens)
		if err != nil {
			fmt.Printf("[read error]: %s\n", err)
			continue
		}

		val, err := eval(sexpr, builtins)
		if err != nil {
			fmt.Printf("[eval error]: %s\n", err)
			continue
		}
		if val != nil {
			lispify(val, os.Stdout)
			fmt.Println("")
		}
	}
}

func lispify(sexpr interface{}, out io.Writer) {
	list, ok := sexpr.([]interface{})
	if ok {
		fmt.Fprintf(out, "(")
		for _, e := range list {
			lispify(e, out)
		}
		fmt.Fprintf(out, ")")
	}
	fmt.Fprintf(out, "%v", sexpr)
}

type envList []map[string]interface{}

func findEnv(name string, envs envList) map[string]interface{} {
	for _, env := range envs {
		if _, ok := env[name]; ok {
			return env
		}
	}

	return nil
}

func main() {
	fmt.Println("Welcome to Hubmle lisp (hit CTRL-D to quit)")
	repl()
	fmt.Println("\nCiao ☺")
	/*
		tokens := tokenize("(define x 1)")
		sexpr, _, _ := readSExpr(tokens)
		eval(sexpr, builtins)
	*/
}
