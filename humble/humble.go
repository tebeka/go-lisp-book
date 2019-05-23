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
		"+": makeBinop(func(a, b float64) (interface{}, error) {
			return a + b, nil
		}),
		"-": makeBinop(func(a, b float64) (interface{}, error) {
			return a - b, nil
		}),
		"*": makeBinop(func(a, b float64) (interface{}, error) {
			return a * b, nil
		}),
		"/": makeBinop(func(a, b float64) (interface{}, error) {
			return a / b, nil
		}),
		">": makeBinop(func(a, b float64) (interface{}, error) {
			return a > b, nil
		}),
		">=": makeBinop(func(a, b float64) (interface{}, error) {
			return a >= b, nil
		}),
		"<": makeBinop(func(a, b float64) (interface{}, error) {
			return a < b, nil
		}),
		"<=": makeBinop(func(a, b float64) (interface{}, error) {
			return a < b, nil
		}),
		"=": makeBinop(func(a, b float64) (interface{}, error) {
			return a == b, nil
		}),
		"!=": makeBinop(func(a, b float64) (interface{}, error) {
			return a != b, nil
		}),
		"if":  function(evalIf),
		"or":  function(evalOr),
		"and": function(evalAnd),
	}

	builtins = envList{env}
}

func makeBinop(fn func(float64, float64) (interface{}, error)) function {
	return func(args []interface{}, env envList) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("wrong number of arguments")
		}
		lhs, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("bad value - %v", args[0])
		}
		rhs, ok := args[1].(float64)
		if !ok {
			return nil, fmt.Errorf("bad value - %v", args[1])
		}

		return fn(lhs, rhs)
	}
}

func tokenize(code string) []string {
	code = commentRe.ReplaceAllString(code, "")
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	return strings.Fields(code)
}

// (if (> 2 1) 10 20)
// (if (> 2 1) 10)
func evalIf(args []interface{}, env envList) (interface{}, error) {
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

// (or (> 1 2) (> 3 4))
func evalOr(args []interface{}, env envList) (interface{}, error) {
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

// (and (> 1 2) (> 3 4))
func evalAnd(args []interface{}, env envList) (interface{}, error) {
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

// (define a 1)
func evalDefine(args []interface{}, env envList) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("malformed define")
	}

	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("can't assign to non-name - %v", args[0])
	}

	val, err := eval(args[1], env)
	if err != nil {
		return nil, err
	}

	env[0][name] = val
	return nil, nil
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

type function func(values []interface{}, env envList) (interface{}, error)

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
	// define is special since we create a new name
	if ok && name == "define" {
		return evalDefine(rest, env)
	}

	val, err := eval(op, env)
	if err != nil {
		return nil, err
	}
	// function invocation
	fn, ok := val.(function)
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

	return fn(args, env)
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

func printSExpr(sexpr interface{}, indent int) {
	if list, ok := sexpr.([]interface{}); ok {
		for _, e := range list {
			printSExpr(e, indent+2)
		}
		return
	}
	fmt.Printf("%*s%v\n", indent, " ", sexpr)
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
